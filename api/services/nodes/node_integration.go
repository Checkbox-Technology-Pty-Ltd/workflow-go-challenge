package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// IntegrationNode calls an external API based on its metadata configuration.
// Raw metadata is preserved for ToJSON(); parsed fields are used by Execute().
type IntegrationNode struct {
	base BaseFields

	// Parsed from metadata for execution
	APIEndpoint     string       `json:"apiEndpoint"`
	InputVariables  []string     `json:"inputVariables"`
	OutputVariables []string     `json:"outputVariables"`
	Options         []CityOption `json:"options"`
}

type CityOption struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// NewIntegrationNode constructs itself from the database fields.
// Metadata is parsed into typed fields for Execute(), while the raw
// bytes are kept on base for lossless ToJSON() passthrough.
func NewIntegrationNode(base BaseFields) (*IntegrationNode, error) {
	n := &IntegrationNode{base: base}
	if err := json.Unmarshal(base.Metadata, n); err != nil {
		return nil, fmt.Errorf("invalid integration metadata: %w", err)
	}
	return n, nil
}

// ToJSON returns the React Flow representation.
// Metadata is the raw DB value — no reconstruction, no data loss.
func (n *IntegrationNode) ToJSON() NodeJSON {
	return NodeJSON{
		ID:       n.base.ID,
		Type:     n.base.NodeType,
		Position: n.base.Position,
		Data: NodeData{
			Label:       n.base.Label,
			Description: n.base.Description,
			Metadata:    n.base.Metadata,
		},
	}
}

// Execute resolves the city from context, looks up coordinates,
// and calls the weather API to fetch the current temperature.
func (n *IntegrationNode) Execute(ctx context.Context, nCtx *NodeContext) (*ExecutionResult, error) {
	city, ok := nCtx.Variables["city"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required input variable: city")
	}

	var opt *CityOption
	for i := range n.Options {
		if strings.EqualFold(n.Options[i].City, city) {
			opt = &n.Options[i]
			break
		}
	}
	if opt == nil {
		return nil, fmt.Errorf("unsupported city: %s", city)
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true",
		opt.Lat, opt.Lon,
	)

	slog.Info("calling weather API", "city", city, "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned %d: %s", resp.StatusCode, string(body))
	}

	var weather struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
		} `json:"current_weather"`
	}
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %w", err)
	}

	temp := weather.CurrentWeather.Temperature
	slog.Info("weather API result", "city", city, "temperature", temp)

	return &ExecutionResult{
		Status: "completed",
		Output: map[string]any{
			"temperature": temp,
			"location":    city,
		},
	}, nil
}
