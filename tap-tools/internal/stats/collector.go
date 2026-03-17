package stats

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/castrojo/tap-tools/internal/github"
)

// Collect gathers statistics for the tap rooted at tapDir.
// It scans Casks/ and Formula/ for .rb files, parses each, and optionally
// checks version freshness against upstream GitHub releases.
func Collect(tapDir string, checkFreshness bool) (*TapStats, error) {
	stats := &TapStats{
		GeneratedAt: time.Now().UTC(),
		TapName:     "castrojo/homebrew-tap",
		TapURL:      "https://github.com/castrojo/homebrew-tap",
	}

	// Discover packages.
	packages, err := scanDirectory(tapDir)
	if err != nil {
		return nil, fmt.Errorf("scanning tap directory: %w", err)
	}

	// Enrich with git last-modified dates.
	for i := range packages {
		packages[i].LastUpdated = gitLastModified(tapDir, packages[i].FilePath)
	}

	// Check version freshness via GitHub API.
	if checkFreshness {
		client, err := github.NewClientWithTokenCheck()
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Skipping freshness check: %v\n", err)
		} else {
			for i := range packages {
				if packages[i].SourceOwner == "" || packages[i].SourceRepo == "" {
					continue
				}
				latest, err := client.GetLatestRelease(packages[i].SourceOwner, packages[i].SourceRepo)
				if err != nil {
					continue
				}
				packages[i].LatestVersion = NormalizeVersion(latest.TagName)
				packages[i].FreshnessKnown = true
				packages[i].IsStale = packages[i].LatestVersion != NormalizeVersion(packages[i].Version)
			}
		}
	}

	stats.Packages = packages
	stats.Summary = computeSummary(packages)
	return stats, nil
}

// scanDirectory walks Casks/ and Formula/ directories.
func scanDirectory(tapDir string) ([]Package, error) {
	var packages []Package

	dirs := []struct {
		dir      string
		pkgType  string
	}{
		{filepath.Join(tapDir, "Casks"), "cask"},
		{filepath.Join(tapDir, "Formula"), "formula"},
	}

	for _, d := range dirs {
		entries, err := os.ReadDir(d.dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", d.dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rb") {
				continue
			}
			// Skip placeholder files.
			if entry.Name() == ".gitkeep" {
				continue
			}

			filePath := filepath.Join(d.dir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Could not read %s: %v\n", filePath, err)
				continue
			}

			pkg, err := ParseRubyFile(string(content), filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Could not parse %s: %v\n", filePath, err)
				continue
			}

			// Use detected type but trust directory as ground truth.
			if pkg.Type == "" {
				pkg.Type = d.pkgType
			}

			packages = append(packages, *pkg)
		}
	}

	return packages, nil
}

// gitLastModified returns the date of the last git commit that touched filePath.
func gitLastModified(tapDir, filePath string) string {
	cmd := exec.Command("git", "-C", tapDir, "log", "-1", "--format=%as", "--", filePath)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// computeSummary aggregates counts from a slice of packages.
func computeSummary(packages []Package) Summary {
	s := Summary{}
	for _, p := range packages {
		s.TotalPackages++
		switch p.Type {
		case "cask":
			s.TotalCasks++
		case "formula":
			s.TotalFormulas++
		}
		if p.FreshnessKnown {
			if p.IsStale {
				s.StaleCount++
			} else {
				s.CurrentCount++
			}
		} else {
			s.UnknownCount++
		}
	}
	return s
}
