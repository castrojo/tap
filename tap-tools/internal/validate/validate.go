package validate

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ValidateResult holds validation results
type ValidateResult struct {
	AuditPassed bool
	StylePassed bool
	Fixed       bool
	Errors      []string
}

// ValidateFile validates a formula or cask file using brew audit and brew style
// Note: brew audit is skipped during generation since it requires the package to be in a tap
func ValidateFile(filePath string, isCask bool, autoFix bool) (*ValidateResult, error) {
	result := &ValidateResult{
		AuditPassed: true,
		StylePassed: true,
		Fixed:       false,
		Errors:      []string{},
	}

	// Skip brew audit - it requires the file to be in a tapped repository
	// The pre-commit hook will run audit after the file is committed to the tap

	// Run brew style (with --fix if autoFix is true)
	if err := runStyle(filePath, autoFix); err != nil {
		result.StylePassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("style check failed: %v", err))
		// Return error only if style check failed
		return result, fmt.Errorf("validation failed: %s", strings.Join(result.Errors, "; "))
	}

	if autoFix {
		result.Fixed = true
	}

	return result, nil
}

func runAudit(filePath string, isCask bool) error {
	args := []string{"audit", "--strict", "--online"}
	if isCask {
		args = append(args, "--cask")
	}
	args = append(args, filePath)

	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runStyle(filePath string, fix bool) error {
	args := []string{"style"}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, filePath)

	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
