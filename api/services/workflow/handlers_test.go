package workflow

import (
	"context"
	"errors"
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
}

func TestWeatherHandler(t *testing.T) {
	t.Run("with weather function", func(t *testing.T) {
		handler := &WeatherHandler{
			weatherFn: func(ctx context.Context, city string) (float64, error) {
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
		}

		step, err := handler.Execute(ec, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ec.Temperature != 28.5 {
			t.Errorf("expected temperature 28.5 in context, got %v", ec.Temperature)
		}

		if step.Output.APIResponse == nil {
			t.Fatal("expected APIResponse in output")
		}
		if step.Output.APIResponse.Data.(map[string]any)["temperature"] != 28.5 {
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
		}

		step, err := handler.Execute(ec, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should fall back to default temperature
		if ec.Temperature != 25.0 {
			t.Errorf("expected default temperature 25.0, got %v", ec.Temperature)
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
			FormData:   FormData{City: "Sydney"},
		}

		_, err := handler.Execute(ec, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ec.Temperature != 25.0 {
			t.Errorf("expected default temperature 25.0, got %v", ec.Temperature)
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
			ec := &ExecutionContext{
				StepNumber:  4,
				Temperature: tt.temperature,
				FormData: FormData{
					Operator:  tt.operator,
					Threshold: tt.threshold,
				},
			}

			step, err := handler.Execute(ec, nil)
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

	ec := &ExecutionContext{
		StepNumber:  5,
		Temperature: 30.0,
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

	if step.NodeType != "email" {
		t.Errorf("expected step type 'email', got %q", step.NodeType)
	}
	if step.Output.EmailContent == nil {
		t.Fatal("expected EmailContent in output")
	}
	if step.Output.EmailContent.To != "alice@example.com" {
		t.Errorf("expected email to 'alice@example.com', got %q", step.Output.EmailContent.To)
	}
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
