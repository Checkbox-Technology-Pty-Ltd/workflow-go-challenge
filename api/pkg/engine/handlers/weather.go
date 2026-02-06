package handlers

import (
	"context"
	"fmt"
	"time"

	"workflow-code-test/api/pkg/engine"
)

// WeatherFunc is a function type for fetching weather data
type WeatherFunc func(ctx context.Context, city string) (float64, error)

// WeatherHandler handles weather/integration nodes.
// It reads "form.city" from state and fetches the temperature.
type WeatherHandler struct {
	weatherFn WeatherFunc
}

// NewWeatherHandler creates a new WeatherHandler with the given weather function
func NewWeatherHandler(weatherFn WeatherFunc) *WeatherHandler {
	return &WeatherHandler{weatherFn: weatherFn}
}

func (h *WeatherHandler) NodeType() string { return "integration" }

func (h *WeatherHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	city := ec.GetString("form.city")
	if city == "" {
		return engine.ExecutionStep{}, fmt.Errorf("city not provided in form data")
	}

	if h.weatherFn == nil {
		return engine.ExecutionStep{}, fmt.Errorf("weather client not configured")
	}

	temperature, err := h.weatherFn(ec.Ctx, city)
	if err != nil {
		return engine.ExecutionStep{}, fmt.Errorf("failed to fetch weather for %s: %w", city, err)
	}

	// Store temperature in state for downstream handlers
	ec.Set("weather.temperature", temperature)

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "integration",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   WeatherNodeDuration,
		Output: map[string]interface{}{
			"message": fmt.Sprintf("Fetched weather data for %s", city),
			"apiResponse": map[string]interface{}{
				"endpoint":   "https://api.open-meteo.com/v1/forecast",
				"method":     "GET",
				"statusCode": 200,
				"data": map[string]interface{}{
					"temperature": temperature,
					"city":        city,
				},
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
