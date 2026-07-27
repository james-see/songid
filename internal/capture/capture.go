package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// CaptureConfig controls microphone recording.
type CaptureConfig struct {
	Duration   time.Duration
	SampleRate int
	Channels   int
	Device     string // optional device name override
}

// DefaultConfig returns sensible defaults for microphone capture.
func DefaultConfig() CaptureConfig {
	return CaptureConfig{
		Duration:   8 * time.Second,
		SampleRate: 44100,
		Channels:   1,
	}
}

// FromMicrophone records audio from the system microphone to a temp WAV file.
// It uses ffmpeg under the hood for cross-platform audio capture.
// On macOS it uses avfoundation, on Linux it uses alsa/pulseaudio.
func FromMicrophone(cfg CaptureConfig) (string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg is required for microphone capture (install: brew install ffmpeg / apt install ffmpeg)")
	}

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("songid_capture_%d.wav", time.Now().Unix()))

	var cmd *exec.Cmd
	durationSec := fmt.Sprintf("%.0f", cfg.Duration.Seconds())

	switch runtime.GOOS {
	case "darwin":
		// avfoundation device ":0" = default audio input
		device := ":0"
		if cfg.Device != "" {
			device = cfg.Device
		}
		cmd = exec.Command("ffmpeg",
			"-y",
			"-f", "avfoundation",
			"-i", device,
			"-t", durationSec,
			"-ar", fmt.Sprintf("%d", cfg.SampleRate),
			"-ac", fmt.Sprintf("%d", cfg.Channels),
			outputPath,
		)
	case "linux":
		// Try pulseaudio first, fall back to alsa
		inputFmt := "pulse"
		device := "default"
		if cfg.Device != "" {
			device = cfg.Device
		}
		cmd = exec.Command("ffmpeg",
			"-y",
			"-f", inputFmt,
			"-i", device,
			"-t", durationSec,
			"-ar", fmt.Sprintf("%d", cfg.SampleRate),
			"-ac", fmt.Sprintf("%d", cfg.Channels),
			outputPath,
		)
	default:
		return "", fmt.Errorf("unsupported OS for microphone capture: %s", runtime.GOOS)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg capture failed: %w\noutput: %s", err, string(output))
	}

	// Verify the file was created
	if _, err := os.Stat(outputPath); err != nil {
		return "", fmt.Errorf("capture file was not created: %w", err)
	}

	return outputPath, nil
}

// IsAvailable checks if ffmpeg is installed.
func IsAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// Cleanup removes a temporary capture file.
func Cleanup(path string) {
	os.Remove(path)
}