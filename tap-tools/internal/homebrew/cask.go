package homebrew

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/castrojo/tap-tools/internal/generator"
)

// CaskData represents data for generating a Homebrew cask
type CaskData struct {
	Token       string // Cask name (always with -linux suffix)
	Version     string
	SHA256      string
	URL         string
	Description string
	Homepage    string
	AppName     string // Original app name
	BinaryPath  string // Path to binary in archive
	BinaryName  string // Name of binary to install

	// Desktop integration
	HasDesktopFile    bool
	DesktopFilePath   string
	DesktopFileSource string // Original path in archive
	HasIcon           bool
	IconPath          string
	IconSource        string // Original path in archive

	// XDG directories to create
	XDGDirs []string

	// Zap configuration
	ZapTrash []string

	// Generation metadata
	SourceURL string // Repository URL for regeneration instructions
}

// caskTemplate is the template for generating Homebrew casks
const caskTemplate = `# typed: strict
# frozen_string_literal: true

cask "{{ .Token }}" do
  version "{{ .Version }}"
  sha256 "{{ .SHA256 }}"

  url "{{ .URL }}"
  name "{{ .AppName }}"
  desc "{{ cleanDesc .Description }}"
  homepage "{{ if .Homepage }}{{ .Homepage }}{{ else }}https://github.com/{{ .AppName }}{{ end }}"

  # Linux-only cask
  depends_on formula: "bash"
{{- if or .HasDesktopFile .HasIcon }}

  preflight do
    {{- if .XDGDirs }}
    # Create XDG directories
    xdg_data_home = ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")
    {{- range .XDGDirs }}
    system_command "mkdir", args: ["-p", "#{xdg_data_home}/{{ . }}"]
    {{- end }}
    {{- end }}
    {{- if .HasDesktopFile }}

    # Fix desktop file paths
    desktop_file = staged_path.join("{{ .DesktopFileSource }}")
    if desktop_file.exist?
      content = desktop_file.read
      content.gsub!(%r{Exec=.*}, "Exec=#{HOMEBREW_PREFIX}/bin/{{ .BinaryName }}")
      {{- if .HasIcon }}
      content.gsub!(%r{Icon=.*}, "Icon=#{xdg_data_home}/icons/{{ .IconSource }}")
      {{- end }}
      desktop_file.write(content)
    end
    {{- end }}
  end
  {{- end }}

  {{- if .BinaryPath }}
  binary "{{ .BinaryPath }}", target: "{{ .BinaryName }}"
  {{- end }}
  {{- if .HasDesktopFile }}
  artifact "{{ .DesktopFileSource }}", target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/applications/{{ .DesktopFilePath }}"
  {{- end }}
  {{- if .HasIcon }}
  artifact "{{ .IconSource }}", target: "#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/icons/{{ .IconPath }}"
  {{- end }}

  {{- if .ZapTrash }}

  zap trash: [
    {{- range $i, $path := sortStrings .ZapTrash }}
    {{- if $i }},
    {{- end }}
    "{{ $path }}"
    {{- end }},
  ]
  {{- end }}
end
`

// cleanDesc removes leading articles and trailing periods from descriptions
func cleanDesc(desc string) string {
	// Remove leading articles
	desc = strings.TrimPrefix(desc, "A ")
	desc = strings.TrimPrefix(desc, "An ")
	desc = strings.TrimPrefix(desc, "The ")

	// Remove trailing period
	desc = strings.TrimSuffix(desc, ".")

	return desc
}

// sortStrings returns a sorted copy of a string slice
func sortStrings(strs []string) []string {
	sorted := make([]string, len(strs))
	copy(sorted, strs)
	// Simple bubble sort for small arrays
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// GenerateCask generates a Homebrew cask from the provided data
func GenerateCask(data *CaskData) (string, error) {
	// Parse template with custom functions
	tmpl, err := template.New("cask").Funcs(template.FuncMap{
		"cleanDesc":   cleanDesc,
		"sortStrings": sortStrings,
	}).Parse(caskTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer

	// Write generation header first
	if data.SourceURL != "" {
		if err := generator.WriteHeader(&buf, "tap-cask", data.SourceURL); err != nil {
			return "", fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Then write cask content
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// NewCaskData creates a new CaskData with sensible defaults
func NewCaskData(token, version, sha256, url string) *CaskData {
	return &CaskData{
		Token:    token,
		Version:  version,
		SHA256:   sha256,
		URL:      url,
		XDGDirs:  []string{},
		ZapTrash: []string{},
	}
}

// AddXDGDir adds an XDG directory to create in preflight
func (c *CaskData) AddXDGDir(dir string) {
	c.XDGDirs = append(c.XDGDirs, dir)
}

// AddZapTrash adds a path to remove on uninstall
func (c *CaskData) AddZapTrash(path string) {
	c.ZapTrash = append(c.ZapTrash, path)
}

// SetDesktopFile configures desktop file integration
func (c *CaskData) SetDesktopFile(sourcePathInArchive, targetFilename string) {
	c.HasDesktopFile = true
	c.DesktopFileSource = sourcePathInArchive
	c.DesktopFilePath = targetFilename
	c.AddXDGDir("applications")
}

// SetIcon configures icon integration
func (c *CaskData) SetIcon(sourcePathInArchive, targetFilename string) {
	c.HasIcon = true
	c.IconSource = sourcePathInArchive
	c.IconPath = targetFilename
	c.AddXDGDir("icons")
}

// InferZapTrash infers common config/cache paths to add to zap trash
func (c *CaskData) InferZapTrash() {
	// Convert app name to lowercase with hyphens for common config patterns
	appSlug := strings.ToLower(c.AppName)
	appSlug = strings.ReplaceAll(appSlug, " ", "-")
	appSlug = strings.ReplaceAll(appSlug, "_", "-")

	// Use XDG environment variables for trash paths
	commonPaths := []string{
		fmt.Sprintf(`#{ENV.fetch("XDG_CONFIG_HOME", "#{Dir.home}/.config")}/%s`, appSlug),
		fmt.Sprintf(`#{ENV.fetch("XDG_CACHE_HOME", "#{Dir.home}/.cache")}/%s`, appSlug),
		fmt.Sprintf(`#{ENV.fetch("XDG_DATA_HOME", "#{Dir.home}/.local/share")}/%s`, appSlug),
	}

	for _, path := range commonPaths {
		c.AddZapTrash(path)
	}
}
