package app

import (
	"context"
	"io"
	"time"

	"github.com/jo-hoe/google-sheets/gs"
	"github.com/jo-hoe/whatsapp-reminder/internal/config"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/configstore"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/management"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/reminder"
)

type AppConfig struct {
	Ctx                  context.Context
	SpreadSheetId        string
	SheetName            string
	ServiceAccountSecret []byte
	TimeLocation         *time.Location
	RetentionTime        time.Duration
	Email                config.EmailConfig
}

func Start(config *AppConfig) error {
	readerCreation := func() (writer io.Reader, err error) {
		return gs.OpenSheet(config.Ctx, config.SpreadSheetId, config.SheetName, gs.O_RDONLY, config.ServiceAccountSecret)
	}

	writerCreation := func() (writer io.Writer, err error) {
		return gs.OpenSheet(config.Ctx, config.SpreadSheetId, config.SheetName, gs.O_RDWR|gs.O_TRUNC, config.ServiceAccountSecret)
	}

	store := configstore.NewCSVConfigStore(readerCreation, writerCreation, *config.TimeLocation)
	mailClient := reminder.NewMailClient(config.Email)
	reminderService := reminder.NewEmailReminderService(mailClient, config.Email.From, config.Email.To, config.Ctx)
	manager := management.NewReminderManagementService(store, reminderService, config.RetentionTime, *config.TimeLocation)

	return manager.Process()
}
