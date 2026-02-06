package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestStartHandler(t *testing.T) {
	handler := &StartHandler{}

	if handler.NodeType() != "start" {
		t.Errorf("expected node type 'start', got %q", handler.NodeType())
	}

	ec := &ExecutionContext{
		StepNumber: 1,
		StartTime:  time.Now(),
	}

	step, err := handler.Execute(ec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "start" {
		t.Errorf("expected step type 'start', got %q", step.NodeType)
	}
	if step.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", step.Status)
	}
	if step.Output.Message != "Workflow started" {
		t.Errorf("unexpected message: %q", step.Output.Message)
	}
}

func TestEndHandler(t *testing.T) {
	handler := &EndHandler{}

	if handler.NodeType() != "end" {
		t.Errorf("expected node type 'end', got %q", handler.NodeType())
	}

	ec := &ExecutionContext{StepNumber: 5}

	step, err := handler.Execute(ec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "end" {
		t.Errorf("expected step type 'end', got %q", step.NodeType)
	}
	if step.Output.Message != "Workflow completed" {
		t.Errorf("unexpected message: %q", step.Output.Message)
	}
}

func TestFormHandler(t *testing.T) {
	handler := &FormHandler{}

	if handler.NodeType() != "form" {
		t.Errorf("expected node type 'form', got %q", handler.NodeType())
	}

	ec := &ExecutionContext{
		StepNumber: 2,
		FormData: FormData{
			Name:  "Alice",
			Email: "alice@example.com",
			City:  "Sydney",
		},
	}

	step, err := handler.Execute(ec, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if step.NodeType != "form" {
		t.Errorf("expected step type 'form', got %q", step.NodeType)
	}
	if step.Output.FormData == nil {
		t.Fatal("expected FormData in output")
	}
	if step.Output.FormData.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", step.Output.FormData.Name)
	}

	// Verify form data is stored in State for downstream handlers
	if ec.State["name"] != "Alice" {
		t.Errorf("expected State['name'] = 'Alice', got %v", ec.State["name"])
	}
	if ec.State["email"] != "alice@example.com" {
		t.Errorf("expected State['email'] = 'alice@example.com', got %v", ec.State["email"])
	}
	if ec.State["city"] != "Sydney" {
		t.Errorf("expected State['city'] = 'Sydney', got %v", ec.State["city"])
	}
}

func TestWeatherHandler(t *testing.T) {
	// WeatherHandler reads city from FormData (user input), not node metadata
	node := &Node{}

	t.Run("with weather function", func(t *testing.T) {
		handler := &WeatherHandler{
			weatherFn: func(ctx context.Context, city string) (float64, error) {
				if city != "Sydney" {
					t.Errorf("expected city 'Sydney', got %q", city)
				}
				return 28.5, nil
			},
		}

		if handler.NodeType() != "integration" {
			t.Errorf("expected node type 'integration', got %q", handler.NodeType())
		}

		ec := &ExecutionContext{
			Ctx:        context.Background(),
			StepNumber: 3,
			FormData:   FormData{City: "Sydney"},
			State:      make(map[string]interface{}),
		}

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ec.State["temperature"].(float64) != 28.5 {
			t.Errorf("expected temperature 28.5 in context, got %v", ec.State["temperature"])
		}

		if step.Output.APIResponse == nil {
			t.Fatal("expected APIResponse in output")
		}
		if step.Output.APIResponse.Data.(map[string]any)["temperature"].(float64) != 28.5 {
			t.Errorf("unexpected temperature in response")
		}
	})

	t.Run("with weather function error", func(t *testing.T) {
		handler := &WeatherHandler{
			weatherFn: func(ctx context.Context, city string) (float64, error) {
				return 0, errors.New("API error")
			},
		}

		ec := &ExecutionContext{
			Ctx:        context.Background(),
			StepNumber: 3,
			FormData:   FormData{City: "Sydney"},
			State:      make(map[string]interface{}),
		}

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should fall back to default temperature
		if ec.State["temperature"].(float64) != 25.0 {
			t.Errorf("expected default temperature 25.0, got %v", ec.State["temperature"])
		}
		if step.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", step.Status)
		}
	})

	t.Run("without weather function", func(t *testing.T) {
		handler := &WeatherHandler{}

		ec := &ExecutionContext{
			Ctx:        context.Background(),
			StepNumber: 3,
			FormData:   FormData{City: "Melbourne"},
			State:      make(map[string]interface{}),
		}

		_, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ec.State["temperature"].(float64) != 25.0 {
			t.Errorf("expected default temperature 25.0, got %v", ec.State["temperature"])
		}
	})

	t.Run("missing city in form data", func(t *testing.T) {
		handler := &WeatherHandler{}

		ec := &ExecutionContext{
			Ctx:        context.Background(),
			StepNumber: 3,
			FormData:   FormData{}, // No city
			State:      make(map[string]interface{}),
		}

		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Error("expected error when city is missing")
		}
	})
}

func TestConditionHandler(t *testing.T) {
	handler := &ConditionHandler{}

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
			nodeJSON := fmt.Sprintf(`{"operator": "%s", "threshold": %.1f}`, tt.operator, tt.threshold)
			node := &Node{}
			if err := json.Unmarshal([]byte(nodeJSON), &node.Metadata); err != nil {
				t.Fatalf("failed to unmarshal node metadata: %v", err)
			}

			ec := &ExecutionContext{
				StepNumber: 4,
				State:      map[string]interface{}{"temperature": tt.temperature},
			}

			step, err := handler.Execute(ec, node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if step.Output.ConditionResult == nil {
				t.Fatal("expected ConditionResult in output")
			}
			if step.Output.ConditionResult.Result != tt.wantResult {
				t.Errorf("expected result %v, got %v", tt.wantResult, step.Output.ConditionResult.Result)
			}
		})
	}
}

func TestEmailHandler(t *testing.T) {
	handler := &EmailHandler{}

	if handler.NodeType() != "email" {
		t.Errorf("expected node type 'email', got %q", handler.NodeType())
	}

	t.Run("sends email to user from form data", func(t *testing.T) {
		// Subject from node metadata (workflow config), recipient from FormData (user input)
		nodeJSON := `{"subject": "Weather Alert"}`
		node := &Node{}
		if err := json.Unmarshal([]byte(nodeJSON), &node.Metadata); err != nil {
			t.Fatalf("failed to unmarshal node metadata: %v", err)
		}

		ec := &ExecutionContext{
			StepNumber: 5,
			FormData: FormData{
				Name:  "Alice",
				Email: "alice@example.com",
				City:  "Sydney",
			},
			State: map[string]interface{}{"temperature": 30.0},
		}

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if step.NodeType != "email" {
			t.Errorf("expected step type 'email', got %q", step.NodeType)
		}
		if step.Output.EmailContent == nil {
			t.Fatal("expected EmailContent in output")
		}
		if step.Output.EmailContent.To != "alice@example.com" {
			t.Errorf("expected email to 'alice@example.com', got %q", step.Output.EmailContent.To)
		}
		if step.Output.EmailContent.Subject != "Weather Alert" {
			t.Errorf("expected subject 'Weather Alert', got %q", step.Output.EmailContent.Subject)
		}
		// Body should include user's name, city, and temperature
		expectedBody := "Hi Alice, weather alert for Sydney! Temperature is 30.0°C!"
		if step.Output.EmailContent.Body != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, step.Output.EmailContent.Body)
		}
	})

	t.Run("uses default subject when not in metadata", func(t *testing.T) {
		node := &Node{} // No metadata

		ec := &ExecutionContext{
			StepNumber: 5,
			FormData: FormData{
				Name:  "Bob",
				Email: "bob@example.com",
				City:  "Melbourne",
			},
			State: map[string]interface{}{"temperature": 22.0},
		}

		step, err := handler.Execute(ec, node)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if step.Output.EmailContent.Subject != "Weather Alert" {
			t.Errorf("expected default subject 'Weather Alert', got %q", step.Output.EmailContent.Subject)
		}
	})

	t.Run("missing email in form data", func(t *testing.T) {
		node := &Node{}

		ec := &ExecutionContext{
			StepNumber: 5,
			FormData:   FormData{Name: "Alice"}, // No email
			State:      map[string]interface{}{"temperature": 30.0},
		}

		_, err := handler.Execute(ec, node)
		if err == nil {
			t.Error("expected error when email is missing")
		}
	})
}

func TestHandlerRegistry(t *testing.T) {
	registry := NewHandlerRegistry()

	registry.Register(&StartHandler{})
	registry.Register(&EndHandler{})

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
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry(nil)

	expectedTypes := []string{"start", "end", "form", "integration", "condition", "email"}

	for _, nodeType := range expectedTypes {
		if _, ok := registry.Get(nodeType); !ok {
			t.Errorf("expected default registry to have handler for %q", nodeType)
		}
	}
}
