package flood

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetFloodRisk_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"daily":{"river_discharge":[250.0]}}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	result, err := client.GetFloodRisk(context.Background(), -27.47, 153.03)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Discharge != 250.0 {
		t.Errorf("expected discharge 250.0, got %v", result.Discharge)
	}
	if result.RiskLevel != "moderate" {
		t.Errorf("expected risk 'moderate', got %q", result.RiskLevel)
	}
}

func TestGetFloodRisk_HighRisk(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"daily":{"river_discharge":[750.0]}}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	result, err := client.GetFloodRisk(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RiskLevel != "high" {
		t.Errorf("expected risk 'high', got %q", result.RiskLevel)
	}
}

func TestGetFloodRisk_EmptyDischarge(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"daily":{"river_discharge":[]}}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	result, err := client.GetFloodRisk(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Discharge != 0 {
		t.Errorf("expected 0 discharge for empty array, got %v", result.Discharge)
	}
	if result.RiskLevel != "low" {
		t.Errorf("expected risk 'low', got %q", result.RiskLevel)
	}
}

func TestGetFloodRisk_ServerError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`service down`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetFloodRisk(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFloodRisk_MalformedJSON(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{broken`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetFloodRisk(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestGetFloodRisk_Timeout(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetFloodRisk(ctx, 0, 0)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestClassifyRisk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		discharge float64
		want      string
	}{
		{0, "low"},
		{100, "low"},
		{101, "moderate"},
		{500, "moderate"},
		{501, "high"},
	}

	for _, tt := range tests {
		got := classifyRisk(tt.discharge)
		if got != tt.want {
			t.Errorf("classifyRisk(%v) = %q, want %q", tt.discharge, got, tt.want)
		}
	}
}
