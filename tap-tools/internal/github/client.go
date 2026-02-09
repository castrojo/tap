package github

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	ghRelease, _, err := c.gh.Repositories.GetLatestRelease(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	return c.convertRelease(ghRelease), nil
}

// GetAllReleases fetches all releases (including prereleases)
func (c *Client) GetAllReleases(owner, repo string) ([]*Release, error) {
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
