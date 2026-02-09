package homebrew

import (
	"strings"
	"testing"
)

func TestPackageNameToClassName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lowercase",
			input:    "jq",
			expected: "Jq",
		},
		{
			name:     "Single word",
			input:    "ripgrep",
			expected: "Ripgrep",
		},
		{
			name:     "Hyphenated name",
			input:    "go-task",
			expected: "GoTask",
		},
		{
			name:     "Underscore name",
			input:    "node_exporter",
			expected: "NodeExporter",
		},
		{
			name:     "Multiple hyphens",
			input:    "visual-studio-code",
			expected: "VisualStudioCode",
		},
		{
			name:     "Mixed separators",
			input:    "my-cool_app",
			expected: "MyCoolApp",
		},
		{
			name:     "Already capitalized",
			input:    "MyApp",
			expected: "MyApp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PackageNameToClassName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateFormula(t *testing.T) {
	t.Run("Simple formula", func(t *testing.T) {
		data := &FormulaData{
			ClassName:    "Jq",
			PackageName:  "jq",
			Version:      "1.7.1",
			SHA256:       "abc123",
			URL:          "https://github.com/jqlang/jq/releases/download/v1.7.1/jq-1.7.1.tar.gz",
			Description:  "Command-line JSON processor",
			Homepage:     "https://jqlang.github.io/jq",
			License:      "MIT",
			BuildSystem:  "Makefile",
			Dependencies: []string{},
			InstallBlock: "def install\n    bin.install \"jq\"\n  end",
			TestBlock:    "test do\n    system \"#{bin}/jq\", \"--version\"\n  end",
		}

		result, err := GenerateFormula(data)
		if err != nil {
			t.Fatalf("Failed to generate formula: %v", err)
		}

		// Check essential parts
		if !strings.Contains(result, "class Jq < Formula") {
			t.Error("Formula should contain class definition")
		}
		if !strings.Contains(result, "desc \"Command-line JSON processor\"") {
			t.Error("Formula should contain description")
		}
		if !strings.Contains(result, "homepage \"https://jqlang.github.io/jq\"") {
			t.Error("Formula should contain homepage")
		}
		if !strings.Contains(result, "url \"https://github.com/jqlang/jq") {
			t.Error("Formula should contain URL")
		}
		if !strings.Contains(result, "sha256 \"abc123\"") {
			t.Error("Formula should contain SHA256")
		}
		if !strings.Contains(result, "license \"MIT\"") {
			t.Error("Formula should contain license")
		}
		if !strings.Contains(result, "def install") {
			t.Error("Formula should contain install block")
		}
		if !strings.Contains(result, "test do") {
			t.Error("Formula should contain test block")
		}
	})

	t.Run("Formula with dependencies", func(t *testing.T) {
		data := &FormulaData{
			ClassName:    "Mytool",
			PackageName:  "mytool",
			Version:      "1.0.0",
			SHA256:       "def456",
			URL:          "https://example.com/mytool-1.0.0.tar.gz",
			Description:  "Example tool",
			Homepage:     "https://example.com",
			License:      "Apache-2.0",
			BuildSystem:  "Go",
			Dependencies: []string{"go", "openssl@3"},
			InstallBlock: "def install\n    system \"go\", \"build\"\n  end",
			TestBlock:    "test do\n    system \"#{bin}/mytool\", \"--help\"\n  end",
		}

		result, err := GenerateFormula(data)
		if err != nil {
			t.Fatalf("Failed to generate formula: %v", err)
		}

		if !strings.Contains(result, "depends_on \"go\"") {
			t.Error("Formula should contain go dependency")
		}
		if !strings.Contains(result, "depends_on \"openssl@3\"") {
			t.Error("Formula should contain openssl dependency")
		}
	})

	t.Run("Formula without license", func(t *testing.T) {
		data := &FormulaData{
			ClassName:    "Notool",
			PackageName:  "notool",
			Version:      "1.0.0",
			SHA256:       "xyz789",
			URL:          "https://example.com/notool-1.0.0.tar.gz",
			Description:  "Tool without license",
			Homepage:     "https://example.com",
			License:      "",
			BuildSystem:  "Binary",
			Dependencies: []string{},
			InstallBlock: "def install\n    bin.install \"notool\"\n  end",
			TestBlock:    "test do\n    system \"#{bin}/notool\", \"--version\"\n  end",
		}

		result, err := GenerateFormula(data)
		if err != nil {
			t.Fatalf("Failed to generate formula: %v", err)
		}

		// The template should not render the license line when license is empty
		// But there might be whitespace, so check for 'license "' specifically
		if strings.Contains(result, "license \"") {
			t.Errorf("Formula should not contain license line when license is empty. Got:\n%s", result)
		}
	})
}

func TestNewFormulaData(t *testing.T) {
	t.Run("Go project", func(t *testing.T) {
		repoFiles := []string{
			"main.go",
			"go.mod",
			"go.sum",
			"README.md",
		}

		data, err := NewFormulaData(
			"mytool",
			"1.0.0",
			"abc123",
			"https://example.com/mytool-1.0.0.tar.gz",
			"My awesome tool",
			"https://example.com",
			"MIT",
			repoFiles,
			"mytool",
		)

		if err != nil {
			t.Fatalf("Failed to create formula data: %v", err)
		}

		if data.ClassName != "Mytool" {
			t.Errorf("Expected class name 'Mytool', got %s", data.ClassName)
		}

		if data.BuildSystem != "Go" {
			t.Errorf("Expected build system 'Go', got %s", data.BuildSystem)
		}

		if len(data.Dependencies) == 0 {
			t.Error("Go projects should have go dependency")
		}

		if !strings.Contains(data.InstallBlock, "go") {
			t.Error("Go project install block should contain 'go' command")
		}

		if !strings.Contains(data.TestBlock, "mytool") {
			t.Error("Test block should reference binary name")
		}
	})

	t.Run("Rust project", func(t *testing.T) {
		repoFiles := []string{
			"src/main.rs",
			"Cargo.toml",
			"Cargo.lock",
		}

		data, err := NewFormulaData(
			"rust-app",
			"2.0.0",
			"def456",
			"https://example.com/rust-app-2.0.0.tar.gz",
			"Rust application",
			"https://example.com",
			"Apache-2.0",
			repoFiles,
			"rust-app",
		)

		if err != nil {
			t.Fatalf("Failed to create formula data: %v", err)
		}

		if data.ClassName != "RustApp" {
			t.Errorf("Expected class name 'RustApp', got %s", data.ClassName)
		}

		if data.BuildSystem != "Rust" {
			t.Errorf("Expected build system 'Rust', got %s", data.BuildSystem)
		}

		if !strings.Contains(data.InstallBlock, "cargo") {
			t.Error("Rust project install block should contain 'cargo' command")
		}
	})

	t.Run("CMake project", func(t *testing.T) {
		repoFiles := []string{
			"src/main.c",
			"CMakeLists.txt",
			"README.md",
		}

		data, err := NewFormulaData(
			"cmake-tool",
			"0.5.0",
			"xyz789",
			"https://example.com/cmake-tool-0.5.0.tar.gz",
			"CMake-based tool",
			"https://example.com",
			"GPL-3.0",
			repoFiles,
			"cmake-tool",
		)

		if err != nil {
			t.Fatalf("Failed to create formula data: %v", err)
		}

		if data.BuildSystem != "CMake" {
			t.Errorf("Expected build system 'CMake', got %s", data.BuildSystem)
		}

		if !strings.Contains(data.InstallBlock, "cmake") {
			t.Error("CMake project install block should contain 'cmake' command")
		}
	})

	t.Run("No build system detected", func(t *testing.T) {
		repoFiles := []string{
			"README.md",
			"LICENSE",
		}

		_, err := NewFormulaData(
			"unknown",
			"1.0.0",
			"abc123",
			"https://example.com/unknown-1.0.0.tar.gz",
			"Unknown project",
			"https://example.com",
			"MIT",
			repoFiles,
			"unknown",
		)

		if err == nil {
			t.Error("Expected error when no build system is detected")
		}

		if !strings.Contains(err.Error(), "could not detect build system") {
			t.Errorf("Expected 'could not detect build system' error, got: %v", err)
		}
	})
}

func TestNewFormulaDataSimple(t *testing.T) {
	t.Run("Simple binary formula", func(t *testing.T) {
		data := NewFormulaDataSimple(
			"simple-tool",
			"3.0.0",
			"hash123",
			"https://example.com/simple-tool-3.0.0-linux-x64.tar.gz",
			"Simple binary tool",
			"https://example.com",
			"BSD-3-Clause",
			"simple-tool",
		)

		if data.ClassName != "SimpleTool" {
			t.Errorf("Expected class name 'SimpleTool', got %s", data.ClassName)
		}

		if data.BuildSystem != "Binary" {
			t.Errorf("Expected build system 'Binary', got %s", data.BuildSystem)
		}

		if len(data.Dependencies) != 0 {
			t.Error("Binary-only formula should have no dependencies")
		}

		if !strings.Contains(data.InstallBlock, "bin.install \"simple-tool\"") {
			t.Error("Install block should use bin.install")
		}

		if !strings.Contains(data.TestBlock, "simple-tool") {
			t.Error("Test block should reference binary name")
		}

		if !strings.Contains(data.TestBlock, "--version") {
			t.Error("Test block should test --version flag")
		}
	})

	t.Run("Binary with different name", func(t *testing.T) {
		data := NewFormulaDataSimple(
			"my-package",
			"1.0.0",
			"abc123",
			"https://example.com/package-1.0.0.tar.gz",
			"Package with different binary name",
			"https://example.com",
			"MIT",
			"actual-binary-name",
		)

		if !strings.Contains(data.InstallBlock, "actual-binary-name") {
			t.Error("Install block should use the provided binary name")
		}

		if !strings.Contains(data.TestBlock, "actual-binary-name") {
			t.Error("Test block should use the provided binary name")
		}
	})
}

func TestGenerateFormulaIntegration(t *testing.T) {
	t.Run("Full Go project formula", func(t *testing.T) {
		repoFiles := []string{"main.go", "go.mod", "README.md"}

		data, err := NewFormulaData(
			"ripgrep",
			"14.0.0",
			"abcdef1234567890",
			"https://github.com/BurntSushi/ripgrep/archive/v14.0.0.tar.gz",
			"Recursively search directories for a regex pattern",
			"https://github.com/BurntSushi/ripgrep",
			"MIT",
			repoFiles,
			"rg",
		)

		if err != nil {
			t.Fatalf("Failed to create formula data: %v", err)
		}

		formula, err := GenerateFormula(data)
		if err != nil {
			t.Fatalf("Failed to generate formula: %v", err)
		}

		// Verify the complete formula structure
		expectedParts := []string{
			"class Ripgrep < Formula",
			"desc \"Recursively search directories for a regex pattern\"",
			"homepage \"https://github.com/BurntSushi/ripgrep\"",
			"url \"https://github.com/BurntSushi/ripgrep/archive/v14.0.0.tar.gz\"",
			"sha256 \"abcdef1234567890\"",
			"license \"MIT\"",
			"depends_on \"go\"",
			"def install",
			"test do",
			"end",
		}

		for _, part := range expectedParts {
			if !strings.Contains(formula, part) {
				t.Errorf("Formula missing expected part: %s", part)
			}
		}

		// Verify formula is valid Ruby syntax (basic check)
		if !strings.HasPrefix(formula, "class") {
			t.Error("Formula should start with 'class'")
		}

		if !strings.HasSuffix(strings.TrimSpace(formula), "end") {
			t.Error("Formula should end with 'end'")
		}
	})
}
