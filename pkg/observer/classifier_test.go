package observer

import (
	"axom-observer/pkg/config"
	"axom-observer/pkg/models"
	"testing"
)

func TestGetMetric(t *testing.T) {
	sig := models.Signal{
		LatencyMS: 123.4,
		Status:    200,
		Metadata:  map[string]interface{}{"tokens_used": 42},
	}
	if got := getMetric(sig, "latency_ms"); got != 123.4 {
		t.Errorf("latency_ms: got %v, want 123.4", got)
	}
	if got := getMetric(sig, "status"); got != 200 {
		t.Errorf("status: got %v, want 200", got)
	}
	if got := getMetric(sig, "tokens_used"); got != 42 {
		t.Errorf("tokens_used: got %v, want 42", got)
	}
}

func TestOutcomeDetection(t *testing.T) {
	rules := &config.Rules{
		OutcomeDetection: config.OutcomeDetection{
			SuccessConditions: []struct {
				Protocol     string "yaml:\"protocol\""
				StatusCodes  []int  "yaml:\"status_codes\""
				ContentMatch string "yaml:\"content_match\""
			}{
				{Protocol: "http", StatusCodes: []int{200}, ContentMatch: ""},
			},
		},
	}
	classifier := NewBehaviorClassifier(rules)
	sig := models.Signal{Protocol: "http", Status: 200}
	alerts := classifier.Analyze(sig)
	found := false
	for _, a := range alerts {
		if a == "outcome_success" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected outcome_success alert, got %v", alerts)
	}
}
