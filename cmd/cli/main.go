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
	configPath := flag.String("config", "", "Path to configuration file (default: ./config.yaml)")
	flag.Parse()

	log.Println("starting WhatsApp Reminder CLI...")

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

	log.Println("running reminder...")
	start := time.Now()
	err = app.Start(appConfig)
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.Canceled {
			log.Printf("reminder execution cancelled after %v due to shutdown signal", duration)
			os.Exit(0)
		}
		log.Printf("reminder execution failed after %v: %v", duration, err)
		os.Exit(1)
	}

	log.Printf("reminder execution completed successfully in %v", duration)
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
