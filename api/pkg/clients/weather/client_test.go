package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetTemperature_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"current_weather":{"temperature":28.5}}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	temp, err := client.GetTemperature(context.Background(), -33.87, 151.21)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if temp != 28.5 {
		t.Errorf("expected 28.5, got %v", temp)
	}
}

func TestGetTemperature_ServerError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetTemperature(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTemperature_MalformedJSON(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetTemperature(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestGetTemperature_Timeout(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.GetTemperature(ctx, 0, 0)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGetTemperature_EmptyResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"current_weather":{}}`))
	}))
	defer server.Close()

	client := &OpenMeteoClient{baseURL: server.URL, httpClient: server.Client()}
	temp, err := client.GetTemperature(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if temp != 0 {
		t.Errorf("expected 0 for missing temperature, got %v", temp)
	}
}

func TestNewOpenMeteoClient_DefaultHTTPClient(t *testing.T) {
	t.Parallel()
	client := NewOpenMeteoClient(nil)
	if client.httpClient != http.DefaultClient {
		t.Error("expected http.DefaultClient when nil passed")
	}
}
