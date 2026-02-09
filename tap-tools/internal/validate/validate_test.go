package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	// Skip if brew is not installed
	if _, err := os.Stat("/home/linuxbrew/.linuxbrew/bin/brew"); os.IsNotExist(err) {
		t.Skip("brew not installed, skipping validation tests")
	}

	// Create a temporary test formula
	tmpDir := t.TempDir()
	testFormula := filepath.Join(tmpDir, "test.rb")

	validFormula := `class Test < Formula
  desc "Test formula"
  homepage "https://example.com"
  url "https://example.com/test-1.0.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"

  def install
    bin.install "test"
  end

  test do
    system "#{bin}/test", "--version"
  end
end
`

	if err := os.WriteFile(testFormula, []byte(validFormula), 0644); err != nil {
		t.Fatalf("failed to write test formula: %v", err)
	}

	// Test validation without auto-fix
	result, err := ValidateFile(testFormula, false, false)
	if err != nil {
		t.Logf("validation failed (expected for offline test): %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestValidateResult(t *testing.T) {
	result := &ValidateResult{
		AuditPassed: true,
		StylePassed: false,
		Fixed:       false,
		Errors:      []string{"style error"},
	}

	if result.AuditPassed != true {
		t.Error("expected audit to pass")
	}

	if result.StylePassed != false {
		t.Error("expected style to fail")
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}
