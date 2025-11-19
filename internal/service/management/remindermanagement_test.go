package management

import (
	"reflect"
	"testing"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/configstore"
	"github.com/jo-hoe/whatsapp-reminder/internal/service/reminder"
)

func TestReminderManagementService_Process(t *testing.T) {
	now := time.Now()

	alreadyProcessedItem := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  now.Add(-1 * time.Hour),
		DueTime:      now.Add(-2 * time.Hour),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}

	notDueItem := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  time.Time{},
		DueTime:      now.Add(time.Hour * 2),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}

	itemToProcess := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  time.Time{},
		DueTime:      now.Add(time.Hour * -1),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "9876543210",
			MailAddress: "tset@mail.com",
			MessageText: "ollah",
		},
	}
	mockStore := &configstore.ConfigStoreMock{
		ReadStore: []configstore.ConfigEntry{alreadyProcessedItem, notDueItem, itemToProcess},
	}
	mockReminder := &reminder.ReminderMock{}
	service := NewReminderManagementService(mockStore, mockReminder, getDefaultRetention(t), *getDefaultTestLocation(t))

	err := service.Process()
	if err != nil {
		t.Errorf("found error %+v", err)
	}
	if len(mockReminder.RemindResult) != 1 || mockReminder.RemindResult[0] != itemToProcess.WhatsappReminderConfig {
		t.Errorf("reminder was not sent as expected")
	}
	if len(mockStore.ReadStore) != 3 || mockStore.ReadStore[1].ProcessTime.IsZero() {
		t.Errorf("config store did not have expected state '%+v'", mockStore.ReadStore)
	}
}

func TestReminderManagementService_Process_Order(t *testing.T) {
	// Prepare
	now := time.Now()

	laterItem := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  time.Time{},
		DueTime:      now.Add(time.Hour * 2),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}

	earlierItem := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  time.Time{},
		DueTime:      now.Add(time.Hour * 1),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}
	mockStore := &configstore.ConfigStoreMock{
		ReadStore: []configstore.ConfigEntry{laterItem, earlierItem},
	}
	mockReminder := &reminder.ReminderMock{}

	// Test
	service := NewReminderManagementService(mockStore, mockReminder, getDefaultRetention(t), *getDefaultTestLocation(t))

	// Assert
	err := service.Process()
	if err != nil {
		t.Errorf("found error %+v", err)
	}
	if len(mockReminder.RemindResult) != 0 {
		t.Errorf("reminder was not sent as expected")
	}
	expected := []configstore.ConfigEntry{earlierItem, laterItem}
	if len(mockStore.ReadStore) != len(expected) || !reflect.DeepEqual(mockStore.ReadStore, expected) {
		t.Errorf("config store did not have expected state expected %+v\nactual %+v", expected, mockStore.ReadStore)
	}
}

func TestReminderManagementService_Process_Retention(t *testing.T) {
	// Prepare
	now := time.Now()

	itemBeforeRetention := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  now.Add(-23 * time.Hour),
		DueTime:      now.Add(-25 * time.Hour),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}

	itemAfterRetention := configstore.ConfigEntry{
		CreationTime: now.Add(-72 * time.Hour),
		ProcessTime:  now.Add(-25 * time.Hour),
		DueTime:      now.Add(-23 * time.Hour),
		WhatsappReminderConfig: dto.WhatsappReminderConfig{
			PhoneNumber: "0123456789",
			MailAddress: "test@mail.com",
			MessageText: "hallo",
		},
	}
	mockStore := &configstore.ConfigStoreMock{
		ReadStore: []configstore.ConfigEntry{itemAfterRetention, itemBeforeRetention},
	}
	mockReminder := &reminder.ReminderMock{}
	retention, err := time.ParseDuration("24h")
	if err != nil {
		t.Errorf("found error %+v", err)
	}

	// Test
	service := NewReminderManagementService(mockStore, mockReminder, retention, *getDefaultTestLocation(t))

	// Assert
	err = service.Process()
	if err != nil {
		t.Errorf("found error %+v", err)
	}
	if len(mockReminder.RemindResult) != 0 {
		t.Errorf("reminder was not sent as expected")
	}
	expected := []configstore.ConfigEntry{itemBeforeRetention}
	if len(mockStore.ReadStore) != len(expected) || !reflect.DeepEqual(mockStore.ReadStore, expected) {
		t.Errorf("config store did not have expected state expected %+v\nactual %+v", expected, mockStore.ReadStore)
	}
}

func getDefaultTestLocation(t *testing.T) *time.Location {
	location, err := time.LoadLocation("Europe/Berlin")

	if err != nil {
		t.Errorf("found error %+v", err)
	}

	return location
}

func getDefaultRetention(t *testing.T) time.Duration {
	duration, err := time.ParseDuration("72h")

	if err != nil {
		t.Errorf("found error %+v", err)
	}

	return duration
}
