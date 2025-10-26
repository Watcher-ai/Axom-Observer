package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"axom-observer/pkg/models"
	"axom-observer/pkg/observer"
)

// getEnvWithDefault gets environment variable with fallback
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Parse command line flags
	var (
		customerID   = flag.String("customer-id", getEnvWithDefault("CUSTOMER_ID", ""), "Customer identifier (Agent Name)")
		agentID      = flag.String("agent-id", getEnvWithDefault("AGENT_ID", ""), "AI agent identifier")
		clientID     = flag.String("client-id", getEnvWithDefault("CLIENT_ID", ""), "Client ID for authentication")
		clientSecret = flag.String("client-secret", getEnvWithDefault("CLIENT_SECRET", ""), "Client Secret for authentication")
		agentSecret  = flag.String("agent-secret", getEnvWithDefault("AGENT_SECRET", ""), "Agent Secret for API authentication")
		backendURL   = flag.String("backend-url", getEnvWithDefault("BACKEND_URL", "http://localhost:8080/api/v1/signals"), "Backend URL for signals")
		httpPort     = flag.String("http-port", "8888", "HTTP proxy port")
		httpsPort    = flag.String("https-port", "8443", "HTTPS proxy port")
	)
	flag.Parse()

	// Validate required fields
	if *customerID == "" || *agentID == "" || *clientID == "" || *clientSecret == "" || *agentSecret == "" {
		logger := log.New(os.Stdout, "observer: ", log.LstdFlags)
		logger.Println("❌ Missing required configuration!")
		logger.Println("Please provide the following environment variables:")
		logger.Println("  CUSTOMER_ID    - Your Agent Name")
		logger.Println("  AGENT_ID       - Your Agent ID")
		logger.Println("  CLIENT_ID      - Your Client ID")
		logger.Println("  CLIENT_SECRET  - Your Client Secret")
		logger.Println("  AGENT_SECRET   - Your Agent Secret (API Key)")
		logger.Println("")
		logger.Println("You can set these by:")
		logger.Println("  1. Creating a .env file with your values")
		logger.Println("  2. Running: docker-compose up")
		logger.Println("  3. Or using export commands")
		os.Exit(1)
	}

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger := log.New(os.Stdout, "observer: ", log.LstdFlags)
	logger.Println("🚀 Starting Axom AI Observer")
	logger.Printf("📡 Customer ID: %s", *customerID)
	logger.Printf("🤖 Agent ID: %s", *agentID)
	logger.Printf("🔑 Client ID: %s", *clientID)
	logger.Printf("🔐 Client Secret: %s", maskSecret(*clientSecret))
	logger.Printf("🔑 Agent Secret: %s", maskSecret(*agentSecret))
	logger.Printf("🌐 Backend URL: %s", *backendURL)
	logger.Printf("🔗 HTTP Port: %s", *httpPort)
	logger.Printf("🔒 HTTPS Port: %s", *httpsPort)

	// Create signal channel
	signalCh := make(chan models.Signal, 100)

	// Create comprehensive AI traffic monitor
	aiMonitor := observer.NewAITrafficMonitor(signalCh, logger, *customerID, *agentID)

	// Create signal sender
	signalSender := observer.NewSignalSender(
		*agentSecret,  // Use agent secret as API key for authentication
		*backendURL,   // Backend URL
		10,            // Batch size
		5*time.Second, // Flush interval
	)

	// Start AI traffic monitor
	if err := aiMonitor.Start(ctx); err != nil {
		logger.Fatalf("Failed to start AI traffic monitor: %v", err)
	}

	// Start signal processing
	go processSignals(ctx, signalCh, signalSender)

	logger.Println("✅ Observer started successfully")
	logger.Printf("📡 Listening for AI API traffic on HTTP port %s and HTTPS port %s", *httpPort, *httpsPort)
	logger.Printf("📊 Sending signals to backend at %s", *backendURL)
	logger.Println("🔍 Monitoring all major AI providers: OpenAI, Anthropic, Google AI, Cohere, Hugging Face, Azure OpenAI")

	<-ctx.Done()
	logger.Println("🛑 Shutdown initiated...")

	// Stop AI traffic monitor
	if err := aiMonitor.Stop(ctx); err != nil {
		logger.Printf("Error stopping AI traffic monitor: %v", err)
	}

	time.Sleep(1 * time.Second) // Allow final flush
}

func processSignals(
	ctx context.Context,
	signalCh <-chan models.Signal,
	sender *observer.SignalSender,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-signalCh:
			log.Printf("📡 Processing signal: %s %s -> %s (latency: %.2fms)",
				sig.Protocol, sig.Operation, sig.Destination.IP, sig.LatencyMS)

			// Extract provider information
			if provider, ok := sig.Metadata["provider"].(string); ok {
				log.Printf("🤖 AI Provider: %s", provider)
			}

			// Extract model information
			if model, ok := sig.Metadata["model"].(string); ok {
				log.Printf("🧠 Model: %s", model)
			}

			// Extract token usage
			if totalTokens, ok := sig.Metadata["total_tokens"].(int); ok {
				log.Printf("🔢 Total Tokens: %d", totalTokens)
			}

			if sig.IsTaskComplete() {
				log.Printf("✅ Task completed: %s - Outcome: %s", sig.TaskID, sig.Outcome)
			}

			if err := sender.Send(sig); err != nil {
				log.Printf("❌ Failed to send signal: %v", err)
			} else {
				log.Printf("✅ Signal sent successfully")
			}
		}
	}
}

// maskSecret masks sensitive information for logging
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
