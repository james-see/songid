package fingerprint

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// FingerprintResult holds the output of fpcalc.
type FingerprintResult struct {
	Duration     float64 `json:"duration"`
	Fingerprint  string  `json:"fingerprint"`
}

// Generate runs fpcalc on the given audio file and returns the fingerprint.
// fpcalc is the Chromaprint CLI tool (brew install chromaprint / apt install chromaprint).
func Generate(audioPath string) (*FingerprintResult, error) {
	cmd := exec.Command("fpcalc", "-json", audioPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("fpcalc failed: %w\noutput: %s", err, string(output))
	}

	var result FingerprintResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse fpcalc output: %w", err)
	}

	if result.Fingerprint == "" {
		return nil, fmt.Errorf("fpcalc returned empty fingerprint")
	}

	return &result, nil
}

// IsAvailable checks if fpcalc is installed and accessible.
func IsAvailable() bool {
	cmd := exec.Command("fpcalc", "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// Version returns the fpcalc version string.
func Version() (string, error) {
	cmd := exec.Command("fpcalc", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}