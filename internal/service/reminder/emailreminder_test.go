package reminder

import (
	"context"
	"strings"
	"testing"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
)

// MockMailClient is a mock implementation of MailClient for testing
type MockMailClient struct {
	SentMails []MailRequest
	SendError error
}

func (m *MockMailClient) SendMail(ctx context.Context, request MailRequest) (*MailResponse, error) {
	if m.SendError != nil {
		return nil, m.SendError
	}
	m.SentMails = append(m.SentMails, request)
	return &MailResponse{
		To:      request.To,
		Subject: request.Subject,
		Message: "Email sent successfully",
	}, nil
}

func (m *MockMailClient) HealthCheck(ctx context.Context) error {
	return nil
}

func Test_Remind(t *testing.T) {
	mock := &MockMailClient{
		SentMails: make([]MailRequest, 0),
	}
	service := NewEmailReminderService(mock, "sender@test.com", "Test Sender", context.Background())
	testSet := make([]dto.WhatsappReminderConfig, 0)
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MessageText: "Text 1",
		MailAddress: "a@mail.com",
	})
	testSet = append(testSet, dto.WhatsappReminderConfig{
		PhoneNumber: "0123",
		MessageText: "Text 2",
		MailAddress: "a@mail.com",
	})
	testSet = append(testSet, dto.WhatsappReminderConfig{
		PhoneNumber: "0123",
		MessageText: "Text 3",
		MailAddress: "b@mail.com",
	})

	actual := service.Remind(testSet)

	if len(mock.SentMails) != 2 {
		t.Errorf("Expected 2 mails but found %d", len(mock.SentMails))
	}

	assertContains(t, "a@mail.com", mock.SentMails[0].To, mock.SentMails[1].To)
	assertContains(t, "b@mail.com", mock.SentMails[0].To, mock.SentMails[1].To)
	assertContains(t, "a@mail.com", actual[0].MailAddress, actual[1].MailAddress, actual[2].MailAddress)
	assertContains(t, "b@mail.com", actual[0].MailAddress, actual[1].MailAddress, actual[2].MailAddress)
}

func Test_messageConfigsToMailRequest(t *testing.T) {
	mock := &MockMailClient{}
	service := NewEmailReminderService(mock, "sender@test.com", "Test Sender", context.Background())

	testSet := make([]dto.WhatsappReminderConfig, 0)
	var longMessageBuilder strings.Builder
	for i := 1; i < 100; i++ {
		longMessageBuilder.WriteString("b")
	}
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MessageText: "a",
		PhoneNumber: "012",
		MailAddress: "test@mail.com",
	})
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MessageText: longMessageBuilder.String(),
		PhoneNumber: "007",
		MailAddress: "test@mail.com",
	})

	actual := service.messageConfigsToMailRequest(testSet)

	if strings.Count(actual.HtmlContent, "<li>") != len(testSet) {
		t.Errorf("Expected %d list items but found %d", len(testSet), strings.Count(actual.HtmlContent, "<li>"))
	}
	if strings.Count(actual.HtmlContent, "...") != 1 {
		t.Errorf("Expected one element to be cut off but found %d elements", strings.Count(actual.HtmlContent, "..."))
	}
	if actual.From != "sender@test.com" {
		t.Errorf("Expected sender@test.com but got %s", actual.From)
	}
	if actual.FromName != "Test Sender" {
		t.Errorf("Expected Test Sender but got %s", actual.FromName)
	}
}

func Test_groupByMailAddress(t *testing.T) {
	testSet := make([]dto.WhatsappReminderConfig, 0)
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MailAddress: "a",
	})
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MailAddress: "b",
	})
	testSet = append(testSet, dto.WhatsappReminderConfig{
		MailAddress: "b",
	})

	actual := groupByMailAddress(testSet)

	if len(actual) != 2 {
		t.Errorf("Expected 2 item but found %d", len(actual))
	}
	if len(actual["a"]) != 1 {
		t.Errorf("Expected a to have 1 item but found %d", len(actual))
	}
	if len(actual["b"]) != 2 {
		t.Errorf("Expected a to have 2 items but found %d", len(actual))
	}
}

func assertContains(t *testing.T, item any, searchSpace ...any) {
	found := false

	for _, element := range searchSpace {
		if item == element {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Did not found %+v in '%+v'", item, searchSpace)
	}
}
