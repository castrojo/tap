package stats

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/castrojo/tap-tools/internal/github"
	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

// getTapInfoFromGit reads the git remote URL and returns (tapName, tapURL, error).
// tapName is "owner/repo" and tapURL is "https://github.com/owner/repo".
func getTapInfoFromGit(tapDir string) (string, string, error) {
	cmd := exec.Command("git", "-C", tapDir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("reading git remote: %w", err)
	}
	raw := strings.TrimSpace(string(out))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(raw, "git@github.com:") {
		path := strings.TrimPrefix(raw, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			tapName := parts[0] + "/" + parts[1]
			return tapName, "https://github.com/" + tapName, nil
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo[.git]
	if strings.Contains(raw, "github.com/") {
		idx := strings.Index(raw, "github.com/")
		path := raw[idx+len("github.com/"):]
		path = strings.TrimSuffix(path, ".git")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			tapName := parts[0] + "/" + parts[1]
			return tapName, "https://github.com/" + tapName, nil
		}
	}

	return "", "", fmt.Errorf("could not parse GitHub remote from %q", raw)
}

// Collect gathers statistics for the tap rooted at tapDir.
// It scans Casks/ and Formula/ for .rb files, parses each, and optionally
// checks version freshness against upstream GitHub releases and fetches
// Homebrew OS analytics.
func Collect(tapDir string, checkFreshness bool, fetchOSStats bool) (*TapStats, error) {
	tapName, tapURL, err := getTapInfoFromGit(tapDir)
	if err != nil {
		// Fall back to a placeholder so the rest of the output is still useful.
		fmt.Fprintf(os.Stderr, "⚠️  Could not detect tap info from git remote: %v\n", err)
		tapName = "unknown/tap"
		tapURL = ""
	}

	stats := &TapStats{
		GeneratedAt: time.Now().UTC(),
		TapName:     tapName,
		TapURL:      tapURL,
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

	// Check version freshness and fetch tap traffic via GitHub API.
	if checkFreshness {
		client, err := github.NewClientWithTokenCheck()
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Skipping freshness check: %v\n", err)
		} else {
			// Per-package freshness checks.
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

			// Tap-level traffic: how many unique IPs cloned/tapped this repo.
			parts := strings.SplitN(tapName, "/", 2)
			if len(parts) == 2 {
				fmt.Fprintln(os.Stderr, infoStyle.Render("→ Fetching tap traffic…"))
				count, uniques, err := client.GetTrafficClones(parts[0], parts[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Could not fetch tap traffic: %v\n", err)
				} else {
					stats.Traffic = &TapTraffic{
						Count:   count,
						Uniques: uniques,
						Window:  "14 days",
					}
				}
			}
		}
	}

	// Sort packages alphabetically by name.
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	stats.Packages = packages
	stats.Summary = computeSummary(packages)

	// Fetch Homebrew OS version analytics (30d / 90d / 365d).
	if fetchOSStats {
		fmt.Fprintln(os.Stderr, infoStyle.Render("→ Fetching Homebrew OS analytics…"))
		osStats, err := FetchOSStats(10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Could not fetch OS stats from formulae.brew.sh: %v\n", err)
		} else {
			stats.OSStats = osStats
		}
	}

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

			// Use a path relative to the tap root so the JSON output doesn't
			// expose the user's local directory structure.
			relPath, err := filepath.Rel(tapDir, filePath)
			if err != nil {
				relPath = filePath // shouldn't happen, but safe fallback
			}

			pkg, err := ParseRubyFile(string(content), relPath)
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



