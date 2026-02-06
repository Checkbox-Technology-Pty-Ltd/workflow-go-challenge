package handlers

// Simulated duration constants for each node type (in milliseconds).
// These represent expected/typical execution times for each handler.
// In a production system, these would be replaced with actual timing measurements.
const (
	StartNodeDuration     = 10  // Minimal processing for workflow initialization
	EndNodeDuration       = 5   // Minimal processing for workflow completion
	FormNodeDuration      = 15  // Time to process form data
	ConditionNodeDuration = 5   // Time to evaluate condition logic
	EmailNodeDuration     = 50  // Simulated email send time
	WeatherNodeDuration   = 150 // Simulated external API call latency
)

// Default values for node configuration
const (
	DefaultEmailSubject = "Weather Alert"
)
