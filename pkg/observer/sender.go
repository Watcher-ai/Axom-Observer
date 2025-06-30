package observer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"axom-observer/pkg/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Environment variables (documented for production):
//   AXOM_API_KEY           - Required. API key for backend authentication.
//   AXOM_BACKEND_URL       - Optional. Override backend URL. Default: https://api.axom.ai/ingest
//   AXOM_SKIP_TLS_VERIFY   - Optional. Set to "1" to skip TLS verification (testing only!)
//   AXOM_BATCH_SIZE        - Optional. Batch size for sending signals. Default: 50
//   AXOM_FLUSH_INTERVAL    - Optional. Flush interval in seconds. Default: 10
//   AXOM_METRICS_ENABLED   - Optional. Set to "0" to disable Prometheus metrics server. Default: enabled.

var (
	signalsSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "axom_signals_sent_total",
		Help: "Total number of signals sent to backend",
	})
	signalsDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "axom_signals_dropped_total",
		Help: "Total number of signals dropped after retries",
	})
	metricsServerStarted = false
)

func init() {
	prometheus.MustRegister(signalsSent, signalsDropped)
	// Only start metrics server if enabled (default: true)
	if os.Getenv("AXOM_METRICS_ENABLED") != "0" && !metricsServerStarted {
		metricsServerStarted = true
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			server := &http.Server{Addr: ":2112", Handler: mux}
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Prometheus metrics server error: %v", err)
			}
		}()
	}
	log.Println("[observer] SignalSender initialized. Prometheus metrics enabled:", os.Getenv("AXOM_METRICS_ENABLED") != "0")
}

type SignalSender struct {
	apiKey        string
	url           string
	client        *http.Client
	batchSize     int
	flushInterval time.Duration
}

// NewSignalSender creates a new SignalSender with config values.
func NewSignalSender(apiKey, url string, batchSize int, flushInterval time.Duration) *SignalSender {
	if url == "" {
		url = os.Getenv("AXOM_BACKEND_URL")
		if url == "" {
			url = "http://localhost:8000/ingest"
		}
	}
	skipTLS := os.Getenv("AXOM_SKIP_TLS_VERIFY") == "1"
	client := &http.Client{Timeout: 10 * time.Second}
	if skipTLS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	if batchSize <= 0 {
		if v := os.Getenv("AXOM_BATCH_SIZE"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				batchSize = n
			}
		}
		if batchSize <= 0 {
			batchSize = 50
		}
	}
	if flushInterval <= 0 {
		if v := os.Getenv("AXOM_FLUSH_INTERVAL"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				flushInterval = time.Duration(n) * time.Second
			}
		}
		if flushInterval <= 0 {
			flushInterval = 10 * time.Second
		}
	}
	return &SignalSender{
		apiKey:        apiKey,
		url:           url,
		client:        client,
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

func (s *SignalSender) Start(ctx context.Context, ch <-chan models.Signal) {
	batch := make([]models.Signal, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()
	flush := func() {
		if len(batch) > 0 {
			s.sendBatchWithRetry(batch)
			batch = batch[:0]
		}
	}
	for {
		select {
		case sig := <-ch:
			sig.Redact("authorization", "api_key")
			batch = append(batch, sig)
			if len(batch) >= s.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

// sendBatchWithRetry sends a batch with exponential backoff on 429/5xx errors.
func (s *SignalSender) sendBatchWithRetry(signals []models.Signal) {
	const maxRetries = 5
	const baseDelay = 2 * time.Second
	var attempt int
	log.Printf("[observer] Attempting to send batch of %d signals to %s", len(signals), s.url)
	for {
		err, retry, status := s.sendBatchOnce(signals)
		if err == nil {
			log.Printf("[observer] Successfully sent batch of %d signals", len(signals))
			return
		}
		if !retry || attempt >= maxRetries {
			log.Printf("[observer] Failed to send batch after %d attempts (last status: %d): %v", attempt+1, status, err)
			signalsDropped.Add(float64(len(signals)))
			return
		}
		delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
		log.Printf("[observer] Batch send failed with status %d, retrying in %v (attempt %d/%d)...", status, delay, attempt+1, maxRetries)
		time.Sleep(delay)
		attempt++
	}
}

// sendBatchOnce sends a batch and returns (error, shouldRetry, statusCode)
func (s *SignalSender) sendBatchOnce(signals []models.Signal) (error, bool, int) {
	body, err := json.Marshal(signals)
	if err != nil {
		log.Printf("Failed to marshal batch: %v", err)
		return err, false, 0
	}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(body))
	if err != nil {
		log.Printf("Failed to create batch request: %v", err)
		return err, false, 0
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Failed to send batch: %v", err)
		return err, true, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		signalsSent.Add(float64(len(signals)))
		return nil, false, resp.StatusCode
	}
	log.Printf("Batch HTTP error: %s", resp.Status)
	// Retry on 429 and 5xx
	if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
		return &httpStatusError{StatusCode: resp.StatusCode}, true, resp.StatusCode
	}
	signalsDropped.Add(float64(len(signals)))
	return &httpStatusError{StatusCode: resp.StatusCode}, false, resp.StatusCode
}

// For compatibility with main.go (single send, not used in batch mode)
func (s *SignalSender) Send(sig models.Signal) error {
	sig.Redact()
	return s.SendBatchCompat([]models.Signal{sig})
}

func (s *SignalSender) SendBatchCompat(signals []models.Signal) error {
	body, err := json.Marshal(signals)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return &httpStatusError{StatusCode: resp.StatusCode}
	}
	return nil
}

type httpStatusError struct {
	StatusCode int
}

func (e *httpStatusError) Error() string {
	return "HTTP error: " + http.StatusText(e.StatusCode)
}
