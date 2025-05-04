package observer

import (
	"axom-observer/pkg/models"
	"net/http"
	    "net/http/httptest"
	"testing"
)

func TestSendBatchEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
	sender := &SignalSender{
		apiKey: "dummy",
		url:    server.URL,
		client: &http.Client{},       
	}
	// Should not panic or error on empty batch
	sender.sendBatchWithRetry([]models.Signal{})
}
