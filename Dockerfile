# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the container application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o whatsapp-reminder ./cmd/container

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/whatsapp-reminder .
RUN chown appuser:appuser /app/whatsapp-reminder

USER appuser

CMD ["./whatsapp-reminder"]
