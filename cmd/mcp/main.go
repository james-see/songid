package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/james-see/songid/internal/capture"
	"github.com/james-see/songid/internal/fingerprint"
	"github.com/james-see/songid/pkg/api"
)

// ListenInput is the input schema for identify_from_microphone.
type ListenInput struct {
	Duration string `jsonschema:"recording duration (e.g. '8s', '10s'). default: 8s" json:"duration,omitempty"`
}

// FileInput is the input schema for identify_from_file.
type FileInput struct {
	Path string `jsonschema:"path to the audio file to identify" json:"path"`
}

// DoctorInput is the input schema for doctor (no params).
type DoctorInput struct{}

// MatchOutput is the output schema for identification tools.
type MatchOutput struct {
	Matched      bool    `json:"matched"`
	Title        string  `json:"title,omitempty"`
	Artist       string  `json:"artist,omitempty"`
	Score        float64 `json:"score,omitempty"`
	RecordingID  string  `json:"recording_id,omitempty"`
	Message      string  `json:"message,omitempty"`
}

// DoctorOutput is the output schema for the doctor tool.
type DoctorOutput struct {
	Fpcalc  string `json:"fpcalc"`
	Ffmpeg  string `json:"ffmpeg"`
	APIKey  string `json:"api_key"`
}

func main() {
	apiKey := flag.String("api-key", "", "AcoustID API key (or set ACOUSTID_API_KEY env var)")
	flag.Parse()

	key := *apiKey
	if key == "" {
		key = os.Getenv("ACOUSTID_API_KEY")
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "songid",
		Version: "0.1.0",
	}, nil)

	// identify_from_microphone
	mcp.AddTool(server,
		&mcp.Tool{
			Description: "Record audio from the system microphone and identify the song using Chromaprint fingerprinting and AcoustID lookup.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input ListenInput) (*mcp.CallToolResult, MatchOutput, error) {
			if key == "" {
				return errorResult("AcoustID API key is required (register at https://acoustid.org/applications, set ACOUSTID_API_KEY env var)"), MatchOutput{}, nil
			}

			durStr := input.Duration
			if durStr == "" {
				durStr = "8s"
			}
			dur, err := time.ParseDuration(durStr)
			if err != nil {
				return errorResult(fmt.Sprintf("Invalid duration: %s", durStr)), MatchOutput{}, nil
			}

			if !fingerprint.IsAvailable() {
				return errorResult("fpcalc (chromaprint) is required. Install: brew install chromaprint"), MatchOutput{}, nil
			}
			if !capture.IsAvailable() {
				return errorResult("ffmpeg is required. Install: brew install ffmpeg"), MatchOutput{}, nil
			}

			cfg := capture.CaptureConfig{
				Duration:   dur,
				SampleRate: 44100,
				Channels:   1,
			}

			identifier := api.New(api.Options{AcoustIDAPIKey: key})
			result, err := identifier.FromMicrophone(cfg)
			if err != nil {
				return errorResult(fmt.Sprintf("Identification failed: %v", err)), MatchOutput{}, nil
			}

			if result == nil {
				return nil, MatchOutput{Matched: false, Message: "No match found."}, nil
			}

			return nil, MatchOutput{
				Matched:     true,
				Title:       result.Title,
				Artist:      result.Artist,
				Score:       result.Score,
				RecordingID: result.RecordingID,
			}, nil
		},
	)

	// identify_from_file
	mcp.AddTool(server,
		&mcp.Tool{
			Description: "Identify a song from an audio file path using Chromaprint fingerprinting and AcoustID lookup.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input FileInput) (*mcp.CallToolResult, MatchOutput, error) {
			if key == "" {
				return errorResult("AcoustID API key is required (register at https://acoustid.org/applications, set ACOUSTID_API_KEY env var)"), MatchOutput{}, nil
			}

			if input.Path == "" {
				return errorResult("path parameter is required"), MatchOutput{}, nil
			}

			if _, err := os.Stat(input.Path); err != nil {
				return errorResult(fmt.Sprintf("File not found: %s", input.Path)), MatchOutput{}, nil
			}

			if !fingerprint.IsAvailable() {
				return errorResult("fpcalc (chromaprint) is required. Install: brew install chromaprint"), MatchOutput{}, nil
			}

			identifier := api.New(api.Options{AcoustIDAPIKey: key})
			result, err := identifier.FromFile(input.Path)
			if err != nil {
				return errorResult(fmt.Sprintf("Identification failed: %v", err)), MatchOutput{}, nil
			}

			if result == nil {
				return nil, MatchOutput{Matched: false, Message: "No match found."}, nil
			}

			return nil, MatchOutput{
				Matched:     true,
				Title:       result.Title,
				Artist:      result.Artist,
				Score:       result.Score,
				RecordingID: result.RecordingID,
			}, nil
		},
	)

	// doctor
	mcp.AddTool(server,
		&mcp.Tool{
			Description: "Check if songid dependencies (fpcalc, ffmpeg, API key) are installed and configured.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input DoctorInput) (*mcp.CallToolResult, DoctorOutput, error) {
			out := DoctorOutput{}

			if fingerprint.IsAvailable() {
				ver, _ := fingerprint.Version()
				out.Fpcalc = "OK (" + ver + ")"
			} else {
				out.Fpcalc = "MISSING (install: brew install chromaprint)"
			}

			if capture.IsAvailable() {
				out.Ffmpeg = "OK"
			} else {
				out.Ffmpeg = "MISSING (install: brew install ffmpeg)"
			}

			if key != "" {
				out.APIKey = "OK"
			} else {
				out.APIKey = "NOT SET (register at https://acoustid.org/applications)"
			}

			return nil, out, nil
		},
	)

	// Run via stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "songid MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Error: " + msg},
		},
		IsError: true,
	}
}

// suppress unused import warning for json (used in case we need raw marshaling)
var _ = json.Marshal