package workflow

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestNode_ToResponse(t *testing.T) {
	workflowID := uuid.New()
	label := "Start"
	desc := "Begin workflow"

	node := Node{
		WorkflowID:  workflowID,
		NodeID:      "start",
		NodeType:    "start",
		Label:       &label,
		Description: &desc,
		XPos:        100.5,
		YPos:        200.5,
		Metadata:    json.RawMessage(`{"hasHandles":{"source":true,"target":false}}`),
	}

	resp := node.ToResponse()

	if resp.ID != "start" {
		t.Errorf("expected ID 'start', got %q", resp.ID)
	}
	if resp.Type != "start" {
		t.Errorf("expected Type 'start', got %q", resp.Type)
	}
	if resp.Position.X != 100.5 {
		t.Errorf("expected Position.X 100.5, got %f", resp.Position.X)
	}
	if resp.Position.Y != 200.5 {
		t.Errorf("expected Position.Y 200.5, got %f", resp.Position.Y)
	}
	if resp.Data.Label != "Start" {
		t.Errorf("expected Data.Label 'Start', got %q", resp.Data.Label)
	}
	if resp.Data.Description != "Begin workflow" {
		t.Errorf("expected Data.Description 'Begin workflow', got %q", resp.Data.Description)
	}
	if string(resp.Data.Metadata) != `{"hasHandles":{"source":true,"target":false}}` {
		t.Errorf("unexpected metadata: %s", resp.Data.Metadata)
	}
}

func TestNode_ToResponse_NilFields(t *testing.T) {
	node := Node{
		WorkflowID:  uuid.New(),
		NodeID:      "test",
		NodeType:    "action",
		Label:       nil,
		Description: nil,
		XPos:        0,
		YPos:        0,
		Metadata:    nil,
	}

	resp := node.ToResponse()

	if resp.Data.Label != "" {
		t.Errorf("expected empty label for nil, got %q", resp.Data.Label)
	}
	if resp.Data.Description != "" {
		t.Errorf("expected empty description for nil, got %q", resp.Data.Description)
	}
}

func TestEdge_ToResponse(t *testing.T) {
	workflowID := uuid.New()
	sourceHandle := "true"

	edge := Edge{
		WorkflowID:   workflowID,
		EdgeID:       "e1",
		SourceID:     "condition",
		TargetID:     "email",
		SourceHandle: &sourceHandle,
		EdgeProps: json.RawMessage(`{
			"type": "smoothstep",
			"animated": true,
			"style": {"stroke": "#10b981", "strokeWidth": 3},
			"label": "Condition Met",
			"labelStyle": {"fill": "#10b981"}
		}`),
	}

	resp := edge.ToResponse()

	if resp.ID != "e1" {
		t.Errorf("expected ID 'e1', got %q", resp.ID)
	}
	if resp.Source != "condition" {
		t.Errorf("expected Source 'condition', got %q", resp.Source)
	}
	if resp.Target != "email" {
		t.Errorf("expected Target 'email', got %q", resp.Target)
	}
	if resp.SourceHandle != "true" {
		t.Errorf("expected SourceHandle 'true', got %q", resp.SourceHandle)
	}
	if resp.Type != "smoothstep" {
		t.Errorf("expected Type 'smoothstep', got %q", resp.Type)
	}
	if !resp.Animated {
		t.Error("expected Animated true, got false")
	}
	if resp.Label != "Condition Met" {
		t.Errorf("expected Label 'Condition Met', got %q", resp.Label)
	}
}

func TestEdge_ToResponse_NilSourceHandle(t *testing.T) {
	edge := Edge{
		WorkflowID:   uuid.New(),
		EdgeID:       "e2",
		SourceID:     "start",
		TargetID:     "form",
		SourceHandle: nil,
		EdgeProps:    json.RawMessage(`{"type": "smoothstep", "animated": false}`),
	}

	resp := edge.ToResponse()

	if resp.SourceHandle != "" {
		t.Errorf("expected empty SourceHandle for nil, got %q", resp.SourceHandle)
	}
	if resp.Animated {
		t.Error("expected Animated false, got true")
	}
}

func TestEdge_ToResponse_EmptyEdgeProps(t *testing.T) {
	edge := Edge{
		WorkflowID:   uuid.New(),
		EdgeID:       "e3",
		SourceID:     "a",
		TargetID:     "b",
		SourceHandle: nil,
		EdgeProps:    nil,
	}

	resp := edge.ToResponse()

	if resp.Type != "" {
		t.Errorf("expected empty Type, got %q", resp.Type)
	}
	if resp.Animated {
		t.Error("expected Animated false for empty props")
	}
}

func TestEdge_ToResponse_InvalidJSON(t *testing.T) {
	edge := Edge{
		WorkflowID:   uuid.New(),
		EdgeID:       "e4",
		SourceID:     "x",
		TargetID:     "y",
		SourceHandle: nil,
		EdgeProps:    json.RawMessage(`{invalid json`),
	}

	// Should not panic, just return empty values
	resp := edge.ToResponse()

	if resp.ID != "e4" {
		t.Errorf("expected ID 'e4', got %q", resp.ID)
	}
	if resp.Type != "" {
		t.Errorf("expected empty Type for invalid JSON, got %q", resp.Type)
	}
}
