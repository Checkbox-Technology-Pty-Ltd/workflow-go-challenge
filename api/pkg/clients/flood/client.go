package flood

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Result holds flood risk data for a location.
type Result struct {
	Discharge float64 // river discharge in m³/s
	RiskLevel string  // "low", "moderate", "high"
}

// Client defines the interface for fetching flood risk data.
type Client interface {
	GetFloodRisk(ctx context.Context, lat, lon float64) (*Result, error)
}

// OpenMeteoClient fetches flood data from the Open-Meteo Flood API.
type OpenMeteoClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOpenMeteoClient(httpClient *http.Client) *OpenMeteoClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OpenMeteoClient{
		baseURL:    "https://flood-api.open-meteo.com/v1/flood",
		httpClient: httpClient,
	}
}

func (c *OpenMeteoClient) GetFloodRisk(ctx context.Context, lat, lon float64) (*Result, error) {
	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&daily=river_discharge", c.baseURL, lat, lon)

	slog.Debug("calling flood API", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flood API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flood API returned %d: %s", resp.StatusCode, string(body))
	}

	var data struct {
		Daily struct {
			RiverDischarge []float64 `json:"river_discharge"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse flood response: %w", err)
	}

	discharge := 0.0
	if len(data.Daily.RiverDischarge) > 0 {
		discharge = data.Daily.RiverDischarge[0]
	}

	risk := classifyRisk(discharge)

	return &Result{
		Discharge: discharge,
		RiskLevel: risk,
	}, nil
}

func classifyRisk(discharge float64) string {
	switch {
	case discharge > 500:
		return "high"
	case discharge > 100:
		return "moderate"
	default:
		return "low"
	}
}
