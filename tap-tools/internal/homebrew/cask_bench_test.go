package homebrew

import (
	"testing"
)

func BenchmarkGenerateCask(b *testing.B) {
	data := &CaskData{
		Token:       "sublime-text-linux",
		Version:     "4200",
		SHA256:      "abc123def456",
		URL:         "https://example.com/sublime_text_build_4200_x64.tar.gz",
		AppName:     "Sublime Text",
		Description: "Text editor for code, markup and prose",
		Homepage:    "https://www.sublimetext.com",
		BinaryName:  "sublime_text",
		BinaryPath:  "sublime_text",
		ZapTrash:    []string{"~/.config/sublime-text"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateCask(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewCaskData(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCaskData("sublime-text", "4200", "abc123def456", "https://example.com/sublime_text_build_4200_x64.tar.gz")
	}
}
