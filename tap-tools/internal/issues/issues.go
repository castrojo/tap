// Package issues provides GitHub issue parsing and handling for package requests
package issues

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// PackageType represents the type of package requested
type PackageType string

const (
	PackageTypeFormula PackageType = "formula"
	PackageTypeCask    PackageType = "cask"
	PackageTypeUnknown PackageType = "unknown"
)

// IssueRequest represents a parsed package request from a GitHub issue
type IssueRequest struct {
	Number      int         // Issue number
	Title       string      // Issue title
	Body        string      // Issue body
	RepoURL     string      // Repository URL to package
	Description string      // Package description (optional)
	PackageType PackageType // Detected package type (formula or cask)
	PackageName string      // Derived package name
	State       string      // Issue state (open/closed)
	URL         string      // Issue URL
}

// Client wraps GitHub API client for issue operations
type Client struct {
	gh *github.Client
}

// NewClient creates a new issues client
// Uses GITHUB_TOKEN environment variable if available
func NewClient() *Client {
	var client *github.Client

	if token := getGitHubToken(); token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Client{gh: client}
}

// getGitHubToken returns GitHub token from environment
func getGitHubToken() string {
	// Try common environment variables
	for _, env := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := os.Getenv(env); token != "" {
			return token
		}
	}
	return ""
}

// GetIssue fetches and parses a GitHub issue
func (c *Client) GetIssue(owner, repo string, number int) (*IssueRequest, error) {
	ctx := context.Background()

	issue, _, err := c.gh.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue: %w", err)
	}

	return c.parseIssue(issue, number)
}

// parseIssue extracts package request information from an issue
func (c *Client) parseIssue(issue *github.Issue, number int) (*IssueRequest, error) {
	body := issue.GetBody()

	// Extract repository URL
	repoURL := extractRepositoryURL(body)
	if repoURL == "" {
		return nil, fmt.Errorf("could not find repository URL in issue body")
	}

	// Validate it's a GitHub URL
	if !strings.Contains(repoURL, "github.com") {
		return nil, fmt.Errorf("repository URL must be a GitHub URL: %s", repoURL)
	}

	// Extract package name from repository URL
	packageName := extractPackageNameFromURL(repoURL)
	if packageName == "" {
		return nil, fmt.Errorf("could not derive package name from repository URL: %s", repoURL)
	}

	// Extract description (optional)
	description := extractDescription(body)

	// Detect package type
	packageType := detectPackageType(body, issue.GetTitle())

	return &IssueRequest{
		Number:      number,
		Title:       issue.GetTitle(),
		Body:        body,
		RepoURL:     repoURL,
		Description: description,
		PackageType: packageType,
		PackageName: packageName,
		State:       issue.GetState(),
		URL:         issue.GetHTMLURL(),
	}, nil
}

// extractRepositoryURL extracts the repository URL from issue body
// Looks for patterns like:
// ### Repository or Homepage URL
// https://github.com/owner/repo
func extractRepositoryURL(body string) string {
	// Try multiple patterns
	patterns := []string{
		`###.*(?:Repository|URL|Homepage).*\n+([^\n]+github\.com[^\s\n]+)`,
		`(?:Repository|URL|Homepage).*\n+([^\n]+github\.com[^\s\n]+)`,
		`(https?://github\.com/[^\s\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) > 1 {
			url := strings.TrimSpace(matches[1])
			// Clean up common trailing characters
			url = strings.TrimSuffix(url, ".")
			url = strings.TrimSuffix(url, ",")
			url = strings.TrimSuffix(url, ")")
			url = strings.TrimSuffix(url, "]")
			return url
		}
	}

	return ""
}

// extractDescription extracts the description from issue body
// Looks for patterns like:
// ### Description
// Package description here
func extractDescription(body string) string {
	pattern := `###.*Description.*\n+([^\n#]+)`
	re := regexp.MustCompile(`(?i)` + pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractPackageNameFromURL derives package name from repository URL
// Example: https://github.com/user/My_Cool-App -> my-cool-app
func extractPackageNameFromURL(url string) string {
	// Extract repository name from URL
	re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/\.]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return ""
	}

	repoName := matches[2]

	// Normalize: lowercase, replace underscores with hyphens
	name := strings.ToLower(repoName)
	name = strings.ReplaceAll(name, "_", "-")

	return name
}

// detectPackageType attempts to determine if this should be a formula or cask
// Priority:
// 1. Explicit type hint in issue body
// 2. Keywords in title/body
// 3. Default to formula (most common)
func detectPackageType(body, title string) PackageType {
	combined := strings.ToLower(body + " " + title)

	// Check for explicit type hints
	if strings.Contains(combined, "type: cask") || strings.Contains(combined, "type: gui") {
		return PackageTypeCask
	}
	if strings.Contains(combined, "type: formula") || strings.Contains(combined, "type: cli") {
		return PackageTypeFormula
	}

	// Check for GUI/application indicators
	guiKeywords := []string{
		"gui", "desktop", "application", " app",
		"electron", "tauri", "qt", "gtk",
		"visual", "editor", "ide",
	}
	for _, keyword := range guiKeywords {
		if strings.Contains(combined, keyword) {
			return PackageTypeCask
		}
	}

	// Check for CLI indicators
	cliKeywords := []string{
		"cli", "command-line", "terminal", "shell",
		"tool", "utility", "binary",
	}
	for _, keyword := range cliKeywords {
		if strings.Contains(combined, keyword) {
			return PackageTypeFormula
		}
	}

	// Default to formula (most packages are CLI tools)
	return PackageTypeFormula
}

// DetectPackageTypeFromRepo uses GitHub API to detect package type from repository
func (c *Client) DetectPackageTypeFromRepo(owner, repo string) (PackageType, error) {
	ctx := context.Background()

	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return PackageTypeUnknown, fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Check topics and description
	topics := repository.Topics
	description := repository.GetDescription()

	combined := strings.ToLower(strings.Join(topics, " ") + " " + description)

	// Check for GUI indicators
	guiKeywords := []string{
		"gui", "desktop", "application", "app",
		"electron", "tauri", "qt", "gtk",
	}
	for _, keyword := range guiKeywords {
		if strings.Contains(combined, keyword) {
			return PackageTypeCask, nil
		}
	}

	// Check for CLI indicators
	cliKeywords := []string{
		"cli", "command-line", "terminal", "tool",
	}
	for _, keyword := range cliKeywords {
		if strings.Contains(combined, keyword) {
			return PackageTypeFormula, nil
		}
	}

	// Default to formula
	return PackageTypeFormula, nil
}

// CreatePullRequest creates a pull request for the package
func (c *Client) CreatePullRequest(owner, repo, head, base, title, body string) (string, error) {
	ctx := context.Background()

	pr, _, err := c.gh.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	})

	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr.GetHTMLURL(), nil
}

// CommentOnIssue adds a comment to an issue
func (c *Client) CommentOnIssue(owner, repo string, number int, body string) error {
	ctx := context.Background()

	_, _, err := c.gh.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: github.String(body),
	})

	if err != nil {
		return fmt.Errorf("failed to comment on issue: %w", err)
	}

	return nil
}
