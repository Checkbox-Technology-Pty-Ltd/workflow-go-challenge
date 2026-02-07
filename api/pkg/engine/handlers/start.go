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
	startTime := time.Now()
	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "start",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
		Output:     map[string]interface{}{"message": "Workflow started"},
		Timestamp:  startTime.Format(time.RFC3339),
	}, nil
}
