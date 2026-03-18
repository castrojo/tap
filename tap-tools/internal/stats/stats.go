package stats

import "time"

// TapStats represents statistics about the entire tap.
type TapStats struct {
	GeneratedAt time.Time  `json:"generated_at"`
	TapName     string     `json:"tap_name"`
	TapURL      string     `json:"tap_url"`
	Summary     Summary    `json:"summary"`
	Packages    []Package  `json:"packages"`
	OSStats     []OSStats  `json:"os_stats,omitempty"`
}

// GeneratedAtStr returns a human-readable timestamp for templates.
func (t *TapStats) GeneratedAtStr() string {
	return t.GeneratedAt.UTC().Format("2006-01-02 15:04 UTC")
}

// Summary contains aggregate statistics.
type Summary struct {
	TotalPackages int `json:"total_packages"`
	TotalCasks    int `json:"total_casks"`
	TotalFormulas int `json:"total_formulas"`
	CurrentCount  int `json:"current_count"`
	StaleCount    int `json:"stale_count"`
	UnknownCount  int `json:"unknown_count"`
}

// Package represents a single cask or formula entry.
type Package struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Type        string `json:"type"` // "cask" or "formula"
	Version     string `json:"version"`
	SHA256      string `json:"sha256,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	FilePath    string `json:"file_path,omitempty"`

	// Source repository (populated when URL is a GitHub releases URL).
	SourceOwner string `json:"source_owner,omitempty"`
	SourceRepo  string `json:"source_repo,omitempty"`

	// Freshness check results.
	LatestVersion  string `json:"latest_version,omitempty"`
	IsStale        bool   `json:"is_stale"`
	FreshnessKnown bool   `json:"freshness_known"`

	// DownloadCount is the GitHub release asset download count for the asset
	// that matches this package's URL. It serves as an install-count proxy.
	// Zero means no data is available (non-GitHub packages, or not yet fetched).
	DownloadCount int64 `json:"download_count"`

	LastUpdated string `json:"last_updated,omitempty"`
}

// StatusString returns a lowercase status label.
func (p *Package) StatusString() string {
	if !p.FreshnessKnown {
		return "unknown"
	}
	if p.IsStale {
		return "stale"
	}
	return "current"
}

// StatusEmoji returns a single emoji representing status.
func (p *Package) StatusEmoji() string {
	if !p.FreshnessKnown {
		return "❓"
	}
	if p.IsStale {
		return "⚠️"
	}
	return "✅"
}
