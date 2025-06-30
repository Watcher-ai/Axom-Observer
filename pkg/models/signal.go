package models

import "time"

// Signal represents a captured AI API interaction for billing and monitoring
type Signal struct {
	// Core identification
	ID         string `json:"id"`                // Unique signal identifier
	CustomerID string `json:"customer_id"`       // Customer identifier
	AgentID    string `json:"agent_id"`          // AI agent identifier
	TaskID     string `json:"task_id,omitempty"` // Business task identifier for outcome-based billing

	// Timing and performance
	Timestamp time.Time `json:"timestamp"`  // When the signal was captured
	LatencyMS float64   `json:"latency_ms"` // Request latency in milliseconds

	// Network information
	Protocol    string   `json:"protocol"`    // HTTP/HTTPS
	Source      Endpoint `json:"source"`      // Client endpoint
	Destination Endpoint `json:"destination"` // AI service endpoint

	// AI operation details
	Operation string                 `json:"operation"` // chat_completion, embedding, etc.
	Status    int                    `json:"status"`    // HTTP status code
	Metadata  map[string]interface{} `json:"metadata"`  // AI-specific data (tokens, model, etc.)

	// Task and outcome tracking
	TaskType    string                 `json:"task_type,omitempty"`    // Business task type
	Outcome     string                 `json:"outcome,omitempty"`      // success, failure, partial
	OutcomeData map[string]interface{} `json:"outcome_data,omitempty"` // Outcome-specific metrics

	// Resource usage
	CPUUsage    float64 `json:"cpu_usage,omitempty"`    // CPU usage percentage
	MemoryUsage float64 `json:"memory_usage,omitempty"` // Memory usage in MB
	GPUUsage    float64 `json:"gpu_usage,omitempty"`    // GPU usage percentage

	// Database operations (if applicable)
	DBOperation string  `json:"db_operation,omitempty"`  // SELECT, INSERT, etc.
	DBTable     string  `json:"db_table,omitempty"`      // Database table name
	DBLatencyMS float64 `json:"db_latency_ms,omitempty"` // Database operation latency

	// Alerts and monitoring
	Alerts []Alert `json:"alerts,omitempty"` // Any alerts triggered

	// Raw data for debugging (optional)
	RawRequest  []byte `json:"raw_request,omitempty"`  // Original request body
	RawResponse []byte `json:"raw_response,omitempty"` // Original response body
}

// Endpoint represents a network endpoint
type Endpoint struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname,omitempty"`
}

// Alert represents a monitoring alert
type Alert struct {
	Type      string                 `json:"type"`      // error, warning, info
	Message   string                 `json:"message"`   // Alert description
	Severity  string                 `json:"severity"`  // low, medium, high, critical
	Metadata  map[string]interface{} `json:"metadata"`  // Alert-specific data
	Timestamp time.Time              `json:"timestamp"` // When alert was triggered
}

// Task represents a business task that groups related AI operations
type Task struct {
	ID          string                 `json:"id"`                     // Unique task identifier
	CustomerID  string                 `json:"customer_id"`            // Customer identifier
	AgentID     string                 `json:"agent_id"`               // AI agent identifier
	Type        string                 `json:"type"`                   // Task type (e.g., "cold_call", "support_ticket")
	Status      string                 `json:"status"`                 // pending, in_progress, completed, failed
	CreatedAt   time.Time              `json:"created_at"`             // Task creation time
	CompletedAt *time.Time             `json:"completed_at,omitempty"` // Task completion time
	Outcome     string                 `json:"outcome,omitempty"`      // success, failure, partial
	Metadata    map[string]interface{} `json:"metadata"`               // Task-specific data
	Signals     []string               `json:"signals"`                // Associated signal IDs
}

// BillingMetrics represents aggregated billing information
type BillingMetrics struct {
	CustomerID string    `json:"customer_id"`
	AgentID    string    `json:"agent_id"`
	Period     string    `json:"period"` // daily, weekly, monthly
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`

	// Usage metrics
	TotalSignals int     `json:"total_signals"`
	TotalTokens  int     `json:"total_tokens"`
	TotalLatency float64 `json:"total_latency_ms"`

	// Operation breakdown
	Operations map[string]int `json:"operations"` // Operation type counts
	Models     map[string]int `json:"models"`     // Model usage counts

	// Outcome-based metrics
	SuccessfulTasks int            `json:"successful_tasks"`
	FailedTasks     int            `json:"failed_tasks"`
	TaskTypes       map[string]int `json:"task_types"` // Task type counts

	// Resource usage
	TotalCPUUsage    float64 `json:"total_cpu_usage"`
	TotalMemoryUsage float64 `json:"total_memory_usage"`

	// Cost estimates (if applicable)
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
	Currency      string  `json:"currency,omitempty"`

	Metadata map[string]interface{} `json:"metadata"`
}

// Redact sensitive fields from the signal before export
func (s *Signal) Redact(fields ...string) {
	if s.Metadata != nil {
		for _, field := range fields {
			if _, ok := s.Metadata[field]; ok {
				s.Metadata[field] = "[REDACTED]"
			}
		}
	}
	if s.OutcomeData != nil {
		for _, field := range fields {
			if _, ok := s.OutcomeData[field]; ok {
				s.OutcomeData[field] = "[REDACTED]"
			}
		}
	}
}

// SetOutcome updates the signal with task outcome information
func (s *Signal) SetOutcome(outcome string, outcomeData map[string]interface{}) {
	s.Outcome = outcome
	s.OutcomeData = outcomeData
}

// IsTaskComplete checks if this signal represents a completed task
func (s *Signal) IsTaskComplete() bool {
	return s.Outcome != "" && s.TaskID != ""
}
