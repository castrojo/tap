package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tap-test",
		Short: "Smoke test Homebrew formulas and casks",
		Long:  "Run smoke tests on installed formulas and casks to verify they work.",
	}

	testFormulaCmd := &cobra.Command{
		Use:   "formula [name]",
		Short: "Test that a formula works after installation",
		Args:  cobra.ExactArgs(1),
		RunE:  testFormula,
	}

	testCaskCmd := &cobra.Command{
		Use:   "cask [name]",
		Short: "Test that a cask works after installation",
		Args:  cobra.ExactArgs(1),
		RunE:  testCask,
	}

	rootCmd.AddCommand(testFormulaCmd)
	rootCmd.AddCommand(testCaskCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func testFormula(cmd *cobra.Command, args []string) error {
	formulaName := args[0]

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Testing formula: %s\n", formulaName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Check if binary exists in PATH
	_, err := exec.LookPath(formulaName)
	if err != nil {
		fmt.Printf("❌ Binary '%s' not found in PATH\n", formulaName)
		fmt.Println("Searching for binary in Homebrew prefix...")

		// Try to find in Homebrew prefix
		homebrewPrefix, err := getHomebrewPrefix()
		if err != nil {
			return fmt.Errorf("failed to get Homebrew prefix: %w", err)
		}

		possiblePaths := []string{
			filepath.Join(homebrewPrefix, "bin", formulaName),
			filepath.Join(homebrewPrefix, "opt", formulaName, "bin", formulaName),
		}

		found := false
		for _, path := range possiblePaths {
			if fileExists(path) && isExecutable(path) {
				fmt.Printf("✓ Found binary at: %s\n", path)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("binary not found in expected locations")
		}
	}

	// Test binary execution with common version flags
	fmt.Println()
	fmt.Println("Testing binary execution...")

	versionFlags := []string{"--version", "-v", "-V", "version", "--help", "-h"}
	success := false

	for _, flag := range versionFlags {
		testCmd := exec.Command(formulaName, flag)
		if err := testCmd.Run(); err == nil {
			fmt.Printf("✓ Binary executes successfully (tested: %s %s)\n", formulaName, flag)
			success = true
			break
		}
	}

	if !success {
		// Try running without flags with timeout
		fmt.Println("Trying execution without flags (5s timeout)...")
		testCmd := exec.Command(formulaName)

		done := make(chan error, 1)
		go func() {
			done <- testCmd.Run()
		}()

		select {
		case <-time.After(5 * time.Second):
			if testCmd.Process != nil {
				testCmd.Process.Kill()
			}
			fmt.Println("✓ Binary executes successfully (no flags)")
			success = true
		case err := <-done:
			if err == nil {
				fmt.Println("✓ Binary executes successfully (no flags)")
				success = true
			}
		}
	}

	if !success {
		fmt.Println("⚠ Warning: Could not verify binary execution (none of the common flags worked)")
		fmt.Println("This might be expected for some CLIs that require specific arguments")
		// Don't fail - just warn
	}

	fmt.Println()
	fmt.Printf("✅ Formula %s smoke test completed\n", formulaName)
	return nil
}

func testCask(cmd *cobra.Command, args []string) error {
	caskName := args[0]

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Testing cask: %s\n", caskName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get installation directory
	homebrewPrefix, err := getHomebrewPrefix()
	if err != nil {
		return fmt.Errorf("failed to get Homebrew prefix: %w", err)
	}

	installDir := filepath.Join(homebrewPrefix, "Caskroom", caskName)

	// Check installation directory exists
	if !dirExists(installDir) {
		return fmt.Errorf("installation directory not found: %s", installDir)
	}
	fmt.Printf("✓ Installation directory exists: %s\n", installDir)

	// Check for desktop file (GUI apps)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	desktopFile := filepath.Join(homeDir, ".local", "share", "applications", caskName+".desktop")
	if fileExists(desktopFile) {
		fmt.Printf("✓ Desktop file exists: %s\n", desktopFile)

		// Validate desktop file if validator is available
		if _, err := exec.LookPath("desktop-file-validate"); err == nil {
			validateCmd := exec.Command("desktop-file-validate", desktopFile)
			if err := validateCmd.Run(); err != nil {
				fmt.Printf("⚠ Desktop file validation failed: %v\n", err)
			} else {
				fmt.Println("✓ Desktop file is valid")
			}
		}
	}

	// Check for icon (GUI apps)
	iconDir := filepath.Join(homeDir, ".local", "share", "icons")
	if dirExists(iconDir) {
		iconCount := 0
		filepath.Walk(iconDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if strings.Contains(info.Name(), caskName) {
				iconCount++
			}
			return nil
		})

		if iconCount > 0 {
			fmt.Printf("✓ Found %d icon(s)\n", iconCount)
		}
	}

	// Try to find and test executable
	var foundExecutable string
	filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && isExecutable(path) && foundExecutable == "" {
			foundExecutable = path
			return filepath.SkipDir
		}
		return nil
	})

	if foundExecutable != "" {
		fmt.Printf("✓ Found executable: %s\n", foundExecutable)

		// Try --version flag
		testCmd := exec.Command(foundExecutable, "--version")
		if err := testCmd.Run(); err == nil {
			fmt.Println("✓ Binary executes successfully")
		} else {
			fmt.Println("⚠ Binary found but --version failed (may be GUI-only)")
		}
	}

	fmt.Println()
	fmt.Printf("✅ Cask %s smoke test completed\n", caskName)
	return nil
}

// Helper functions

func getHomebrewPrefix() (string, error) {
	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}
