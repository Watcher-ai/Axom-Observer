package models

import (
	"testing"
)

func TestRedact(t *testing.T) {
	s := &Signal{
		Metadata: map[string]interface{}{
			"authorization": "secret",
			"api_key":       "key",
			"other":         "visible",
		},
	}
	s.Redact("authorization", "api_key")
	if s.Metadata["authorization"] != "[REDACTED]" {
		t.Errorf("authorization not redacted")
	}
	if s.Metadata["api_key"] != "[REDACTED]" {
		t.Errorf("api_key not redacted")
	}
	if s.Metadata["other"] != "visible" {
		t.Errorf("other field should not be redacted")
	}
}
