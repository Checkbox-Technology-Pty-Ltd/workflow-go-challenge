package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentTemperature(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lon            float64
		serverResponse string
		serverStatus   int
		wantTemp       float64
		wantErr        bool
	}{
		{
			name:           "successful response",
			lat:            -33.8688,
			lon:            151.2093,
			serverResponse: `{"current_weather":{"temperature":28.5}}`,
			serverStatus:   http.StatusOK,
			wantTemp:       28.5,
			wantErr:        false,
		},
		{
			name:           "server error",
			lat:            -33.8688,
			lon:            151.2093,
			serverResponse: `{"error":"internal error"}`,
			serverStatus:   http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "invalid json response",
			lat:            -33.8688,
			lon:            151.2093,
			serverResponse: `not valid json`,
			serverStatus:   http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewOpenMeteoClientWithHTTP(server.Client(), server.URL)
			temp, err := client.GetCurrentTemperature(context.Background(), tt.lat, tt.lon)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentTemperature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && temp != tt.wantTemp {
				t.Errorf("GetCurrentTemperature() = %v, want %v", temp, tt.wantTemp)
			}
		})
	}
}

func TestGetTemperatureForCity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"current_weather":{"temperature":22.3}}`))
	}))
	defer server.Close()

	client := NewOpenMeteoClientWithHTTP(server.Client(), server.URL)

	t.Run("known city", func(t *testing.T) {
		temp, err := client.GetTemperatureForCity(context.Background(), "Sydney")
		if err != nil {
			t.Errorf("GetTemperatureForCity() error = %v", err)
			return
		}
		if temp != 22.3 {
			t.Errorf("GetTemperatureForCity() = %v, want 22.3", temp)
		}
	})

	t.Run("unknown city", func(t *testing.T) {
		_, err := client.GetTemperatureForCity(context.Background(), "UnknownCity")
		if err == nil {
			t.Error("GetTemperatureForCity() expected error for unknown city")
		}
	})
}

func TestCityCoordinates(t *testing.T) {
	expectedCities := []string{"Sydney", "Melbourne", "Brisbane", "Perth", "Adelaide"}

	for _, city := range expectedCities {
		if _, ok := CityCoordinates[city]; !ok {
			t.Errorf("CityCoordinates missing expected city: %s", city)
		}
	}
}
