package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/app"
	"github.com/jo-hoe/whatsapp-reminder/internal/config"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/reminder"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file (default: /app/config.yaml or ./config.yaml)")
	flag.Parse()

	// If config path not provided via flag, check environment variable
	if *configPath == "" {
		if envConfigPath := os.Getenv("CONFIG_PATH"); envConfigPath != "" {
			*configPath = envConfigPath
		}
	}

	log.Println("Starting WhatsApp Reminder Container...")

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	// Load configuration from YAML file
	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")

	// Create app configuration with cancellable context
	appConfig, err := createAppConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create app configuration: %v", err)
	}

	// Validate configuration and perform health checks
	if err := validateAndHealthCheck(ctx, appConfig); err != nil {
		log.Fatalf("Validation or health check failed: %v", err)
	}

	// Execute reminder once and exit
	log.Println("Executing reminder...")
	if err := runReminder(appConfig); err != nil {
		// Check if it was a graceful shutdown
		if ctx.Err() == context.Canceled {
			log.Println("Reminder execution cancelled due to shutdown signal")
			os.Exit(0)
		}
		log.Printf("Reminder execution failed: %v", err)
		os.Exit(1)
	}

	log.Println("Container execution completed successfully. Exiting.")
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

func validateAndHealthCheck(ctx context.Context, appConfig *app.AppConfig) error {
	log.Println("Validating configuration...")

	// Check required fields
	if appConfig.SpreadSheetId == "" {
		return fmt.Errorf("spreadsheet ID is required")
	}
	if appConfig.SheetName == "" {
		return fmt.Errorf("sheet name is required")
	}
	if appConfig.MailServiceURL == "" {
		return fmt.Errorf("mail service URL is required")
	}
	if appConfig.OriginAddress == "" {
		return fmt.Errorf("origin address is required")
	}
	if appConfig.OriginName == "" {
		return fmt.Errorf("origin name is required")
	}
	if len(appConfig.ServiceAccountSecret) == 0 {
		return fmt.Errorf("service account secret is required")
	}

	log.Println("Configuration validated successfully")

	// Perform health check on mail service with cancellable context
	log.Printf("Checking mail service health at %s...", appConfig.MailServiceURL)

	mailClient := reminder.NewMailClient(appConfig.MailServiceURL, 10*time.Second)

	healthCheckCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := mailClient.HealthCheck(healthCheckCtx); err != nil {
		return fmt.Errorf("mail service health check failed: %w", err)
	}

	log.Println("Mail service is healthy")
	return nil
}

func runReminder(config *app.AppConfig) error {
	start := time.Now()
	err := app.Start(config)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Reminder execution failed after %v: %v", duration, err)
		return err
	}

	log.Printf("Reminder execution completed successfully in %v", duration)
	return nil
}
