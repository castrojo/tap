package platform

import (
	"testing"
)

func BenchmarkDetectPlatform(b *testing.B) {
	filename := "sublime_text_build_4200_x64.tar.gz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectPlatform(filename)
	}
}

func BenchmarkFilterLinuxAssets(b *testing.B) {
	assets := []*Asset{
		{Name: "app-linux-x64.tar.gz", Platform: "linux", Format: "tar.gz", Priority: 1},
		{Name: "app-darwin-arm64.tar.gz", Platform: "macos", Format: "tar.gz", Priority: 1},
		{Name: "app-linux-amd64.deb", Platform: "linux", Format: "deb", Priority: 2},
		{Name: "app-windows-x64.zip", Platform: "windows", Format: "zip", Priority: 3},
		{Name: "app-linux-arm64.tar.gz", Platform: "linux", Format: "tar.gz", Priority: 1},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterLinuxAssets(assets)
	}
}

func BenchmarkSelectBestAsset(b *testing.B) {
	assets := []*Asset{
		{Name: "app-linux-amd64.deb", Platform: "linux", Arch: "x86_64", Format: "deb", Priority: 2},
		{Name: "app-linux-x64.tar.gz", Platform: "linux", Arch: "x86_64", Format: "tar.gz", Priority: 1},
		{Name: "app-linux-arm64.tar.gz", Platform: "linux", Arch: "arm64", Format: "tar.gz", Priority: 1},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SelectBestAsset(assets)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNormalizePackageName(b *testing.B) {
	testCases := []string{
		"sublime-text",
		"Sublime_Text",
		"My_Cool_App",
		"App___With___Underscores",
		"App@#$%Special",
		"--leading-trailing--",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testCases {
			_ = NormalizePackageName(name)
		}
	}
}

func BenchmarkEnsureLinuxSuffix(b *testing.B) {
	testCases := []string{
		"sublime-text",
		"sublime-text-linux",
		"app",
		"tool-linux",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testCases {
			_ = EnsureLinuxSuffix(name)
		}
	}
}
