package platform

import (
	"fmt"
	"regexp"
	"strings"
)

// Platform represents a software platform
type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformUnknown Platform = "unknown"
)

// Architecture represents CPU architecture
type Architecture string

const (
	ArchX86_64  Architecture = "x86_64"
	ArchAMD64   Architecture = "amd64"
	ArchARM64   Architecture = "arm64"
	ArchARM     Architecture = "arm"
	ArchUnknown Architecture = "unknown"
)

// Format represents package format
type Format string

const (
	FormatTarGz    Format = "tar.gz"
	FormatTarXz    Format = "tar.xz"
	FormatTarBz2   Format = "tar.bz2"
	FormatTgz      Format = "tgz"
	FormatDeb      Format = "deb"
	FormatRpm      Format = "rpm"
	FormatAppImage Format = "appimage"
	FormatUnknown  Format = "unknown"
)

// Priority levels for package formats (lower is better)
const (
	PriorityTarball = 1 // .tar.gz, .tar.xz, .tgz
	PriorityDeb     = 2 // .deb
	PriorityOther   = 3 // Everything else
)

// Asset represents a release asset with detected metadata
type Asset struct {
	Name        string
	URL         string
	DownloadURL string
	Size        int64
	Platform    Platform
	Arch        Architecture
	Format      Format
	Priority    int
	IsSource    bool
	IsChecksum  bool
}

// DetectPlatform analyzes a filename and returns asset metadata
func DetectPlatform(filename string) *Asset {
	lower := strings.ToLower(filename)

	asset := &Asset{
		Name:     filename,
		Platform: detectPlatformFromFilename(lower),
		Arch:     detectArchFromFilename(lower),
		Format:   detectFormatFromFilename(lower),
	}

	// Check if it's a source archive
	asset.IsSource = isSourceArchive(lower)

	// Check if it's a checksum file
	asset.IsChecksum = isChecksumFile(lower)

	// Assign priority based on format
	asset.Priority = getPriority(asset.Format)

	return asset
}

// detectPlatformFromFilename detects the platform from filename
// For Linux-only tap, we only detect Linux formats
func detectPlatformFromFilename(filename string) Platform {
	// Check format first - .deb and .rpm are Linux-specific
	if strings.HasSuffix(filename, ".deb") || strings.HasSuffix(filename, ".rpm") {
		return PlatformLinux
	}

	// Linux patterns
	linuxPatterns := []string{
		"linux", "ubuntu", "debian", "fedora", "rhel",
		"centos", "alpine", "arch", "opensuse",
	}
	for _, pattern := range linuxPatterns {
		if strings.Contains(filename, pattern) {
			return PlatformLinux
		}
	}

	// Reject non-Linux patterns (macOS and Windows)
	nonLinuxPatterns := []string{
		"macos", "darwin", "osx", "mac",
		"windows", "win32", "win64",
	}
	for _, pattern := range nonLinuxPatterns {
		if strings.Contains(filename, pattern) {
			return PlatformUnknown // Explicitly mark as unknown/rejected
		}
	}

	return PlatformUnknown
}

// detectArchFromFilename detects the architecture from filename
func detectArchFromFilename(filename string) Architecture {
	// x86_64 / AMD64 patterns
	x64Patterns := []string{
		"x86_64", "x86-64", "amd64", "x64",
	}
	for _, pattern := range x64Patterns {
		if strings.Contains(filename, pattern) {
			return ArchX86_64
		}
	}

	// ARM64 patterns
	arm64Patterns := []string{
		"arm64", "aarch64", "armv8",
	}
	for _, pattern := range arm64Patterns {
		if strings.Contains(filename, pattern) {
			return ArchARM64
		}
	}

	// ARM patterns
	armPatterns := []string{
		"armv7", "armhf", "arm",
	}
	for _, pattern := range armPatterns {
		if strings.Contains(filename, pattern) {
			return ArchARM
		}
	}

	return ArchUnknown
}

// detectFormatFromFilename detects the package format from filename
func detectFormatFromFilename(filename string) Format {
	switch {
	case strings.HasSuffix(filename, ".tar.gz"):
		return FormatTarGz
	case strings.HasSuffix(filename, ".tar.xz"):
		return FormatTarXz
	case strings.HasSuffix(filename, ".tar.bz2"):
		return FormatTarBz2
	case strings.HasSuffix(filename, ".tgz"):
		return FormatTgz
	case strings.HasSuffix(filename, ".deb"):
		return FormatDeb
	case strings.HasSuffix(filename, ".rpm"):
		return FormatRpm
	case strings.HasSuffix(strings.ToLower(filename), ".appimage"):
		return FormatAppImage
	default:
		return FormatUnknown
	}
}

// isSourceArchive checks if the filename looks like a source code archive
func isSourceArchive(filename string) bool {
	sourcePatterns := []string{
		"source", "src", "sources",
	}
	for _, pattern := range sourcePatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// isChecksumFile checks if the filename is a checksum file
func isChecksumFile(filename string) bool {
	checksumPatterns := []string{
		"checksum", "sha256", "sha512", "md5",
		"sums.txt", "checksums.txt",
	}
	for _, pattern := range checksumPatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// getPriority returns the priority for a given format
func getPriority(format Format) int {
	switch format {
	case FormatTarGz, FormatTarXz, FormatTarBz2, FormatTgz:
		return PriorityTarball
	case FormatDeb:
		return PriorityDeb
	default:
		return PriorityOther
	}
}

// FilterLinuxAssets filters assets to only include Linux packages
// Excludes: source archives, checksums, non-Linux platforms
func FilterLinuxAssets(assets []*Asset) []*Asset {
	var filtered []*Asset

	for _, asset := range assets {
		// Skip source archives and checksums
		if asset.IsSource || asset.IsChecksum {
			continue
		}

		// Skip explicitly non-Linux platforms
		if asset.Platform == PlatformUnknown && !isLikelyLinux(asset) {
			continue
		}

		// Skip unknown formats (unless it's explicitly Linux)
		if asset.Format == FormatUnknown && asset.Platform != PlatformLinux {
			continue
		}

		// Include Linux or likely Linux packages
		if asset.Platform == PlatformLinux {
			filtered = append(filtered, asset)
		}
	}

	return filtered
}

// isLikelyLinux checks if an asset is likely for Linux based on format
func isLikelyLinux(asset *Asset) bool {
	// Tarballs could be universal, so we include them
	return asset.Format == FormatTarGz ||
		asset.Format == FormatTarXz ||
		asset.Format == FormatTarBz2 ||
		asset.Format == FormatTgz
}

// SelectBestAsset selects the best asset from a list based on priority
// Priority order: tarball > deb > other
// If multiple assets have the same priority, prefer x86_64/amd64
func SelectBestAsset(assets []*Asset) (*Asset, error) {
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets to select from")
	}

	// Find the highest priority (lowest number)
	bestPriority := PriorityOther + 1
	for _, asset := range assets {
		if asset.Priority < bestPriority {
			bestPriority = asset.Priority
		}
	}

	// Filter assets with the best priority
	var candidates []*Asset
	for _, asset := range assets {
		if asset.Priority == bestPriority {
			candidates = append(candidates, asset)
		}
	}

	// If only one candidate, return it
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Prefer x86_64/amd64 architecture
	for _, asset := range candidates {
		if asset.Arch == ArchX86_64 || asset.Arch == ArchAMD64 {
			return asset, nil
		}
	}

	// Return the first candidate
	return candidates[0], nil
}

// NormalizePackageName normalizes a repository name to a package name
// Example: "My_Cool_App" -> "my-cool-app"
func NormalizePackageName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace underscores and spaces with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any non-alphanumeric characters except hyphens
	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "")

	// Remove consecutive hyphens
	re = regexp.MustCompile(`-+`)
	name = re.ReplaceAllString(name, "-")

	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	return name
}

// EnsureLinuxSuffix ensures the cask name has a -linux suffix
func EnsureLinuxSuffix(name string) string {
	if strings.HasSuffix(name, "-linux") {
		return name
	}
	return name + "-linux"
}
