package models

import "time"

type Signal struct {
	Timestamp   time.Time              `json:"timestamp"`
	Protocol    string                 `json:"protocol"`
	Source      Endpoint               `json:"source"`
	Destination Endpoint               `json:"destination"`
	Operation   string                 `json:"operation"`
	Status      int                    `json:"status"`
	LatencyMS   float64                `json:"latency_ms"`
	Metadata    map[string]interface{} `json:"metadata"`
	Alerts      []string               `json:"alerts,omitempty"`
	RawRequest  []byte                 `json:"raw_request,omitempty"`
	RawResponse []byte                 `json:"raw_response,omitempty"`
	CPUUsage    float64                `json:"cpu_usage,omitempty"`    // percent
	MemUsage    float64                `json:"mem_usage,omitempty"`    // MB or percent
	GPUUsage    float64                `json:"gpu_usage,omitempty"`    // percent (if available)
	DBOperation string                 `json:"db_operation,omitempty"` // e.g. SELECT, INSERT
	DBTable     string                 `json:"db_table,omitempty"`
	DBLatencyMS float64                `json:"db_latency_ms,omitempty"`
}

type Endpoint struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname,omitempty"`
}

// Redact sensitive fields from the signal before export.
// Accepts a list of field names to redact.
func (s *Signal) Redact(fields ...string) {
	if s.Metadata != nil {
		for _, field := range fields {
			if _, ok := s.Metadata[field]; ok {
				s.Metadata[field] = "[REDACTED]"
			}
		}
	}
	// Optionally redact RawRequest/RawResponse if needed
}
