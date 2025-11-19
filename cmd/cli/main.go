package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/app"
	"github.com/jo-hoe/whatsapp-reminder/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file (default: ./config.yaml)")
	flag.Parse()

	log.Println("starting WhatsApp Reminder CLI...")

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		log.Printf("received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	// Load configuration from YAML file
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.Printf("configuration loaded successfully")

	// Create app configuration with cancellable context
	appConfig, err := createAppConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app configuration: %v", err)
	}

	// Run the reminder
	log.Println("running reminder...")
	start := time.Now()
	err = app.Start(appConfig)
	duration := time.Since(start)

	if err != nil {
		// Check if it was a graceful shutdown
		if ctx.Err() == context.Canceled {
			log.Printf("reminder execution cancelled after %v due to shutdown signal", duration)
			os.Exit(0)
		}
		log.Printf("reminder execution failed after %v: %v", duration, err)
		os.Exit(1)
	}

	log.Printf("reminder execution completed successfully in %v", duration)
}

func createAppConfig(ctx context.Context, config *config.Config) (*app.AppConfig, error) {
	// Parse time location
	timeLocation, err := time.LoadLocation(config.App.TimeLocation)
	if err != nil {
		return nil, err
	}

	// Parse retention time
	retentionTime, err := time.ParseDuration(config.App.RetentionTime)
	if err != nil {
		return nil, err
	}

	// Get service account secret
	serviceAccountSecret, err := config.GetServiceAccountSecret()
	if err != nil {
		return nil, err
	}

	return &app.AppConfig{
		Ctx:                  ctx,
		SpreadSheetId:        config.GoogleSheets.SpreadsheetID,
		SheetName:            config.GoogleSheets.SheetName,
		ServiceAccountSecret: serviceAccountSecret,
		RetentionTime:        retentionTime,
		TimeLocation:         timeLocation,
		MailServiceURL:       config.Email.ServiceURL,
		OriginAddress:        config.GetEmailOriginAddress(),
		OriginName:           config.GetEmailOriginName(),
	}, nil
}
