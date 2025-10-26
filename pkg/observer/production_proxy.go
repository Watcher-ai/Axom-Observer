package observer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"axom-observer/pkg/models"

	"github.com/AdguardTeam/gomitmproxy"
)

// ProductionProxy provides production-grade MITM proxy capabilities
type ProductionProxy struct {
	proxy        *gomitmproxy.Proxy
	signalCh     chan<- models.Signal
	logger       *log.Logger
	customerID   string
	agentID      string
	taskDetector *TaskDetector
	certCache    map[string]*tls.Certificate
	certMutex    sync.RWMutex
}

// NewProductionProxy creates a new production-grade MITM proxy
func NewProductionProxy(port string, signalCh chan<- models.Signal, logger *log.Logger, customerID, agentID string) *ProductionProxy {
	return &ProductionProxy{
		signalCh:     signalCh,
		logger:       logger,
		customerID:   customerID,
		agentID:      agentID,
		taskDetector: NewTaskDetector(signalCh, logger, customerID, agentID),
		certCache:    make(map[string]*tls.Certificate),
	}
}

// Start starts the production proxy
func (p *ProductionProxy) Start(ctx context.Context) error {
	p.logger.Println("🚀 Starting Production MITM Proxy")

	// Create proxy configuration with built-in CA
	config := gomitmproxy.Config{
		ListenAddr: &net.TCPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: 8443, // Default HTTPS port
		},
		OnRequest:  p.handleRequest,
		OnResponse: p.handleResponse,
	}

	// Create proxy instance
	p.proxy = gomitmproxy.NewProxy(config)

	// Start proxy in goroutine
	go func() {
		if err := p.proxy.Start(); err != nil {
			p.logger.Printf("Production proxy error: %v", err)
		}
	}()

	p.logger.Println("✅ Production MITM Proxy started successfully")
	return nil
}

// Stop stops the production proxy
func (p *ProductionProxy) Stop(ctx context.Context) error {
	if p.proxy != nil {
		p.proxy.Close()
	}
	return nil
}

// handleRequest processes incoming requests
func (p *ProductionProxy) handleRequest(session *gomitmproxy.Session) (*http.Request, *http.Response) {
	req := session.Request()
	startTime := time.Now()

	// Try to detect AI provider, but proceed regardless
	aiProvider := p.detectAIProvider(req.URL.Host, req.URL.Path)
	if aiProvider == nil {
		aiProvider = &AIProvider{Name: "Unknown", Domains: []string{req.URL.Host}, APIPatterns: []string{req.URL.Path}}
	}

	p.logger.Printf("📡 Request detected: %s %s -> %s",
		aiProvider.Name, req.Method, req.URL.String())

	// Capture request body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		p.logger.Printf("Failed to read request body: %v", err)
		return nil, nil
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse request
	aiRequest := p.parseAIRequest(req, bodyBytes, aiProvider)

	// Store request data in session for response handling
	session.SetProp("ai_provider", aiProvider)
	session.SetProp("ai_request", aiRequest)
	session.SetProp("start_time", startTime)

	// Pass through the request
	return nil, nil
}

// handleResponse processes outgoing responses
func (p *ProductionProxy) handleResponse(session *gomitmproxy.Session) *http.Response {
	resp := session.Response()
	req := session.Request()

	aiProviderVal, _ := session.GetProp("ai_provider")
	aiProvider, _ := aiProviderVal.(*AIProvider)
	if aiProvider == nil {
		aiProvider = &AIProvider{Name: "Unknown", Domains: []string{req.URL.Host}, APIPatterns: []string{req.URL.Path}}
	}
	startTimeVal, _ := session.GetProp("start_time")
	startTime, ok := startTimeVal.(time.Time)
	if !ok {
		startTime = time.Now()
	}
	aiRequestVal, _ := session.GetProp("ai_request")
	aiRequest, _ := aiRequestVal.(map[string]interface{})
	if aiRequest == nil {
		aiRequest = make(map[string]interface{})
	}

	p.logger.Printf("📡 Response detected: %s %s -> %s (status: %d)",
		aiProvider.Name, req.Method, req.URL.String(), resp.StatusCode)

	// Capture response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Printf("Failed to read response body: %v", err)
		return nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse response
	aiResponse := p.parseAIResponse(bodyBytes, aiProvider)

	// Calculate latency
	latency := time.Since(startTime)

	// Create signal
	signal := p.createSignal(req, aiRequest, aiResponse, resp.StatusCode, latency, aiProvider)

	// Send signal
	select {
	case p.signalCh <- signal:
		p.logger.Printf("📡 Production signal captured: %s %s -> %s (latency: %.2fms)",
			aiProvider.Name, signal.Operation, req.URL.Host, signal.LatencyMS)
	default:
		p.logger.Printf("Signal channel full, dropping signal")
	}

	// Pass through the response
	return nil
}

// detectAIProvider detects which AI provider this request is for
func (p *ProductionProxy) detectAIProvider(host, path string) *AIProvider {
	for _, provider := range knownAIProviders {
		for _, domain := range provider.Domains {
			matchPattern := strings.ReplaceAll(domain, "*", "")
			if strings.Contains(host, matchPattern) {
				for _, apiPattern := range provider.APIPatterns {
					if strings.Contains(path, apiPattern) {
						return &provider
					}
				}
			}
		}
	}
	return nil
}

// parseAIRequest parses the AI request based on provider
func (p *ProductionProxy) parseAIRequest(r *http.Request, bodyBytes []byte, provider *AIProvider) map[string]interface{} {
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
func (p *ProductionProxy) parseAIResponse(bodyBytes []byte, provider *AIProvider) map[string]interface{} {
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
func (p *ProductionProxy) parseOpenAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific fields
	if stream, ok := jsonData["stream"].(bool); ok {
		request["stream"] = stream
	}
	if n, ok := jsonData["n"].(float64); ok {
		request["n"] = int(n)
	}
}

// parseAnthropicRequest parses Anthropic-specific request fields
func (p *ProductionProxy) parseAnthropicRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Anthropic-specific fields
	if max_tokens, ok := jsonData["max_tokens"].(float64); ok {
		request["max_tokens"] = int(max_tokens)
	}
	if system, ok := jsonData["system"].(string); ok {
		request["system"] = system
	}
}

// parseGoogleAIRequest parses Google AI-specific request fields
func (p *ProductionProxy) parseGoogleAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Google AI-specific fields
	if generationConfig, ok := jsonData["generationConfig"].(map[string]interface{}); ok {
		request["generation_config"] = generationConfig
	}
}

// parseOpenAIResponse parses OpenAI-specific response fields
func (p *ProductionProxy) parseOpenAIResponse(response map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific response parsing
	if id, ok := jsonData["id"].(string); ok {
		response["id"] = id
	}
}

// parseAnthropicResponse parses Anthropic-specific response fields
func (p *ProductionProxy) parseAnthropicResponse(response map[string]interface{}, jsonData map[string]interface{}) {
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
func (p *ProductionProxy) createSignal(
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
func (p *ProductionProxy) determineOperation(path string, request map[string]interface{}, provider *AIProvider) string {
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

// generateSignalID generates a unique signal ID
func (p *ProductionProxy) generateSignalID() string {
	return fmt.Sprintf("signal_%d", time.Now().UnixNano())
}

// truncateString truncates a string to max length
func (p *ProductionProxy) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
