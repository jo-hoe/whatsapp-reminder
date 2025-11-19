package reminder

import "github.com/jo-hoe/whatsapp-reminder/internal/dto"

type ReminderService interface {
	Remind(messageConfigs []dto.WhatsappReminderConfig) (result []dto.WhatsappReminderConfig)
}

type ReminderMock struct {
	RemindResult []dto.WhatsappReminderConfig
}

func (service *ReminderMock) Remind(messageConfigs []dto.WhatsappReminderConfig) (result []dto.WhatsappReminderConfig) {
	service.RemindResult = messageConfigs

	return service.RemindResult
}
