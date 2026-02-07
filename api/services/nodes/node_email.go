package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// EmailNode composes and "sends" an email using a template from metadata.
// Variable placeholders like {{city}} in the template are resolved from
// the runtime context. Actual sending is simulated for this challenge.
type EmailNode struct {
	base BaseFields

	InputVariables  []string      `json:"inputVariables"`
	OutputVariables []string      `json:"outputVariables"`
	EmailTemplate   EmailTemplate `json:"emailTemplate"`
}

type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewEmailNode(base BaseFields) (*EmailNode, error) {
	n := &EmailNode{base: base}
	if err := json.Unmarshal(base.Metadata, n); err != nil {
		return nil, fmt.Errorf("invalid email metadata: %w", err)
	}
	return n, nil
}

func (n *EmailNode) ToJSON() NodeJSON {
	return NodeJSON{
		ID:       n.base.ID,
		Type:     n.base.NodeType,
		Position: n.base.Position,
		Data: NodeData{
			Label:       n.base.Label,
			Description: n.base.Description,
			Metadata:    n.base.Metadata,
		},
	}
}

// Execute resolves template placeholders from context variables and
// simulates sending the email. Returns the composed email as output.
func (n *EmailNode) Execute(_ context.Context, nCtx *NodeContext) (*ExecutionResult, error) {
	subject := resolveTemplate(n.EmailTemplate.Subject, nCtx.Variables)
	body := resolveTemplate(n.EmailTemplate.Body, nCtx.Variables)

	email, _ := nCtx.Variables["email"].(string)

	slog.Info("sending email", "to", email, "subject", subject)

	return &ExecutionResult{
		Status: "completed",
		Output: map[string]any{
			"emailDraft": map[string]any{
				"to":      email,
				"from":    "weather-alerts@example.com",
				"subject": subject,
				"body":    body,
			},
			"deliveryStatus": "sent",
			"emailSent":      true,
		},
	}, nil
}

// resolveTemplate replaces {{key}} placeholders with values from variables.
func resolveTemplate(tmpl string, vars map[string]any) string {
	result := tmpl
	for key, val := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}
	return result
}
