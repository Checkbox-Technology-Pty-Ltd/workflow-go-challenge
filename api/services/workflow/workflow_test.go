package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
