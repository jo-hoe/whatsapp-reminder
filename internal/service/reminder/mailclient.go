package reminder

import (
	"context"
	"time"

	mailclient "github.com/jo-hoe/go-mail-service/pkg/client"
)

// MailClientInterface defines the interface for mail clients
type MailClientInterface interface {
	SendMail(ctx context.Context, request MailRequest) (*MailResponse, error)
	HealthCheck(ctx context.Context) error
}

// MailClient is a wrapper around the go-mail-service client
type MailClient struct {
	client *mailclient.Client
}

// MailRequest represents the request to send an email
type MailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"content"`
	From        string `json:"from,omitempty"`
	FromName    string `json:"fromName,omitempty"`
}

// MailResponse represents the response from sending an email
type MailResponse struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"content"`
	From        string `json:"from,omitempty"`
	FromName    string `json:"fromName,omitempty"`
}

// NewMailClient creates a new mail service client using the official go-mail-service client library
func NewMailClient(baseURL string, timeout time.Duration) *MailClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := mailclient.NewClient(baseURL, mailclient.WithTimeout(timeout))

	return &MailClient{
		client: client,
	}
}

// SendMail sends an email using the mail service
func (c *MailClient) SendMail(ctx context.Context, request MailRequest) (*MailResponse, error) {
	// Convert our MailRequest to the library's MailRequest
	libRequest := mailclient.MailRequest{
		To:          request.To,
		Subject:     request.Subject,
		HtmlContent: request.HtmlContent,
		From:        request.From,
		FromName:    request.FromName,
	}

	// Call the library's SendMail method
	libResponse, err := c.client.SendMail(ctx, libRequest)
	if err != nil {
		return nil, err
	}

	// Convert the library's response to our MailResponse
	response := &MailResponse{
		To:          libResponse.To,
		Subject:     libResponse.Subject,
		HtmlContent: libResponse.HtmlContent,
		From:        libResponse.From,
		FromName:    libResponse.FromName,
	}

	return response, nil
}

// HealthCheck performs a health check against the mail service
func (c *MailClient) HealthCheck(ctx context.Context) error {
	return c.client.HealthCheck(ctx)
}
