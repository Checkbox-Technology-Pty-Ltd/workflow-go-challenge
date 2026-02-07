package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-code-test/api/pkg/engine"
)

// SMSFunc is a function type for sending SMS messages
type SMSFunc func(ctx context.Context, phone, message string) error

// SMSHandler handles SMS notification nodes.
// It reads recipient phone from form data, message template from metadata,
// and builds a message using user data and temperature.
type SMSHandler struct {
	smsFn SMSFunc
}

// NewSMSHandler creates a new SMSHandler with the given SMS function
func NewSMSHandler(smsFn SMSFunc) *SMSHandler {
	return &SMSHandler{smsFn: smsFn}
}

func (h *SMSHandler) NodeType() string { return "sms" }

func (h *SMSHandler) Execute(ec *engine.ExecutionContext, node *engine.Node) (engine.ExecutionStep, error) {
	startTime := time.Now()

	// Get recipient phone from form data
	phone := ec.GetString("form.phone")
	if phone == "" {
		return engine.ExecutionStep{}, fmt.Errorf("recipient phone not provided in form data")
	}

	// Parse metadata for message template
	var metadata struct {
		Template string `json:"template"`
	}

	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return engine.ExecutionStep{}, fmt.Errorf("failed to parse SMS node metadata: %w", err)
		}
	}

	// Build message using user data
	name := ec.GetString("form.name")
	city := ec.GetString("form.city")
	temperature := ec.GetFloat("weather.temperature")

	// Use template from metadata or default
	message := metadata.Template
	if message == "" {
		message = fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", name, city, temperature)
	}

	// Send the SMS via injected client
	if h.smsFn == nil {
		return engine.ExecutionStep{}, fmt.Errorf("SMS client not configured")
	}

	if err := h.smsFn(ec.Ctx, phone, message); err != nil {
		return engine.ExecutionStep{}, fmt.Errorf("failed to send SMS: %w", err)
	}

	timestamp := time.Now().Format(time.RFC3339)
	duration := time.Since(startTime).Milliseconds()

	return engine.ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "sms",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   duration,
		Output: map[string]interface{}{
			"message": "SMS sent",
			"smsContent": map[string]interface{}{
				"to":        phone,
				"message":   message,
				"timestamp": timestamp,
			},
		},
		Timestamp: timestamp,
	}, nil
}
