package observer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"axom-observer/pkg/models"
)

// HTTPSProxy handles HTTPS traffic with MITM capabilities
type HTTPSProxy struct {
	port         string
	signalCh     chan<- models.Signal
	logger       *log.Logger
	customerID   string
	agentID      string
	taskDetector *TaskDetector
	server       *http.Server
	caCert       *x509.Certificate
	caKey        *rsa.PrivateKey
}

// NewHTTPSProxy creates a new HTTPS proxy
func NewHTTPSProxy(port string, signalCh chan<- models.Signal, logger *log.Logger, customerID, agentID string) *HTTPSProxy {
	return &HTTPSProxy{
		port:         port,
		signalCh:     signalCh,
		logger:       logger,
		customerID:   customerID,
		agentID:      agentID,
		taskDetector: NewTaskDetector(signalCh, logger, customerID, agentID),
	}
}

// Start starts the HTTPS proxy
func (p *HTTPSProxy) Start(ctx context.Context) error {
	p.logger.Printf("Starting HTTPS proxy on port %s", p.port)

	// Load or generate CA certificate and key
	if err := p.loadOrGenerateCA(); err != nil {
		return fmt.Errorf("failed to load or generate CA: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleRequest)

	p.server = &http.Server{
		Addr:    ":" + p.port,
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Printf("HTTPS proxy error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTPS proxy
func (p *HTTPSProxy) Stop(ctx context.Context) error {
	if p.server != nil {
		return p.server.Shutdown(ctx)
	}
	return nil
}

// loadOrGenerateCA loads a CA from disk or generates and saves a new one
func (p *HTTPSProxy) loadOrGenerateCA() error {
	certPath := "certs/ca.crt"
	keyPath := "certs/ca.key"

	// Check if cert and key files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		p.logger.Println("No CA certificate found, generating a new one...")
		return p.generateAndSaveCA()
	}

	p.logger.Println("Loading CA certificate from", certPath)
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read CA cert: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA key: %w", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse CA key pair: %w", err)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	p.caCert = x509Cert
	p.caKey = cert.PrivateKey.(*rsa.PrivateKey)

	p.logger.Println("✅ CA loaded successfully.")
	return nil
}

// generateAndSaveCA generates a new CA and saves it to disk
func (p *HTTPSProxy) generateAndSaveCA() error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Axom AI Observer CA"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return err
	}

	p.caCert = cert
	p.caKey = privateKey

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll("certs", 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Save certificate to file
	certOut, err := os.Create("certs/ca.crt")
	if err != nil {
		return fmt.Errorf("failed to open ca.crt for writing: %w", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	p.logger.Println("📄 CA certificate saved to certs/ca.crt")

	// Save key to file
	keyOut, err := os.OpenFile("certs/ca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open ca.key for writing: %w", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	keyOut.Close()
	p.logger.Println("🔑 CA private key saved to certs/ca.key")

	return nil
}

// handleRequest handles incoming HTTPS requests
func (p *HTTPSProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Handle CONNECT method for HTTPS tunneling
	if r.Method == "CONNECT" {
		p.handleCONNECT(w, r)
		return
	}

	// Handle regular HTTPS requests
	p.handleHTTPSRequest(w, r)
}

// handleCONNECT handles CONNECT requests for HTTPS tunneling
func (p *HTTPSProxy) handleCONNECT(w http.ResponseWriter, r *http.Request) {
	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Send 200 OK to client
	clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	// Create TLS config for the client connection
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{p.generateCert(r.Host)},
	}

	// Upgrade client connection to TLS
	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	// Handle the TLS connection
	p.handleTLSConnection(tlsConn, r.Host)
}

// handleTLSConnection handles TLS connections
func (p *HTTPSProxy) handleTLSConnection(tlsConn *tls.Conn, host string) {
	// Accept the TLS connection
	if err := tlsConn.Handshake(); err != nil {
		p.logger.Printf("TLS handshake failed: %v", err)
		return
	}

	// Read HTTP request from TLS connection
	req, err := http.ReadRequest(bufio.NewReader(tlsConn))
	if err != nil {
		p.logger.Printf("Failed to read request from TLS: %v", err)
		return
	}

	// Set the host
	req.URL.Host = host
	req.URL.Scheme = "https"

	// Handle the request
	p.processHTTPSRequest(req, tlsConn)
}

// handleHTTPSRequest handles regular HTTPS requests
func (p *HTTPSProxy) handleHTTPSRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Check if this is an AI API call
	aiProvider := p.detectAIProvider(r.URL.Host, r.URL.Path)
	if aiProvider == nil {
		// Not an AI API call, forward as-is
		p.forwardHTTPSRequest(w, r)
		return
	}

	// Capture request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		p.logger.Printf("Failed to read request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Parse AI request
	aiRequest := p.parseAIRequest(r, bodyBytes, aiProvider)

	// Forward request to actual AI service
	resp, err := p.forwardAIRequest(r, bodyBytes)
	if err != nil {
		p.logger.Printf("Failed to forward AI request: %v", err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Capture response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Printf("Failed to read response body: %v", err)
	}

	// Parse AI response
	aiResponse := p.parseAIResponse(respBodyBytes, aiProvider)

	// Calculate latency
	latency := time.Since(startTime)

	// Create signal
	signal := p.createSignal(r, aiRequest, aiResponse, resp.StatusCode, latency, aiProvider)

	// Detect task if this is a new task
	if task := p.taskDetector.DetectTask(signal); task != nil {
		signal.TaskID = task.ID
		signal.TaskType = task.Type
		signal.Metadata["task_confidence"] = task.Metadata["confidence"]
	}

	// Send signal
	select {
	case p.signalCh <- signal:
		p.logger.Printf("📡 HTTPS AI signal captured: %s %s -> %s (latency: %.2fms)",
			aiProvider.Name, signal.Operation, r.URL.Host, signal.LatencyMS)
	default:
		p.logger.Printf("Signal channel full, dropping signal")
	}

	// Return response to client
	w.WriteHeader(resp.StatusCode)
	w.Write(respBodyBytes)
}

// processHTTPSRequest processes HTTPS requests
func (p *HTTPSProxy) processHTTPSRequest(req *http.Request, tlsConn *tls.Conn) {
	startTime := time.Now()

	// Check if this is an AI API call
	aiProvider := p.detectAIProvider(req.URL.Host, req.URL.Path)
	if aiProvider == nil {
		// Not an AI API call, forward as-is
		p.forwardTLSRequest(req, tlsConn)
		return
	}

	// Capture request body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		p.logger.Printf("Failed to read request body: %v", err)
		return
	}
	req.Body.Close()

	// Parse AI request
	aiRequest := p.parseAIRequest(req, bodyBytes, aiProvider)

	// Forward request to actual AI service
	resp, err := p.forwardAIRequest(req, bodyBytes)
	if err != nil {
		p.logger.Printf("Failed to forward AI request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Capture response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Printf("Failed to read response body: %v", err)
	}

	// Parse AI response
	aiResponse := p.parseAIResponse(respBodyBytes, aiProvider)

	// Calculate latency
	latency := time.Since(startTime)

	// Create signal
	signal := p.createSignal(req, aiRequest, aiResponse, resp.StatusCode, latency, aiProvider)

	// Detect task if this is a new task
	if task := p.taskDetector.DetectTask(signal); task != nil {
		signal.TaskID = task.ID
		signal.TaskType = task.Type
		signal.Metadata["task_confidence"] = task.Metadata["confidence"]
	}

	// Send signal
	select {
	case p.signalCh <- signal:
		p.logger.Printf("📡 TLS AI signal captured: %s %s -> %s (latency: %.2fms)",
			aiProvider.Name, signal.Operation, req.URL.Host, signal.LatencyMS)
	default:
		p.logger.Printf("Signal channel full, dropping signal")
	}

	// Write response to TLS connection
	resp.Write(tlsConn)
}

// generateCert generates a certificate for the given hostname
func (p *HTTPSProxy) generateCert(hostname string) tls.Certificate {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		p.logger.Printf("Failed to generate private key: %v", err)
		return tls.Certificate{}
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Axom AI Observer"},
			Country:      []string{"US"},
		},
		DNSNames:    []string{hostname},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, p.caCert, &privateKey.PublicKey, p.caKey)
	if err != nil {
		p.logger.Printf("Failed to create certificate: %v", err)
		return tls.Certificate{}
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		p.logger.Printf("Failed to parse certificate: %v", err)
		return tls.Certificate{}
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}
}

// detectAIProvider detects which AI provider this request is for
func (p *HTTPSProxy) detectAIProvider(host, path string) *AIProvider {
	for _, provider := range knownAIProviders {
		for _, domain := range provider.Domains {
			// Handle wildcard domains for services like Azure
			matchPattern := strings.ReplaceAll(domain, "*", "")
			if (strings.HasPrefix(domain, "*") && strings.HasSuffix(host, matchPattern)) ||
				(!strings.HasPrefix(domain, "*") && host == domain) {
				for _, apiPattern := range provider.APIPatterns {
					if strings.HasPrefix(path, apiPattern) {
						return &provider
					}
				}
			}
		}
	}
	return nil
}

// parseAIRequest parses the AI request based on provider
func (p *HTTPSProxy) parseAIRequest(r *http.Request, bodyBytes []byte, provider *AIProvider) map[string]interface{} {
	request := make(map[string]interface{})

	// Common fields
	request["provider"] = provider.Name
	request["endpoint"] = r.URL.Path
	request["method"] = r.Method

	// Parse JSON body if available
	if len(bodyBytes) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			// Extract model
			if model, ok := jsonData["model"].(string); ok {
				request["model"] = model
			}

			// Extract messages for chat completions
			if messages, ok := jsonData["messages"].([]interface{}); ok {
				request["messages"] = messages
				if len(messages) > 0 {
					if msg, ok := messages[0].(map[string]interface{}); ok {
						if content, ok := msg["content"].(string); ok {
							request["prompt_preview"] = p.truncateString(content, 100)
						}
					}
				}
			}

			// Extract other common fields
			for _, field := range []string{"max_tokens", "temperature", "top_p", "frequency_penalty", "presence_penalty"} {
				if value, ok := jsonData[field]; ok {
					request[field] = value
				}
			}

			// Provider-specific parsing
			switch provider.Name {
			case "OpenAI":
				p.parseOpenAIRequest(request, jsonData)
			case "Anthropic":
				p.parseAnthropicRequest(request, jsonData)
			case "Google AI":
				p.parseGoogleAIRequest(request, jsonData)
			}
		}
	}

	return request
}

// parseAIResponse parses the AI response based on provider
func (p *HTTPSProxy) parseAIResponse(bodyBytes []byte, provider *AIProvider) map[string]interface{} {
	response := make(map[string]interface{})

	if len(bodyBytes) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			// Extract usage information
			if usage, ok := jsonData["usage"].(map[string]interface{}); ok {
				response["usage"] = usage
			}

			// Extract choices/response
			if choices, ok := jsonData["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if message, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := message["content"].(string); ok {
							response["response_preview"] = p.truncateString(content, 100)
						}
					}
				}
			}

			// Provider-specific parsing
			switch provider.Name {
			case "OpenAI":
				p.parseOpenAIResponse(response, jsonData)
			case "Anthropic":
				p.parseAnthropicResponse(response, jsonData)
			}
		}
	}

	return response
}

// parseOpenAIRequest parses OpenAI-specific request fields
func (p *HTTPSProxy) parseOpenAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific fields
	if stream, ok := jsonData["stream"].(bool); ok {
		request["stream"] = stream
	}
	if n, ok := jsonData["n"].(float64); ok {
		request["n"] = int(n)
	}
}

// parseAnthropicRequest parses Anthropic-specific request fields
func (p *HTTPSProxy) parseAnthropicRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Anthropic-specific fields
	if max_tokens, ok := jsonData["max_tokens"].(float64); ok {
		request["max_tokens"] = int(max_tokens)
	}
	if system, ok := jsonData["system"].(string); ok {
		request["system"] = system
	}
}

// parseGoogleAIRequest parses Google AI-specific request fields
func (p *HTTPSProxy) parseGoogleAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Google AI-specific fields
	if generationConfig, ok := jsonData["generationConfig"].(map[string]interface{}); ok {
		request["generation_config"] = generationConfig
	}
}

// parseOpenAIResponse parses OpenAI-specific response fields
func (p *HTTPSProxy) parseOpenAIResponse(response map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific response parsing
	if id, ok := jsonData["id"].(string); ok {
		response["id"] = id
	}
}

// parseAnthropicResponse parses Anthropic-specific response fields
func (p *HTTPSProxy) parseAnthropicResponse(response map[string]interface{}, jsonData map[string]interface{}) {
	// Anthropic-specific response parsing
	if content, ok := jsonData["content"].([]interface{}); ok && len(content) > 0 {
		if contentItem, ok := content[0].(map[string]interface{}); ok {
			if text, ok := contentItem["text"].(string); ok {
				response["response_preview"] = p.truncateString(text, 100)
			}
		}
	}
}

// createSignal creates a signal from the AI request/response
func (p *HTTPSProxy) createSignal(
	r *http.Request,
	request map[string]interface{},
	response map[string]interface{},
	statusCode int,
	latency time.Duration,
	provider *AIProvider,
) models.Signal {

	// Determine operation type
	operation := p.determineOperation(r.URL.Path, request, provider)

	// Extract metadata
	metadata := make(map[string]interface{})
	for k, v := range request {
		metadata[k] = v
	}
	for k, v := range response {
		metadata[k] = v
	}

	// Add provider information
	metadata["provider"] = provider.Name
	metadata["endpoint"] = r.URL.Path

	// Extract usage information
	if usage, ok := response["usage"].(map[string]interface{}); ok {
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			metadata["prompt_tokens"] = int(promptTokens)
		}
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			metadata["completion_tokens"] = int(completionTokens)
		}
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			metadata["total_tokens"] = int(totalTokens)
		}
	}

	return models.Signal{
		ID:          p.generateSignalID(),
		CustomerID:  p.customerID,
		AgentID:     p.agentID,
		Timestamp:   time.Now(),
		Protocol:    "https",
		LatencyMS:   float64(latency.Milliseconds()),
		Metadata:    metadata,
		Source:      models.Endpoint{IP: "127.0.0.1", Port: 0},
		Destination: models.Endpoint{IP: r.URL.Host, Port: 443},
		Operation:   operation,
		Status:      statusCode,
	}
}

// determineOperation determines the operation type
func (p *HTTPSProxy) determineOperation(path string, request map[string]interface{}, provider *AIProvider) string {
	// Check path patterns
	if strings.Contains(path, "/chat/completions") || strings.Contains(path, "/messages") {
		return "chat_completion"
	}
	if strings.Contains(path, "/completions") || strings.Contains(path, "/generate") {
		return "text_completion"
	}
	if strings.Contains(path, "/embeddings") || strings.Contains(path, "/embed") {
		return "embedding"
	}
	if strings.Contains(path, "/images/generations") {
		return "image_generation"
	}
	if strings.Contains(path, "/audio/transcriptions") {
		return "audio_transcription"
	}
	if strings.Contains(path, "/audio/translations") {
		return "audio_translation"
	}
	if strings.Contains(path, "/moderations") {
		return "moderation"
	}

	// Default based on provider
	return "ai_request"
}

// forwardAIRequest forwards the request to the actual AI service
func (p *HTTPSProxy) forwardAIRequest(r *http.Request, bodyBytes []byte) (*http.Response, error) {
	// Create new request to actual AI service
	req, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Copy headers
	req.Header = r.Header

	// Create HTTP client with TLS
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	return client.Do(req)
}

// forwardHTTPSRequest forwards non-AI HTTPS requests
func (p *HTTPSProxy) forwardHTTPSRequest(w http.ResponseWriter, r *http.Request) {
	// Simple forwarding for non-AI requests
	http.Error(w, "Not an AI API endpoint", http.StatusNotFound)
}

// forwardTLSRequest forwards TLS requests
func (p *HTTPSProxy) forwardTLSRequest(req *http.Request, tlsConn *tls.Conn) {
	// Forward to actual service
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		p.logger.Printf("Failed to forward TLS request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Copy response to TLS connection
	resp.Write(tlsConn)
}

// generateSignalID generates a unique signal ID
func (p *HTTPSProxy) generateSignalID() string {
	return fmt.Sprintf("signal_%d", time.Now().UnixNano())
}

// truncateString truncates a string to max length
func (p *HTTPSProxy) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
