package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/castrojo/tap-tools/internal/buildsystem"
	"github.com/castrojo/tap-tools/internal/checksum"
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
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

var rootCmd = &cobra.Command{
	Use:   "tap-formula",
	Short: "Generate Homebrew formulas for Linux",
	Long: `tap-formula generates Homebrew formulas for Linux CLI tools and libraries.

It fetches release information from GitHub, detects the build system,
downloads assets, verifies checksums, and generates properly formatted formula files.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate [repo-url]",
	Short: "Generate a new formula from GitHub repository",
	Long: `Generate a new formula from a GitHub repository.

The tool automatically detects the build system (Go, Rust, CMake, etc.)
and generates appropriate installation instructions.

Examples:
  tap-formula generate https://github.com/BurntSushi/ripgrep
  tap-formula generate BurntSushi/ripgrep
  tap-formula generate https://github.com/user/repo --name my-tool`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

var (
	flagName       string
	flagOutput     string
	flagBinary     string
	flagFromSource bool
)

func init() {
	generateCmd.Flags().StringVar(&flagName, "name", "", "Override package name")
	generateCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output file path (default: Formula/<name>.rb)")
	generateCmd.Flags().StringVar(&flagBinary, "binary", "", "Binary name (defaults to package name)")
	generateCmd.Flags().BoolVar(&flagFromSource, "from-source", false, "Generate formula for building from source (use source tarball)")

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

	// Determine package name
	packageName := flagName
	if packageName == "" {
		packageName = platform.NormalizePackageName(repo)
	}
	fmt.Println(infoStyle.Render(fmt.Sprintf("  Package: %s", packageName)))

	// Determine binary name
	binaryName := flagBinary
	if binaryName == "" {
		binaryName = packageName
	}

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
	fmt.Println(infoStyle.Render(fmt.Sprintf("  License: %s", repository.License)))

	// Get latest release
	fmt.Println(titleStyle.Render("\n🔍 Finding latest release..."))
	release, err := client.GetLatestRelease(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	version := release.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:] // Remove 'v' prefix
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Version: %s", version)))

	// Select asset
	fmt.Println(titleStyle.Render("\n🔍 Analyzing release assets..."))

	var selectedAsset *platform.Asset
	var downloadURL string

	if flagFromSource {
		// Use source tarball
		downloadURL = fmt.Sprintf("https://github.com/%s/%s/archive/v%s.tar.gz", owner, repo, version)
		fmt.Println(infoStyle.Render("  Using source tarball (--from-source)"))
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ URL: %s", downloadURL)))
	} else {
		// Try to find pre-built Linux binary
		var assets []*platform.Asset
		for _, ghAsset := range release.Assets {
			asset := platform.DetectPlatform(ghAsset.Name)
			if asset != nil {
				asset.URL = ghAsset.URL
				asset.DownloadURL = ghAsset.BrowserDownloadURL
				asset.Size = ghAsset.Size
				assets = append(assets, asset)
			}
		}

		// Filter Linux assets only
		linuxAssets := platform.FilterLinuxAssets(assets)

		if len(linuxAssets) == 0 {
			fmt.Println(warnStyle.Render("⚠ No Linux binaries found in releases"))
			fmt.Println(infoStyle.Render("  Falling back to source tarball"))
			downloadURL = fmt.Sprintf("https://github.com/%s/%s/archive/v%s.tar.gz", owner, repo, version)
			flagFromSource = true
		} else {
			fmt.Println(infoStyle.Render(fmt.Sprintf("  Found %d Linux asset(s)", len(linuxAssets))))

			// Select best asset
			var err error
			selectedAsset, err = platform.SelectBestAsset(linuxAssets)
			if err != nil {
				return fmt.Errorf("failed to select asset: %w", err)
			}

			downloadURL = selectedAsset.DownloadURL
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Selected: %s (%s - Priority %d)",
				selectedAsset.Name, selectedAsset.Format, selectedAsset.Priority)))
		}
	}

	// Download and calculate checksum
	fmt.Println(titleStyle.Render("\n⬇️  Downloading asset..."))
	data, err := checksum.DownloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Downloaded %.1f MB", float64(len(data))/(1024*1024))))

	// Calculate SHA256
	fmt.Println(titleStyle.Render("\n🔐 Calculating SHA256..."))
	sha256 := checksum.CalculateSHA256(data)
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ SHA256: %s", sha256)))

	// Generate formula based on whether we're building from source
	fmt.Println(titleStyle.Render("\n📝 Generating formula..."))

	var formula string

	if flagFromSource {
		// Fetch repository files to detect build system
		fmt.Println(infoStyle.Render("  Detecting build system from repository..."))

		// Get repository tree to detect build system
		repoFiles, err := client.GetRepoFiles(owner, repo)
		if err != nil {
			fmt.Println(warnStyle.Render(fmt.Sprintf("  ⚠ Could not fetch repository files: %v", err)))
			fmt.Println(infoStyle.Render("  Generating simple formula template"))

			// Fallback to simple formula
			formulaData := homebrew.NewFormulaDataSimple(
				packageName,
				version,
				sha256,
				downloadURL,
				repository.Description,
				repository.Homepage,
				repository.License,
				binaryName,
			)

			formula, err = homebrew.GenerateFormula(formulaData)
			if err != nil {
				return fmt.Errorf("failed to generate formula: %w", err)
			}
		} else {
			// Detect build system
			buildSys := buildsystem.Detect(repoFiles)
			if buildSys == nil {
				fmt.Println(warnStyle.Render("  ⚠ Could not detect build system"))
				fmt.Println(infoStyle.Render("  Generating simple formula template"))

				formulaData := homebrew.NewFormulaDataSimple(
					packageName,
					version,
					sha256,
					downloadURL,
					repository.Description,
					repository.Homepage,
					repository.License,
					binaryName,
				)

				formula, err = homebrew.GenerateFormula(formulaData)
				if err != nil {
					return fmt.Errorf("failed to generate formula: %w", err)
				}
			} else {
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ Detected build system: %s", buildSys.Name())))

				formulaData, err := homebrew.NewFormulaData(
					packageName,
					version,
					sha256,
					downloadURL,
					repository.Description,
					repository.Homepage,
					repository.License,
					repoFiles,
					binaryName,
				)
				if err != nil {
					return fmt.Errorf("failed to create formula data: %w", err)
				}

				formula, err = homebrew.GenerateFormula(formulaData)
				if err != nil {
					return fmt.Errorf("failed to generate formula: %w", err)
				}
			}
		}
	} else {
		// Pre-built binary - simple install
		formulaData := homebrew.NewFormulaDataSimple(
			packageName,
			version,
			sha256,
			downloadURL,
			repository.Description,
			repository.Homepage,
			repository.License,
			binaryName,
		)

		formula, err = homebrew.GenerateFormula(formulaData)
		if err != nil {
			return fmt.Errorf("failed to generate formula: %w", err)
		}
	}

	// Determine output path
	outputPath := flagOutput
	if outputPath == "" {
		// Default to Formula/<name>.rb in current directory
		outputPath = filepath.Join("Formula", packageName+".rb")
	}

	// Ensure Formula directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write formula
	if err := os.WriteFile(outputPath, []byte(formula), 0644); err != nil {
		return fmt.Errorf("failed to write formula: %w", err)
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Created: %s", outputPath)))

	// Validate the generated formula
	fmt.Println(titleStyle.Render("\n🔍 Validating generated formula..."))
	result, err := validate.ValidateFile(outputPath, false, true)
	if err != nil {
		fmt.Println(errorStyle.Render("✗ Validation failed:"))
		if result != nil {
			for _, errMsg := range result.Errors {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  - %s", errMsg)))
			}
		}
		return fmt.Errorf("generated formula failed validation")
	}

	if result.Fixed {
		fmt.Println(successStyle.Render("✓ Validation passed (style issues auto-fixed)"))
	} else {
		fmt.Println(successStyle.Render("✓ Validation passed"))
	}

	// Print next steps
	fmt.Println(titleStyle.Render("\n✅ Done! Next steps:"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   1. Review %s", outputPath)))
	if flagFromSource {
		fmt.Println(infoStyle.Render("   2. Test: HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source " + packageName))
	} else {
		fmt.Println(infoStyle.Render("   2. Verify binary paths and adjust if needed"))
		fmt.Println(infoStyle.Render("   3. Test: brew install " + packageName))
	}
	fmt.Println(infoStyle.Render("   4. Commit and push"))

	return nil
}
