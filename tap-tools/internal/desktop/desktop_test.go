package desktop

import (
	"testing"
)

func TestDetectDesktopFile(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		want    string
		wantErr bool
	}{
		{
			name:    "Simple desktop file",
			files:   []string{"app/app.desktop", "app/bin/app"},
			want:    "app/app.desktop",
			wantErr: false,
		},
		{
			name:    "Nested desktop file",
			files:   []string{"app/share/applications/app.desktop", "app/bin/app"},
			want:    "app/share/applications/app.desktop",
			wantErr: false,
		},
		{
			name:    "Case insensitive",
			files:   []string{"app/App.DESKTOP", "app/bin/app"},
			want:    "app/App.DESKTOP",
			wantErr: false,
		},
		{
			name:    "No desktop file",
			files:   []string{"app/bin/app", "app/README.md"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectDesktopFile(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectDesktopFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Path != tt.want {
				t.Errorf("DetectDesktopFile() = %v, want %v", got.Path, tt.want)
			}
		})
	}
}

func TestDetectIcon(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		want    string // Path we expect to be selected
		wantErr bool
	}{
		{
			name: "Prefer larger icon",
			files: []string{
				"app/icons/16x16/app.png",
				"app/icons/128x128/app.png",
				"app/icons/48x48/app.png",
			},
			want:    "app/icons/128x128/app.png",
			wantErr: false,
		},
		{
			name: "SVG icon",
			files: []string{
				"app/icons/scalable/app.svg",
				"app/bin/app",
			},
			want:    "app/icons/scalable/app.svg",
			wantErr: false,
		},
		{
			name: "Icon in share directory",
			files: []string{
				"app/share/icons/hicolor/128x128/apps/app.png",
			},
			want:    "app/share/icons/hicolor/128x128/apps/app.png",
			wantErr: false,
		},
		{
			name:    "No icon",
			files:   []string{"app/bin/app", "app/README.md"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectIcon(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectIcon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Path != tt.want {
				t.Errorf("DetectIcon() = %v, want %v", got.Path, tt.want)
			}
		})
	}
}

func TestExtractIconSize(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"app/icons/128x128/app.png", "128x128"},
		{"app/icons/hicolor/128x128/apps/app.png", "hicolor"}, // Returns first match
		{"app/share/pixmaps/app.png", "unknown"},
		{"app/icons/scalable/app.svg", "scalable"},
		{"app/icons/hicolor/scalable/apps/app.svg", "hicolor"}, // Returns first match
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractIconSize(tt.path)
			if got != tt.want {
				t.Errorf("extractIconSize(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestSelectBestIcon(t *testing.T) {
	tests := []struct {
		name       string
		candidates []*IconInfo
		want       string // Filename of expected selection
	}{
		{
			name: "Prefer larger size",
			candidates: []*IconInfo{
				{Filename: "app-48.png", Size: "48x48"},
				{Filename: "app-128.png", Size: "128x128"},
				{Filename: "app-64.png", Size: "64x64"},
			},
			want: "app-128.png",
		},
		{
			name: "Prefer SVG over PNG when same size",
			candidates: []*IconInfo{
				{Filename: "app-128.png", Size: "128x128"},
				{Filename: "app-128.svg", Size: "128x128"},
			},
			want: "app-128.svg",
		},
		{
			name: "Prefer hicolor",
			candidates: []*IconInfo{
				{Filename: "app.png", Size: "unknown"},
				{Filename: "app-hicolor.png", Size: "hicolor"},
			},
			want: "app-hicolor.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectBestIcon(tt.candidates)
			if got.Filename != tt.want {
				t.Errorf("selectBestIcon() = %v, want %v", got.Filename, tt.want)
			}
		})
	}
}

func TestGenerateXDGPaths(t *testing.T) {
	tests := []struct {
		name           string
		hasDesktopFile bool
		hasIcon        bool
		wantCount      int
	}{
		{"Both", true, true, 2},
		{"Desktop only", true, false, 1},
		{"Icon only", false, true, 1},
		{"Neither", false, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateXDGPaths(tt.hasDesktopFile, tt.hasIcon)
			if len(got) != tt.wantCount {
				t.Errorf("GenerateXDGPaths() returned %d paths, want %d", len(got), tt.wantCount)
			}
		})
	}
}
