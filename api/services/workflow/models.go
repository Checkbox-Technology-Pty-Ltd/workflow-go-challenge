package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Database Models - map directly to PostgreSQL tables
// =============================================================================

// Workflow represents a row in the workflows table
type Workflow struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Version     int       `db:"version"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Node represents a row in the nodes table
type Node struct {
	WorkflowID  uuid.UUID       `db:"workflow_id"`
	NodeID      string          `db:"node_id"`
	NodeType    string          `db:"node_type"`
	Label       *string         `db:"label"`
	Description *string         `db:"description"`
	XPos        float64         `db:"x_pos"`
	YPos        float64         `db:"y_pos"`
	Metadata    json.RawMessage `db:"metadata"`
}

// Edge represents a row in the edges table
type Edge struct {
	WorkflowID   uuid.UUID       `db:"workflow_id"`
	EdgeID       string          `db:"edge_id"`
	SourceID     string          `db:"source_id"`
	TargetID     string          `db:"target_id"`
	SourceHandle *string         `db:"source_handle"`
	EdgeProps    json.RawMessage `db:"edge_props"`
}

// WorkflowExecution represents a row in the workflow_executions table
type WorkflowExecution struct {
	ID             uuid.UUID       `db:"id"`
	WorkflowID     *uuid.UUID      `db:"workflow_id"`
	Status         string          `db:"status"`
	ExecutedAt     time.Time       `db:"executed_at"`
	FinalContext   json.RawMessage `db:"final_context"`
	ExecutionTrace json.RawMessage `db:"execution_trace"`
}

// =============================================================================
// API Response Models - match the JSON structure expected by the frontend
// =============================================================================

// WorkflowResponse is the API response for GET /workflows/{id}
type WorkflowResponse struct {
	ID    uuid.UUID      `json:"id"`
	Nodes []NodeResponse `json:"nodes"`
	Edges []EdgeResponse `json:"edges"`
}

// NodeResponse represents a node in the API response
type NodeResponse struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Position NodePosition     `json:"position"`
	Data     NodeDataResponse `json:"data"`
}

// NodePosition represents x/y coordinates
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeDataResponse contains the node's display and configuration data
type NodeDataResponse struct {
	Label       string          `json:"label,omitempty"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// EdgeResponse represents an edge in the API response
type EdgeResponse struct {
	ID           string          `json:"id"`
	Source       string          `json:"source"`
	Target       string          `json:"target"`
	Type         string          `json:"type"`
	SourceHandle string          `json:"sourceHandle,omitempty"`
	Animated     bool            `json:"animated"`
	Style        json.RawMessage `json:"style,omitempty"`
	Label        string          `json:"label,omitempty"`
	LabelStyle   json.RawMessage `json:"labelStyle,omitempty"`
}

// =============================================================================
// Conversion functions - DB models to API response models
// =============================================================================

// ToResponse converts a Node to a NodeResponse
func (n *Node) ToResponse() NodeResponse {
	label := ""
	if n.Label != nil {
		label = *n.Label
	}
	desc := ""
	if n.Description != nil {
		desc = *n.Description
	}

	return NodeResponse{
		ID:   n.NodeID,
		Type: n.NodeType,
		Position: NodePosition{
			X: n.XPos,
			Y: n.YPos,
		},
		Data: NodeDataResponse{
			Label:       label,
			Description: desc,
			Metadata:    n.Metadata,
		},
	}
}

// ToResponse converts an Edge to an EdgeResponse
func (e *Edge) ToResponse() EdgeResponse {
	resp := EdgeResponse{
		ID:     e.EdgeID,
		Source: e.SourceID,
		Target: e.TargetID,
	}

	if e.SourceHandle != nil {
		resp.SourceHandle = *e.SourceHandle
	}

	// Parse edge_props to extract individual fields
	if len(e.EdgeProps) > 0 {
		var props struct {
			Type       string          `json:"type"`
			Animated   bool            `json:"animated"`
			Style      json.RawMessage `json:"style"`
			Label      string          `json:"label"`
			LabelStyle json.RawMessage `json:"labelStyle"`
		}
		if err := json.Unmarshal(e.EdgeProps, &props); err == nil {
			resp.Type = props.Type
			resp.Animated = props.Animated
			resp.Style = props.Style
			resp.Label = props.Label
			resp.LabelStyle = props.LabelStyle
		}
	}

	return resp
}
