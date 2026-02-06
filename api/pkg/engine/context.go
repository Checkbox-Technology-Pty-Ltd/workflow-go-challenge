package engine

import (
	"context"
	"time"
)

// ExecutionContext holds state that is passed between node handlers during execution.
// It provides a generic State map for sharing data between handlers.
type ExecutionContext struct {
	// Ctx is the Go context for cancellation and timeouts
	Ctx context.Context

	// State holds arbitrary data shared between handlers
	// Handlers can read and write to this map to pass data downstream
	State map[string]interface{}

	// StartTime is when the workflow execution began
	StartTime time.Time

	// StepNumber is the current step in the execution sequence
	StepNumber int
}

// NewExecutionContext creates a new ExecutionContext with initialized State
func NewExecutionContext(ctx context.Context) *ExecutionContext {
	return &ExecutionContext{
		Ctx:       ctx,
		State:     make(map[string]interface{}),
		StartTime: time.Now(),
	}
}

// Get retrieves a value from State, returning nil if not found
func (ec *ExecutionContext) Get(key string) interface{} {
	if ec.State == nil {
		return nil
	}
	return ec.State[key]
}

// Set stores a value in State
func (ec *ExecutionContext) Set(key string, value interface{}) {
	if ec.State == nil {
		ec.State = make(map[string]interface{})
	}
	ec.State[key] = value
}

// GetString retrieves a string value from State
func (ec *ExecutionContext) GetString(key string) string {
	if v, ok := ec.State[key].(string); ok {
		return v
	}
	return ""
}

// GetFloat retrieves a float64 value from State
func (ec *ExecutionContext) GetFloat(key string) float64 {
	if v, ok := ec.State[key].(float64); ok {
		return v
	}
	return 0
}
