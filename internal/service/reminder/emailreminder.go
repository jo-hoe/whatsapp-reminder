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
	mailClient    MailClientInterface
	originAddress string
	originName    string
	ctx           context.Context
}

func NewEmailReminderService(
	mailClient MailClientInterface,
	originAddress,
	originName string,
	ctx context.Context) *EmailReminderService {
	return &EmailReminderService{
		mailClient:    mailClient,
		originAddress: originAddress,
		originName:    originName,
		ctx:           ctx,
	}
}

func (service *EmailReminderService) Remind(messageConfigs []dto.WhatsappReminderConfig) (result []dto.WhatsappReminderConfig) {
	configsGroupedByMail := groupByMailAddress(messageConfigs)

	result = make([]dto.WhatsappReminderConfig, 0)

	for _, configs := range configsGroupedByMail {
		mailRequest := service.messageConfigsToMailRequest(configs)
		_, err := service.mailClient.SendMail(service.ctx, mailRequest)
		if err != nil {
			log.Printf("could not send due to error %v: %v", err, configs)
		} else {
			result = append(result, configs...)
		}
	}

	return result
}

func groupByMailAddress(messageConfigs []dto.WhatsappReminderConfig) map[string][]dto.WhatsappReminderConfig {
	result := make(map[string][]dto.WhatsappReminderConfig, 0)

	for _, messageConfig := range messageConfigs {
		if _, ok := result[messageConfig.MailAddress]; !ok {
			result[messageConfig.MailAddress] = make([]dto.WhatsappReminderConfig, 0)
		}
		result[messageConfig.MailAddress] = append(result[messageConfig.MailAddress], messageConfig)
	}

	return result
}

func (service *EmailReminderService) messageConfigsToMailRequest(messageConfigs []dto.WhatsappReminderConfig) MailRequest {
	result := MailRequest{
		To:       messageConfigs[0].MailAddress,
		Subject:  "WhatsApp Reminder",
		From:     service.originAddress,
		FromName: service.originName,
	}

	// build content
	var stringBuilder strings.Builder
	stringBuilder.WriteString(mailStart)

	for _, messageConfig := range messageConfigs {
		whatsappLink := whatsapp.CreateWhatsappLink(messageConfig.PhoneNumber, messageConfig.MessageText)

		htmlEscapedText := html.EscapeString(messageConfig.MessageText)
		if len(htmlEscapedText) > (61) {
			htmlEscapedText = htmlEscapedText[:61] + "..."
		}

		number := messageConfig.PhoneNumber
		if len(number) == 0 {
			number = "no number provided"
		}

		stringBuilder.WriteString(fmt.Sprintf(mailItem, whatsappLink, htmlEscapedText, number))
	}
	stringBuilder.WriteString(mailEnd)
	result.HtmlContent = stringBuilder.String()

	return result
}
