package workflow

import (
	"context"
	"time"
)

// ExecutionContext holds state that is passed between node handlers during execution
type ExecutionContext struct {
	Ctx         context.Context
	FormData    FormData
	Temperature float64
	StartTime   time.Time
	StepNumber  int
}

// NodeHandler defines the interface for handling different node types
type NodeHandler interface {
	// Execute processes the node and returns an ExecutionStep
	// The handler may modify the ExecutionContext (e.g., setting temperature)
	Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error)

	// NodeType returns the type of node this handler processes
	NodeType() string
}

// HandlerRegistry maps node types to their handlers
type HandlerRegistry struct {
	handlers map[string]NodeHandler
}

// NewHandlerRegistry creates a new empty handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]NodeHandler),
	}
}

// Register adds a handler for a specific node type
func (r *HandlerRegistry) Register(handler NodeHandler) {
	r.handlers[handler.NodeType()] = handler
}

// Get returns the handler for a given node type
func (r *HandlerRegistry) Get(nodeType string) (NodeHandler, bool) {
	h, ok := r.handlers[nodeType]
	return h, ok
}

// DefaultRegistry creates a registry with all standard handlers
func DefaultRegistry(weatherFn func(ctx context.Context, city string) (float64, error)) *HandlerRegistry {
	registry := NewHandlerRegistry()
	registry.Register(&StartHandler{})
	registry.Register(&EndHandler{})
	registry.Register(&FormHandler{})
	registry.Register(&WeatherHandler{weatherFn: weatherFn})
	registry.Register(&ConditionHandler{})
	registry.Register(&EmailHandler{})
	return registry
}
