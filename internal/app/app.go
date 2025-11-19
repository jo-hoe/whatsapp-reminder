package app

import (
	"context"
	"io"
	"time"

	"github.com/jo-hoe/google-sheets/gs"
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
	MailServiceURL       string
	OriginAddress        string
	OriginName           string
}

func Start(config *AppConfig) error {
	readerCreation := func() (writer io.Reader, err error) {
		return gs.OpenSheet(config.Ctx, config.SpreadSheetId, config.SheetName, gs.O_RDONLY, config.ServiceAccountSecret)
	}

	writerCreation := func() (writer io.Writer, err error) {
		return gs.OpenSheet(config.Ctx, config.SpreadSheetId, config.SheetName, gs.O_RDWR|gs.O_TRUNC, config.ServiceAccountSecret)
	}

	store := configstore.NewCSVConfigStore(readerCreation, writerCreation, *config.TimeLocation)
	mailClient := reminder.NewMailClient(config.MailServiceURL, 30*time.Second)
	reminderService := reminder.NewEmailReminderService(mailClient, config.OriginAddress, config.OriginName, config.Ctx)
	manager := management.NewReminderManagementService(store, reminderService, config.RetentionTime, *config.TimeLocation)

	return manager.Process()
}
