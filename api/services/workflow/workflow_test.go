package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func TestHandleGetWorkflow(t *testing.T) {
	validID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	label := "Start"
	desc := "Test node"

	tests := []struct {
		name           string
		workflowID     string
		mockSetup      func(*MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "valid workflow returns 200",
			workflowID: validID.String(),
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return &Workflow{ID: validID, Name: "Test Workflow"}, nil
				}
				m.GetNodesByWorkflowIDFunc = func(ctx context.Context, id uuid.UUID) ([]Node, error) {
					return []Node{
						{WorkflowID: validID, NodeID: "start", NodeType: "start", Label: &label, Description: &desc, XPos: 0, YPos: 0},
					}, nil
				}
				m.GetEdgesByWorkflowIDFunc = func(ctx context.Context, id uuid.UUID) ([]Edge, error) {
					return []Edge{}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid returns 400",
			workflowID:     "not-a-uuid",
			mockSetup:      func(m *MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid workflow id",
		},
		{
			name:       "workflow not found returns 404",
			workflowID: validID.String(),
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "workflow not found",
		},
		{
			name:       "database error returns 500",
			workflowID: validID.String(),
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return nil, errors.New("connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
		{
			name:       "nodes query error returns 500",
			workflowID: validID.String(),
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return &Workflow{ID: validID}, nil
				}
				m.GetNodesByWorkflowIDFunc = func(ctx context.Context, id uuid.UUID) ([]Node, error) {
					return nil, errors.New("query failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
		{
			name:       "edges query error returns 500",
			workflowID: validID.String(),
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return &Workflow{ID: validID}, nil
				}
				m.GetNodesByWorkflowIDFunc = func(ctx context.Context, id uuid.UUID) ([]Node, error) {
					return []Node{}, nil
				}
				m.GetEdgesByWorkflowIDFunc = func(ctx context.Context, id uuid.UUID) ([]Edge, error) {
					return nil, errors.New("query failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRepository{}
			tt.mockSetup(mock)

			svc := &Service{repo: mock}

			req := httptest.NewRequest(http.MethodGet, "/workflows/"+tt.workflowID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.workflowID})
			rec := httptest.NewRecorder()

			svc.HandleGetWorkflow(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" {
				body := rec.Body.String()
				if body != tt.expectedBody+"\n" {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var resp WorkflowResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
			}
		})
	}
}

func TestHandleExecuteWorkflow(t *testing.T) {
	validID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	validRequestBody := `{
		"formData": {
			"name": "Alice",
			"email": "alice@example.com",
			"city": "Sydney",
			"operator": "greater_than",
			"threshold": 25
		},
		"condition": {
			"operator": "greater_than",
			"threshold": 25
		}
	}`

	tests := []struct {
		name           string
		workflowID     string
		requestBody    string
		mockSetup      func(*MockRepository)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *ExecutionResponse)
	}{
		{
			name:        "valid execution returns 200",
			workflowID:  validID.String(),
			requestBody: validRequestBody,
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return &Workflow{ID: validID, Name: "Test Workflow"}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *ExecutionResponse) {
				if resp.Status != "completed" {
					t.Errorf("expected status 'completed', got %q", resp.Status)
				}
				if len(resp.Steps) < 5 {
					t.Errorf("expected at least 5 steps, got %d", len(resp.Steps))
				}
			},
		},
		{
			name:           "invalid uuid returns 400",
			workflowID:     "not-a-uuid",
			requestBody:    validRequestBody,
			mockSetup:      func(m *MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid workflow id",
		},
		{
			name:           "invalid request body returns 400",
			workflowID:     validID.String(),
			requestBody:    "not json",
			mockSetup:      func(m *MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:        "workflow not found returns 404",
			workflowID:  validID.String(),
			requestBody: validRequestBody,
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "workflow not found",
		},
		{
			name:        "database error returns 500",
			workflowID:  validID.String(),
			requestBody: validRequestBody,
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return nil, errors.New("connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
		{
			name:       "condition not met skips email step",
			workflowID: validID.String(),
			requestBody: `{
				"formData": {
					"name": "Bob",
					"email": "bob@example.com",
					"city": "Melbourne",
					"operator": "greater_than",
					"threshold": 30
				},
				"condition": {
					"operator": "greater_than",
					"threshold": 30
				}
			}`,
			mockSetup: func(m *MockRepository) {
				m.GetWorkflowFunc = func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
					return &Workflow{ID: validID, Name: "Test Workflow"}, nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *ExecutionResponse) {
				hasEmail := false
				for _, step := range resp.Steps {
					if step.NodeType == "email" {
						hasEmail = true
					}
				}
				if hasEmail {
					t.Error("expected no email step when condition not met")
				}
			},
		},
	}

	// Test that execution is persisted
	t.Run("execution is persisted to database", func(t *testing.T) {
		var savedExec *WorkflowExecution
		mock := &MockRepository{
			GetWorkflowFunc: func(ctx context.Context, id uuid.UUID) (*Workflow, error) {
				return &Workflow{ID: validID, Name: "Test Workflow"}, nil
			},
			CreateExecutionFunc: func(ctx context.Context, exec *WorkflowExecution) error {
				savedExec = exec
				return nil
			},
		}

		svc := &Service{repo: mock}

		req := httptest.NewRequest(http.MethodPost, "/workflows/"+validID.String()+"/execute", strings.NewReader(validRequestBody))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"id": validID.String()})
		rec := httptest.NewRecorder()

		svc.HandleExecuteWorkflow(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if savedExec == nil {
			t.Fatal("expected CreateExecution to be called")
		}

		if savedExec.WorkflowID == nil || *savedExec.WorkflowID != validID {
			t.Errorf("expected workflow ID %s, got %v", validID, savedExec.WorkflowID)
		}

		if savedExec.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", savedExec.Status)
		}

		if len(savedExec.FinalContext) == 0 {
			t.Error("expected FinalContext to contain form data")
		}

		if len(savedExec.ExecutionTrace) == 0 {
			t.Error("expected ExecutionTrace to contain steps")
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRepository{}
			tt.mockSetup(mock)

			svc := &Service{repo: mock}

			req := httptest.NewRequest(http.MethodPost, "/workflows/"+tt.workflowID+"/execute", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": tt.workflowID})
			rec := httptest.NewRecorder()

			svc.HandleExecuteWorkflow(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" {
				body := rec.Body.String()
				if body != tt.expectedBody+"\n" {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.checkResponse != nil && tt.expectedStatus == http.StatusOK {
				var resp ExecutionResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				tt.checkResponse(t, &resp)
			}
		})
	}
}
