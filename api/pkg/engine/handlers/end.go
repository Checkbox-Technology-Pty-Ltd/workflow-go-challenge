package handlers

import (
	"time"

	"workflow-code-test/api/pkg/engine"
)

// EndHandler handles end nodes
type EndHandler struct{}

// NewEndHandler creates a new EndHandler
func NewEndHandler() *EndHandler {
	return &EndHandler{}
}

func (h *EndHandler) NodeType() string { return "end" }

func (h *EndHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "end",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   EndNodeDuration,
		Output:     map[string]interface{}{"message": "Workflow completed"},
		Timestamp:  time.Now().Format(time.RFC3339),
	}, nil
}
