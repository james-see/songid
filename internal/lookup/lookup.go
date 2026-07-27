package lookup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Client is an AcoustID API client.
type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

// DefaultAPIKey is the AcoustID example key from the public docs.
// It works for lookups out of the box. For production use, register
// your own app at https://acoustid.org/applications and set ACOUSTID_API_KEY.
const DefaultAPIKey = "igwQBqQvQU"

// New creates a new AcoustID lookup client.
func New(apiKey string) *Client {
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.acoustid.org/v2",
		HTTP:    &http.Client{},
	}
}

// Recording represents a matched recording from AcoustID.
type Recording struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artists  []Artist `json:"artists"`
	Duration int     `json:"duration"`
}

// Artist represents an artist in a recording.
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Result represents a single lookup result.
type Result struct {
	ID         string      `json:"id"`
	Score      float64     `json:"score"`
	Recordings []Recording `json:"recordings"`
}

// Response is the AcoustID API response.
type Response struct {
	Status  string   `json:"status"`
	Results []Result `json:"results"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError represents an API error.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Lookup queries the AcoustID API with a fingerprint and duration.
// meta specifies what metadata to return (e.g. "recordings", "releases", "releasegroups").
func (c *Client) Lookup(fingerprint string, duration float64, meta string) (*Response, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("AcoustID API key is required (register at https://acoustid.org/applications)")
	}

	params := url.Values{
		"client":      {c.APIKey},
		"duration":    {strconv.FormatFloat(duration, 'f', -1, 64)},
		"fingerprint": {fingerprint},
	}
	if meta != "" {
		params.Set("meta", meta)
	}

	resp, err := c.HTTP.Get(c.BaseURL + "/lookup?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("AcoustID request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode AcoustID response: %w", err)
	}

	if apiResp.Status != "ok" {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("AcoustID error (code %d): %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("AcoustID returned non-ok status: %s", apiResp.Status)
	}

	return &apiResp, nil
}

// BestMatch returns the highest-scoring result with recordings, or nil if no match.
func (r *Response) BestMatch() *Result {
	var best *Result
	for i := range r.Results {
		if len(r.Results[i].Recordings) == 0 {
			continue
		}
		if best == nil || r.Results[i].Score > best.Score {
			best = &r.Results[i]
		}
	}
	return best
}