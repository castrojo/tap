package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	gh  *github.Client
	ctx context.Context
}

// Repository represents a GitHub repository
type Repository struct {
	Owner       string
	Name        string
	Description string
	Homepage    string
	License     string
	Stars       int
}

// Release represents a GitHub release
type Release struct {
	TagName     string
	Name        string
	Body        string
	Prerelease  bool
	Draft       bool
	PublishedAt string
	Assets      []*Asset
}

// Asset represents a release asset
type Asset struct {
	Name               string
	URL                string
	DownloadURL        string
	Size               int64
	BrowserDownloadURL string
}

// detectEnvironment returns the execution environment
func detectEnvironment() string {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "github-actions"
	}
	if os.Getenv("CODESPACES") == "true" {
		return "codespaces"
	}
	return "local"
}

// checkGitHubToken verifies GITHUB_TOKEN is set and provides helpful error messages
func checkGitHubToken() error {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return nil
	}

	env := detectEnvironment()

	switch env {
	case "github-actions":
		return fmt.Errorf(`GITHUB_TOKEN not found in GitHub Actions environment

This is unexpected. The token should be available automatically.

Debugging steps:
  1. Check workflow permissions: gh api repos/{owner}/{repo}/actions/permissions
  2. Verify workflow has 'contents: read' or higher
  3. Check if GITHUB_TOKEN is being explicitly unset
  4. Ensure token is passed in workflow: env: GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

Current rate limit: 60 requests/hour (unauthenticated)
With token: 15,000 requests/hour (GitHub Actions)`)

	case "codespaces":
		return fmt.Errorf(`GITHUB_TOKEN not found in Codespaces environment

Codespaces should have GITHUB_TOKEN available automatically.

Possible solutions:
  1. Wait for token to be injected (may be delayed)
  2. Use 'gh auth token' as workaround: export GITHUB_TOKEN=$(gh auth token)
  3. Contact repository admin to verify Codespaces permissions

Current rate limit: 60 requests/hour (unauthenticated)
With token: 5,000 requests/hour`)

	default: // local
		return fmt.Errorf(`GITHUB_TOKEN environment variable not set

All tap-tools require GitHub API access with authentication.

Solutions:
  1. Use gh CLI (recommended): export GITHUB_TOKEN=$(gh auth token)
  2. Create personal access token:
     - Go to: https://github.com/settings/tokens
     - Create token with 'repo' scope (read access)
     - Export: export GITHUB_TOKEN=ghp_your_token_here

Current rate limit: 60 requests/hour (unauthenticated)
With token: 5,000 requests/hour

Check rate limit: gh api rate_limit`)
	}
}

// NewClient creates a new GitHub client
// It will use the GITHUB_TOKEN environment variable if set
func NewClient() *Client {
	ctx := context.Background()
	var client *github.Client

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Client{
		gh:  client,
		ctx: ctx,
	}
}

// NewClientWithTokenCheck creates a new GitHub client and verifies GITHUB_TOKEN is set
// Returns an error with helpful context-specific message if token is missing
func NewClientWithTokenCheck() (*Client, error) {
	if err := checkGitHubToken(); err != nil {
		return nil, err
	}
	return NewClient(), nil
}

// CheckRateLimit monitors GitHub API rate limit and warns if running low
func (c *Client) CheckRateLimit() error {
	rateLimit, _, err := c.gh.RateLimits(c.ctx)
	if err != nil {
		// Warn that the rate limit check failed, but don't block execution.
		fmt.Fprintf(os.Stderr, "⚠️  Could not check GitHub API rate limit: %v\n", err)
		return nil
	}

	remaining := rateLimit.Core.Remaining
	limit := rateLimit.Core.Limit
	resetTime := rateLimit.Core.Reset.Time

	// Warn if less than 100 requests remaining or less than 10% of limit
	threshold := 100
	if limit < 1000 {
		threshold = limit / 10
	}

	if remaining < threshold {
		fmt.Fprintf(os.Stderr, "⚠️  GitHub API rate limit low: %d/%d remaining\n", remaining, limit)
		fmt.Fprintf(os.Stderr, "   Resets at: %s (in %s)\n",
			resetTime.Format(time.RFC3339),
			time.Until(resetTime).Round(time.Minute))

		if limit == 60 {
			fmt.Fprintf(os.Stderr, "   ℹ️  Using unauthenticated rate limit (60/hour)\n")
			fmt.Fprintf(os.Stderr, "   💡 Set GITHUB_TOKEN to increase limit to 5,000/hour\n")
			fmt.Fprintf(os.Stderr, "      Run: export GITHUB_TOKEN=$(gh auth token)\n")
		}
	}

	return nil
}

// ParseRepoURL extracts owner and repo name from a GitHub URL
// Supports: https://github.com/owner/repo, github.com/owner/repo, owner/repo
func ParseRepoURL(url string) (owner, repo string, err error) {
	// Remove trailing slashes
	url = strings.TrimRight(url, "/")

	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	// Split into parts
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s (expected format: owner/repo)", url)
	}

	owner = parts[0]
	repo = parts[1]

	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub URL: owner or repo cannot be empty")
	}

	return owner, repo, nil
}

// GetRepository fetches repository metadata
func (c *Client) GetRepository(owner, repo string) (*Repository, error) {
	// Check rate limit before making API call
	c.CheckRateLimit()

	ghRepo, _, err := c.gh.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	license := ""
	if ghRepo.License != nil && ghRepo.License.SPDXID != nil {
		license = *ghRepo.License.SPDXID
	}

	return &Repository{
		Owner:       owner,
		Name:        repo,
		Description: ghRepo.GetDescription(),
		Homepage:    ghRepo.GetHomepage(),
		License:     license,
		Stars:       ghRepo.GetStargazersCount(),
	}, nil
}

// GetLatestRelease fetches the latest release (excluding prereleases and drafts)
func (c *Client) GetLatestRelease(owner, repo string) (*Release, error) {
	// Check rate limit before making API call
	c.CheckRateLimit()

	ghRelease, _, err := c.gh.Repositories.GetLatestRelease(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	return c.convertRelease(ghRelease), nil
}

// GetAllReleases fetches all releases (including prereleases)
func (c *Client) GetAllReleases(owner, repo string) ([]*Release, error) {
	// Check rate limit before making API call
	c.CheckRateLimit()

	opts := &github.ListOptions{PerPage: 100}
	ghReleases, _, err := c.gh.Repositories.ListReleases(c.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	releases := make([]*Release, 0, len(ghReleases))
	for _, ghRelease := range ghReleases {
		releases = append(releases, c.convertRelease(ghRelease))
	}

	return releases, nil
}

// convertRelease converts a GitHub release to our internal representation
func (c *Client) convertRelease(ghRelease *github.RepositoryRelease) *Release {
	assets := make([]*Asset, 0, len(ghRelease.Assets))
	for _, ghAsset := range ghRelease.Assets {
		assets = append(assets, &Asset{
			Name:               ghAsset.GetName(),
			URL:                ghAsset.GetURL(),
			DownloadURL:        ghAsset.GetBrowserDownloadURL(),
			Size:               int64(ghAsset.GetSize()),
			BrowserDownloadURL: ghAsset.GetBrowserDownloadURL(),
		})
	}

	publishedAt := ""
	if ghRelease.PublishedAt != nil {
		publishedAt = ghRelease.PublishedAt.Format("2006-01-02")
	}

	return &Release{
		TagName:     ghRelease.GetTagName(),
		Name:        ghRelease.GetName(),
		Body:        ghRelease.GetBody(),
		Prerelease:  ghRelease.GetPrerelease(),
		Draft:       ghRelease.GetDraft(),
		PublishedAt: publishedAt,
		Assets:      assets,
	}
}

// GetRepoFiles fetches the list of files in the repository root
// Used for build system detection
func (c *Client) GetRepoFiles(owner, repo string) ([]string, error) {
	// Check rate limit before making API call
	c.CheckRateLimit()

	_, dirContent, _, err := c.gh.Repositories.GetContents(c.ctx, owner, repo, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository contents: %w", err)
	}

	files := make([]string, 0, len(dirContent))
	for _, content := range dirContent {
		if content.GetType() == "file" {
			files = append(files, content.GetName())
		}
	}

	return files, nil
}
