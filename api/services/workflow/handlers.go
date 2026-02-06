package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// StartHandler handles start nodes
type StartHandler struct{}

func (h *StartHandler) NodeType() string { return "start" }

func (h *StartHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "start",
		Status:     "completed",
		Duration:   10,
		Output:     StepOutput{Message: "Workflow started"},
		Timestamp:  ec.StartTime.Format(time.RFC3339),
	}, nil
}

// EndHandler handles end nodes
type EndHandler struct{}

func (h *EndHandler) NodeType() string { return "end" }

func (h *EndHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "end",
		Status:     "completed",
		Duration:   5,
		Output:     StepOutput{Message: "Workflow completed"},
		Timestamp:  time.Now().Format(time.RFC3339),
	}, nil
}

// FormHandler handles form nodes
type FormHandler struct{}

func (h *FormHandler) NodeType() string { return "form" }

func (h *FormHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	// Store form data in state for downstream handlers
	if ec.State == nil {
		ec.State = make(map[string]interface{})
	}
	ec.State["name"] = ec.FormData.Name
	ec.State["email"] = ec.FormData.Email
	ec.State["city"] = ec.FormData.City

	formData := ec.FormData
	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "form",
		Status:     "completed",
		Duration:   15,
		Output: StepOutput{
			Message:  "User input collected",
			FormData: &formData,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// WeatherHandler handles weather/integration nodes
type WeatherHandler struct {
	weatherFn func(ctx context.Context, city string) (float64, error)
}

func (h *WeatherHandler) NodeType() string { return "integration" }

func (h *WeatherHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	// Read city from user input (FormData), not from node metadata
	city := ec.FormData.City
	if city == "" {
		return ExecutionStep{}, fmt.Errorf("city not provided in form data")
	}

	temperature := 25.0 // default
	if h.weatherFn != nil {
		temp, err := h.weatherFn(ec.Ctx, city)
		if err == nil {
			temperature = temp
		}
	}

	// Store temperature in state for condition handler
	if ec.State == nil {
		ec.State = make(map[string]interface{})
	}
	ec.State["temperature"] = temperature

	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "integration",
		Status:     "completed",
		Duration:   150,
		Output: StepOutput{
			Message: fmt.Sprintf("Fetched weather data for %s", city),
			APIResponse: &APIResponse{
				Endpoint:   "https://api.open-meteo.com/v1/forecast",
				Method:     "GET",
				StatusCode: 200,
				Data: map[string]any{
					"temperature": temperature,
					"city":        city,
				},
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// ConditionHandler handles condition nodes
type ConditionHandler struct{}

func (h *ConditionHandler) NodeType() string { return "condition" }

func (h *ConditionHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	var metadata map[string]interface{}
	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return ExecutionStep{}, fmt.Errorf("failed to unmarshal node metadata: %w", err)
		}
	}

	temp, ok := ec.State["temperature"].(float64)
	if !ok {
		return ExecutionStep{}, fmt.Errorf("temperature not found in execution context state")
	}
	temperature := temp

	operator, ok := metadata["operator"].(string)
	if !ok {
		return ExecutionStep{}, fmt.Errorf("operator not found or is not a string in node metadata")
	}

	threshold, ok := metadata["threshold"].(float64)
	if !ok {
		return ExecutionStep{}, fmt.Errorf("threshold not found or is not a float64 in node metadata")
	}

	result := evaluateCondition(temperature, operator, threshold)

	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "condition",
		Status:     "completed",
		Duration:   5,
		Output: StepOutput{
			Message: fmt.Sprintf("Condition evaluated: temperature %.1f°C %s %.1f°C", temperature, operator, threshold),
			ConditionResult: &ConditionResult{
				Expression:  fmt.Sprintf("temperature %s %.1f", operator, threshold),
				Result:      result,
				Temperature: temperature,
				Operator:    operator,
				Threshold:   threshold,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// EmailHandler handles email nodes
type EmailHandler struct{}

func (h *EmailHandler) NodeType() string { return "email" }

func (h *EmailHandler) Execute(ec *ExecutionContext, node *Node) (ExecutionStep, error) {
	// Read recipient from user input (FormData)
	to := ec.FormData.Email
	if to == "" {
		return ExecutionStep{}, fmt.Errorf("recipient email not provided in form data")
	}

	// Read subject and template from node metadata (workflow config)
	var metadata map[string]interface{}
	if len(node.Metadata) > 0 {
		if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
			return ExecutionStep{}, fmt.Errorf("failed to unmarshal node metadata: %w", err)
		}
	}

	subject, ok := metadata["subject"].(string)
	if !ok {
		subject = "Weather Alert" // Default subject
	}

	// Build email body using user's name, city, and temperature
	name := ec.FormData.Name
	city := ec.FormData.City
	temperature, _ := ec.State["temperature"].(float64)

	body := fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", name, city, temperature)

	timestamp := time.Now().Format(time.RFC3339)

	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "email",
		Status:     "completed",
		Duration:   50,
		Output: StepOutput{
			Message: "Alert email sent",
			EmailContent: &EmailContent{
				To:        to,
				Subject:   subject,
				Body:      body,
				Timestamp: timestamp,
			},
		},
		Timestamp: timestamp,
	}, nil
}

func evaluateCondition(temperature float64, operator string, threshold float64) bool {
	switch operator {
	case "greater_than":
		return temperature > threshold
	case "less_than":
		return temperature < threshold
	case "equals":
		return temperature == threshold
	case "greater_than_or_equal":
		return temperature >= threshold
	case "less_than_or_equal":
		return temperature <= threshold
	default:
		return false
	}
}
