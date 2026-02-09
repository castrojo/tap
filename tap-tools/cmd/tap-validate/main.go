package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/castrojo/tap-tools/internal/validate"
	"github.com/spf13/cobra"
)

var (
	fixStyle bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tap-validate",
		Short: "Validate Homebrew formulas and casks",
		Long:  "Run brew audit and brew style on formulas and casks in this tap.",
	}

	validateAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Validate all formulas and casks",
		RunE:  validateAll,
	}

	validateFileCmd := &cobra.Command{
		Use:   "file [path]",
		Short: "Validate a specific formula or cask file",
		Args:  cobra.ExactArgs(1),
		RunE:  validateFileCmd,
	}

	validateAllCmd.Flags().BoolVar(&fixStyle, "fix", false, "Automatically fix style issues")
	validateFileCmd.Flags().BoolVar(&fixStyle, "fix", false, "Automatically fix style issues")

	rootCmd.AddCommand(validateAllCmd)
	rootCmd.AddCommand(validateFileCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func validateAll(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	var failed int

	// Validate formulas
	formulaDir := filepath.Join(repoRoot, "Formula")
	if _, err := os.Stat(formulaDir); err == nil {
		formulas, err := filepath.Glob(filepath.Join(formulaDir, "*.rb"))
		if err != nil {
			return fmt.Errorf("failed to find formulas: %w", err)
		}

		if len(formulas) > 0 {
			fmt.Println("→ Validating formulas...")
			for _, formula := range formulas {
				name := strings.TrimSuffix(filepath.Base(formula), ".rb")
				fmt.Printf("  Checking %s...\n", name)

				result, err := validate.ValidateFile(formula, false, fixStyle)
				if err != nil {
					fmt.Printf("  ✗ %s failed validation\n", name)
					if result != nil {
						for _, errMsg := range result.Errors {
							fmt.Printf("    - %s\n", errMsg)
						}
					}
					failed++
				} else {
					if result.Fixed {
						fmt.Printf("  ✓ %s passed (style issues auto-fixed)\n", name)
					} else {
						fmt.Printf("  ✓ %s passed\n", name)
					}
				}
			}
		} else {
			fmt.Println("→ No formulas to validate")
		}
	}

	fmt.Println()

	// Validate casks
	caskDir := filepath.Join(repoRoot, "Casks")
	if _, err := os.Stat(caskDir); err == nil {
		casks, err := filepath.Glob(filepath.Join(caskDir, "*.rb"))
		if err != nil {
			return fmt.Errorf("failed to find casks: %w", err)
		}

		if len(casks) > 0 {
			fmt.Println("→ Validating casks...")
			for _, cask := range casks {
				name := strings.TrimSuffix(filepath.Base(cask), ".rb")
				fmt.Printf("  Checking %s...\n", name)

				result, err := validate.ValidateFile(cask, true, fixStyle)
				if err != nil {
					fmt.Printf("  ✗ %s failed validation\n", name)
					if result != nil {
						for _, errMsg := range result.Errors {
							fmt.Printf("    - %s\n", errMsg)
						}
					}
					failed++
				} else {
					if result.Fixed {
						fmt.Printf("  ✓ %s passed (style issues auto-fixed)\n", name)
					} else {
						fmt.Printf("  ✓ %s passed\n", name)
					}
				}
			}
		} else {
			fmt.Println("→ No casks to validate")
		}
	}

	fmt.Println()

	if failed == 0 {
		fmt.Println("✓ All checks passed!")
		return nil
	}

	return fmt.Errorf("✗ %d check(s) failed", failed)
}

func validateFileCmd(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Determine if it's a cask or formula
	isCask := strings.Contains(filePath, "Casks")

	name := strings.TrimSuffix(filepath.Base(filePath), ".rb")
	fmt.Printf("→ Validating %s...\n", name)

	result, err := validate.ValidateFile(filePath, isCask, fixStyle)
	if err != nil {
		fmt.Println("✗ Validation failed")
		if result != nil {
			for _, errMsg := range result.Errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		return err
	}

	if result.Fixed {
		fmt.Println("✓ Validation passed (style issues auto-fixed)")
	} else {
		fmt.Println("✓ Validation passed")
	}

	return nil
}

func findRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
