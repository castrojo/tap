package platform

import (
	"testing"
)

func TestDetectPlatformFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     Platform
	}{
		// Linux
		{"app-linux-x64.tar.gz", PlatformLinux},
		{"tool_ubuntu_amd64.deb", PlatformLinux},
		{"program-debian.tar.xz", PlatformLinux},
		{"binary-fedora-x86_64.rpm", PlatformLinux},
		// Non-Linux (should be rejected/unknown)
		{"app-macos-arm64.tar.gz", PlatformUnknown},
		{"tool-darwin-x64.tar.gz", PlatformUnknown},
		{"app-windows-x64.tar.gz", PlatformUnknown},
		{"tool-win64.tar.gz", PlatformUnknown},
		// Unknown (generic)
		{"generic-binary.tar.gz", PlatformUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectPlatformFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("detectPlatformFromFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectArchFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     Architecture
	}{
		{"app-x86_64.tar.gz", ArchX86_64},
		{"tool-amd64.deb", ArchX86_64},
		{"program-x64.zip", ArchX86_64},
		{"binary-arm64.tar.gz", ArchARM64},
		{"tool-aarch64.deb", ArchARM64},
		{"app-armv7.tar.gz", ArchARM},
		{"generic.tar.gz", ArchUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectArchFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("detectArchFromFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectFormatFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     Format
	}{
		{"app.tar.gz", FormatTarGz},
		{"tool.tar.xz", FormatTarXz},
		{"program.tgz", FormatTgz},
		{"binary.deb", FormatDeb},
		{"package.rpm", FormatRpm},
		{"app.AppImage", FormatAppImage},
		{"unknown", FormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := detectFormatFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("detectFormatFromFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     *Asset
	}{
		{
			name:     "Linux tarball x86_64",
			filename: "sublime-text-linux-x64.tar.gz",
			want: &Asset{
				Name:       "sublime-text-linux-x64.tar.gz",
				Platform:   PlatformLinux,
				Arch:       ArchX86_64,
				Format:     FormatTarGz,
				Priority:   PriorityTarball,
				IsSource:   false,
				IsChecksum: false,
			},
		},
		{
			name:     "Linux deb amd64",
			filename: "app_1.0.0_amd64.deb",
			want: &Asset{
				Name:       "app_1.0.0_amd64.deb",
				Platform:   PlatformLinux,
				Arch:       ArchX86_64,
				Format:     FormatDeb,
				Priority:   PriorityDeb,
				IsSource:   false,
				IsChecksum: false,
			},
		},
		{
			name:     "Source tarball",
			filename: "app-1.0.0-source.tar.gz",
			want: &Asset{
				Name:       "app-1.0.0-source.tar.gz",
				Platform:   PlatformUnknown,
				Arch:       ArchUnknown,
				Format:     FormatTarGz,
				Priority:   PriorityTarball,
				IsSource:   true,
				IsChecksum: false,
			},
		},
		{
			name:     "Checksum file",
			filename: "sha256sums.txt",
			want: &Asset{
				Name:       "sha256sums.txt",
				Platform:   PlatformUnknown,
				Arch:       ArchUnknown,
				Format:     FormatUnknown,
				Priority:   PriorityOther,
				IsSource:   false,
				IsChecksum: true,
			},
		},
		{
			name:     "Non-Linux tarball (should be rejected)",
			filename: "app-macos.tar.gz",
			want: &Asset{
				Name:       "app-macos.tar.gz",
				Platform:   PlatformUnknown,
				Arch:       ArchUnknown,
				Format:     FormatTarGz,
				Priority:   PriorityTarball,
				IsSource:   false,
				IsChecksum: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatform(tt.filename)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Platform != tt.want.Platform {
				t.Errorf("Platform = %v, want %v", got.Platform, tt.want.Platform)
			}
			if got.Arch != tt.want.Arch {
				t.Errorf("Arch = %v, want %v", got.Arch, tt.want.Arch)
			}
			if got.Format != tt.want.Format {
				t.Errorf("Format = %v, want %v", got.Format, tt.want.Format)
			}
			if got.Priority != tt.want.Priority {
				t.Errorf("Priority = %v, want %v", got.Priority, tt.want.Priority)
			}
			if got.IsSource != tt.want.IsSource {
				t.Errorf("IsSource = %v, want %v", got.IsSource, tt.want.IsSource)
			}
			if got.IsChecksum != tt.want.IsChecksum {
				t.Errorf("IsChecksum = %v, want %v", got.IsChecksum, tt.want.IsChecksum)
			}
		})
	}
}

func TestFilterLinuxAssets(t *testing.T) {
	assets := []*Asset{
		{Name: "app-linux-x64.tar.gz", Platform: PlatformLinux, Format: FormatTarGz, IsSource: false, IsChecksum: false},
		{Name: "app-source.tar.gz", Platform: PlatformUnknown, Format: FormatTarGz, IsSource: true, IsChecksum: false},
		{Name: "checksums.txt", Platform: PlatformUnknown, Format: FormatUnknown, IsSource: false, IsChecksum: true},
		{Name: "app-macos.tar.gz", Platform: PlatformUnknown, Format: FormatTarGz, IsSource: false, IsChecksum: false},
		{Name: "app-windows.tar.gz", Platform: PlatformUnknown, Format: FormatTarGz, IsSource: false, IsChecksum: false},
		{Name: "app_amd64.deb", Platform: PlatformLinux, Format: FormatDeb, IsSource: false, IsChecksum: false},
	}

	filtered := FilterLinuxAssets(assets)

	// Should only have 2 Linux assets
	if len(filtered) != 2 {
		t.Errorf("FilterLinuxAssets() returned %d assets, want 2", len(filtered))
	}

	// Check that we have the right ones
	wantNames := map[string]bool{
		"app-linux-x64.tar.gz": true,
		"app_amd64.deb":        true,
	}

	for _, asset := range filtered {
		if !wantNames[asset.Name] {
			t.Errorf("FilterLinuxAssets() included unexpected asset: %s", asset.Name)
		}
	}
}

func TestSelectBestAsset(t *testing.T) {
	tests := []struct {
		name    string
		assets  []*Asset
		want    string
		wantErr bool
	}{
		{
			name:    "Empty list",
			assets:  []*Asset{},
			want:    "",
			wantErr: true,
		},
		{
			name: "Prefer tarball over deb",
			assets: []*Asset{
				{Name: "app_amd64.deb", Priority: PriorityDeb, Arch: ArchX86_64},
				{Name: "app-linux-x64.tar.gz", Priority: PriorityTarball, Arch: ArchX86_64},
			},
			want:    "app-linux-x64.tar.gz",
			wantErr: false,
		},
		{
			name: "Prefer x86_64 when multiple tarballs",
			assets: []*Asset{
				{Name: "app-linux-arm64.tar.gz", Priority: PriorityTarball, Arch: ArchARM64},
				{Name: "app-linux-x64.tar.gz", Priority: PriorityTarball, Arch: ArchX86_64},
			},
			want:    "app-linux-x64.tar.gz",
			wantErr: false,
		},
		{
			name: "Single asset",
			assets: []*Asset{
				{Name: "app.tar.gz", Priority: PriorityTarball, Arch: ArchUnknown},
			},
			want:    "app.tar.gz",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectBestAsset(tt.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectBestAsset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want {
				t.Errorf("SelectBestAsset() = %v, want %v", got.Name, tt.want)
			}
		})
	}
}

func TestNormalizePackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sublime-text", "sublime-text"},
		{"Sublime_Text", "sublime-text"},
		{"My Cool App", "my-cool-app"},
		{"App___With___Underscores", "app-with-underscores"},
		{"App@#$%Special", "appspecial"},
		{"--leading-trailing--", "leading-trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizePackageName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePackageName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureLinuxSuffix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sublime-text", "sublime-text-linux"},
		{"sublime-text-linux", "sublime-text-linux"},
		{"app", "app-linux"},
		{"tool-linux", "tool-linux"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EnsureLinuxSuffix(tt.input)
			if got != tt.want {
				t.Errorf("EnsureLinuxSuffix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
