package observer

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"axom-observer/pkg/models"
)

// TaskDetector provides comprehensive AI task detection
type TaskDetector struct {
	logger     *log.Logger
	taskRules  []TaskRule
	signalCh   chan<- models.Signal
	customerID string
	agentID    string
}

// TaskRule defines a pattern for detecting tasks
type TaskRule struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Provider    string            `json:"provider"`
	Patterns    []TaskPattern     `json:"patterns"`
	Outcomes    []OutcomeRule     `json:"outcomes"`
	Timeout     time.Duration     `json:"timeout"`
	Metadata    map[string]string `json:"metadata"`
}

// TaskPattern defines how to detect a task
type TaskPattern struct {
	Type       string            `json:"type"`       // "prompt", "response", "model", "endpoint"
	Conditions map[string]string `json:"conditions"` // field -> regex pattern
	Confidence float64           `json:"confidence"` // 0.0 to 1.0
	Required   bool              `json:"required"`   // if true, must match
}

// OutcomeRule defines how to determine task outcome
type OutcomeRule struct {
	Name       string            `json:"name"`
	Conditions map[string]string `json:"conditions"`
	Outcome    string            `json:"outcome"` // "success", "failure", "partial"
	Score      float64           `json:"score"`   // 0.0 to 1.0
}

// NewTaskDetector creates a new task detector
func NewTaskDetector(signalCh chan<- models.Signal, logger *log.Logger, customerID, agentID string) *TaskDetector {
	detector := &TaskDetector{
		logger:     logger,
		signalCh:   signalCh,
		customerID: customerID,
		agentID:    agentID,
	}

	// Initialize with comprehensive task rules
	detector.initializeTaskRules()

	return detector
}

// initializeTaskRules initializes comprehensive task detection rules
func (d *TaskDetector) initializeTaskRules() {
	d.taskRules = []TaskRule{
		// Sales and Marketing Tasks
		{
			Name:        "cold_calling",
			Description: "Cold calling and lead generation",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(cold call|lead generation|prospecting|sales call|outreach)",
					},
					Confidence: 0.8,
					Required:   true,
				},
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(script|dialogue|conversation|pitch)",
					},
					Confidence: 0.6,
					Required:   false,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "meeting_scheduled",
					Conditions: map[string]string{
						"response": "(?i)(schedule|meeting|appointment|calendar)",
					},
					Outcome: "success",
					Score:   1.0,
				},
				{
					Name: "lead_qualified",
					Conditions: map[string]string{
						"response": "(?i)(qualified|interested|budget|decision maker)",
					},
					Outcome: "success",
					Score:   0.8,
				},
			},
			Timeout: 10 * time.Minute,
		},
		{
			Name:        "email_marketing",
			Description: "Email marketing and campaigns",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(email|newsletter|campaign|blast|sequence)",
					},
					Confidence: 0.9,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "email_generated",
					Conditions: map[string]string{
						"response": "(?i)(subject|body|signature|call to action)",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 5 * time.Minute,
		},

		// Customer Support Tasks
		{
			Name:        "customer_support",
			Description: "Customer support and help desk",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(support|help|issue|problem|ticket|complaint)",
					},
					Confidence: 0.8,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "issue_resolved",
					Conditions: map[string]string{
						"response": "(?i)(resolved|fixed|solved|working|resolved)",
					},
					Outcome: "success",
					Score:   1.0,
				},
				{
					Name: "escalated",
					Conditions: map[string]string{
						"response": "(?i)(escalate|manager|supervisor|higher level)",
					},
					Outcome: "partial",
					Score:   0.5,
				},
			},
			Timeout: 15 * time.Minute,
		},

		// Content Creation Tasks
		{
			Name:        "content_creation",
			Description: "Content creation and writing",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(write|create|generate|compose|draft)",
					},
					Confidence: 0.7,
					Required:   true,
				},
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(blog|article|post|content|copy)",
					},
					Confidence: 0.6,
					Required:   false,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "content_created",
					Conditions: map[string]string{
						"response": "(?i)(\\w{50,})", // At least 50 characters
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 10 * time.Minute,
		},

		// Data Analysis Tasks
		{
			Name:        "data_analysis",
			Description: "Data analysis and insights",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(analyze|analysis|insights|data|metrics|statistics)",
					},
					Confidence: 0.8,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "insights_generated",
					Conditions: map[string]string{
						"response": "(?i)(trend|pattern|insight|finding|conclusion)",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 15 * time.Minute,
		},

		// Code Generation Tasks
		{
			Name:        "code_generation",
			Description: "Code generation and programming",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(code|program|function|script|algorithm)",
					},
					Confidence: 0.9,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "code_generated",
					Conditions: map[string]string{
						"response": "(?i)(def |function |class |import |const |let |var )",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 10 * time.Minute,
		},

		// Translation Tasks
		{
			Name:        "translation",
			Description: "Language translation",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(translate|translation|language|convert)",
					},
					Confidence: 0.9,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "translation_complete",
					Conditions: map[string]string{
						"response": "(?i)(\\w{10,})", // At least 10 characters
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 5 * time.Minute,
		},

		// Image Generation Tasks
		{
			Name:        "image_generation",
			Description: "Image generation and creation",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "endpoint",
					Conditions: map[string]string{
						"path": "(?i)(image|generation|dall|midjourney)",
					},
					Confidence: 0.9,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "image_created",
					Conditions: map[string]string{
						"response": "(?i)(url|image|png|jpg|jpeg)",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 2 * time.Minute,
		},

		// Meeting Scheduling Tasks
		{
			Name:        "meeting_scheduling",
			Description: "Meeting scheduling and calendar management",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(schedule|meeting|appointment|calendar|book)",
					},
					Confidence: 0.8,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "meeting_scheduled",
					Conditions: map[string]string{
						"response": "(?i)(scheduled|booked|confirmed|calendar)",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 5 * time.Minute,
		},

		// Research Tasks
		{
			Name:        "research",
			Description: "Research and information gathering",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(research|find|search|investigate|look up)",
					},
					Confidence: 0.8,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "research_complete",
					Conditions: map[string]string{
						"response": "(?i)(\\w{50,})", // At least 50 characters
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 10 * time.Minute,
		},

		// Summarization Tasks
		{
			Name:        "summarization",
			Description: "Text summarization and extraction",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(summarize|summary|extract|key points|main points)",
					},
					Confidence: 0.9,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "summary_created",
					Conditions: map[string]string{
						"response": "(?i)(\\w{30,})", // At least 30 characters
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 5 * time.Minute,
		},

		// Sentiment Analysis Tasks
		{
			Name:        "sentiment_analysis",
			Description: "Sentiment analysis and emotion detection",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(sentiment|emotion|feeling|tone|mood)",
					},
					Confidence: 0.8,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "sentiment_detected",
					Conditions: map[string]string{
						"response": "(?i)(positive|negative|neutral|happy|sad|angry)",
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 3 * time.Minute,
		},

		// QA Tasks
		{
			Name:        "question_answering",
			Description: "Question answering and knowledge retrieval",
			Provider:    "any",
			Patterns: []TaskPattern{
				{
					Type: "prompt",
					Conditions: map[string]string{
						"content": "(?i)(what|how|why|when|where|who|which)",
					},
					Confidence: 0.7,
					Required:   true,
				},
			},
			Outcomes: []OutcomeRule{
				{
					Name: "answer_provided",
					Conditions: map[string]string{
						"response": "(?i)(\\w{20,})", // At least 20 characters
					},
					Outcome: "success",
					Score:   1.0,
				},
			},
			Timeout: 5 * time.Minute,
		},
	}
}

// DetectTask detects if a signal represents a task
func (d *TaskDetector) DetectTask(signal models.Signal) *models.Task {
	for _, rule := range d.taskRules {
		if d.matchesTaskRule(signal, rule) {
			task := &models.Task{
				ID:         d.generateTaskID(signal.CustomerID, signal.AgentID, rule.Name),
				CustomerID: signal.CustomerID,
				AgentID:    signal.AgentID,
				Type:       rule.Name,
				Status:     "in_progress",
				CreatedAt:  signal.Timestamp,
				Metadata: map[string]interface{}{
					"description": rule.Description,
					"provider":    signal.Metadata["provider"],
					"model":       signal.Metadata["model"],
					"confidence":  d.calculateConfidence(signal, rule),
				},
				Signals: []string{signal.ID},
			}

			d.logger.Printf("🎯 Task detected: %s (%s) - Confidence: %.2f",
				rule.Name, rule.Description, task.Metadata["confidence"])

			return task
		}
	}

	return nil
}

// matchesTaskRule checks if a signal matches a task rule
func (d *TaskDetector) matchesTaskRule(signal models.Signal, rule TaskRule) bool {
	// Check provider if specified
	if rule.Provider != "any" {
		if provider, ok := signal.Metadata["provider"].(string); ok {
			if provider != rule.Provider {
				return false
			}
		}
	}

	// Check all patterns
	for _, pattern := range rule.Patterns {
		matches := d.matchesPattern(signal, pattern)
		if pattern.Required && !matches {
			return false
		}
	}

	return true
}

// matchesPattern checks if a signal matches a specific pattern
func (d *TaskDetector) matchesPattern(signal models.Signal, pattern TaskPattern) bool {
	switch pattern.Type {
	case "prompt":
		if prompt, ok := signal.Metadata["prompt_preview"].(string); ok {
			return d.matchesConditions(prompt, pattern.Conditions)
		}
	case "response":
		if response, ok := signal.Metadata["response_preview"].(string); ok {
			return d.matchesConditions(response, pattern.Conditions)
		}
	case "model":
		if model, ok := signal.Metadata["model"].(string); ok {
			return d.matchesConditions(model, pattern.Conditions)
		}
	case "endpoint":
		if endpoint, ok := signal.Metadata["endpoint"].(string); ok {
			return d.matchesConditions(endpoint, pattern.Conditions)
		}
	}

	return false
}

// matchesConditions checks if text matches all conditions
func (d *TaskDetector) matchesConditions(text string, conditions map[string]string) bool {
	for pattern := range conditions {
		matched, err := regexp.MatchString(pattern, text)
		if err != nil {
			d.logger.Printf("Invalid regex pattern %s: %v", pattern, err)
			continue
		}
		if !matched {
			return false
		}
	}
	return true
}

// calculateConfidence calculates confidence score for task detection
func (d *TaskDetector) calculateConfidence(signal models.Signal, rule TaskRule) float64 {
	totalConfidence := 0.0
	matchedPatterns := 0

	for _, pattern := range rule.Patterns {
		if d.matchesPattern(signal, pattern) {
			totalConfidence += pattern.Confidence
			matchedPatterns++
		}
	}

	if matchedPatterns == 0 {
		return 0.0
	}

	return totalConfidence / float64(matchedPatterns)
}

// DetermineOutcome determines the outcome of a completed task
func (d *TaskDetector) DetermineOutcome(task *models.Task, signals []models.Signal) (string, map[string]interface{}) {
	// Find the rule for this task type
	var rule *TaskRule
	for _, r := range d.taskRules {
		if r.Name == task.Type {
			rule = &r
			break
		}
	}

	if rule == nil {
		return "unknown", map[string]interface{}{"reason": "no_rule_found"}
	}

	// Check all outcome rules
	bestOutcome := "unknown"
	bestScore := 0.0
	outcomeData := make(map[string]interface{})

	for _, outcomeRule := range rule.Outcomes {
		score := d.evaluateOutcomeRule(signals, outcomeRule)
		if score > bestScore {
			bestScore = score
			bestOutcome = outcomeRule.Outcome
			outcomeData["outcome_rule"] = outcomeRule.Name
			outcomeData["confidence"] = score
		}
	}

	// Add task metadata
	outcomeData["task_type"] = task.Type
	outcomeData["total_signals"] = len(signals)
	outcomeData["duration_minutes"] = time.Since(task.CreatedAt).Minutes()

	return bestOutcome, outcomeData
}

// evaluateOutcomeRule evaluates how well signals match an outcome rule
func (d *TaskDetector) evaluateOutcomeRule(signals []models.Signal, rule OutcomeRule) float64 {
	matches := 0
	total := 0

	for _, signal := range signals {
		if response, ok := signal.Metadata["response_preview"].(string); ok {
			total++
			if d.matchesConditions(response, rule.Conditions) {
				matches++
			}
		}
	}

	if total == 0 {
		return 0.0
	}

	return float64(matches) / float64(total) * rule.Score
}

// generateTaskID generates a unique task ID
func (d *TaskDetector) generateTaskID(customerID, agentID, taskType string) string {
	return fmt.Sprintf("%s_%s_%s_%d", customerID, agentID, taskType, time.Now().Unix())
}
