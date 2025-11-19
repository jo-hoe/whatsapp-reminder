package reminder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MailClientInterface defines the interface for mail clients
type MailClientInterface interface {
	SendMail(ctx context.Context, request MailRequest) (*MailResponse, error)
	HealthCheck(ctx context.Context) error
}

// MailClient is an HTTP client for the go-mail-service
type MailClient struct {
	baseURL    string
	httpClient *http.Client
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
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response from the mail service
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("mail service error (HTTP %d): %s", e.Code, e.Message)
}

// NewMailClient creates a new mail service client
func NewMailClient(baseURL string, timeout time.Duration) *MailClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &MailClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendMail sends an email using the mail service
func (c *MailClient) SendMail(ctx context.Context, request MailRequest) (*MailResponse, error) {
	// Validate required fields
	if request.To == "" {
		return nil, fmt.Errorf("recipient email address is required")
	}
	if request.Subject == "" {
		return nil, fmt.Errorf("email subject is required")
	}
	if request.HtmlContent == "" {
		return nil, fmt.Errorf("email content is required")
	}

	// Marshal request body
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/v1/sendmail"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
			fmt.Printf("failed to close response body: %v\n", closeErr)
		}
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			// If we can't parse the error response, create a generic one
			return nil, ErrorResponse{
				Message: string(respBody),
				Code:    resp.StatusCode,
			}
		}
		errorResp.Code = resp.StatusCode
		return nil, errorResp
	}

	// Parse success response
	var mailResp MailResponse
	if err := json.Unmarshal(respBody, &mailResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &mailResp, nil
}

// HealthCheck performs a health check against the mail service
func (c *MailClient) HealthCheck(ctx context.Context) error {
	rootReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/", nil)
	if err != nil {
		return fmt.Errorf("failed to create root health check request: %w", err)
	}

	rootResp, err := c.httpClient.Do(rootReq)
	if err != nil {
		return fmt.Errorf("root health check failed: %w", err)
	}
	defer rootResp.Body.Close()

	if rootResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rootResp.Body)
		return fmt.Errorf("root health check failed with status %d: %s", rootResp.StatusCode, string(body))
	}

	return nil
}
