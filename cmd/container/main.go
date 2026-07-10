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
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file (default: /app/config.yaml or ./config.yaml)")
	flag.Parse()

	if *configPath == "" {
		if envConfigPath := os.Getenv("CONFIG_PATH"); envConfigPath != "" {
			*configPath = envConfigPath
		}
	}

	log.Println("starting WhatsApp Reminder container...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.Printf("configuration loaded successfully")

	appConfig, err := createAppConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app configuration: %v", err)
	}

	if err := validateConfig(appConfig); err != nil {
		log.Fatalf("configuration validation failed: %v", err)
	}

	log.Println("executing reminder...")
	if err := runReminder(appConfig); err != nil {
		if ctx.Err() == context.Canceled {
			log.Println("reminder execution cancelled due to shutdown signal")
			os.Exit(0)
		}
		log.Printf("reminder execution failed: %v", err)
		os.Exit(1)
	}

	log.Println("container execution completed successfully, exiting")
}

func createAppConfig(ctx context.Context, cfg *config.Config) (*app.AppConfig, error) {
	timeLocation, err := time.LoadLocation(cfg.App.TimeLocation)
	if err != nil {
		return nil, err
	}

	retentionTime, err := time.ParseDuration(cfg.App.RetentionTime)
	if err != nil {
		return nil, err
	}

	serviceAccountSecret, err := cfg.GetServiceAccountSecret()
	if err != nil {
		return nil, err
	}

	return &app.AppConfig{
		Ctx:                  ctx,
		SpreadSheetId:        cfg.GoogleSheets.SpreadsheetID,
		SheetName:            cfg.GoogleSheets.SheetName,
		ServiceAccountSecret: serviceAccountSecret,
		RetentionTime:        retentionTime,
		TimeLocation:         timeLocation,
		Email:                cfg.Email,
	}, nil
}

func validateConfig(appConfig *app.AppConfig) error {
	if appConfig.SpreadSheetId == "" {
		return fmt.Errorf("spreadsheet ID is required")
	}
	if appConfig.SheetName == "" {
		return fmt.Errorf("sheet name is required")
	}
	if len(appConfig.ServiceAccountSecret) == 0 {
		return fmt.Errorf("service account secret is required")
	}
	return nil
}

func runReminder(config *app.AppConfig) error {
	start := time.Now()
	err := app.Start(config)
	duration := time.Since(start)

	if err != nil {
		log.Printf("reminder execution failed after %v: %v", duration, err)
		return err
	}

	log.Printf("reminder execution completed successfully in %v", duration)
	return nil
}
