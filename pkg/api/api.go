// Package api provides the public songid API for reuse in other Go projects.
package api

import (
	"fmt"

	"github.com/james-see/songid/internal/capture"
	"github.com/james-see/songid/internal/fingerprint"
	"github.com/james-see/songid/internal/lookup"
)

// Identifiable holds a song identification result.
type Identifiable struct {
	Title       string
	Artist      string
	Score       float64
	RecordingID string
}

// Options configures the identifier.
type Options struct {
	AcoustIDAPIKey string
	Meta           string // metadata fields to request from AcoustID (default: "recordings")
}

// Identifier identifies songs from audio.
type Identifier struct {
	client *lookup.Client
	meta   string
}

// New creates a new Identifier with the given options.
func New(opts Options) *Identifier {
	meta := opts.Meta
	if meta == "" {
		meta = "recordings"
	}
	return &Identifier{
		client: lookup.New(opts.AcoustIDAPIKey),
		meta:   meta,
	}
}

// FromFile identifies a song from an audio file path.
func (id *Identifier) FromFile(path string) (*Identifiable, error) {
	fp, err := fingerprint.Generate(path)
	if err != nil {
		return nil, fmt.Errorf("fingerprinting failed: %w", err)
	}

	resp, err := id.client.Lookup(fp.Fingerprint, fp.Duration, id.meta)
	if err != nil {
		return nil, fmt.Errorf("lookup failed: %w", err)
	}

	best := resp.BestMatch()
	if best == nil || len(best.Recordings) == 0 {
		return nil, nil // no match found
	}

	rec := best.Recordings[0]
	artistName := ""
	if len(rec.Artists) > 0 {
		artistName = rec.Artists[0].Name
	}

	return &Identifiable{
		Title:       rec.Title,
		Artist:      artistName,
		Score:       best.Score,
		RecordingID: rec.ID,
	}, nil
}

// FromMicrophone records audio from the mic and identifies it.
func (id *Identifier) FromMicrophone(cfg capture.CaptureConfig) (*Identifiable, error) {
	if !capture.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg is required for microphone capture")
	}

	audioPath, err := capture.FromMicrophone(cfg)
	if err != nil {
		return nil, err
	}
	defer capture.Cleanup(audioPath)

	return id.FromFile(audioPath)
}