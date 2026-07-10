package reminder

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/jo-hoe/whatsapp-reminder/internal/config"
)

// MailClientInterface defines the interface for mail clients
type MailClientInterface interface {
	SendMail(ctx context.Context, request MailRequest) error
}

// MailRequest represents the data needed to send an email
type MailRequest struct {
	To          string
	Subject     string
	HtmlContent string
	From        string
}

// MailClient sends email via SMTP
type MailClient struct {
	cfg config.EmailConfig
}

// NewMailClient creates a new SMTP mail client
func NewMailClient(cfg config.EmailConfig) *MailClient {
	return &MailClient{cfg: cfg}
}

// SendMail sends an HTML email over SMTP, respecting context cancellation
func (c *MailClient) SendMail(ctx context.Context, request MailRequest) error {
	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)

	done := make(chan error, 1)
	go func() {
		done <- c.send(addr, request)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (c *MailClient) send(addr string, request MailRequest) error {
	conn, err := net.DialTimeout("tcp", addr, c.cfg.Timeout)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	client, err := smtp.NewClient(conn, c.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	if err := c.negotiate(client); err != nil {
		return err
	}

	if err := client.Mail(request.From); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := client.Rcpt(request.To); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}

	return c.writeBody(client, request)
}

func (c *MailClient) negotiate(client *smtp.Client) error {
	if c.cfg.StartTLS {
		tlsCfg := &tls.Config{
			ServerName: c.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if c.cfg.Auth.Required {
		auth := smtp.PlainAuth("", c.cfg.Auth.Username, c.cfg.Auth.Password, c.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	return nil
}

func (c *MailClient) writeBody(client *smtp.Client, r MailRequest) error {
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		r.From, r.To, r.Subject, r.HtmlContent,
	)
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}
