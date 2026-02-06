package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client defines the interface for weather data retrieval
type Client interface {
	GetCurrentTemperature(ctx context.Context, lat, lon float64) (float64, error)
}

// OpenMeteoClient implements Client using the Open-Meteo API
type OpenMeteoClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewOpenMeteoClient creates a new OpenMeteoClient with default settings
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.open-meteo.com",
	}
}

// NewOpenMeteoClientWithHTTP creates an OpenMeteoClient with a custom HTTP client and base URL
func NewOpenMeteoClientWithHTTP(client *http.Client, baseURL string) *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: client,
		baseURL:    baseURL,
	}
}

type openMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

// GetCurrentTemperature fetches the current temperature for given coordinates
func (c *OpenMeteoClient) GetCurrentTemperature(ctx context.Context, lat, lon float64) (float64, error) {
	url := fmt.Sprintf("%s/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true", c.baseURL, lat, lon)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.CurrentWeather.Temperature, nil
}

// CityCoordinates maps city names to their lat/lon coordinates
var CityCoordinates = map[string]struct {
	Lat float64
	Lon float64
}{
	"Sydney":    {Lat: -33.8688, Lon: 151.2093},
	"Melbourne": {Lat: -37.8136, Lon: 144.9631},
	"Brisbane":  {Lat: -27.4698, Lon: 153.0251},
	"Perth":     {Lat: -31.9505, Lon: 115.8605},
	"Adelaide":  {Lat: -34.9285, Lon: 138.6007},
}

// GetTemperatureForCity is a convenience function that looks up coordinates and fetches temperature
func (c *OpenMeteoClient) GetTemperatureForCity(ctx context.Context, city string) (float64, error) {
	coords, ok := CityCoordinates[city]
	if !ok {
		return 0, fmt.Errorf("unknown city: %s", city)
	}
	return c.GetCurrentTemperature(ctx, coords.Lat, coords.Lon)
}
