package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestBuildGraph(t *testing.T) {
	workflowID := uuid.New()

	t.Run("valid workflow", func(t *testing.T) {
		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "end-1"},
		}

		graph, err := BuildGraph(nodes, edges)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if graph.startNode != "start-1" {
			t.Errorf("expected start node 'start-1', got %q", graph.startNode)
		}

		node, ok := graph.GetNode("start-1")
		if !ok {
			t.Fatal("expected to find start-1 node")
		}
		if node.NodeType != "start" {
			t.Errorf("expected node type 'start', got %q", node.NodeType)
		}
	})

	t.Run("no start node", func(t *testing.T) {
		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
		}
		edges := []Edge{}

		_, err := BuildGraph(nodes, edges)
		if err == nil {
			t.Error("expected error for workflow with no start node")
		}
	})
}

func TestWorkflowGraph_GetNextNode(t *testing.T) {
	workflowID := uuid.New()
	trueHandle := "true"
	falseHandle := "false"

	nodes := []Node{
		{WorkflowID: workflowID, NodeID: "condition-1", NodeType: "condition"},
		{WorkflowID: workflowID, NodeID: "email-1", NodeType: "email"},
		{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
	}
	edges := []Edge{
		{WorkflowID: workflowID, EdgeID: "e1", SourceID: "condition-1", TargetID: "email-1", SourceHandle: &trueHandle},
		{WorkflowID: workflowID, EdgeID: "e2", SourceID: "condition-1", TargetID: "end-1", SourceHandle: &falseHandle},
	}

	// Add a start node to pass validation
	nodes = append([]Node{{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"}}, nodes...)

	graph, _ := BuildGraph(nodes, edges)

	t.Run("condition true", func(t *testing.T) {
		result := true
		nextID, ok := graph.GetNextNode("condition-1", &result)
		if !ok {
			t.Fatal("expected to find next node")
		}
		if nextID != "email-1" {
			t.Errorf("expected 'email-1', got %q", nextID)
		}
	})

	t.Run("condition false", func(t *testing.T) {
		result := false
		nextID, ok := graph.GetNextNode("condition-1", &result)
		if !ok {
			t.Fatal("expected to find next node")
		}
		if nextID != "end-1" {
			t.Errorf("expected 'end-1', got %q", nextID)
		}
	})

	t.Run("no condition result", func(t *testing.T) {
		nextID, ok := graph.GetNextNode("condition-1", nil)
		if !ok {
			t.Fatal("expected to find next node")
		}
		// Should return first edge
		if nextID != "email-1" && nextID != "end-1" {
			t.Errorf("expected 'email-1' or 'end-1', got %q", nextID)
		}
	})

	t.Run("no outgoing edges", func(t *testing.T) {
		_, ok := graph.GetNextNode("end-1", nil)
		if ok {
			t.Error("expected no next node for end node")
		}
	})
}

func TestExecutor_Execute(t *testing.T) {
	workflowID := uuid.New()

	t.Run("simple workflow", func(t *testing.T) {
		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "end-1"},
		}

		graph, _ := BuildGraph(nodes, edges)
		registry := DefaultRegistry(nil)
		executor := NewExecutor(registry)

		steps, err := executor.Execute(context.Background(), graph, FormData{})
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
		trueHandle := "true"
		falseHandle := "false"

		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "form-1", NodeType: "form"},
			{WorkflowID: workflowID, NodeID: "weather-1", NodeType: "integration"},
			{WorkflowID: workflowID, NodeID: "condition-1", NodeType: "condition"},
			{WorkflowID: workflowID, NodeID: "email-1", NodeType: "email"},
			{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "form-1"},
			{WorkflowID: workflowID, EdgeID: "e2", SourceID: "form-1", TargetID: "weather-1"},
			{WorkflowID: workflowID, EdgeID: "e3", SourceID: "weather-1", TargetID: "condition-1"},
			{WorkflowID: workflowID, EdgeID: "e4", SourceID: "condition-1", TargetID: "email-1", SourceHandle: &trueHandle},
			{WorkflowID: workflowID, EdgeID: "e5", SourceID: "condition-1", TargetID: "end-1", SourceHandle: &falseHandle},
			{WorkflowID: workflowID, EdgeID: "e6", SourceID: "email-1", TargetID: "end-1"},
		}

		graph, _ := BuildGraph(nodes, edges)

		// Use a weather function that returns a high temperature
		weatherFn := func(ctx context.Context, city string) (float64, error) {
			return 30.0, nil
		}
		registry := DefaultRegistry(weatherFn)
		executor := NewExecutor(registry)

		formData := FormData{
			Name:      "Alice",
			Email:     "alice@example.com",
			City:      "Sydney",
			Operator:  "greater_than",
			Threshold: 25.0,
		}

		steps, err := executor.Execute(context.Background(), graph, formData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have: start, form, weather, condition, email, end
		if len(steps) != 6 {
			t.Errorf("expected 6 steps, got %d", len(steps))
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
		trueHandle := "true"
		falseHandle := "false"

		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "form-1", NodeType: "form"},
			{WorkflowID: workflowID, NodeID: "weather-1", NodeType: "integration"},
			{WorkflowID: workflowID, NodeID: "condition-1", NodeType: "condition"},
			{WorkflowID: workflowID, NodeID: "email-1", NodeType: "email"},
			{WorkflowID: workflowID, NodeID: "end-1", NodeType: "end"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "form-1"},
			{WorkflowID: workflowID, EdgeID: "e2", SourceID: "form-1", TargetID: "weather-1"},
			{WorkflowID: workflowID, EdgeID: "e3", SourceID: "weather-1", TargetID: "condition-1"},
			{WorkflowID: workflowID, EdgeID: "e4", SourceID: "condition-1", TargetID: "email-1", SourceHandle: &trueHandle},
			{WorkflowID: workflowID, EdgeID: "e5", SourceID: "condition-1", TargetID: "end-1", SourceHandle: &falseHandle},
			{WorkflowID: workflowID, EdgeID: "e6", SourceID: "email-1", TargetID: "end-1"},
		}

		graph, _ := BuildGraph(nodes, edges)

		// Use a weather function that returns a low temperature
		weatherFn := func(ctx context.Context, city string) (float64, error) {
			return 20.0, nil
		}
		registry := DefaultRegistry(weatherFn)
		executor := NewExecutor(registry)

		formData := FormData{
			Name:      "Bob",
			Email:     "bob@example.com",
			City:      "Melbourne",
			Operator:  "greater_than",
			Threshold: 25.0,
		}

		steps, err := executor.Execute(context.Background(), graph, formData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have: start, form, weather, condition, end (no email)
		if len(steps) != 5 {
			t.Errorf("expected 5 steps, got %d", len(steps))
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
		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "unknown-1", NodeType: "unknown_type"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "unknown-1"},
		}

		graph, _ := BuildGraph(nodes, edges)
		registry := DefaultRegistry(nil)
		executor := NewExecutor(registry)

		_, err := executor.Execute(context.Background(), graph, FormData{})
		if err == nil {
			t.Error("expected error for unknown node type")
		}
	})

	t.Run("handler error", func(t *testing.T) {
		nodes := []Node{
			{WorkflowID: workflowID, NodeID: "start-1", NodeType: "start"},
			{WorkflowID: workflowID, NodeID: "weather-1", NodeType: "integration"},
		}
		edges := []Edge{
			{WorkflowID: workflowID, EdgeID: "e1", SourceID: "start-1", TargetID: "weather-1"},
		}

		graph, _ := BuildGraph(nodes, edges)

		// Create a registry with a failing handler
		registry := NewHandlerRegistry()
		registry.Register(&StartHandler{})
		registry.Register(&failingHandler{})

		executor := NewExecutor(registry)

		steps, err := executor.Execute(context.Background(), graph, FormData{})
		if err == nil {
			t.Error("expected error from failing handler")
		}

		// Should have start step before error
		if len(steps) < 1 {
			t.Error("expected at least start step before error")
		}
	})
}

type failingHandler struct{}

func (h *failingHandler) NodeType() string { return "integration" }
func (h *failingHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "integration",
	}, errors.New("handler failed")
}
