package handlers

import (
	"time"

	"workflow-code-test/api/pkg/engine"
)

// StartHandler handles start nodes
type StartHandler struct{}

// NewStartHandler creates a new StartHandler
func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

func (h *StartHandler) NodeType() string { return "start" }

func (h *StartHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "start",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   10,
		Output:     map[string]interface{}{"message": "Workflow started"},
		Timestamp:  ec.StartTime.Format(time.RFC3339),
	}, nil
}
