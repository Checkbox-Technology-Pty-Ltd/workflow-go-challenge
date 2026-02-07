package nodes

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFormNode_Execute(t *testing.T) {
	t.Parallel()
	base := BaseFields{
		ID:       "form",
		NodeType: "form",
		Metadata: json.RawMessage(`{"inputFields":["name","city"],"outputVariables":["name","city"]}`),
	}

	tests := []struct {
		name      string
		variables map[string]any
		wantErr   string
		checkOut  func(t *testing.T, result *ExecutionResult)
	}{
		{
			name:      "all fields present",
			variables: map[string]any{"name": "Alice", "city": "Sydney"},
			checkOut: func(t *testing.T, r *ExecutionResult) {
				if r.Status != "completed" {
					t.Errorf("expected completed, got %q", r.Status)
				}
				if r.Output["name"] != "Alice" {
					t.Errorf("expected name=Alice, got %v", r.Output["name"])
				}
				if r.Output["city"] != "Sydney" {
					t.Errorf("expected city=Sydney, got %v", r.Output["city"])
				}
			},
		},
		{
			name:      "missing required field",
			variables: map[string]any{"name": "Alice"},
			wantErr:   "missing required form field: city",
		},
		{
			name:      "empty variables",
			variables: map[string]any{},
			wantErr:   "missing required form field: name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			node, err := NewFormNode(base)
			if err != nil {
				t.Fatalf("failed to create form node: %v", err)
			}

			nCtx := &NodeContext{Variables: tt.variables}
			result, err := node.Execute(context.Background(), nCtx)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkOut != nil {
				tt.checkOut(t, result)
			}
		})
	}
}

func TestConditionNode_Execute(t *testing.T) {
	t.Parallel()
	base := BaseFields{
		ID:       "condition",
		NodeType: "condition",
		Metadata: json.RawMessage(`{"conditionVariable":"temperature","conditionExpression":"temperature > threshold","outputVariables":["conditionMet"]}`),
	}

	tests := []struct {
		name      string
		variables map[string]any
		wantErr   string
		wantMet   bool
		wantBranch string
	}{
		{
			name:       "greater_than met",
			variables:  map[string]any{"temperature": 30.0, "operator": "greater_than", "threshold": 25.0},
			wantMet:    true,
			wantBranch: "true",
		},
		{
			name:       "greater_than not met",
			variables:  map[string]any{"temperature": 20.0, "operator": "greater_than", "threshold": 25.0},
			wantMet:    false,
			wantBranch: "false",
		},
		{
			name:       "less_than met",
			variables:  map[string]any{"temperature": 10.0, "operator": "less_than", "threshold": 25.0},
			wantMet:    true,
			wantBranch: "true",
		},
		{
			name:       "equal_to met",
			variables:  map[string]any{"temperature": 25.0, "operator": "equal_to", "threshold": 25.0},
			wantMet:    true,
			wantBranch: "true",
		},
		{
			name:       "greater_than_or_equal at boundary",
			variables:  map[string]any{"temperature": 25.0, "operator": "greater_than_or_equal", "threshold": 25.0},
			wantMet:    true,
			wantBranch: "true",
		},
		{
			name:       "less_than_or_equal at boundary",
			variables:  map[string]any{"temperature": 25.0, "operator": "less_than_or_equal", "threshold": 25.0},
			wantMet:    true,
			wantBranch: "true",
		},
		{
			name:      "unsupported operator",
			variables: map[string]any{"temperature": 30.0, "operator": "not_equal", "threshold": 25.0},
			wantErr:   "unsupported operator: not_equal",
		},
		{
			name:      "missing condition variable",
			variables: map[string]any{"operator": "greater_than", "threshold": 25.0},
			wantErr:   "missing or invalid variable: temperature",
		},
		{
			name:       "defaults to greater_than with threshold 25",
			variables:  map[string]any{"temperature": 30.0},
			wantMet:    true,
			wantBranch: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			node, err := NewConditionNode(base)
			if err != nil {
				t.Fatalf("failed to create condition node: %v", err)
			}

			nCtx := &NodeContext{Variables: tt.variables}
			result, err := node.Execute(context.Background(), nCtx)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Branch != tt.wantBranch {
				t.Errorf("expected branch %q, got %q", tt.wantBranch, result.Branch)
			}
			met, ok := result.Output["conditionMet"].(bool)
			if !ok || met != tt.wantMet {
				t.Errorf("expected conditionMet=%v, got %v", tt.wantMet, result.Output["conditionMet"])
			}
		})
	}
}

func TestConditionNode_CustomVariable(t *testing.T) {
	t.Parallel()
	base := BaseFields{
		ID:       "condition",
		NodeType: "condition",
		Metadata: json.RawMessage(`{"conditionVariable":"discharge","outputVariables":["conditionMet"]}`),
	}

	node, err := NewConditionNode(base)
	if err != nil {
		t.Fatalf("failed to create condition node: %v", err)
	}

	nCtx := &NodeContext{Variables: map[string]any{
		"discharge": 500.0,
		"operator":  "greater_than",
		"threshold": 100.0,
	}}
	result, err := node.Execute(context.Background(), nCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Branch != "true" {
		t.Errorf("expected branch 'true', got %q", result.Branch)
	}
}

func TestSentinelNode_Execute(t *testing.T) {
	t.Parallel()
	node := NewSentinelNode(BaseFields{ID: "start", NodeType: "start"})

	result, err := node.Execute(context.Background(), &NodeContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected completed, got %q", result.Status)
	}
}

func TestNodeFactory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		nodeType string
		metadata string
		wantErr  bool
	}{
		{name: "start", nodeType: "start", metadata: `{}`},
		{name: "end", nodeType: "end", metadata: `{}`},
		{name: "form", nodeType: "form", metadata: `{"inputFields":["name"]}`},
		{name: "condition", nodeType: "condition", metadata: `{"conditionVariable":"temp"}`},
		{name: "unknown type", nodeType: "foobar", metadata: `{}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base := BaseFields{
				ID:       tt.name,
				NodeType: tt.nodeType,
				Metadata: json.RawMessage(tt.metadata),
			}
			_, err := New(base, Deps{})

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    float64
		wantOK  bool
	}{
		{name: "float64", input: 42.5, want: 42.5, wantOK: true},
		{name: "float32", input: float32(42.5), want: 42.5, wantOK: true},
		{name: "int", input: 42, want: 42.0, wantOK: true},
		{name: "json.Number", input: json.Number("42.5"), want: 42.5, wantOK: true},
		{name: "string fails", input: "42.5", want: 0, wantOK: false},
		{name: "nil fails", input: nil, want: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tt.input)
			if ok != tt.wantOK {
				t.Errorf("toFloat64(%v): ok=%v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("toFloat64(%v)=%v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
