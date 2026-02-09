package homebrew

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/castrojo/tap-tools/internal/buildsystem"
)

// FormulaData represents data for generating a Homebrew formula
type FormulaData struct {
	ClassName    string   // Ruby class name (PascalCase)
	PackageName  string   // Package name (lowercase with hyphens)
	Version      string   // Version number
	SHA256       string   // SHA256 checksum
	URL          string   // Download URL
	Description  string   // Short description
	Homepage     string   // Project homepage
	License      string   // SPDX license ID
	BuildSystem  string   // Detected build system name
	Dependencies []string // Formula dependencies
	InstallBlock string   // Ruby code for install method
	TestBlock    string   // Ruby code for test method
}

// formulaTemplate is the template for generating Homebrew formulas
const formulaTemplate = `# typed: strict
# frozen_string_literal: true

# {{ cleanDesc .Description }}
class {{ .ClassName }} < Formula
  desc "{{ cleanDesc .Description }}"
  homepage "{{ if .Homepage }}{{ .Homepage }}{{ else }}https://github.com/{{ .PackageName }}{{ end }}"
  url "{{ .URL }}"
  sha256 "{{ .SHA256 }}"
{{- if .License }}

  license "{{ .License }}"
{{- end }}
{{- if .Dependencies }}

{{- range .Dependencies }}
  depends_on "{{ . }}"
{{- end }}
{{- end }}

  {{ .InstallBlock }}

  {{ .TestBlock }}
end
`

// GenerateFormula generates a Homebrew formula from FormulaData
func GenerateFormula(data *FormulaData) (string, error) {
	tmpl, err := template.New("formula").Funcs(template.FuncMap{
		"cleanDesc": cleanDesc,
	}).Parse(formulaTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse formula template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute formula template: %w", err)
	}

	return buf.String(), nil
}

// PackageNameToClassName converts a package name to a Ruby class name
// Examples:
//   - "jq" -> "Jq"
//   - "ripgrep" -> "Ripgrep"
//   - "go-task" -> "GoTask"
//   - "node_exporter" -> "NodeExporter"
func PackageNameToClassName(name string) string {
	// Replace hyphens and underscores with spaces for splitting
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Split into words
	words := strings.Fields(name)

	// Capitalize first letter of each word
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	// Join without spaces
	return strings.Join(words, "")
}

// NewFormulaData creates FormulaData with automatic build system detection
func NewFormulaData(packageName, version, sha256, url, description, homepage, license string, repoFiles []string, binaryName string) (*FormulaData, error) {
	// Detect build system
	bs := buildsystem.Detect(repoFiles)
	if bs == nil {
		return nil, fmt.Errorf("could not detect build system from repository files")
	}

	// Generate install block
	installOpts := buildsystem.InstallOptions{
		BinaryName: binaryName,
		Prefix:     "#{prefix}",
	}
	installBlock := bs.GenerateInstallBlock(installOpts)

	// Generate test block
	testBlock := bs.GenerateTestBlock(binaryName)

	// Get dependencies
	dependencies := bs.GenerateDependencies()

	// Build with dependencies
	var buildDeps []string
	for _, dep := range dependencies {
		buildDeps = append(buildDeps, dep)
	}

	return &FormulaData{
		ClassName:    PackageNameToClassName(packageName),
		PackageName:  packageName,
		Version:      version,
		SHA256:       sha256,
		URL:          url,
		Description:  description,
		Homepage:     homepage,
		License:      license,
		BuildSystem:  bs.Name(),
		Dependencies: buildDeps,
		InstallBlock: installBlock,
		TestBlock:    testBlock,
	}, nil
}

// NewFormulaDataSimple creates FormulaData for simple binary-only packages
// (no build system, just extract and install)
func NewFormulaDataSimple(packageName, version, sha256, url, description, homepage, license, binaryName string) *FormulaData {
	installBlock := fmt.Sprintf(`def install
    bin.install "%s"
  end`, binaryName)

	testBlock := fmt.Sprintf(`test do
    system "#{bin}/%s", "--version"
  end`, binaryName)

	return &FormulaData{
		ClassName:    PackageNameToClassName(packageName),
		PackageName:  packageName,
		Version:      version,
		SHA256:       sha256,
		URL:          url,
		Description:  description,
		Homepage:     homepage,
		License:      license,
		BuildSystem:  "Binary",
		Dependencies: []string{},
		InstallBlock: installBlock,
		TestBlock:    testBlock,
	}
}
