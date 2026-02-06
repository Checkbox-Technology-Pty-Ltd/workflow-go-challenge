package handlers

import (
	"time"

	"workflow-code-test/api/pkg/engine"
)

// FormHandler handles form nodes.
// It expects form data to be pre-populated in the ExecutionContext State
// with keys like "form.name", "form.email", "form.city".
type FormHandler struct{}

// NewFormHandler creates a new FormHandler
func NewFormHandler() *FormHandler {
	return &FormHandler{}
}

func (h *FormHandler) NodeType() string { return "form" }

func (h *FormHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	startTime := time.Now()

	// Form data should already be in state (populated by the workflow service)
	// This handler just records that form input was collected

	formData := map[string]interface{}{
		"name":  ec.GetString("form.name"),
		"email": ec.GetString("form.email"),
		"city":  ec.GetString("form.city"),
	}

	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "form",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
		Output: map[string]interface{}{
			"message":  "User input collected",
			"formData": formData,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
