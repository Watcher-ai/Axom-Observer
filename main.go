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

func main() {
	// Parse command line flags
	var (
		customerID = flag.String("customer-id", "default-customer", "Customer identifier")
		agentID    = flag.String("agent-id", "default-agent", "AI agent identifier")
		backendURL = flag.String("backend-url", "http://localhost:8080/api/v1/signals", "Backend URL for signals")
		httpPort   = flag.String("http-port", "8888", "HTTP proxy port")
		httpsPort  = flag.String("https-port", "8443", "HTTPS proxy port")
	)
	flag.Parse()

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
	logger.Printf("🌐 Backend URL: %s", *backendURL)
	logger.Printf("🔗 HTTP Port: %s", *httpPort)
	logger.Printf("🔒 HTTPS Port: %s", *httpsPort)

	// Create signal channel
	signalCh := make(chan models.Signal, 100)

	// Create comprehensive AI traffic monitor
	aiMonitor := observer.NewAITrafficMonitor(signalCh, logger, *customerID, *agentID)

	// Create signal sender
	signalSender := observer.NewSignalSender(
		"dummy-token", // API key (not used in current implementation)
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
