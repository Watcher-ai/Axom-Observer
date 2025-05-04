package observer

import (
	"axom-observer/pkg/config"
	"axom-observer/pkg/models"
	"regexp"
	"strconv"
	"strings"
)

type BehaviorClassifier struct {
	rules        *config.Rules
	redactFields []string // configurable fields to redact
}

func NewBehaviorClassifier(rules *config.Rules) *BehaviorClassifier {
	return &BehaviorClassifier{
		rules:        rules,
		redactFields: []string{"authorization", "api_key"}, // extend as needed/configurable
	}
}

func (c *BehaviorClassifier) Analyze(signal models.Signal) []string {
	var alerts []string

	// Outcome-based pricing signal: check outcome detection rules
	for _, cond := range c.rules.OutcomeDetection.SuccessConditions {
		if cond.Protocol == signal.Protocol {
			statusMatch := false
			for _, code := range cond.StatusCodes {
				if signal.Status == code {
					statusMatch = true
					break
				}
			}
			contentMatch := false
			if cond.ContentMatch != "" && len(signal.RawResponse) > 0 {
				matched, _ := regexp.Match(cond.ContentMatch, signal.RawResponse)
				contentMatch = matched
			}
			if statusMatch || contentMatch {
				alerts = append(alerts, "outcome_success")
			}
		}
	}

	// Example: check all behavior profiles from config
	for _, profile := range c.rules.BehaviorProfiles {
		if evalCondition(profile.Condition, signal) {
			alerts = append(alerts, profile.Name)
		}
	}

	return alerts
}

func evalCondition(cond string, signal models.Signal) bool {
	parts := strings.Split(cond, " ")
	if len(parts) < 3 {
		return false
	}
	switch parts[1] {
	case ">":
		return getMetric(signal, parts[0]) > parseValue(parts[2])
	case "==":
		return getMetric(signal, parts[0]) == parseValue(parts[2])
	}
	return false
}

// getMetric extracts a float64 metric from the signal by key.
// Extend this function to support more metrics as needed.
func getMetric(signal models.Signal, key string) float64 {
	switch key {
	case "latency_ms":
		return signal.LatencyMS
	case "status":
		return float64(signal.Status)
	// Add more cases as needed, e.g., tokens_used, etc.
	default:
		if v, ok := signal.Metadata[key]; ok {
			switch val := v.(type) {
			case float64:
				return val
			case int:
				return float64(val)
			case string:
				f, _ := strconv.ParseFloat(val, 64)
				return f
			}
		}
	}
	return 0
}

// parseValue parses a string to float64 for comparison in evalCondition.
func parseValue(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)
	return f
}
