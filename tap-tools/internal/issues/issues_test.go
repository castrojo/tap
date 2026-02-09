package issues

import (
	"testing"
)

func TestExtractRepositoryURL(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "Standard template format",
			body: `### Repository or Homepage URL
https://github.com/user/repo

### Description
A cool tool`,
			expected: "https://github.com/user/repo",
		},
		{
			name: "URL with trailing period",
			body: `### Repository URL
https://github.com/user/repo.

Other text`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "Inline URL",
			body:     "Check out this project: https://github.com/user/awesome-tool for details",
			expected: "https://github.com/user/awesome-tool",
		},
		{
			name: "Case insensitive header",
			body: `### repository URL
https://github.com/owner/project`,
			expected: "https://github.com/owner/project",
		},
		{
			name:     "No URL",
			body:     "This issue has no repository URL",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepositoryURL(tt.body)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "Standard description",
			body: `### Description
A fast search tool

### Other section`,
			expected: "A fast search tool",
		},
		{
			name: "Case insensitive",
			body: `### description
Command-line JSON processor`,
			expected: "Command-line JSON processor",
		},
		{
			name: "Multi-line description (takes first line)",
			body: `### Description
This is a tool
that does many things

### URL`,
			expected: "This is a tool",
		},
		{
			name:     "No description",
			body:     "Some body without description section",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDescription(tt.body)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractPackageNameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Simple lowercase",
			url:      "https://github.com/user/ripgrep",
			expected: "ripgrep",
		},
		{
			name:     "Mixed case to lowercase",
			url:      "https://github.com/user/MyAwesomeTool",
			expected: "myawesometool",
		},
		{
			name:     "Underscores to hyphens",
			url:      "https://github.com/user/my_cool_app",
			expected: "my-cool-app",
		},
		{
			name:     "Already has hyphens",
			url:      "https://github.com/user/visual-studio-code",
			expected: "visual-studio-code",
		},
		{
			name:     "SSH URL format",
			url:      "git@github.com:user/project.git",
			expected: "project",
		},
		{
			name:     "Invalid URL",
			url:      "not a github url",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPackageNameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDetectPackageType(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		title    string
		expected PackageType
	}{
		{
			name:     "Explicit formula type",
			body:     "Type: formula\nThis is a CLI tool",
			title:    "Add ripgrep",
			expected: PackageTypeFormula,
		},
		{
			name:     "Explicit cask type",
			body:     "Type: cask\nGUI application",
			title:    "Add VS Code",
			expected: PackageTypeCask,
		},
		{
			name:     "GUI keyword in body",
			body:     "This is a desktop application with a GUI",
			title:    "Package request",
			expected: PackageTypeCask,
		},
		{
			name:     "Electron app",
			body:     "Built with Electron framework",
			title:    "Add app",
			expected: PackageTypeCask,
		},
		{
			name:     "CLI keyword in title",
			body:     "A useful tool",
			title:    "CLI tool for searching",
			expected: PackageTypeFormula,
		},
		{
			name:     "Terminal tool",
			body:     "Command-line utility for terminal",
			title:    "Add tool",
			expected: PackageTypeFormula,
		},
		{
			name:     "No clear indicators - defaults to formula",
			body:     "A useful tool for developers",
			title:    "Package request: mytool",
			expected: PackageTypeFormula,
		},
		{
			name:     "Editor keyword (GUI)",
			body:     "A code editor",
			title:    "Package request",
			expected: PackageTypeCask,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectPackageType(tt.body, tt.title)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPackageTypePriority(t *testing.T) {
	// Test that explicit type hints take priority over keywords
	body := "Type: formula\nThis is a GUI desktop application"
	title := "Add app"

	result := detectPackageType(body, title)
	if result != PackageTypeFormula {
		t.Errorf("Explicit type hint should take priority. Expected formula, got %s", result)
	}
}

func TestExtractRepositoryURLEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "Multiple URLs - takes first",
			body: `### Repository URL
https://github.com/user/repo1
https://github.com/user/repo2`,
			expected: "https://github.com/user/repo1",
		},
		{
			name:     "URL in markdown link",
			body:     "See [project](https://github.com/user/project) for more",
			expected: "https://github.com/user/project",
		},
		{
			name:     "URL with .git suffix",
			body:     "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepositoryURL(tt.body)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPackageNameNormalization(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Multiple underscores",
			url:      "https://github.com/user/my_super_cool_app",
			expected: "my-super-cool-app",
		},
		{
			name:     "Mixed case with underscores",
			url:      "https://github.com/user/My_Cool_App",
			expected: "my-cool-app",
		},
		{
			name:     "Numbers preserved",
			url:      "https://github.com/user/tool_v2",
			expected: "tool-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPackageNameFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
