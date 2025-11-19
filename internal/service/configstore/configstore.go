package configstore

import (
	"errors"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
)

type ConfigEntry struct {
	WhatsappReminderConfig dto.WhatsappReminderConfig
	CreationTime           time.Time
	DueTime                time.Time
	ProcessTime            time.Time
}

type ConfigStore interface {
	OverwriteConfigs(configs []ConfigEntry) error
	GetConfigs() ([]ConfigEntry, error)
}

type ConfigStoreMock struct {
	ReadStore         []ConfigEntry
	StoreConfigResult error
}

func (service *ConfigStoreMock) OverwriteConfigs(configs []ConfigEntry) error {
	service.ReadStore = configs

	return service.StoreConfigResult
}

func (service *ConfigStoreMock) GetConfigs() ([]ConfigEntry, error) {
	if service.ReadStore == nil {
		return nil, errors.New("")
	}

	return service.ReadStore, nil
}
