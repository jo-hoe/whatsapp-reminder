# WhatsApp Reminder

[![Test Status](https://github.com/jo-hoe/whatsapp-reminder/workflows/test/badge.svg)](https://github.com/jo-hoe/whatsapp-reminder/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/whatsapp-reminder/workflows/lint/badge.svg)](https://github.com/jo-hoe/whatsapp-reminder/actions?workflow=lint)

A Go application that sends WhatsApp reminders based on data from Google Sheets. Supports multiple deployment methods: CLI, Container, and Kubernetes.

## Features

- Read reminder data from Google Sheets
- Send email notifications with WhatsApp links
- Configurable retention time for processed reminders
- Multiple deployment options (CLI, Docker, Kubernetes)
- YAML-based configuration

## Configuration

All deployment methods use the same YAML configuration format. Copy `config.yaml.example` to `config.yaml` and update the values:

```yaml
# Google Sheets configuration
googleSheets:
  spreadsheetId: "your_google_spreadsheet_id_here"
  sheetName: "your_sheet_name_here"
  serviceAccountFile: "/path/to/service-account.json"

# Email configuration (go-mail-service)
email:
  serviceUrl: "http://localhost:80"  # URL of the go-mail-service
  originAddress: "your_email@example.com"
  originName: "Your Name"

# Scheduling configuration (container only)
schedule:
  interval: "1h"        # How often to run
  runOnStartup: true    # Run immediately on startup

# Application configuration
app:
  timeLocation: "UTC"   # Timezone
  retentionTime: "24h"  # How long to keep processed reminders
  logLevel: "info"      # Log level
```

## Deployment Methods

### 1. Kubernetes (Scheduled CronJob)

Deploy to Kubernetes using Helm for scheduled execution:

```bash
# Install the Helm chart

# Option 1: Using a service account JSON file
helm install whatsapp-reminder ./charts/whatsapp-reminder \
  --set config.googleSheets.spreadsheetId=YOUR_SPREADSHEET_ID \
  --set config.googleSheets.sheetName=YOUR_SHEET_NAME \
  --set secrets.emailOriginAddress=your-email@example.com \
  --set secrets.emailOriginName="Your Name" \
  --set-file secrets.serviceAccountJson=./service-account.json

# Option 2: Using service account JSON as a string (useful for CI/CD)
helm install whatsapp-reminder ./charts/whatsapp-reminder \
  --set config.googleSheets.spreadsheetId=YOUR_SPREADSHEET_ID \
  --set config.googleSheets.sheetName=YOUR_SHEET_NAME \
  --set secrets.emailOriginAddress=your-email@example.com \
  --set secrets.emailOriginName="Your Name" \
  --set secrets.serviceAccountJson='{"type":"service_account","project_id":"your-project",...}'

# Upgrade an existing deployment
helm upgrade whatsapp-reminder ./charts/whatsapp-reminder \
  -f your-values.yaml

# Uninstall
helm uninstall whatsapp-reminder
```

The Helm chart configures a Kubernetes CronJob that runs on a schedule (default: every hour). You can customize the schedule and other settings in `values.yaml`.

#### Local K3D Testing

For local development and testing with K3D:

```bash
# Prerequisites: Install K3D
# https://k3d.io/#install-script

# Setup: Copy .env.example to .env and fill in your values
cp .env.example .env
# Edit .env with your configuration

# Start K3D cluster and deploy
make start-k3d

# Stop K3D cluster
make stop-k3d

# Restart (useful after code changes)
make restart-k3d
```

The Makefile reads configuration from a `.env` file. Required variables:

- `SPREADSHEET_ID` - Your Google Sheets spreadsheet ID
- `SHEET_NAME` - Name of the sheet to read from
- `EMAIL_ORIGIN_ADDRESS` - Email address to send from
- `EMAIL_ORIGIN_NAME` - Display name for emails

Optional variables (see `.env.example` for defaults):

- `SCHEDULE` - Cron expression for job scheduling
- `TIME_LOCATION` - Timezone for processing
- `RETENTION_TIME` - How long to keep processed reminders
- `EMAIL_SERVICE_URL` - URL of the mail service

**Service Account Authentication:**

Set `SERVICE_ACCOUNT_JSON_BASE64` environment variable in your `.env` file with the base64-encoded JSON content. This approach completely avoids all shell escaping issues with special characters and literal `\n` in the `private_key` field.

**To create the base64-encoded value:**

```bash
# On Linux/Mac:
base64 -w 0 service-account.json

# On Windows (PowerShell):
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("service-account.json"))

# On Windows (Git Bash):
base64 -w 0 service-account.json
```

**Then in your `.env` file (no quotes needed):**

```bash
SERVICE_ACCOUNT_JSON_BASE64=eyJ0eXBlIjoic2VydmljZV9hY2NvdW50IiwicHJvamVjdF9pZCI6InlvdXItcHJvamVjdCIsLi4ufQ==
```

**Note about the private_key field:** When you view the original JSON, the `private_key` field contains literal `\n` characters (backslash-n) instead of actual newlines. This is normal for service account JSON files. The base64 encoding handles this correctly, and the Google API client will process it properly.

### 2. CLI (One-time execution)

Build and run the CLI version for one-time execution:

```bash
# Build
go build ./cmd/cli

# Run with default config (./config.yaml)
./cli

# Run with custom config path
./cli -config /path/to/config.yaml
```

### 3. Container (Scheduled execution)

Build and run as a container for scheduled execution:

```bash
# Build container
docker build -t whatsapp-reminder .

# Run container
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/service-account.json:/app/service-account.json \
  whatsapp-reminder

# Or use docker-compose
docker-compose up
```

## Email Service

This application requires the [go-mail-service](https://github.com/jo-hoe/go-mail-service) to send emails. You can run it locally or deploy it as needed. The mail service supports multiple providers including SendGrid and Mailjet.

For local development, you can run the mail service using Docker:

```bash
docker run -p 80:80 --env-file .env ghcr.io/jo-hoe/go-mail-service:latest
```

See the [go-mail-service documentation](https://github.com/jo-hoe/go-mail-service) for more details on configuration and deployment.

## Linting

Project uses `golangci-lint` for linting.

### Installation

<https://golangci-lint.run/usage/install/>

#### Execution

```bash
# Run linting
make lint
```

# Or directly

```bash
golangci-lint run ./...
```
