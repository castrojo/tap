package homebrew

import (
	"strings"
	"testing"
)

func TestGenerateCask(t *testing.T) {
	data := &CaskData{
		Token:       "sublime-text-linux",
		Version:     "4200",
		SHA256:      "abc123def456",
		URL:         "https://example.com/sublime.tar.gz",
		Description: "Text editor",
		Homepage:    "https://sublimetext.com",
		License:     "MIT",
		AppName:     "Sublime Text",
		BinaryPath:  "sublime_text/sublime_text",
		BinaryName:  "sublime-text",
	}

	cask, err := GenerateCask(data)
	if err != nil {
		t.Fatalf("GenerateCask() error = %v", err)
	}

	// Check that essential parts are present
	required := []string{
		`cask "sublime-text-linux"`,
		`version "4200"`,
		`sha256 "abc123def456"`,
		`url "https://example.com/sublime.tar.gz"`,
		`name "Sublime Text"`,
		`desc "Text editor"`,
		`homepage "https://sublimetext.com"`,
		`license "MIT"`,
		`binary "sublime_text/sublime_text", target: "sublime-text"`,
	}

	for _, req := range required {
		if !strings.Contains(cask, req) {
			t.Errorf("Generated cask missing required content: %q", req)
		}
	}
}

func TestGenerateCaskWithDesktopFile(t *testing.T) {
	data := &CaskData{
		Token:       "test-app-linux",
		Version:     "1.0.0",
		SHA256:      "abc123",
		URL:         "https://example.com/app.tar.gz",
		Description: "Test app",
		Homepage:    "https://example.com",
		AppName:     "Test App",
		BinaryPath:  "app/bin/app",
		BinaryName:  "test-app",
	}

	data.SetDesktopFile("app/app.desktop", "test-app.desktop")
	data.SetIcon("app/icons/128x128/app.png", "test-app.png")

	cask, err := GenerateCask(data)
	if err != nil {
		t.Fatalf("GenerateCask() error = %v", err)
	}

	// Check for desktop integration
	required := []string{
		"preflight do",
		"mkdir",
		".local/share/applications",
		".local/share/icons",
		"desktop_file",
		`artifact "app/app.desktop"`,
		`artifact "app/icons/128x128/app.png"`,
	}

	for _, req := range required {
		if !strings.Contains(cask, req) {
			t.Errorf("Generated cask missing desktop integration: %q", req)
		}
	}
}

func TestNewCaskData(t *testing.T) {
	data := NewCaskData("test-linux", "1.0.0", "abc123", "https://example.com/test.tar.gz")

	if data.Token != "test-linux" {
		t.Errorf("Token = %q, want %q", data.Token, "test-linux")
	}
	if data.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", data.Version, "1.0.0")
	}
	if data.SHA256 != "abc123" {
		t.Errorf("SHA256 = %q, want %q", data.SHA256, "abc123")
	}
	if data.URL != "https://example.com/test.tar.gz" {
		t.Errorf("URL = %q, want %q", data.URL, "https://example.com/test.tar.gz")
	}
}

func TestInferZapTrash(t *testing.T) {
	data := NewCaskData("test-linux", "1.0.0", "abc", "https://example.com")
	data.AppName = "My Test App"

	data.InferZapTrash()

	if len(data.ZapTrash) == 0 {
		t.Error("InferZapTrash() did not add any paths")
	}

	// Should have common XDG paths
	found := false
	for _, path := range data.ZapTrash {
		if strings.Contains(path, "config") || strings.Contains(path, "cache") {
			found = true
			break
		}
	}
	if !found {
		t.Error("InferZapTrash() did not add config/cache paths")
	}
}

func TestSetDesktopFile(t *testing.T) {
	data := NewCaskData("test-linux", "1.0.0", "abc", "https://example.com")

	data.SetDesktopFile("app/test.desktop", "test.desktop")

	if !data.HasDesktopFile {
		t.Error("SetDesktopFile() did not set HasDesktopFile")
	}
	if data.DesktopFileSource != "app/test.desktop" {
		t.Errorf("DesktopFileSource = %q, want %q", data.DesktopFileSource, "app/test.desktop")
	}
	if data.DesktopFilePath != "test.desktop" {
		t.Errorf("DesktopFilePath = %q, want %q", data.DesktopFilePath, "test.desktop")
	}
}

func TestSetIcon(t *testing.T) {
	data := NewCaskData("test-linux", "1.0.0", "abc", "https://example.com")

	data.SetIcon("app/icons/icon.png", "test-icon.png")

	if !data.HasIcon {
		t.Error("SetIcon() did not set HasIcon")
	}
	if data.IconSource != "app/icons/icon.png" {
		t.Errorf("IconSource = %q, want %q", data.IconSource, "app/icons/icon.png")
	}
	if data.IconPath != "test-icon.png" {
		t.Errorf("IconPath = %q, want %q", data.IconPath, "test-icon.png")
	}
}
