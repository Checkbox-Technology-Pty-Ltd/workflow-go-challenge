package workflow

import (
	"context"
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
	city := ec.FormData.City
	temperature := 25.0 // default

	if h.weatherFn != nil {
		temp, err := h.weatherFn(ec.Ctx, city)
		if err == nil {
			temperature = temp
		}
	}

	// Store temperature in context for condition handler
	ec.Temperature = temperature

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
	temperature := ec.Temperature
	operator := ec.FormData.Operator
	threshold := ec.FormData.Threshold

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
	formData := ec.FormData
	temperature := ec.Temperature
	timestamp := time.Now().Format(time.RFC3339)

	return ExecutionStep{
		StepNumber: ec.StepNumber,
		NodeType:   "email",
		Status:     "completed",
		Duration:   50,
		Output: StepOutput{
			Message: "Alert email sent",
			EmailContent: &EmailContent{
				To:        formData.Email,
				Subject:   "Weather Alert",
				Body:      fmt.Sprintf("Hi %s, weather alert for %s! Temperature is %.1f°C!", formData.Name, formData.City, temperature),
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
