package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// DownloadFile downloads a file from the given URL and returns its content
func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// CalculateSHA256 calculates the SHA256 checksum of the given data
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// VerifyChecksum verifies that the calculated checksum matches the expected one
func VerifyChecksum(data []byte, expected string) error {
	calculated := CalculateSHA256(data)
	if calculated != strings.ToLower(expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, calculated)
	}
	return nil
}

// FindUpstreamChecksum searches for upstream checksums in common locations
// Returns a map of filename -> checksum
func FindUpstreamChecksum(releaseURL string) (map[string]string, error) {
	// Common checksum file patterns
	patterns := []string{
		"checksums.txt",
		"sha256sums.txt",
		"SHA256SUMS",
		"SHA256SUMS.txt",
		"checksums.sha256",
	}

	// Extract base URL from release URL
	// e.g., https://github.com/owner/repo/releases/download/v1.0.0/
	baseURL := releaseURL
	if idx := strings.LastIndex(releaseURL, "/"); idx != -1 {
		baseURL = releaseURL[:idx+1]
	}

	// Try each pattern
	for _, pattern := range patterns {
		checksumURL := baseURL + pattern
		data, err := DownloadFile(checksumURL)
		if err != nil {
			continue // Try next pattern
		}

		// Parse checksum file
		checksums := parseChecksumFile(string(data))
		if len(checksums) > 0 {
			return checksums, nil
		}
	}

	return nil, fmt.Errorf("no upstream checksums found")
}

// parseChecksumFile parses a checksum file in various formats
// Supports:
// - "checksum  filename" (two spaces, common in sha256sum output)
// - "checksum *filename" (asterisk for binary mode)
// - "checksum filename" (single space)
func parseChecksumFile(content string) map[string]string {
	checksums := make(map[string]string)

	// Regular expression to match checksum lines
	// Matches: <64-char hex> <whitespace or *> <filename>
	re := regexp.MustCompile(`([a-fA-F0-9]{64})\s+[\*]?(.+)`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			checksum := strings.ToLower(matches[1])
			filename := strings.TrimSpace(matches[2])
			checksums[filename] = checksum
		}
	}

	return checksums
}

// VerifyFromUpstream downloads a file and verifies it against upstream checksums
func VerifyFromUpstream(downloadURL, filename string, releaseURL string) (sha256sum string, verified bool, err error) {
	// Download the file
	data, err := DownloadFile(downloadURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to download file: %w", err)
	}

	// Calculate checksum
	calculated := CalculateSHA256(data)

	// Try to find upstream checksum
	upstreamChecksums, err := FindUpstreamChecksum(releaseURL)
	if err != nil {
		// No upstream checksum found, but we still have the calculated one
		return calculated, false, nil
	}

	// Look for this file in upstream checksums
	if expected, found := upstreamChecksums[filename]; found {
		if calculated != expected {
			return calculated, false, fmt.Errorf("checksum mismatch: expected %s, got %s", expected, calculated)
		}
		return calculated, true, nil
	}

	// File not in upstream checksums, but we have calculated one
	return calculated, false, nil
}
