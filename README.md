# songid

Identify songs from audio files or microphone input using open-source audio fingerprinting. No subscriptions, no third-party paid services.

songid uses [Chromaprint](https://github.com/acoustid/chromaprint) for audio fingerprinting and [AcoustID](https://acoustid.org/) for lookup — both free and open source. AcoustID is powered by open data from [MusicBrainz](https://musicbrainz.org/).

## Install

```bash
# Install dependencies
brew install chromaprint ffmpeg    # macOS
# apt install chromaprint ffmpeg    # Linux (Debian/Ubuntu)

# Install songid CLI
go install github.com/james-see/songid/cmd/songid@latest

# Install songid MCP server
go install github.com/james-see/songid/cmd/mcp@latest
```

## Get a free AcoustID API key

1. Register at https://acoustid.org/applications
2. Create an application
3. Copy the API key
4. Set it as an environment variable:

```bash
export ACOUSTID_API_KEY=your_key_here
```

## CLI Usage

```bash
# Identify from microphone (records 8 seconds by default)
songid listen
songid listen --duration 10s

# Identify from an audio file
songid file ~/Downloads/song.mp3

# Generate fingerprint without lookup
songid fingerprint ~/Downloads/song.wav

# Check dependencies
songid doctor

# JSON output
songid listen --json
songid file ~/Downloads/song.mp3 --json
```

## MCP Server

songid includes an MCP server for integration with AI agent harnesses (goose, Hermes, Claude, etc.).

```bash
# Run the MCP server
songid-mcp
# or with API key
songid-mcp --api-key your_key_here
```

### MCP Tools

| Tool | Description |
|------|-------------|
| `identify_from_microphone` | Record from system microphone and identify the song |
| `identify_from_file` | Identify a song from an audio file path |
| `doctor` | Check if dependencies (fpcalc, ffmpeg, API key) are installed |

### Goose extension config

Add to your goose `servers.json`:

```json
{
  "songid": {
    "command": "songid-mcp",
    "env": {
      "ACOUSTID_API_KEY": "your_key_here"
    }
  }
}
```

## Go API

```go
import "github.com/james-see/songid/pkg/api"

identifier := api.New(api.Options{
    AcoustIDAPIKey: "your_key_here",
})

// From file
result, err := identifier.FromFile("song.mp3")

// From microphone
cfg := capture.CaptureConfig{
    Duration:   8 * time.Second,
    SampleRate: 44100,
    Channels:   1,
}
result, err := identifier.FromMicrophone(cfg)
```

## How it works

1. **Capture** — ffmpeg records audio from the microphone (or reads from a file)
2. **Fingerprint** — Chromaprint (`fpcalc`) generates a compact audio fingerprint
3. **Lookup** — The fingerprint is sent to the AcoustID web service, which returns matching recordings
4. **Enrich** — Metadata (title, artist) comes from MusicBrainz via AcoustID

## Dependencies

| Dependency | Required | Install |
|-----------|----------|---------|
| [Chromaprint](https://github.com/acoustid/chromaprint) | Yes | `brew install chromaprint` |
| [ffmpeg](https://ffmpeg.org) | For microphone capture only | `brew install ffmpeg` |
| [AcoustID API key](https://acoustid.org/applications) | Yes | Free registration |

## License

MIT

## Author

James Campbell ([@james-see](https://github.com/james-see))