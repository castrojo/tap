package homebrew

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// CaskData represents data for generating a Homebrew cask
type CaskData struct {
	Token       string // Cask name (always with -linux suffix)
	Version     string
	SHA256      string
	URL         string
	Description string
	Homepage    string
	License     string
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
}

// caskTemplate is the template for generating Homebrew casks
const caskTemplate = `cask "{{ .Token }}" do
  version "{{ .Version }}"
  sha256 "{{ .SHA256 }}"

  url "{{ .URL }}"
  name "{{ .AppName }}"
  desc "{{ .Description }}"
  homepage "{{ .Homepage }}"
{{- if .License }}
  license "{{ .License }}"
{{- end }}

  # Linux-only cask
  depends_on formula: "bash"

  {{- if or .HasDesktopFile .HasIcon }}

  preflight do
    {{- if .XDGDirs }}
    # Create XDG directories
    {{- range .XDGDirs }}
    system_command "mkdir", args: ["-p", "{{ . }}"]
    {{- end }}
    {{- end }}
    {{- if .HasDesktopFile }}

    # Fix desktop file paths
    desktop_file = staged_path.join("{{ .DesktopFileSource }}")
    if desktop_file.exist?
      content = desktop_file.read
      content.gsub!(%r{Exec=.*}, "Exec=#{HOMEBREW_PREFIX}/bin/{{ .BinaryName }}")
      {{- if .HasIcon }}
      content.gsub!(%r{Icon=.*}, "Icon=#{ENV.fetch("HOME")}/.local/share/icons/{{ .IconSource }}")
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
  artifact "{{ .DesktopFileSource }}", target: "#{ENV.fetch("HOME")}/.local/share/applications/{{ .DesktopFilePath }}"
  {{- end }}
  {{- if .HasIcon }}
  artifact "{{ .IconSource }}", target: "#{ENV.fetch("HOME")}/.local/share/icons/{{ .IconPath }}"
  {{- end }}

  {{- if .ZapTrash }}

  zap trash: [
    {{- range $i, $path := .ZapTrash }}
    {{- if $i }},{{ end }}
    "{{ $path }}"
    {{- end }}
  ]
  {{- end }}
end
`

// GenerateCask generates a Homebrew cask from the provided data
func GenerateCask(data *CaskData) (string, error) {
	// Parse template
	tmpl, err := template.New("cask").Parse(caskTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
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
	c.AddXDGDir("#{ENV.fetch(\"HOME\")}/.local/share/applications")
}

// SetIcon configures icon integration
func (c *CaskData) SetIcon(sourcePathInArchive, targetFilename string) {
	c.HasIcon = true
	c.IconSource = sourcePathInArchive
	c.IconPath = targetFilename
	c.AddXDGDir("#{ENV.fetch(\"HOME\")}/.local/share/icons")
}

// InferZapTrash infers common config/cache paths to add to zap trash
func (c *CaskData) InferZapTrash() {
	// Convert app name to lowercase with hyphens for common config patterns
	appSlug := strings.ToLower(c.AppName)
	appSlug = strings.ReplaceAll(appSlug, " ", "-")
	appSlug = strings.ReplaceAll(appSlug, "_", "-")

	commonPaths := []string{
		fmt.Sprintf("~/.config/%s", appSlug),
		fmt.Sprintf("~/.cache/%s", appSlug),
		fmt.Sprintf("~/.local/share/%s", appSlug),
	}

	for _, path := range commonPaths {
		c.AddZapTrash(path)
	}
}
