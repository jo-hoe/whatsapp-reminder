package reminder

import (
	"context"
	"strings"
	"testing"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
)

// MockMailClient is a mock implementation of MailClientInterface for testing
type MockMailClient struct {
	SentMails []MailRequest
	SendError error
}

func (m *MockMailClient) SendMail(ctx context.Context, request MailRequest) error {
	if m.SendError != nil {
		return m.SendError
	}
	m.SentMails = append(m.SentMails, request)
	return nil
}

func Test_Remind(t *testing.T) {
	mock := &MockMailClient{
		SentMails: make([]MailRequest, 0),
	}
	to := []string{"recipient@test.com"}
	service := NewEmailReminderService(mock, "sender@test.com", to, context.Background())
	testSet := []dto.WhatsappReminderConfig{
		{MessageText: "Text 1", MailAddress: "a@mail.com"},
		{PhoneNumber: "0123", MessageText: "Text 2", MailAddress: "a@mail.com"},
		{PhoneNumber: "0123", MessageText: "Text 3", MailAddress: "b@mail.com"},
	}

	actual := service.Remind(testSet)

	if len(mock.SentMails) != 1 {
		t.Errorf("Expected 1 mail (one per config recipient) but found %d", len(mock.SentMails))
	}
	if mock.SentMails[0].To != "recipient@test.com" {
		t.Errorf("Expected recipient@test.com but got %s", mock.SentMails[0].To)
	}
	if len(actual) != len(testSet) {
		t.Errorf("Expected %d returned configs but got %d", len(testSet), len(actual))
	}
}

func Test_buildHtmlContent(t *testing.T) {
	mock := &MockMailClient{}
	service := NewEmailReminderService(mock, "sender@test.com", []string{"r@test.com"}, context.Background())

	var longMessageBuilder strings.Builder
	for i := 1; i < 100; i++ {
		longMessageBuilder.WriteString("b")
	}
	testSet := []dto.WhatsappReminderConfig{
		{MessageText: "a", PhoneNumber: "012", MailAddress: "test@mail.com"},
		{MessageText: longMessageBuilder.String(), PhoneNumber: "007", MailAddress: "test@mail.com"},
	}

	actual := service.buildHtmlContent(testSet)

	if strings.Count(actual, "<li>") != len(testSet) {
		t.Errorf("Expected %d list items but found %d", len(testSet), strings.Count(actual, "<li>"))
	}
	if strings.Count(actual, "...") != 1 {
		t.Errorf("Expected one element to be cut off but found %d elements", strings.Count(actual, "..."))
	}
}
