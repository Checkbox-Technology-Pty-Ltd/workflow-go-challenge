package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"workflow-code-test/api/pkg/engine"
)

// EmailHandler handles email nodes.
// It reads recipient from form data, subject from metadata,
// and builds a body using user data and temperature.
type EmailHandler struct{}

// NewEmailHandler creates a new EmailHandler
func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (h *EmailHandler) NodeType() string { return "email" }

func (h *EmailHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	// Get recipient from form data
	to := ec.GetString("form.email")
	if to == "" {
		return engine.ExecutionStep{}, fmt.Errorf("recipient email not provided in form data")
	}

	// Parse metadata for subject
	var metadata struct {
		Subject string `json:"subject"`
	}

	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return engine.ExecutionStep{}, fmt.Errorf("failed to parse email node metadata: %w", err)
		}
	}

	if metadata.Subject == "" {
		metadata.Subject = "Weather Alert"
	}

	// Build email body using user data
	name := ec.GetString("form.name")
	city := ec.GetString("form.city")
	temperature := ec.GetFloat("weather.temperature")

	body := fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", name, city, temperature)

	timestamp := time.Now().Format(time.RFC3339)

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "email",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   50,
		Output: map[string]interface{}{
			"message": "Alert email sent",
			"emailContent": map[string]interface{}{
				"to":        to,
				"subject":   metadata.Subject,
				"body":      body,
				"timestamp": timestamp,
			},
		},
		Timestamp: timestamp,
	}, nil
}
