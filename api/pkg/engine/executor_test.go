package engine

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestGraph(t *testing.T) {
	t.Run("valid graph", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "end-1"})

		if err := graph.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if graph.GetStartNode() != "start-1" {
			t.Errorf("expected start node 'start-1', got %q", graph.GetStartNode())
		}

		node, ok := graph.GetNode("start-1")
		if !ok {
			t.Fatal("expected to find start-1 node")
		}
		if node.Type != "start" {
			t.Errorf("expected node type 'start', got %q", node.Type)
		}
	})

	t.Run("no start node", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "end-1", Type: "end"})

		if err := graph.Validate(); err == nil {
			t.Error("expected error for graph with no start node")
		}
	})

	t.Run("get outgoing edges", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "end-1"})

		edges := graph.GetOutgoingEdges("start-1")
		if len(edges) != 1 {
			t.Errorf("expected 1 edge, got %d", len(edges))
		}
		if edges[0].TargetID != "end-1" {
			t.Errorf("expected target 'end-1', got %q", edges[0].TargetID)
		}
	})
}

func TestExecutor(t *testing.T) {
	t.Run("simple workflow", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "end-1"})

		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
		registry.Register(&mockHandler{nodeType: "end", output: map[string]interface{}{"message": "ended"}})

		executor := NewExecutor(registry)
		steps, err := executor.Execute(context.Background(), graph, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(steps) != 2 {
			t.Errorf("expected 2 steps, got %d", len(steps))
		}

		if steps[0].NodeType != "start" {
			t.Errorf("expected first step to be 'start', got %q", steps[0].NodeType)
		}
		if steps[1].NodeType != "end" {
			t.Errorf("expected last step to be 'end', got %q", steps[1].NodeType)
		}
	})

	t.Run("condition true path", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		condMetadata, _ := json.Marshal(map[string]interface{}{
			"operator":  "greater_than",
			"threshold": 25.0,
		})
		graph.AddNode(&Node{ID: "condition-1", Type: "condition", Metadata: condMetadata})
		graph.AddNode(&Node{ID: "email-1", Type: "email"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})

		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "condition-1"})
		graph.AddEdge(Edge{SourceID: "condition-1", TargetID: "email-1", SourceHandle: "true"})
		graph.AddEdge(Edge{SourceID: "condition-1", TargetID: "end-1", SourceHandle: "false"})
		graph.AddEdge(Edge{SourceID: "email-1", TargetID: "end-1"})

		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
		registry.Register(&mockConditionHandler{result: true})
		registry.Register(&mockHandler{nodeType: "email", output: map[string]interface{}{"message": "email sent"}})
		registry.Register(&mockHandler{nodeType: "end", output: map[string]interface{}{"message": "ended"}})

		executor := NewExecutor(registry)
		steps, err := executor.Execute(context.Background(), graph, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have: start, condition, email, end
		if len(steps) != 4 {
			t.Errorf("expected 4 steps, got %d", len(steps))
		}

		hasEmail := false
		for _, step := range steps {
			if step.NodeType == "email" {
				hasEmail = true
			}
		}
		if !hasEmail {
			t.Error("expected email step when condition is true")
		}
	})

	t.Run("condition false path", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		condMetadata, _ := json.Marshal(map[string]interface{}{
			"operator":  "greater_than",
			"threshold": 25.0,
		})
		graph.AddNode(&Node{ID: "condition-1", Type: "condition", Metadata: condMetadata})
		graph.AddNode(&Node{ID: "email-1", Type: "email"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})

		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "condition-1"})
		graph.AddEdge(Edge{SourceID: "condition-1", TargetID: "email-1", SourceHandle: "true"})
		graph.AddEdge(Edge{SourceID: "condition-1", TargetID: "end-1", SourceHandle: "false"})
		graph.AddEdge(Edge{SourceID: "email-1", TargetID: "end-1"})

		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
		registry.Register(&mockConditionHandler{result: false})
		registry.Register(&mockHandler{nodeType: "email", output: map[string]interface{}{"message": "email sent"}})
		registry.Register(&mockHandler{nodeType: "end", output: map[string]interface{}{"message": "ended"}})

		executor := NewExecutor(registry)
		steps, err := executor.Execute(context.Background(), graph, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have: start, condition, end (no email)
		if len(steps) != 3 {
			t.Errorf("expected 3 steps, got %d", len(steps))
		}

		hasEmail := false
		for _, step := range steps {
			if step.NodeType == "email" {
				hasEmail = true
			}
		}
		if hasEmail {
			t.Error("expected no email step when condition is false")
		}
	})

	t.Run("unknown node type", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "unknown-1", Type: "unknown_type"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "unknown-1"})

		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})

		executor := NewExecutor(registry)
		_, err := executor.Execute(context.Background(), graph, nil)
		if err == nil {
			t.Error("expected error for unknown node type")
		}
	})

	t.Run("handler error", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "fail-1", Type: "fail"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "fail-1"})

		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
		registry.Register(&failingHandler{})

		executor := NewExecutor(registry)
		steps, err := executor.Execute(context.Background(), graph, nil)
		if err == nil {
			t.Error("expected error from failing handler")
		}

		// Should have start step before error
		if len(steps) < 1 {
			t.Error("expected at least start step before error")
		}
	})

	t.Run("initial state is passed to handlers", func(t *testing.T) {
		graph := NewGraph()
		graph.AddNode(&Node{ID: "start-1", Type: "start"})
		graph.AddNode(&Node{ID: "check-1", Type: "check"})
		graph.AddNode(&Node{ID: "end-1", Type: "end"})
		graph.AddEdge(Edge{SourceID: "start-1", TargetID: "check-1"})
		graph.AddEdge(Edge{SourceID: "check-1", TargetID: "end-1"})

		var receivedValue string
		registry := NewRegistry()
		registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
		registry.Register(&stateCheckHandler{
			key:    "test.value",
			onExec: func(val string) { receivedValue = val },
		})
		registry.Register(&mockHandler{nodeType: "end", output: map[string]interface{}{"message": "ended"}})

		executor := NewExecutor(registry)
		initialState := map[string]interface{}{
			"test.value": "hello",
		}
		_, err := executor.Execute(context.Background(), graph, initialState)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedValue != "hello" {
			t.Errorf("expected 'hello', got %q", receivedValue)
		}
	})
}

// Mock handlers for testing

type mockHandler struct {
	nodeType string
	output   map[string]interface{}
}

func (h *mockHandler) NodeType() string { return h.nodeType }
func (h *mockHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   h.nodeType,
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   10,
		Output:     h.output,
	}, nil
}

type mockConditionHandler struct {
	result bool
}

func (h *mockConditionHandler) NodeType() string { return "condition" }
func (h *mockConditionHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "condition",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   5,
		Output: map[string]interface{}{
			"conditionResult": map[string]interface{}{
				"result": h.result,
			},
		},
	}, nil
}

type failingHandler struct{}

func (h *failingHandler) NodeType() string { return "fail" }
func (h *failingHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "fail",
		NodeID:     node.ID,
	}, errors.New("handler failed")
}

type stateCheckHandler struct {
	key    string
	onExec func(string)
}

func (h *stateCheckHandler) NodeType() string { return "check" }
func (h *stateCheckHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	val := ec.GetString(h.key)
	if h.onExec != nil {
		h.onExec(val)
	}
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "check",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   5,
		Output:     map[string]interface{}{"checked": val},
	}, nil
}

func TestExecutor_ContextCancellation(t *testing.T) {
	graph := NewGraph()
	graph.AddNode(&Node{ID: "start-1", Type: "start"})
	graph.AddNode(&Node{ID: "slow-1", Type: "slow"})
	graph.AddNode(&Node{ID: "end-1", Type: "end"})
	graph.AddEdge(Edge{SourceID: "start-1", TargetID: "slow-1"})
	graph.AddEdge(Edge{SourceID: "slow-1", TargetID: "end-1"})

	registry := NewRegistry()
	registry.Register(&mockHandler{nodeType: "start", output: map[string]interface{}{"message": "started"}})
	registry.Register(&slowHandler{})
	registry.Register(&mockHandler{nodeType: "end", output: map[string]interface{}{"message": "ended"}})

	executor := NewExecutor(registry)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := executor.Execute(ctx, graph, nil)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

type slowHandler struct{}

func (h *slowHandler) NodeType() string { return "slow" }
func (h *slowHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "slow",
		NodeID:     node.ID,
		Status:     "completed",
		Duration:   1000,
		Output:     map[string]interface{}{"message": "slow operation"},
	}, nil
}
