package observer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"axom-observer/pkg/models"
)

// AITrafficMonitor provides comprehensive AI traffic monitoring
type AITrafficMonitor struct {
	httpProxy       *HTTPProxy
	productionProxy *ProductionProxy
	taskDetector    *TaskDetector
	logger          *log.Logger
	signalCh        chan<- models.Signal
	customerID      string
	agentID         string
}

// AIProvider represents an AI service provider
type AIProvider struct {
	Name        string
	Domains     []string
	APIPatterns []string
	Models      []string
	TaskTypes   []string
}

// Known AI providers and their patterns
var knownAIProviders = []AIProvider{
	// LLM Providers
	{
		Name:    "OpenAI",
		Domains: []string{"api.openai.com"},
		APIPatterns: []string{
			"/v1/chat/completions", "/v1/completions", "/v1/embeddings",
			"/v1/images/generations", "/v1/audio/transcriptions",
			"/v1/audio/translations", "/v1/moderations",
		},
	},
	{
		Name:    "Anthropic",
		Domains: []string{"api.anthropic.com"},
		APIPatterns: []string{
			"/v1/messages", "/v1/complete",
		},
	},
	{
		Name:    "Google AI",
		Domains: []string{"generativelanguage.googleapis.com", "aiplatform.googleapis.com"},
		APIPatterns: []string{
			"/v1beta/models", "/v1/projects",
		},
	},
	{
		Name:    "Cohere",
		Domains: []string{"api.cohere.ai"},
		APIPatterns: []string{
			"/v1/generate", "/v1/embed", "/v1/classify", "/v1/summarize",
		},
	},
	{
		Name:    "Together AI",
		Domains: []string{"api.together.ai"},
		APIPatterns: []string{
			"/v1/chat/completions", "/v1/completions", "/v1/embeddings", "/inference",
		},
	},
	{
		Name:    "Groq",
		Domains: []string{"api.groq.com"},
		APIPatterns: []string{
			"/openai/v1/chat/completions",
		},
	},
	{
		Name:    "Hugging Face",
		Domains: []string{"api-inference.huggingface.co"},
		APIPatterns: []string{
			"/models/",
		},
	},
	{
		Name:    "Azure OpenAI",
		Domains: []string{"*.openai.azure.com"},
		APIPatterns: []string{
			"/openai/deployments/",
		},
	},
	// STT (Speech-to-Text) Providers
	{
		Name:    "Deepgram",
		Domains: []string{"api.deepgram.com"},
		APIPatterns: []string{
			"/v1/listen", "/v1/speak",
		},
	},
	{
		Name:    "AssemblyAI",
		Domains: []string{"api.assemblyai.com"},
		APIPatterns: []string{
			"/v2/transcript", "/v2/realtime",
		},
	},
	// TTS (Text-to-Speech) Providers
	{
		Name:    "ElevenLabs",
		Domains: []string{"api.elevenlabs.io"},
		APIPatterns: []string{
			"/v1/text-to-speech", "/v1/speech-synthesis",
		},
	},
	{
		Name:    "PlayHT",
		Domains: []string{"api.play.ht"},
		APIPatterns: []string{
			"/api/v2/tts", "/api/v1/convert",
		},
	},
	{
		Name:    "Amazon Polly",
		Domains: []string{"polly.*.amazonaws.com"},
		APIPatterns: []string{
			"/v1/speech",
		},
	},
	{
		Name:    "Azure TTS",
		Domains: []string{"*.cognitiveservices.azure.com"},
		APIPatterns: []string{
			"/sts/v1.0/issueToken", "/cognitiveservices/v1",
		},
	},
	{
		Name:    "Dubverse",
		Domains: []string{"api.dubverse.ai"},
		APIPatterns: []string{
			"/v1/text-to-speech", "/v1/dubbing",
		},
	},
	{
		Name:    "Sarvam AI",
		Domains: []string{"api.sarvam.ai"},
		APIPatterns: []string{
			"/v1/voice/tts", "/v1/llm/o/v1/chat/completions",
		},
	},
	// Phone / Streaming Service Providers
	{
		Name:    "Twilio",
		Domains: []string{"api.twilio.com"},
		APIPatterns: []string{
			"/2010-04-01/Accounts",
		},
	},
	{
		Name:    "Plivo",
		Domains: []string{"api.plivo.com"},
		APIPatterns: []string{
			"/v1/Account",
		},
	},
	{
		Name:    "Vonage",
		Domains: []string{"api.nexmo.com", "api.vonage.com"},
		APIPatterns: []string{
			"/v1/calls", "/v1/voice",
		},
	},
	{
		Name:    "Daily",
		Domains: []string{"api.daily.co"},
		APIPatterns: []string{
			"/v1/rooms", "/v1/meetings",
		},
	},
	{
		Name:    "100ms",
		Domains: []string{"api.100ms.live"},
		APIPatterns: []string{
			"/v2/rooms", "/v2/sessions",
		},
	},
	// Local and Test Services
	{
		Name: "Local AI Services",
		Domains: []string{
			"localhost",
			"127.0.0.1",
			"0.0.0.0",
		},
		APIPatterns: []string{
			"/v1/chat/completions",
			"/v1/completions",
			"/v1/embeddings",
			"/v1/models",
			"/chat",
			"/embed",
		},
	},
}

// NewAITrafficMonitor creates a new AI traffic monitor
func NewAITrafficMonitor(signalCh chan<- models.Signal, logger *log.Logger, customerID, agentID string) *AITrafficMonitor {
	return &AITrafficMonitor{
		logger:       logger,
		signalCh:     signalCh,
		customerID:   customerID,
		agentID:      agentID,
		taskDetector: NewTaskDetector(signalCh, logger, customerID, agentID),
	}
}

// Start starts the AI traffic monitor
func (m *AITrafficMonitor) Start(ctx context.Context) error {
	m.logger.Println("🚀 Starting AI Traffic Monitor")

	// Start HTTP proxy
	m.httpProxy = NewHTTPProxy("8888", m.signalCh, m.logger, m.customerID, m.agentID)
	if err := m.httpProxy.Start(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP proxy: %w", err)
	}

	// Start Production MITM proxy (replaces old HTTPS proxy)
	m.productionProxy = NewProductionProxy("8443", m.signalCh, m.logger, m.customerID, m.agentID)
	if err := m.productionProxy.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Production MITM proxy: %w", err)
	}

	m.logger.Println("✅ AI Traffic Monitor started successfully")
	return nil
}

// Stop stops the AI traffic monitor
func (m *AITrafficMonitor) Stop(ctx context.Context) error {
	m.logger.Println("🛑 Stopping AI Traffic Monitor")

	if m.httpProxy != nil {
		m.httpProxy.Stop(ctx)
	}
	if m.productionProxy != nil {
		m.productionProxy.Stop(ctx)
	}

	return nil
}

// HTTPProxy handles HTTP traffic
type HTTPProxy struct {
	port         string
	signalCh     chan<- models.Signal
	logger       *log.Logger
	customerID   string
	agentID      string
	taskDetector *TaskDetector
	server       *http.Server
}

// NewHTTPProxy creates a new HTTP proxy
func NewHTTPProxy(port string, signalCh chan<- models.Signal, logger *log.Logger, customerID, agentID string) *HTTPProxy {
	return &HTTPProxy{
		port:         port,
		signalCh:     signalCh,
		logger:       logger,
		customerID:   customerID,
		agentID:      agentID,
		taskDetector: NewTaskDetector(signalCh, logger, customerID, agentID),
	}
}

// Start starts the HTTP proxy
func (p *HTTPProxy) Start(ctx context.Context) error {
	p.logger.Printf("Starting HTTP proxy on port %s", p.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleRequest)

	p.server = &http.Server{
		Addr:    ":" + p.port,
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Printf("HTTP proxy error: %v", err)
		}
	}()

	return nil
}

// Stop stops the HTTP proxy
func (p *HTTPProxy) Stop(ctx context.Context) error {
	if p.server != nil {
		return p.server.Shutdown(ctx)
	}
	return nil
}

// handleRequest handles incoming HTTP requests
func (p *HTTPProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Check if this is an AI API call
	aiProvider := p.detectAIProvider(r.URL.Host, r.URL.Path)
	if aiProvider == nil {
		// Not an AI API call, forward as-is
		p.forwardRequest(w, r)
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
		p.logger.Printf("📡 AI signal captured: %s %s -> %s (latency: %.2fms)",
			aiProvider.Name, signal.Operation, r.URL.Host, signal.LatencyMS)
	default:
		p.logger.Printf("Signal channel full, dropping signal")
	}

	// Return response to client
	w.WriteHeader(resp.StatusCode)
	w.Write(respBodyBytes)
}

// detectAIProvider detects which AI provider this request is for
func (p *HTTPProxy) detectAIProvider(host, path string) *AIProvider {
	// First check if this is a direct request to the observer (proxy scenario)
	// In this case, we detect based on path patterns only
	if strings.Contains(host, "localhost") && (strings.Contains(host, "8888") || strings.Contains(host, "8443")) {
		for _, provider := range knownAIProviders {
			for _, pattern := range provider.APIPatterns {
				if strings.Contains(path, pattern) {
					return &provider
				}
			}
		}
	}

	// Original logic for direct AI provider detection
	for _, provider := range knownAIProviders {
		for _, domain := range provider.Domains {
			if strings.Contains(host, strings.ReplaceAll(domain, "*", "")) {
				for _, pattern := range provider.APIPatterns {
					if strings.Contains(path, pattern) {
						return &provider
					}
				}
			}
		}
	}
	return nil
}

// parseAIRequest parses the AI request based on provider
func (p *HTTPProxy) parseAIRequest(r *http.Request, bodyBytes []byte, provider *AIProvider) map[string]interface{} {
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
func (p *HTTPProxy) parseAIResponse(bodyBytes []byte, provider *AIProvider) map[string]interface{} {
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
func (p *HTTPProxy) parseOpenAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific fields
	if stream, ok := jsonData["stream"].(bool); ok {
		request["stream"] = stream
	}
	if n, ok := jsonData["n"].(float64); ok {
		request["n"] = int(n)
	}
}

// parseAnthropicRequest parses Anthropic-specific request fields
func (p *HTTPProxy) parseAnthropicRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Anthropic-specific fields
	if max_tokens, ok := jsonData["max_tokens"].(float64); ok {
		request["max_tokens"] = int(max_tokens)
	}
	if system, ok := jsonData["system"].(string); ok {
		request["system"] = system
	}
}

// parseGoogleAIRequest parses Google AI-specific request fields
func (p *HTTPProxy) parseGoogleAIRequest(request map[string]interface{}, jsonData map[string]interface{}) {
	// Google AI-specific fields
	if generationConfig, ok := jsonData["generationConfig"].(map[string]interface{}); ok {
		request["generation_config"] = generationConfig
	}
}

// parseOpenAIResponse parses OpenAI-specific response fields
func (p *HTTPProxy) parseOpenAIResponse(response map[string]interface{}, jsonData map[string]interface{}) {
	// OpenAI-specific response parsing
	if id, ok := jsonData["id"].(string); ok {
		response["id"] = id
	}
}

// parseAnthropicResponse parses Anthropic-specific response fields
func (p *HTTPProxy) parseAnthropicResponse(response map[string]interface{}, jsonData map[string]interface{}) {
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
func (p *HTTPProxy) createSignal(
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
		Protocol:    "http",
		LatencyMS:   float64(latency.Milliseconds()),
		Metadata:    metadata,
		Source:      models.Endpoint{IP: "127.0.0.1", Port: 0},
		Destination: models.Endpoint{IP: r.URL.Host, Port: 443},
		Operation:   operation,
		Status:      statusCode,
	}
}

// determineOperation determines the operation type
func (p *HTTPProxy) determineOperation(path string, request map[string]interface{}, provider *AIProvider) string {
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
func (p *HTTPProxy) forwardAIRequest(r *http.Request, bodyBytes []byte) (*http.Response, error) {
	// Determine the actual AI service URL based on the request
	var targetURL string

	// For localhost requests, forward to the demo app
	if strings.Contains(r.URL.Host, "localhost") || strings.Contains(r.URL.Host, "127.0.0.1") {
		// Forward to demo app on port 5002
		targetURL = fmt.Sprintf("http://localhost:5002%s", r.URL.Path)
	} else {
		// For external services, use the original URL
		targetURL = r.URL.String()
	}

	// Create new request to actual AI service
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Copy headers
	req.Header = r.Header

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	return client.Do(req)
}

// forwardRequest forwards non-AI requests
func (p *HTTPProxy) forwardRequest(w http.ResponseWriter, r *http.Request) {
	// Simple forwarding for non-AI requests
	http.Error(w, "Not an AI API endpoint", http.StatusNotFound)
}

// generateSignalID generates a unique signal ID
func (p *HTTPProxy) generateSignalID() string {
	return fmt.Sprintf("signal_%d", time.Now().UnixNano())
}

// truncateString truncates a string to max length
func (p *HTTPProxy) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
