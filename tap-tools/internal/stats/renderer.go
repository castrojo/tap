package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
)

// RenderJSON returns a pretty-printed JSON representation of TapStats.
func RenderJSON(stats *TapStats) ([]byte, error) {
	return json.MarshalIndent(stats, "", "  ")
}

// RenderTerminal writes a coloured table to w.
func RenderTerminal(stats *TapStats, w io.Writer) error {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Fprintln(w, titleStyle.Render("🍺 "+stats.TapName+" — Statistics"))
	fmt.Fprintf(w, "  Generated: %s\n\n", mutedStyle.Render(stats.GeneratedAtStr()))

	// Summary block.
	fmt.Fprintln(w, headerStyle.Render("Summary"))
	fmt.Fprintf(w, "  Total packages : %d (casks: %d, formulas: %d)\n",
		stats.Summary.TotalPackages, stats.Summary.TotalCasks, stats.Summary.TotalFormulas)
	fmt.Fprintf(w, "  Up to date     : %s\n", successStyle.Render(fmt.Sprint(stats.Summary.CurrentCount)))
	fmt.Fprintf(w, "  Stale          : %s\n", warnStyle.Render(fmt.Sprint(stats.Summary.StaleCount)))
	fmt.Fprintf(w, "  Unknown        : %s\n\n", mutedStyle.Render(fmt.Sprint(stats.Summary.UnknownCount)))

	// Package table.
	fmt.Fprintln(w, headerStyle.Render("Packages"))

	colWidths := [5]int{30, 8, 12, 12, 45}
	headers := [5]string{"NAME", "TYPE", "VERSION", "LATEST", "DESCRIPTION"}

	// Header row.
	row := ""
	for i, h := range headers {
		row += fmt.Sprintf("%-*s  ", colWidths[i], h)
	}
	fmt.Fprintln(w, "  "+mutedStyle.Render(row))
	fmt.Fprintln(w, "  "+mutedStyle.Render(strings.Repeat("─", 115)))

	for _, p := range stats.Packages {
		latestStr := "—"
		if p.LatestVersion != "" {
			latestStr = p.LatestVersion
		}
		desc := p.Description
		if len(desc) > colWidths[4] {
			desc = desc[:colWidths[4]-1] + "…"
		}

		var statusColour func(...string) string
		switch p.StatusString() {
		case "current":
			statusColour = successStyle.Render
		case "stale":
			statusColour = warnStyle.Render
		default:
			statusColour = mutedStyle.Render
		}

		line := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s",
			colWidths[0], truncate(p.Name, colWidths[0]),
			colWidths[1], p.Type,
			colWidths[2], truncate(p.Version, colWidths[2]),
			colWidths[3], truncate(latestStr, colWidths[3]),
			colWidths[4], desc,
		)
		fmt.Fprintln(w, statusColour(p.StatusEmoji())+" "+line)
	}
	fmt.Fprintln(w)
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.TapName}} — Statistics</title>
<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
  background: #0d1117;
  color: #e6edf3;
  line-height: 1.5;
  min-height: 100vh;
}
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px 16px;
}
header {
  border-bottom: 1px solid #30363d;
  padding-bottom: 16px;
  margin-bottom: 24px;
}
h1 {
  font-size: 24px;
  font-weight: 600;
  color: #e6edf3;
}
h2 {
  font-weight: 600;
  margin: 28px 0 12px;
  text-transform: uppercase;
  letter-spacing: .5px;
  font-size: 12px;
  color: #7d8590;
}
.meta {
  color: #7d8590;
  font-size: 13px;
  margin-top: 4px;
}
.meta a {
  color: #58a6ff;
  text-decoration: none;
}
.meta a:hover {
  text-decoration: underline;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 32px;
}
.stat-card {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 6px;
  padding: 16px 20px;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #e6edf3;
  line-height: 1;
}
.stat-label {
  font-size: 12px;
  color: #7d8590;
  margin-top: 6px;
}
.stat-current .stat-value {
  color: #3fb950;
}
.stat-stale .stat-value {
  color: #d29922;
}
table {
  width: 100%;
  border-collapse: collapse;
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 6px;
  overflow: hidden;
  font-size: 14px;
}
th {
  background: #21262d;
  padding: 10px 16px;
  text-align: left;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .5px;
  color: #7d8590;
  border-bottom: 1px solid #30363d;
  white-space: nowrap;
}
td {
  padding: 10px 16px;
  border-bottom: 1px solid #21262d;
  vertical-align: middle;
}
tr:last-child td {
  border-bottom: none;
}
tr:hover td {
  background: #1c2128;
}
.pkg-name {
  font-weight: 500;
  color: #e6edf3;
}
.pkg-id {
  font-size: 11px;
  color: #7d8590;
  font-family: 'SF Mono', Consolas, monospace;
}
.badge {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}
.badge-current {
  background: #1f4429;
  color: #3fb950;
}
.badge-stale {
  background: #3d2b00;
  color: #d29922;
}
.badge-unknown {
  background: #21262d;
  color: #7d8590;
}
.badge-cask {
  background: #1b2d47;
  color: #58a6ff;
}
.badge-formula {
  background: #2d1b47;
  color: #a855f7;
}
.version {
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 12px;
  color: #e6edf3;
}
.version-stale {
  color: #d29922;
}
a {
  color: #58a6ff;
  text-decoration: none;
}
a:hover {
  text-decoration: underline;
}
.footer {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid #30363d;
  font-size: 12px;
  color: #7d8590;
}
.os-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}
@media (max-width: 900px) {
  .os-grid { grid-template-columns: 1fr; }
}
.os-col-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
}
.os-window-label {
  font-size: 14px;
  font-weight: 600;
  color: #e6edf3;
}
.os-dates {
  font-size: 11px;
  color: #7d8590;
}
.os-table {
  width: 100%;
  border-collapse: collapse;
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 6px;
  overflow: hidden;
  font-size: 13px;
}
.os-table th {
  background: #21262d;
  padding: 8px 10px;
  text-align: left;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .5px;
  color: #7d8590;
  border-bottom: 1px solid #30363d;
  white-space: nowrap;
}
.os-table td {
  padding: 7px 10px;
  border-bottom: 1px solid #21262d;
  vertical-align: middle;
}
.os-table tr:last-child td { border-bottom: none; }
.os-table tr:hover td { background: #1c2128; }
.os-rank {
  color: #7d8590;
  font-size: 11px;
  width: 20px;
  text-align: right;
  padding-right: 6px !important;
}
.os-name { color: #e6edf3; }
.os-pct {
  color: #7d8590;
  font-size: 11px;
  white-space: nowrap;
  text-align: right;
}
.os-bar-cell { width: 60px; padding: 0 8px !important; }
.os-bar-track {
  background: #21262d;
  border-radius: 2px;
  height: 6px;
  overflow: hidden;
}
.os-bar-fill {
  background: #388bfd;
  height: 6px;
  border-radius: 2px;
}
.source-note {
  font-size: 12px;
  color: #7d8590;
  margin-bottom: 16px;
}
.source-note a { color: #58a6ff; }
.dl-count {
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 13px;
  color: #e6edf3;
  white-space: nowrap;
  text-align: right;
}
</style>
</head>
<body>
<div class="container">
  <header>
    <h1>🍺 {{.TapName}}</h1>
    <div class="meta">
      Generated {{.GeneratedAtStr}} ·
      <a href="stats.json">📊 stats.json</a> ·
      <a href="{{.TapURL}}">GitHub</a>
    </div>
  </header>

  <h2>Overview</h2>
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-value">{{.Summary.TotalPackages}}</div>
      <div class="stat-label">Total Packages</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{{.Summary.TotalCasks}}</div>
      <div class="stat-label">Casks</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{{.Summary.TotalFormulas}}</div>
      <div class="stat-label">Formulas</div>
    </div>
    <div class="stat-card stat-current">
      <div class="stat-value">{{.Summary.CurrentCount}}</div>
      <div class="stat-label">Up to Date</div>
    </div>
    <div class="stat-card stat-stale">
      <div class="stat-value">{{.Summary.StaleCount}}</div>
      <div class="stat-label">Stale</div>
    </div>
    <div class="stat-card">
      <div class="stat-value">{{.Summary.UnknownCount}}</div>
      <div class="stat-label">Unknown</div>
    </div>
  </div>

  <h2>Packages</h2>
  <table>
    <thead>
      <tr>
        <th>Package</th>
        <th>Type</th>
        <th>Tap Version</th>
        <th>Latest</th>
        <th>Status</th>
        <th title="GitHub release asset downloads for the current version (all sources, not tap-specific)">Downloads ↓</th>
        <th>Description</th>
      </tr>
    </thead>
    <tbody>
      {{range .Packages}}
      <tr>
        <td>
          <div class="pkg-name"><a href="{{.Homepage}}">{{.DisplayName}}</a></div>
          <div class="pkg-id">{{.Name}}</div>
        </td>
        <td><span class="badge badge-{{.Type}}">{{.Type}}</span></td>
        <td><span class="version">{{.Version}}</span></td>
        <td><span class="version{{if .IsStale}} version-stale{{end}}">{{if .LatestVersion}}{{.LatestVersion}}{{else}}—{{end}}</span></td>
        <td><span class="badge badge-{{.StatusString}}">{{.StatusEmoji}} {{.StatusString}}</span></td>
        <td class="dl-count">{{if .DownloadCount}}{{fmtCount .DownloadCount}}{{else}}<span style="color:#7d8590">—</span>{{end}}</td>
        <td>{{.Description}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
  <p class="source-note" style="margin-top:8px">
    ↓ Downloads = GitHub release asset downloads for the current version (all sources, not tap-specific).
    Resets on each new release. Sorted highest → lowest.
  </p>

  {{if .OSStats}}
  <h2>OS Version Distribution</h2>
  <p class="source-note">
    Data from <a href="https://formulae.brew.sh/analytics/os-version/30d/">Homebrew Analytics</a> ·
    Top 10 per window · Bars normalised to #1 entry
  </p>
  <div class="os-grid">
    {{range .OSStats}}
    <div class="os-col">
      <div class="os-col-header">
        <span class="os-window-label">{{windowLabel .Window}}</span>
        <span class="os-dates">{{.StartDate}} – {{.EndDate}}</span>
      </div>
      <table class="os-table">
        <thead>
          <tr>
            <th class="os-rank">#</th>
            <th>OS</th>
            <th></th>
            <th class="os-pct">%</th>
          </tr>
        </thead>
        <tbody>
          {{range .Items}}
          <tr>
            <td class="os-rank">{{.Rank}}</td>
            <td class="os-name">{{.OSVersion}}</td>
            <td class="os-bar-cell">
              <div class="os-bar-track">
                <div class="os-bar-fill" style="width:{{.BarWidth}}%"></div>
              </div>
            </td>
            <td class="os-pct">{{.Percent}}%</td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{end}}
  </div>
  {{end}}

  <div class="footer">
    Generated by <a href="{{.TapURL}}/tree/main/tap-tools">tap-stats</a> ·
    <a href="stats.json">Download stats.json</a>
  </div>
</div>
</body>
</html>
`

// windowLabel converts "30d" → "Last 30 Days", "365d" → "Last 365 Days", etc.
func windowLabel(w string) string {
	switch w {
	case "30d":
		return "Last 30 Days"
	case "90d":
		return "Last 90 Days"
	case "365d":
		return "Last 365 Days"
	default:
		return w
	}
}

// fmtCount formats an int64 with comma separators (e.g. 51727 → "51,727").
func fmtCount(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	// Insert commas from right to left.
	var result []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}

// RenderHTML returns the stats page as HTML bytes.
func RenderHTML(stats *TapStats) ([]byte, error) {
	tmpl, err := template.New("stats").Funcs(template.FuncMap{
		"GeneratedAtStr": stats.GeneratedAtStr,
		"windowLabel":    windowLabel,
		"fmtCount":       fmtCount,
	}).Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, stats); err != nil {
		return nil, fmt.Errorf("rendering HTML: %w", err)
	}
	return buf.Bytes(), nil
}
