package desktop

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DesktopFileInfo represents a detected .desktop file
type DesktopFileInfo struct {
	Path     string // Relative path in archive
	Filename string // Just the filename
}

// IconInfo represents a detected icon file
type IconInfo struct {
	Path     string // Relative path in archive
	Filename string // Just the filename
	Size     string // Size like "128x128" or "hicolor"
}

// DetectDesktopFile searches for .desktop files in archive file list
func DetectDesktopFile(archiveFiles []string) (*DesktopFileInfo, error) {
	for _, file := range archiveFiles {
		if strings.HasSuffix(strings.ToLower(file), ".desktop") {
			return &DesktopFileInfo{
				Path:     file,
				Filename: filepath.Base(file),
			}, nil
		}
	}
	return nil, fmt.Errorf("no .desktop file found")
}

// DetectIcon searches for icon files in archive file list
// Prefers larger icons (256x256, 128x128) and common formats (png, svg)
func DetectIcon(archiveFiles []string) (*IconInfo, error) {
	var candidates []*IconInfo

	// Common icon extensions
	iconExts := []string{".png", ".svg", ".xpm", ".ico"}

	// Common icon paths
	iconPaths := []string{
		"icons/", "icon/", "pixmaps/", "share/icons/",
		"share/pixmaps/", ".local/share/icons/",
	}

	for _, file := range archiveFiles {
		lower := strings.ToLower(file)

		// Check if it's an icon file
		isIcon := false
		for _, ext := range iconExts {
			if strings.HasSuffix(lower, ext) {
				isIcon = true
				break
			}
		}

		if !isIcon {
			continue
		}

		// Check if it's in an icon directory or named "icon"
		inIconDir := false
		for _, iconPath := range iconPaths {
			if strings.Contains(lower, iconPath) {
				inIconDir = true
				break
			}
		}

		if !inIconDir && !strings.Contains(lower, "icon") {
			continue
		}

		// Extract size if present (e.g., 128x128, 256x256)
		size := extractIconSize(file)

		candidates = append(candidates, &IconInfo{
			Path:     file,
			Filename: filepath.Base(file),
			Size:     size,
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no icon file found")
	}

	// Select best icon (prefer larger sizes, then SVG, then PNG)
	return selectBestIcon(candidates), nil
}

// extractIconSize tries to extract size from icon path (e.g., "128x128")
func extractIconSize(path string) string {
	// Common patterns: icons/128x128/, icons/hicolor/128x128/, 128x128-icon.png
	parts := strings.Split(path, "/")
	for _, part := range parts {
		part = strings.ToLower(part)

		// Match named sizes first
		if part == "hicolor" || part == "scalable" {
			return part
		}

		// Match NxN pattern
		if strings.Contains(part, "x") {
			fields := strings.Split(part, "x")
			if len(fields) == 2 {
				// Check if both parts start with digits
				if len(fields[0]) > 0 && len(fields[1]) > 0 &&
					fields[0][0] >= '0' && fields[0][0] <= '9' &&
					fields[1][0] >= '0' && fields[1][0] <= '9' {
					return part
				}
			}
		}
	}
	return "unknown"
}

// selectBestIcon selects the best icon from candidates
// Priority: larger size > SVG > PNG > other
func selectBestIcon(candidates []*IconInfo) *IconInfo {
	if len(candidates) == 1 {
		return candidates[0]
	}

	// Assign scores to each candidate
	type scoredIcon struct {
		icon  *IconInfo
		score int
	}

	scored := make([]scoredIcon, len(candidates))
	for i, icon := range candidates {
		score := 0

		// Size scoring
		switch icon.Size {
		case "512x512":
			score += 1000
		case "256x256":
			score += 900
		case "hicolor", "scalable":
			score += 850
		case "128x128":
			score += 800
		case "64x64":
			score += 700
		case "48x48":
			score += 600
		case "32x32":
			score += 500
		case "16x16":
			score += 400
		default:
			score += 300 // Unknown size
		}

		// Format scoring
		lower := strings.ToLower(icon.Filename)
		if strings.HasSuffix(lower, ".svg") {
			score += 100
		} else if strings.HasSuffix(lower, ".png") {
			score += 90
		} else {
			score += 50
		}

		scored[i] = scoredIcon{icon: icon, score: score}
	}

	// Find highest score
	best := scored[0]
	for _, s := range scored[1:] {
		if s.score > best.score {
			best = s
		}
	}

	return best.icon
}

// GenerateXDGPaths generates the list of XDG directories to create
func GenerateXDGPaths(hasDesktopFile, hasIcon bool) []string {
	var paths []string
	if hasDesktopFile {
		paths = append(paths, "#{ENV.fetch(\"HOME\")}/.local/share/applications")
	}
	if hasIcon {
		paths = append(paths, "#{ENV.fetch(\"HOME\")}/.local/share/icons")
	}
	return paths
}
