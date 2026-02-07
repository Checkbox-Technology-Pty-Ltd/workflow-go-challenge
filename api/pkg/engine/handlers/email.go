package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-code-test/api/pkg/engine"
)

// EmailFunc is a function type for sending emails
type EmailFunc func(ctx context.Context, to, subject, body string) error

// EmailHandler handles email nodes.
// It reads recipient from form data, subject from metadata,
// and builds a body using user data and temperature.
type EmailHandler struct {
	emailFn EmailFunc
}

// NewEmailHandler creates a new EmailHandler with the given email function
func NewEmailHandler(emailFn EmailFunc) *EmailHandler {
	return &EmailHandler{emailFn: emailFn}
}

func (h *EmailHandler) NodeType() string { return "email" }

func (h *EmailHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	startTime := time.Now()

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
		metadata.Subject = DefaultEmailSubject
	}

	// Build email body using user data
	name := ec.GetString("form.name")
	city := ec.GetString("form.city")
	temperature := ec.GetFloat("weather.temperature")

	body := fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", name, city, temperature)

	// Send the email via injected client
	if h.emailFn != nil {
		if err := h.emailFn(ec.Ctx, to, metadata.Subject, body); err != nil {
			return engine.ExecutionStep{}, fmt.Errorf("failed to send email: %w", err)
		}
	}

	timestamp := time.Now().Format(time.RFC3339)
	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "email",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
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
