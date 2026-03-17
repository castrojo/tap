package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/castrojo/tap-tools/internal/stats"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

var rootCmd = &cobra.Command{
	Use:   "tap-stats",
	Short: "Generate statistics for a Homebrew tap",
	Long: `tap-stats collects package inventory and version-freshness data
for the tap and renders it as a terminal table, JSON, or an HTML page
suitable for GitHub Pages.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Collect and render tap statistics",
	Long: `Collect and render tap statistics.

Examples:
  # Pretty terminal table
  tap-stats generate

  # Write stats.json + index.html to docs/
  tap-stats generate --output docs/

  # Emit JSON to stdout
  tap-stats generate --format json

  # Emit HTML to stdout
  tap-stats generate --format html`,
	RunE: runGenerate,
}

var (
	flagTapDir  string
	flagOutput  string
	flagFormat  string
	flagNoFresh bool
)

func init() {
	generateCmd.Flags().StringVar(&flagTapDir, "tap-dir", ".", "Root of the tap repository")
	generateCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Directory to write stats.json and index.html")
	generateCmd.Flags().StringVar(&flagFormat, "format", "terminal", "Output format: terminal | json | html")
	generateCmd.Flags().BoolVar(&flagNoFresh, "no-freshness", false, "Skip upstream version freshness check")

	rootCmd.AddCommand(generateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("Error: "+err.Error()))
		os.Exit(1)
	}
}

func runGenerate(_ *cobra.Command, _ []string) error {
	tapDir, err := filepath.Abs(flagTapDir)
	if err != nil {
		return fmt.Errorf("resolving tap directory: %w", err)
	}

	fmt.Fprintln(os.Stderr, infoStyle.Render("→ Collecting tap statistics…"))
	tapStats, err := stats.Collect(tapDir, !flagNoFresh)
	if err != nil {
		return fmt.Errorf("collecting statistics: %w", err)
	}

	// If --output is set, write both JSON and HTML files.
	if flagOutput != "" {
		return writeOutputDir(tapStats, flagOutput)
	}

	// Otherwise render to stdout according to --format.
	switch flagFormat {
	case "json":
		data, err := stats.RenderJSON(tapStats)
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "html":
		data, err := stats.RenderHTML(tapStats)
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "terminal":
		return stats.RenderTerminal(tapStats, os.Stdout)

	default:
		return fmt.Errorf("unknown format %q — valid values: terminal, json, html", flagFormat)
	}

	return nil
}

func writeOutputDir(tapStats *stats.TapStats, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write stats.json.
	jsonPath := filepath.Join(outDir, "stats.json")
	jsonData, err := stats.RenderJSON(tapStats)
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", jsonPath, err)
	}
	fmt.Fprintf(os.Stderr, "✓ Wrote %s\n", jsonPath)

	// Write index.html.
	htmlPath := filepath.Join(outDir, "index.html")
	htmlData, err := stats.RenderHTML(tapStats)
	if err != nil {
		return err
	}
	if err := os.WriteFile(htmlPath, htmlData, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", htmlPath, err)
	}
	fmt.Fprintf(os.Stderr, "✓ Wrote %s\n", htmlPath)

	// Print terminal summary too.
	return stats.RenderTerminal(tapStats, os.Stderr)
}
