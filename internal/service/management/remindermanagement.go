package management

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/configstore"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/reminder"
)

type ReminderManagementService struct {
	store           configstore.ConfigStore
	reminder        reminder.ReminderService
	defaultLocation time.Location
	retentionTime   time.Duration
}

func NewReminderManagementService(store configstore.ConfigStore, reminder reminder.ReminderService, retentionTime time.Duration, defaultLocation time.Location) *ReminderManagementService {
	return &ReminderManagementService{
		store:           store,
		reminder:        reminder,
		retentionTime:   retentionTime,
		defaultLocation: defaultLocation,
	}
}

func (service *ReminderManagementService) Process() error {
	configs, err := service.store.GetConfigs()
	if err != nil {
		return fmt.Errorf("could not read config from sheet %+v", err)
	}

	// get all items which should be processed
	itemsToProcess := make([]dto.WhatsappReminderConfig, 0)
	for _, config := range configs {
		// skip items which are already processed or item which are not due yet
		if !config.ProcessTime.IsZero() || config.DueTime.After(time.Now()) {
			continue
		}
		itemsToProcess = append(itemsToProcess, config.WhatsappReminderConfig)
	}

	// get all actually processed items
	processedItems := service.reminder.Remind(itemsToProcess)
	now := time.Now().In(&service.defaultLocation)
	for _, processedItem := range processedItems {
		for idx := range configs {
			if reflect.DeepEqual(processedItem, configs[idx].WhatsappReminderConfig) {
				configs[idx].ProcessTime = now
			}
		}
	}

	configs = service.filterItemByRetention(configs)

	// order by due time
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].DueTime.Before(configs[j].DueTime)
	})

	return service.store.OverwriteConfigs(configs)
}

func (service *ReminderManagementService) filterItemByRetention(configs []configstore.ConfigEntry) (result []configstore.ConfigEntry) {
	result = make([]configstore.ConfigEntry, 0)

	for _, config := range configs {
		// check retentation only if item has been already processed
		if !config.ProcessTime.IsZero() {
			// check if item can be filtered out based on process time comparison with retention time
			if time.Since(config.ProcessTime) > service.retentionTime {
				continue
			}
		}
		result = append(result, config)
	}

	return result
}
