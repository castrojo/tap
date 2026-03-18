package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OSStats holds the top-N OS version leaderboard for one time window.
type OSStats struct {
	Window     string   `json:"window"`      // "30d", "90d", "365d"
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	TotalCount int64    `json:"total_count"`
	Items      []OSItem `json:"items"`
}

// OSItem is one ranked entry in an OS leaderboard.
type OSItem struct {
	Rank      int    `json:"rank"`
	OSVersion string `json:"os_version"`
	Count     string `json:"count"`
	Percent   string `json:"percent"`
	// BarWidth is 0–100, normalised so the #1 entry is always 100.
	// Used by the HTML template to draw proportional bars without JS.
	BarWidth int `json:"bar_width"`
}

// brewOSResponse is the raw shape returned by formulae.brew.sh.
type brewOSResponse struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	TotalCount int64  `json:"total_count"`
	Items      []struct {
		Number    int    `json:"number"`
		OSVersion string `json:"os_version"`
		Count     string `json:"count"`
		Percent   string `json:"percent"`
	} `json:"items"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

// FetchOSStats fetches the top topN OS versions for each Homebrew analytics
// window (30d, 90d, 365d) from formulae.brew.sh.
func FetchOSStats(topN int) ([]OSStats, error) {
	windows := []string{"30d", "90d", "365d"}
	var result []OSStats

	for _, w := range windows {
		url := fmt.Sprintf("https://formulae.brew.sh/api/analytics/os-version/%s.json", w)
		resp, err := httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", url, err)
		}
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
		}

		var data brewOSResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("decoding %s: %w", url, err)
		}

		limit := topN
		if limit > len(data.Items) {
			limit = len(data.Items)
		}

		os := OSStats{
			Window:     w,
			StartDate:  data.StartDate,
			EndDate:    data.EndDate,
			TotalCount: data.TotalCount,
		}

		// Find the top-ranked percent to normalise bar widths.
		maxPct := 0.0
		if limit > 0 {
			maxPct = parsePercent(data.Items[0].Percent)
		}

		for i := 0; i < limit; i++ {
			raw := data.Items[i]
			pct := parsePercent(raw.Percent)
			barWidth := 0
			if maxPct > 0 {
				barWidth = int(pct / maxPct * 100)
			}
			os.Items = append(os.Items, OSItem{
				Rank:      raw.Number,
				OSVersion: raw.OSVersion,
				Count:     raw.Count,
				Percent:   raw.Percent,
				BarWidth:  barWidth,
			})
		}

		result = append(result, os)
	}

	return result, nil
}

// parsePercent converts a percent string like "34.65" to float64.
func parsePercent(s string) float64 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
