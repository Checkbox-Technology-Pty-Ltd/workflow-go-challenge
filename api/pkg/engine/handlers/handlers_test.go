package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"workflow-code-test/api/pkg/engine"
)

func TestStartHandler(t *testing.T) {
	handler := NewStartHandler()

	if handler.NodeType() != "start" {
		t.Errorf("expected node type 'start', got %q", handler.NodeType())
	}

	ec := engine.NewExecutionContext(context.Background())
	ec.StepNumber = 1

	node := &engine.Node{ID: "start-1", Type: "start"}
	step, err := handler.Execute(ec, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "start" {
		t.Errorf("expected step type 'start', got %q", step.NodeType)
	}
	if step.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", step.Status)
	}
	if step.Output["message"] != "Workflow started" {
		t.Errorf("unexpected message: %v", step.Output["message"])
	}
}

func TestEndHandler(t *testing.T) {
	handler := NewEndHandler()

	if handler.NodeType() != "end" {
		t.Errorf("expected node type 'end', got %q", handler.NodeType())
	}

	ec := engine.NewExecutionContext(context.Background())
	ec.StepNumber = 5

	node := &engine.Node{ID: "end-1", Type: "end"}
	step, err := handler.Execute(ec, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "end" {
		t.Errorf("expected step type 'end', got %q", step.NodeType)
	}
	if step.Output["message"] != "Workflow completed" {
		t.Errorf("unexpected message: %v", step.Output["message"])
	}
}

func TestFormHandler(t *testing.T) {
	handler := NewFormHandler()

	if handler.NodeType() != "form" {
		t.Errorf("expected node type 'form', got %q", handler.NodeType())
	}

	ec := engine.NewExecutionContext(context.Background())
	ec.StepNumber = 2
	ec.Set("form.name", "Alice")
	ec.Set("form.email", "alice@example.com")
	ec.Set("form.city", "Sydney")

	node := &engine.Node{ID: "form-1", Type: "form"}
	step, err := handler.Execute(ec, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "form" {
		t.Errorf("expected step type 'form', got %q", step.NodeType)
	}

	formData, ok := step.Output["formData"].(map[string]interface{})
	if !ok {
		t.Fatal("expected formData in output")
	}
	if formData["name"] != "Alice" {
		t.Errorf("expected name 'Alice', got %v", formData["name"])
	}
}

func TestWeatherHandler(t *testing.T) {
	t.Run("with weather function", func(t *testing.T) {
		handler := NewWeatherHandler(func(ctx context.Context, city string) (float64, error) {
			if city != "Sydney" {
				t.Errorf("expected city 'Sydney', got %q", city)
			}
			return 28.5, nil
		})

		if handler.NodeType() != "integration" {
			t.Errorf("expected node type 'integration', got %q", handler.NodeType())
		}

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 3
		ec.Set("form.city", "Sydney")

		node := &engine.Node{ID: "weather-1", Type: "integration"}
		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ec.GetFloat("weather.temperature") != 28.5 {
			t.Errorf("expected temperature 28.5 in context, got %v", ec.GetFloat("weather.temperature"))
		}

		apiResponse, ok := step.Output["apiResponse"].(map[string]interface{})
		if !ok {
			t.Fatal("expected apiResponse in output")
		}
		data, ok := apiResponse["data"].(map[string]interface{})
		if !ok {
			t.Fatal("expected data in apiResponse")
		}
		if data["temperature"].(float64) != 28.5 {
			t.Errorf("unexpected temperature in response")
		}
	})

	t.Run("with weather function error", func(t *testing.T) {
		handler := NewWeatherHandler(func(ctx context.Context, city string) (float64, error) {
			return 0, errors.New("API error")
		})

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 3
		ec.Set("form.city", "Sydney")

		node := &engine.Node{ID: "weather-1", Type: "integration"}
		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Fatal("expected error when weather API fails")
		}
	})

	t.Run("without weather function", func(t *testing.T) {
		handler := NewWeatherHandler(nil)

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 3
		ec.Set("form.city", "Melbourne")

		node := &engine.Node{ID: "weather-1", Type: "integration"}
		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Fatal("expected error when weather client not configured")
		}
	})

	t.Run("missing city in form data", func(t *testing.T) {
		handler := NewWeatherHandler(nil)

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 3
		// No city set

		node := &engine.Node{ID: "weather-1", Type: "integration"}
		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Error("expected error when city is missing")
		}
	})
}

func TestConditionHandler(t *testing.T) {
	handler := NewConditionHandler()

	if handler.NodeType() != "condition" {
		t.Errorf("expected node type 'condition', got %q", handler.NodeType())
	}

	tests := []struct {
		name        string
		temperature float64
		operator    string
		threshold   float64
		wantResult  bool
	}{
		{"greater than true", 30.0, "greater_than", 25.0, true},
		{"greater than false", 20.0, "greater_than", 25.0, false},
		{"less than true", 20.0, "less_than", 25.0, true},
		{"equals true", 25.0, "equals", 25.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, _ := json.Marshal(map[string]interface{}{
				"operator":  tt.operator,
				"threshold": tt.threshold,
			})
			node := &engine.Node{
				ID:       "condition-1",
				Type:     "condition",
				Metadata: metadata,
			}

			ec := engine.NewExecutionContext(context.Background())
			ec.StepNumber = 4
			ec.Set("weather.temperature", tt.temperature)

			step, err := handler.Execute(ec, node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			condResult, ok := step.Output["conditionResult"].(map[string]interface{})
			if !ok {
				t.Fatal("expected conditionResult in output")
			}
			if condResult["result"].(bool) != tt.wantResult {
				t.Errorf("expected result %v, got %v", tt.wantResult, condResult["result"])
			}
		})
	}
}

func TestEmailHandler(t *testing.T) {
	handler := NewEmailHandler()

	if handler.NodeType() != "email" {
		t.Errorf("expected node type 'email', got %q", handler.NodeType())
	}

	t.Run("sends email to user from form data", func(t *testing.T) {
		metadata, _ := json.Marshal(map[string]interface{}{
			"subject": "Weather Alert",
		})
		node := &engine.Node{
			ID:       "email-1",
			Type:     "email",
			Metadata: metadata,
		}

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 5
		ec.Set("form.name", "Alice")
		ec.Set("form.email", "alice@example.com")
		ec.Set("form.city", "Sydney")
		ec.Set("weather.temperature", 30.0)

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if step.NodeType != "email" {
			t.Errorf("expected step type 'email', got %q", step.NodeType)
		}

		emailContent, ok := step.Output["emailContent"].(map[string]interface{})
		if !ok {
			t.Fatal("expected emailContent in output")
		}
		if emailContent["to"] != "alice@example.com" {
			t.Errorf("expected email to 'alice@example.com', got %v", emailContent["to"])
		}
		if emailContent["subject"] != "Weather Alert" {
			t.Errorf("expected subject 'Weather Alert', got %v", emailContent["subject"])
		}
		expectedBody := "Hi Alice, weather alert for Sydney! Temperature is 30.0°C!"
		if emailContent["body"] != expectedBody {
			t.Errorf("expected body %q, got %v", expectedBody, emailContent["body"])
		}
	})

	t.Run("uses default subject when not in metadata", func(t *testing.T) {
		node := &engine.Node{ID: "email-1", Type: "email"}

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 5
		ec.Set("form.name", "Bob")
		ec.Set("form.email", "bob@example.com")
		ec.Set("form.city", "Melbourne")
		ec.Set("weather.temperature", 22.0)

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		emailContent := step.Output["emailContent"].(map[string]interface{})
		if emailContent["subject"] != "Weather Alert" {
			t.Errorf("expected default subject 'Weather Alert', got %v", emailContent["subject"])
		}
	})

	t.Run("missing email in form data", func(t *testing.T) {
		node := &engine.Node{ID: "email-1", Type: "email"}

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 5
		ec.Set("form.name", "Alice")
		// No email set
		ec.Set("weather.temperature", 30.0)

		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Error("expected error when email is missing")
		}
	})

	t.Run("invalid metadata returns error", func(t *testing.T) {
		node := &engine.Node{
			ID:       "email-1",
			Type:     "email",
			Metadata: []byte(`{invalid json`),
		}

		ec := engine.NewExecutionContext(context.Background())
		ec.StepNumber = 5
		ec.Set("form.name", "Alice")
		ec.Set("form.email", "alice@example.com")
		ec.Set("form.city", "Sydney")
		ec.Set("weather.temperature", 30.0)

		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Error("expected error when metadata is invalid JSON")
		}
	})
}

func TestRegistry(t *testing.T) {
	registry := engine.NewRegistry()

	registry.Register(NewStartHandler())
	registry.Register(NewEndHandler())

	t.Run("get registered handler", func(t *testing.T) {
		handler, ok := registry.Get("start")
		if !ok {
			t.Fatal("expected to find 'start' handler")
		}
		if handler.NodeType() != "start" {
			t.Errorf("expected 'start' handler, got %q", handler.NodeType())
		}
	})

	t.Run("get unregistered handler", func(t *testing.T) {
		_, ok := registry.Get("unknown")
		if ok {
			t.Error("expected not to find 'unknown' handler")
		}
	})

	t.Run("node types", func(t *testing.T) {
		types := registry.NodeTypes()
		if len(types) != 2 {
			t.Errorf("expected 2 types, got %d", len(types))
		}
	})
}

// Suppress unused import warning
var _ = time.Now
