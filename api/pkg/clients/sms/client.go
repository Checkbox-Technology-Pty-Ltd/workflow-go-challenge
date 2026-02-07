package sms

import (
	"context"
	"log/slog"
)

// Message represents an SMS to be sent.
type Message struct {
	To   string
	Body string
}

// Result holds the outcome of a send attempt.
type Result struct {
	DeliveryStatus string
	Sent           bool
}

// Client defines the interface for sending SMS messages.
// Swap between a stub (dev/testing) and a real provider (e.g. Twilio).
type Client interface {
	Send(ctx context.Context, msg Message) (*Result, error)
}

// StubClient simulates sending SMS by logging them.
type StubClient struct{}

func NewStubClient() *StubClient {
	return &StubClient{}
}

func (c *StubClient) Send(_ context.Context, msg Message) (*Result, error) {
	slog.Info("sending sms (stub)", "to", msg.To, "body", msg.Body)
	return &Result{
		DeliveryStatus: "sent",
		Sent:           true,
	}, nil
}
