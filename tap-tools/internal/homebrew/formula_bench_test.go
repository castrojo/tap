package homebrew

import (
	"testing"
)

func BenchmarkGenerateFormula(b *testing.B) {
	data := &FormulaData{
		ClassName:    "Ripgrep",
		PackageName:  "ripgrep",
		Version:      "14.0.0",
		SHA256:       "abc123def456",
		URL:          "https://github.com/BurntSushi/ripgrep/releases/download/14.0.0/ripgrep-14.0.0-x86_64-unknown-linux-musl.tar.gz",
		Description:  "Recursively searches directories for a regex pattern",
		Homepage:     "https://github.com/BurntSushi/ripgrep",
		License:      "Unlicense",
		BuildSystem:  "rust",
		InstallBlock: `system "cargo", "install", "--locked", "--root", prefix, "--path", "."`,
		TestBlock:    `system "#{bin}/rg", "--version"`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateFormula(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewFormulaData(b *testing.B) {
	repoFiles := []string{"Cargo.toml", "Cargo.lock", "src/main.rs"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewFormulaData("ripgrep", "14.0.0", "abc123def456",
			"https://github.com/BurntSushi/ripgrep/releases/download/14.0.0/ripgrep-14.0.0-x86_64-unknown-linux-musl.tar.gz",
			"Recursively searches directories for a regex pattern",
			"https://github.com/BurntSushi/ripgrep",
			"Unlicense",
			repoFiles,
			"rg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPackageNameToClassName(b *testing.B) {
	testCases := []string{
		"jq",
		"ripgrep",
		"sublime-text",
		"visual-studio-code",
		"my_cool_app",
		"tool-with-many-hyphens",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testCases {
			_ = PackageNameToClassName(name)
		}
	}
}
