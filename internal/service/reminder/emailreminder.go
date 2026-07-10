package reminder

import (
	"context"
	_ "embed"
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
	"github.com/jo-hoe/whatsapp-reminder/internal/whatsapp"
)

//go:embed template_start.html
var mailStart string

//go:embed template_item.html
var mailItem string

//go:embed template_end.html
var mailEnd string

type EmailReminderService struct {
	mailClient MailClientInterface
	from       string
	to         []string
	ctx        context.Context
}

func NewEmailReminderService(
	mailClient MailClientInterface,
	from string,
	to []string,
	ctx context.Context) *EmailReminderService {
	return &EmailReminderService{
		mailClient: mailClient,
		from:       from,
		to:         to,
		ctx:        ctx,
	}
}

func (service *EmailReminderService) Remind(messageConfigs []dto.WhatsappReminderConfig) (result []dto.WhatsappReminderConfig) {
	log.Printf("sending emails to %d recipient(s) with %d total reminder(s)", len(service.to), len(messageConfigs))

	htmlContent := service.buildHtmlContent(messageConfigs)

	result = make([]dto.WhatsappReminderConfig, 0)
	successCount := 0
	failureCount := 0

	for _, recipient := range service.to {
		req := MailRequest{
			To:          recipient,
			Subject:     "WhatsApp Reminder",
			HtmlContent: htmlContent,
			From:        service.from,
		}
		log.Printf("sending %d reminder(s) to %s", len(messageConfigs), recipient)

		err := service.mailClient.SendMail(service.ctx, req)
		if err != nil {
			failureCount += len(messageConfigs)
			log.Printf("failed to send %d reminder(s) to %s: %v", len(messageConfigs), recipient, err)
		} else {
			successCount += len(messageConfigs)
			log.Printf("successfully sent %d reminder(s) to %s", len(messageConfigs), recipient)
		}
	}

	if successCount > 0 {
		result = append(result, messageConfigs...)
	}

	log.Printf("email sending summary: %d successful, %d failed out of %d total reminder(s)",
		successCount, failureCount, len(messageConfigs))

	return result
}

func (service *EmailReminderService) buildHtmlContent(messageConfigs []dto.WhatsappReminderConfig) string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString(mailStart)

	for _, messageConfig := range messageConfigs {
		whatsappLink := whatsapp.CreateWhatsappLink(messageConfig.PhoneNumber, messageConfig.MessageText)

		htmlEscapedText := html.EscapeString(messageConfig.MessageText)
		if len(htmlEscapedText) > 61 {
			htmlEscapedText = htmlEscapedText[:61] + "..."
		}

		number := messageConfig.PhoneNumber
		if len(number) == 0 {
			number = "no number provided"
		}

		fmt.Fprintf(&stringBuilder, mailItem, whatsappLink, htmlEscapedText, number)
	}
	stringBuilder.WriteString(mailEnd)
	return stringBuilder.String()
}
