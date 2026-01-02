package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bsko/casino-transaction-system/internal/app/consumer"
)

func main() {
	ctx := context.Background()

	app := &consumer.ConsumerApp{}

	if err := app.Initialize(ctx); err != nil {
		if shutdownErr := app.Shutdown(ctx); shutdownErr != nil {
			log.Printf("Shutdown error after init failure: %v", shutdownErr)
		}
		log.Fatalf("Failed to initialize application: %v", err)
	}

	defer func() {
		if err := app.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		} else {
			log.Println("Application shutdown completed successfully")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := app.Exec(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down gracefully...", sig)
	case err := <-errChan:
		log.Printf("Application error: %v. Shutting down...", err)
	}
}
