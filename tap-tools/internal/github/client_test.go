package github

import (
	"testing"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "Full HTTPS URL",
			url:       "https://github.com/castrojo/homebrew-tap",
			wantOwner: "castrojo",
			wantRepo:  "homebrew-tap",
			wantErr:   false,
		},
		{
			name:      "Full HTTPS URL with trailing slash",
			url:       "https://github.com/castrojo/homebrew-tap/",
			wantOwner: "castrojo",
			wantRepo:  "homebrew-tap",
			wantErr:   false,
		},
		{
			name:      "Without protocol",
			url:       "github.com/sublimehq/sublime_text",
			wantOwner: "sublimehq",
			wantRepo:  "sublime_text",
			wantErr:   false,
		},
		{
			name:      "Short format",
			url:       "BurntSushi/ripgrep",
			wantOwner: "BurntSushi",
			wantRepo:  "ripgrep",
			wantErr:   false,
		},
		{
			name:      "With .git suffix",
			url:       "https://github.com/user/repo.git",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Invalid - missing repo",
			url:       "github.com/user",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
		{
			name:      "Invalid - only username",
			url:       "user",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
		{
			name:      "Invalid - empty",
			url:       "",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseRepoURL() owner = %v, want %v", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("ParseRepoURL() repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	// Test that we can create a client without errors
	client := NewClient()
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.gh == nil {
		t.Error("NewClient() created client with nil GitHub client")
	}
	if client.ctx == nil {
		t.Error("NewClient() created client with nil context")
	}
}

// Note: Tests for GetRepository, GetLatestRelease, etc. would require
// mocking the GitHub API or using integration tests with real API calls.
// For unit tests, we focus on testing the parsing and conversion logic.

func TestConvertRelease(t *testing.T) {
	// This is tested indirectly through the other methods
	// We would need to mock github.RepositoryRelease for full coverage
	t.Skip("Requires mocking GitHub API types")
}
