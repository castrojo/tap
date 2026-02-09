package archive

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// FileEntry represents a file in an archive
type FileEntry struct {
	Path string // Full path in archive
	Size int64
	Mode int64
}

// ListFiles lists all files in a tar archive (supports .tar.gz, .tar.xz, .tar.bz2)
// Returns list of file paths found in the archive
func ListFiles(data []byte, filename string) ([]string, error) {
	// Decompress based on extension
	var reader io.Reader = bytes.NewReader(data)
	var err error

	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress gzip: %w", err)
		}
		defer reader.(io.Closer).Close()
	} else if strings.HasSuffix(filename, ".tar.xz") {
		reader, err = xz.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress xz: %w", err)
		}
	} else if strings.HasSuffix(filename, ".tar.bz2") {
		reader = bzip2.NewReader(reader)
	} else if !strings.HasSuffix(filename, ".tar") {
		return nil, fmt.Errorf("unsupported archive format: %s", filename)
	}

	// Read tar entries
	tarReader := tar.NewReader(reader)
	var files []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Only include regular files (not directories)
		if header.Typeflag == tar.TypeReg {
			files = append(files, header.Name)
		}
	}

	return files, nil
}

// DetectBinaries finds executable files in the archive
// Returns paths to potential binary executables
// The list is sorted with most likely binaries first
func DetectBinaries(files []string) []string {
	var binaries []string

	// Common binary locations
	binPaths := []string{"bin/", "usr/bin/", "usr/local/bin/"}

	// Common non-binary file patterns to exclude
	excludePatterns := []string{
		"LICENSE", "README", "CHANGELOG", "COPYING", "AUTHORS",
		"NOTICE", "PATENTS", "VERSION", "MANIFEST", "TODO",
	}

	// Patterns for support files (not main binaries)
	supportPatterns := []string{
		"autocomplete/", "completions/", "bash_completion/",
		"zsh/", "fish/", "man/", "doc/", "docs/",
	}

	for _, file := range files {
		base := filepath.Base(file)
		baseUpper := strings.ToUpper(base)

		// Exclude documentation files
		isDoc := false
		for _, pattern := range excludePatterns {
			if strings.HasPrefix(baseUpper, pattern) {
				isDoc = true
				break
			}
		}
		if isDoc {
			continue
		}

		// Exclude support files (completions, man pages, etc.)
		isSupport := false
		for _, pattern := range supportPatterns {
			if strings.Contains(strings.ToLower(file), pattern) {
				isSupport = true
				break
			}
		}
		if isSupport {
			continue
		}

		// Exclude text files
		ext := strings.ToLower(filepath.Ext(file))
		if ext == ".txt" || ext == ".md" || ext == ".rst" || ext == ".pdf" ||
			ext == ".html" || ext == ".xml" || ext == ".json" || ext == ".yaml" || ext == ".yml" {
			continue
		}

		// Check if in a bin directory
		inBinDir := false
		for _, binPath := range binPaths {
			if strings.Contains(file, binPath) {
				inBinDir = true
				break
			}
		}

		if inBinDir {
			// Exclude shell scripts (unless they're the only thing there)
			if !strings.HasSuffix(base, ".sh") && !strings.HasSuffix(base, ".bash") {
				binaries = append(binaries, file)
			}
		}
	}

	// If no binaries found in standard locations, look for executables anywhere
	if len(binaries) == 0 {
		for _, file := range files {
			base := filepath.Base(file)
			baseUpper := strings.ToUpper(base)

			// Exclude documentation files
			isDoc := false
			for _, pattern := range excludePatterns {
				if strings.HasPrefix(baseUpper, pattern) {
					isDoc = true
					break
				}
			}
			if isDoc {
				continue
			}

			// Exclude support files
			isSupport := false
			for _, pattern := range supportPatterns {
				if strings.Contains(strings.ToLower(file), pattern) {
					isSupport = true
					break
				}
			}
			if isSupport {
				continue
			}

			// Exclude text and data files
			ext := strings.ToLower(filepath.Ext(file))
			if ext == ".txt" || ext == ".md" || ext == ".rst" || ext == ".pdf" ||
				ext == ".html" || ext == ".xml" || ext == ".json" || ext == ".yaml" || ext == ".yml" ||
				ext == ".conf" || ext == ".cfg" || ext == ".ini" {
				continue
			}

			// Heuristic: files without extension (likely binaries)
			// or with known binary extensions
			if ext == "" || ext == ".bin" || ext == ".elf" {
				binaries = append(binaries, file)
			}
		}
	}

	return binaries
}

// SelectBestBinary selects the most likely main binary from a list
// Prefers binaries that match the package name
func SelectBestBinary(binaries []string, packageName string) string {
	if len(binaries) == 0 {
		return ""
	}

	if len(binaries) == 1 {
		return binaries[0]
	}

	pkgLower := strings.ToLower(packageName)

	// Look for exact match
	for _, bin := range binaries {
		base := strings.ToLower(filepath.Base(bin))
		if base == pkgLower {
			return bin
		}
	}

	// Look for partial match
	for _, bin := range binaries {
		base := strings.ToLower(filepath.Base(bin))
		if strings.Contains(base, pkgLower) || strings.Contains(pkgLower, base) {
			return bin
		}
	}

	// Return first binary as fallback
	return binaries[0]
}

// FindRootDirectory finds the common root directory in archive
// Many tarballs wrap everything in app-version/ directory
func FindRootDirectory(files []string) string {
	if len(files) == 0 {
		return ""
	}

	// Get first path component from first file
	firstFile := files[0]
	parts := strings.Split(firstFile, "/")
	if len(parts) < 2 {
		return ""
	}

	candidate := parts[0] + "/"

	// Check if all files start with this prefix
	for _, file := range files[1:] {
		if !strings.HasPrefix(file, candidate) {
			return "" // No common root
		}
	}

	return candidate
}
