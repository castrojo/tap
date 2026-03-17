package stats

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

var (
	caskRe      = regexp.MustCompile(`^cask\s+"([^"]+)"\s+do`)
	formulaRe   = regexp.MustCompile(`^(?:class\s+(\w+)\s+<\s+Formula|formula\s+"([^"]+)"\s+do)`)
	versionRe   = regexp.MustCompile(`^\s+version\s+"([^"]+)"`)
	sha256Re    = regexp.MustCompile(`^\s+sha256\s+"([a-f0-9]{64})"`)
	urlRe       = regexp.MustCompile(`^\s+url\s+"([^"]+)"`)
	nameRe      = regexp.MustCompile(`^\s+name\s+"([^"]+)"`)
	descRe      = regexp.MustCompile(`^\s+desc\s+"([^"]+)"`)
	homepageRe  = regexp.MustCompile(`^\s+homepage\s+"([^"]+)"`)
	githubURLRe = regexp.MustCompile(`github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)`)
)

// ParseRubyFile extracts package metadata from a Homebrew cask or formula file.
func ParseRubyFile(content, filePath string) (*Package, error) {
	pkg := &Package{FilePath: filePath}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case pkg.Name == "":
			if m := caskRe.FindStringSubmatch(line); m != nil {
				pkg.Name = m[1]
				pkg.Type = "cask"
			} else if m := formulaRe.FindStringSubmatch(line); m != nil {
				// class-style formula: class Name < Formula
				if m[1] != "" {
					pkg.Name = strings.ToLower(m[1])
				} else {
					pkg.Name = m[2]
				}
				pkg.Type = "formula"
			}
		case pkg.Version == "":
			if m := versionRe.FindStringSubmatch(line); m != nil {
				pkg.Version = m[1]
			}
		}

		// These can appear in any order after the header.
		if pkg.SHA256 == "" {
			if m := sha256Re.FindStringSubmatch(line); m != nil {
				pkg.SHA256 = m[1]
			}
		}
		if pkg.URL == "" {
			if m := urlRe.FindStringSubmatch(line); m != nil {
				pkg.URL = m[1]
			}
		}
		if pkg.DisplayName == "" {
			if m := nameRe.FindStringSubmatch(line); m != nil {
				pkg.DisplayName = m[1]
			}
		}
		if pkg.Description == "" {
			if m := descRe.FindStringSubmatch(line); m != nil {
				pkg.Description = m[1]
			}
		}
		if pkg.Homepage == "" {
			if m := homepageRe.FindStringSubmatch(line); m != nil {
				pkg.Homepage = m[1]
			}
		}
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("could not find cask/formula name in %s", filePath)
	}
	if pkg.DisplayName == "" {
		pkg.DisplayName = pkg.Name
	}

	// Extract GitHub source repo from the download URL, then homepage as fallback.
	for _, candidate := range []string{pkg.URL, pkg.Homepage} {
		if m := githubURLRe.FindStringSubmatch(candidate); m != nil {
			pkg.SourceOwner = m[1]
			repo := m[2]
			// Strip .git suffix just in case.
			repo = strings.TrimSuffix(repo, ".git")
			pkg.SourceRepo = repo
			break
		}
	}

	return pkg, nil
}

// NormalizeVersion strips a leading 'v' for comparison.
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}
