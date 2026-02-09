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
func ValidateFile(filePath string, isCask bool, autoFix bool) (*ValidateResult, error) {
	result := &ValidateResult{
		AuditPassed: true,
		StylePassed: true,
		Fixed:       false,
		Errors:      []string{},
	}

	// Run brew audit
	if err := runAudit(filePath, isCask); err != nil {
		result.AuditPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("audit failed: %v", err))
	}

	// Run brew style (with --fix if autoFix is true)
	if err := runStyle(filePath, autoFix); err != nil {
		result.StylePassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("style check failed: %v", err))
	} else if autoFix {
		result.Fixed = true
	}

	// If autoFix was enabled and style passed, re-run audit to ensure fixes didn't break anything
	if autoFix && result.Fixed && result.AuditPassed {
		if err := runAudit(filePath, isCask); err != nil {
			result.AuditPassed = false
			result.Errors = append(result.Errors, fmt.Sprintf("audit failed after style fixes: %v", err))
		}
	}

	if len(result.Errors) > 0 {
		return result, fmt.Errorf("validation failed: %s", strings.Join(result.Errors, "; "))
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
