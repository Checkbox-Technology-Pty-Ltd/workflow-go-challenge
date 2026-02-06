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
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// EdgeResponse represents an edge in the API response
type EdgeResponse struct {
	ID           string          `json:"id"`
	Source       string          `json:"source"`
	Target       string          `json:"target"`
	Type         string          `json:"type"`
	SourceHandle string          `json:"sourceHandle,omitempty"`
	Animated     bool            `json:"animated"`
	Style        json.RawMessage `json:"style,omitempty" swaggertype:"object"`
	Label        string          `json:"label,omitempty"`
	LabelStyle   json.RawMessage `json:"labelStyle,omitempty" swaggertype:"object"`
}

// =============================================================================
// Execution Request/Response Models
// =============================================================================

// ExecuteWorkflowRequest is the request body for POST /workflows/{id}/execute
type ExecuteWorkflowRequest struct {
	FormData  FormData  `json:"formData"`
	Condition Condition `json:"condition"`
}

// FormData contains user input values
type FormData struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	City      string  `json:"city"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

// Condition contains the condition parameters
type Condition struct {
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

// ExecutionResponse is the response for workflow execution
type ExecutionResponse struct {
	ExecutionID   string          `json:"executionId"`
	Status        string          `json:"status"`
	StartTime     string          `json:"startTime"`
	EndTime       string          `json:"endTime"`
	TotalDuration int64           `json:"totalDuration,omitempty"`
	Steps         []ExecutionStep `json:"steps"`
}

// ExecutionSummary is a summary of an execution for the history list
type ExecutionSummary struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	Status     string `json:"status"`
	ExecutedAt string `json:"executedAt"`
}

// ExecutionsListResponse is the response for GET /workflows/{id}/executions
type ExecutionsListResponse struct {
	Executions []ExecutionSummary `json:"executions"`
}

// ExecutionStep represents a single step in the execution trace
type ExecutionStep struct {
	StepNumber int            `json:"stepNumber"`
	NodeType   string         `json:"nodeType"`
	Status     string         `json:"status"`
	Duration   int64          `json:"duration"`
	Output     StepOutput     `json:"output"`
	Timestamp  string         `json:"timestamp"`
	Error      string         `json:"error,omitempty"`
}

// StepOutput contains the output of an execution step
type StepOutput struct {
	Message         string           `json:"message"`
	Details         map[string]any   `json:"details,omitempty"`
	FormData        *FormData        `json:"formData,omitempty"`
	APIResponse     *APIResponse     `json:"apiResponse,omitempty"`
	ConditionResult *ConditionResult `json:"conditionResult,omitempty"`
	EmailContent    *EmailContent    `json:"emailContent,omitempty"`
}

// APIResponse contains details of an API call
type APIResponse struct {
	Endpoint   string `json:"endpoint"`
	Method     string `json:"method"`
	StatusCode int    `json:"statusCode"`
	Data       any    `json:"data"`
}

// ConditionResult contains the result of a condition evaluation
type ConditionResult struct {
	Expression  string  `json:"expression"`
	Result      bool    `json:"result"`
	Temperature float64 `json:"temperature"`
	Operator    string  `json:"operator"`
	Threshold   float64 `json:"threshold"`
}

// EmailContent contains email details
type EmailContent struct {
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Timestamp string `json:"timestamp,omitempty"`
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
