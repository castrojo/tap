package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/castrojo/tap-tools/internal/archive"
	"github.com/castrojo/tap-tools/internal/checksum"
	"github.com/castrojo/tap-tools/internal/desktop"
	"github.com/castrojo/tap-tools/internal/github"
	"github.com/castrojo/tap-tools/internal/homebrew"
	"github.com/castrojo/tap-tools/internal/platform"
	"github.com/castrojo/tap-tools/internal/validate"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Styles for pretty output
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

var rootCmd = &cobra.Command{
	Use:   "tap-cask",
	Short: "Generate Homebrew casks for Linux",
	Long: `tap-cask generates Homebrew casks for Linux applications.

It fetches release information from GitHub, downloads assets,
verifies checksums, and generates properly formatted cask files.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate [repo-url]",
	Short: "Generate a new cask from GitHub repository",
	Long: `Generate a new cask from a GitHub repository.

Examples:
  tap-cask generate https://github.com/sublimehq/sublime_text
  tap-cask generate sublimehq/sublime_text
  tap-cask generate https://github.com/user/repo --name my-app`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

var (
	flagName   string
	flagOutput string
)

func init() {
	generateCmd.Flags().StringVar(&flagName, "name", "", "Override package name (will auto-append -linux)")
	generateCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output file path (default: Casks/<name>-linux.rb)")

	rootCmd.AddCommand(generateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("Error: "+err.Error()))
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	repoURL := args[0]

	// Parse repository URL
	fmt.Println(titleStyle.Render("🔍 Parsing repository URL..."))
	owner, repo, err := github.ParseRepoURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Repository: %s/%s", owner, repo)))

	// Create GitHub client
	client := github.NewClient()

	// Fetch repository metadata
	fmt.Println(titleStyle.Render("\n🔍 Fetching repository metadata..."))
	repository, err := client.GetRepository(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Found: %s", repository.Description)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("  Homepage: %s", repository.Homepage)))

	// Get latest release
	fmt.Println(titleStyle.Render("\n🔍 Finding latest release..."))
	release, err := client.GetLatestRelease(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Version: %s", release.TagName)))

	// Detect platform for all assets
	fmt.Println(titleStyle.Render("\n🔍 Analyzing release assets..."))
	var assets []*platform.Asset
	for _, ghAsset := range release.Assets {
		asset := platform.DetectPlatform(ghAsset.Name)
		asset.URL = ghAsset.URL
		asset.DownloadURL = ghAsset.BrowserDownloadURL
		asset.Size = ghAsset.Size
		assets = append(assets, asset)
	}

	// Filter Linux assets
	linuxAssets := platform.FilterLinuxAssets(assets)
	if len(linuxAssets) == 0 {
		return fmt.Errorf("no Linux assets found in release")
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Found %d Linux asset(s)", len(linuxAssets))))

	// Select best asset
	bestAsset, err := platform.SelectBestAsset(linuxAssets)
	if err != nil {
		return fmt.Errorf("failed to select asset: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Selected: %s (Priority %d)", bestAsset.Name, bestAsset.Priority)))

	// Download and calculate checksum
	fmt.Println(titleStyle.Render("\n⬇️  Downloading asset..."))
	data, err := checksum.DownloadFile(bestAsset.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Downloaded %.2f MB", float64(len(data))/1024/1024)))

	// Calculate SHA256
	fmt.Println(titleStyle.Render("\n🔐 Calculating SHA256..."))
	sha256sum := checksum.CalculateSHA256(data)
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ SHA256: %s", sha256sum)))

	// Try to verify with upstream checksums
	fmt.Println(titleStyle.Render("\n🔍 Searching for upstream checksums..."))
	upstreamChecksums, err := checksum.FindUpstreamChecksum(bestAsset.DownloadURL)
	if err != nil {
		fmt.Println(infoStyle.Render("✗ No upstream checksums found (not an error)"))
	} else {
		if expected, found := upstreamChecksums[bestAsset.Name]; found {
			if expected == sha256sum {
				fmt.Println(successStyle.Render("✓ Checksum verified against upstream!"))
			} else {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, sha256sum)
			}
		} else {
			fmt.Println(infoStyle.Render("✗ File not in upstream checksums (not an error)"))
		}
	}

	// Extract archive and inspect contents
	fmt.Println(titleStyle.Render("\n📦 Inspecting archive contents..."))
	files, err := archive.ListFiles(data, bestAsset.Name)
	if err != nil {
		fmt.Println(infoStyle.Render(fmt.Sprintf("✗ Could not list archive contents: %v", err)))
		fmt.Println(infoStyle.Render("  Will use default paths"))
		files = []string{} // Empty list to fall back to defaults
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Found %d files in archive", len(files))))
	}

	// Detect binaries
	var detectedBinaries []string
	if len(files) > 0 {
		detectedBinaries = archive.DetectBinaries(files)
		if len(detectedBinaries) > 0 {
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Detected %d binary file(s)", len(detectedBinaries))))
			for _, bin := range detectedBinaries {
				fmt.Println(infoStyle.Render(fmt.Sprintf("  - %s", bin)))
			}
		} else {
			fmt.Println(infoStyle.Render("✗ No binary files detected"))
		}
	}

	// Detect desktop integration
	fmt.Println(titleStyle.Render("\n🖼️  Detecting desktop integration..."))
	var desktopFile *desktop.DesktopFileInfo
	var icon *desktop.IconInfo

	if len(files) > 0 {
		desktopFile, _ = desktop.DetectDesktopFile(files)
		icon, _ = desktop.DetectIcon(files)

		if desktopFile != nil {
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Found desktop file: %s", desktopFile.Path)))
		} else {
			fmt.Println(infoStyle.Render("✗ No desktop file found"))
		}

		if icon != nil {
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Found icon: %s (size: %s)", icon.Path, icon.Size)))
		} else {
			fmt.Println(infoStyle.Render("✗ No icon found"))
		}
	}

	// Determine package name
	pkgName := flagName
	if pkgName == "" {
		pkgName = platform.NormalizePackageName(repo)
	}
	token := platform.EnsureLinuxSuffix(pkgName)

	// Create cask data
	caskData := homebrew.NewCaskData(token, release.TagName, sha256sum, bestAsset.DownloadURL)
	caskData.AppName = repo
	caskData.Description = repository.Description
	caskData.Homepage = repository.Homepage

	// Set binary path from detection
	if len(detectedBinaries) > 0 {
		// Select the best binary based on package name
		bestBinary := archive.SelectBestBinary(detectedBinaries, pkgName)
		caskData.BinaryPath = bestBinary

		// Extract just the binary name (without path)
		binaryName := filepath.Base(bestBinary)

		// Prefer package name if binary name matches roughly
		if strings.Contains(strings.ToLower(binaryName), strings.ToLower(pkgName)) ||
			strings.Contains(strings.ToLower(pkgName), strings.ToLower(binaryName)) {
			caskData.BinaryName = pkgName
		} else {
			caskData.BinaryName = binaryName
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("  Binary: %s → %s", caskData.BinaryPath, caskData.BinaryName)))
	} else {
		// Fallback to guessing
		rootDir := archive.FindRootDirectory(files)
		if rootDir != "" {
			caskData.BinaryPath = fmt.Sprintf("%s%s", rootDir, pkgName)
		} else {
			caskData.BinaryPath = pkgName
		}
		caskData.BinaryName = pkgName
		fmt.Println(infoStyle.Render(fmt.Sprintf("  Binary (guessed): %s → %s", caskData.BinaryPath, caskData.BinaryName)))
	}

	// Set desktop file if found
	if desktopFile != nil {
		caskData.SetDesktopFile(desktopFile.Path, desktopFile.Filename)
	}

	// Set icon if found
	if icon != nil {
		caskData.SetIcon(icon.Path, icon.Filename)
	}

	// Infer zap trash paths
	caskData.InferZapTrash()

	// Generate cask
	fmt.Println(titleStyle.Render("\n📝 Generating cask..."))
	caskContent, err := homebrew.GenerateCask(caskData)
	if err != nil {
		return fmt.Errorf("failed to generate cask: %w", err)
	}

	// Determine output path
	outputPath := flagOutput
	if outputPath == "" {
		outputPath = filepath.Join("Casks", token+".rb")
	}

	// Write cask file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(caskContent), 0644); err != nil {
		return fmt.Errorf("failed to write cask file: %w", err)
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Created: %s", outputPath)))

	// Validate the generated cask
	fmt.Println(titleStyle.Render("\n🔍 Validating generated cask..."))
	result, err := validate.ValidateFile(outputPath, true, true)
	if err != nil {
		fmt.Println(errorStyle.Render("✗ Validation failed:"))
		for _, errMsg := range result.Errors {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  - %s", errMsg)))
		}
		return fmt.Errorf("generated cask failed validation")
	}

	if result.Fixed {
		fmt.Println(successStyle.Render("✓ Validation passed (style issues auto-fixed)"))
	} else {
		fmt.Println(successStyle.Render("✓ Validation passed"))
	}

	// Print next steps
	fmt.Println(titleStyle.Render("\n✅ Done! Next steps:"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   1. Review %s", outputPath)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   2. Test: brew install --cask castrojo/tap/%s", token)))
	fmt.Println(infoStyle.Render("   3. Commit and push"))

	return nil
}
