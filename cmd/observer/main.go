package main

import (
	"axom-observer/pkg/config"
	"axom-observer/pkg/models"
	"axom-observer/pkg/observer"
	"context"
	"log"
	"os"
	"os/signal"
)

func main() {
	rules, err := config.LoadRules("/etc/axom/rules.yaml")
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	signalCh := make(chan models.Signal, 1000)

	sniffer := observer.NewTrafficSniffer(rules, signalCh)
	classifier := observer.NewBehaviorClassifier(rules)
	sender := observer.NewSignalSender(os.Getenv("AXOM_API_KEY"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go sniffer.Start(ctx)
	go sender.Start(ctx, signalCh)

	for {
		select {
		case sig := <-signalCh:
			alerts := classifier.Analyze(sig)
			if len(alerts) > 0 {
				sig.Alerts = alerts
			}
			sender.Send(sig)
		case <-ctx.Done():
			log.Println("Shutting down observer...")
			return
		}
	}
}
