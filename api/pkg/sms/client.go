package sms

import (
	"context"
	"fmt"
	"log/slog"
)

// Client defines the interface for sending SMS messages
type Client interface {
	Send(ctx context.Context, phone, message string) error
}

// MockClient is a mock implementation that logs instead of sending
type MockClient struct{}

// NewMockClient creates a new MockClient
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Send logs the SMS details instead of actually sending
func (c *MockClient) Send(ctx context.Context, phone, message string) error {
	slog.Info("mock SMS sent",
		"phone", phone,
		"message_length", len(message),
	)
	return nil
}

// TwilioClient sends SMS via Twilio API (placeholder for real implementation)
type TwilioClient struct {
	accountSID string
	authToken  string
	fromNumber string
}

// TwilioConfig holds Twilio configuration
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// NewTwilioClient creates a new Twilio client
func NewTwilioClient(cfg TwilioConfig) *TwilioClient {
	return &TwilioClient{
		accountSID: cfg.AccountSID,
		authToken:  cfg.AuthToken,
		fromNumber: cfg.FromNumber,
	}
}

// Send sends an SMS via Twilio API
func (c *TwilioClient) Send(ctx context.Context, phone, message string) error {
	// In production, this would use the Twilio SDK
	// For now, return an error indicating it's not configured
	if c.accountSID == "" {
		return fmt.Errorf("Twilio not configured")
	}

	// TODO: Implement actual Twilio API call
	// client := twilio.NewRestClient(c.accountSID, c.authToken)
	// params := &openapi.CreateMessageParams{}
	// params.SetTo(phone)
	// params.SetFrom(c.fromNumber)
	// params.SetBody(message)
	// _, err := client.Api.CreateMessage(params)

	slog.Info("Twilio SMS would be sent",
		"to", phone,
		"from", c.fromNumber,
		"message_length", len(message),
	)
	return nil
}
