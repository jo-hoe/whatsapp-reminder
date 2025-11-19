package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	// Google Sheets configuration
	GoogleSheets GoogleSheetsConfig `yaml:"googleSheets"`

	// Email configuration
	Email EmailConfig `yaml:"email"`

	// Scheduling configuration
	Schedule ScheduleConfig `yaml:"schedule"`

	// Application configuration
	App AppConfig `yaml:"app"`
}

type GoogleSheetsConfig struct {
	SpreadsheetID      string `yaml:"spreadsheetId"`
	SheetName          string `yaml:"sheetName"`
	ServiceAccountFile string `yaml:"serviceAccountFile"`
}

type EmailConfig struct {
	ServiceURL    string `yaml:"serviceUrl"`
	OriginAddress string `yaml:"originAddress,omitempty"`
	OriginName    string `yaml:"originName,omitempty"`
}

type ScheduleConfig struct {
	Interval     string `yaml:"interval"`
	RunOnStartup bool   `yaml:"runOnStartup"`
}

type AppConfig struct {
	TimeLocation  string `yaml:"timeLocation"`
	RetentionTime string `yaml:"retentionTime"`
	LogLevel      string `yaml:"logLevel"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Set default config path if not provided
	if configPath == "" {
		configPath = "/app/config.yaml"
		// Check if config exists in current directory for development
		if _, err := os.Stat("./config.yaml"); err == nil {
			configPath = "./config.yaml"
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Schedule.Interval == "" {
		config.Schedule.Interval = "1h"
	}
	if config.App.TimeLocation == "" {
		config.App.TimeLocation = "UTC"
	}
	if config.App.RetentionTime == "" {
		config.App.RetentionTime = "24h"
	}
	if config.App.LogLevel == "" {
		config.App.LogLevel = "info"
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate checks that all required configuration fields are present
func (c *Config) validate() error {
	if c.GoogleSheets.SpreadsheetID == "" {
		return fmt.Errorf("googleSheets.spreadsheetId is required")
	}
	if c.GoogleSheets.SheetName == "" {
		return fmt.Errorf("googleSheets.sheetName is required")
	}
	if c.GoogleSheets.ServiceAccountFile == "" {
		return fmt.Errorf("googleSheets.serviceAccountFile is required")
	}
	if c.Email.ServiceURL == "" {
		return fmt.Errorf("email.serviceUrl is required")
	}
	// Email origin can come from env vars or config
	if c.GetEmailOriginAddress() == "" {
		return fmt.Errorf("email.originAddress is required (via config or EMAIL_ORIGIN_ADDRESS env var)")
	}
	if c.GetEmailOriginName() == "" {
		return fmt.Errorf("email.originName is required (via config or EMAIL_ORIGIN_NAME env var)")
	}

	// Validate duration formats
	if _, err := time.ParseDuration(c.Schedule.Interval); err != nil {
		return fmt.Errorf("invalid schedule.interval: %w", err)
	}
	if _, err := time.ParseDuration(c.App.RetentionTime); err != nil {
		return fmt.Errorf("invalid app.retentionTime: %w", err)
	}

	// Validate timezone
	if _, err := time.LoadLocation(c.App.TimeLocation); err != nil {
		return fmt.Errorf("invalid app.timeLocation: %w", err)
	}

	return nil
}

// GetServiceAccountSecret returns the service account secret from file
func (c *Config) GetServiceAccountSecret() ([]byte, error) {
	data, err := os.ReadFile(c.GoogleSheets.ServiceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file %s: %w", c.GoogleSheets.ServiceAccountFile, err)
	}
	return data, nil
}

// GetEmailOriginAddress returns the email origin address from config or env var
// Environment variable takes precedence over config file
func (c *Config) GetEmailOriginAddress() string {
	if addr := os.Getenv("EMAIL_ORIGIN_ADDRESS"); addr != "" {
		return addr
	}
	return c.Email.OriginAddress
}

// GetEmailOriginName returns the email origin name from config or env var
// Environment variable takes precedence over config file
func (c *Config) GetEmailOriginName() string {
	if name := os.Getenv("EMAIL_ORIGIN_NAME"); name != "" {
		return name
	}
	return c.Email.OriginName
}
