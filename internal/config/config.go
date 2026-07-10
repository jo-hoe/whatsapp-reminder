package config

import (
	"fmt"
	"os"
	"path/filepath"
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

type SMTPAuthConfig struct {
	Required bool   `yaml:"required"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type EmailConfig struct {
	Host     string         `yaml:"host"`
	Port     int            `yaml:"port"`
	From     string         `yaml:"from"`
	To       []string       `yaml:"to"`
	Auth     SMTPAuthConfig `yaml:"auth"`
	StartTLS bool           `yaml:"startTLS"`
	Timeout  time.Duration  `yaml:"timeout"`
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

	data, err := os.ReadFile(filepath.Clean(configPath))
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
	if config.Email.Port == 0 {
		config.Email.Port = 587
	}
	if config.Email.Timeout == 0 {
		config.Email.Timeout = 30 * time.Second
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
	if c.Email.Host == "" {
		return fmt.Errorf("email.host is required")
	}
	if c.Email.Port == 0 {
		return fmt.Errorf("email.port is required")
	}
	if c.Email.From == "" {
		return fmt.Errorf("email.from is required")
	}
	if len(c.Email.To) == 0 {
		return fmt.Errorf("email.to must have at least one recipient")
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
