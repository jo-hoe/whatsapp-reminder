package configstore

import (
	"encoding/csv"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
)

type CSVConfigStore struct {
	mutex           sync.RWMutex
	openReader      openReader
	openWriter      openWriter
	defaultLocation time.Location
}

type openReader func() (reader io.Reader, err error)
type openWriter func() (writer io.Writer, err error)

var header = []string{"Timestamp", "Message Text", "Send Date", "Send Time", "Phone Number", "Mail Address", "Process Time"}

func NewCSVConfigStore(openReader openReader, openWriter openWriter, defaultLocation time.Location) *CSVConfigStore {
	return &CSVConfigStore{
		openReader:      openReader,
		openWriter:      openWriter,
		defaultLocation: defaultLocation,
	}
}

func (service *CSVConfigStore) OverwriteConfigs(configs []ConfigEntry) error {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	writer, err := service.openWriter()
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	err = csvWriter.Write(header)
	if err != nil {
		return err
	}

	data := make([][]string, 0)
	for _, config := range configs {
		row := make([]string, 7)
		row[0] = config.CreationTime.Format("02/01/2006 15:04:05")
		row[1] = config.WhatsappReminderConfig.MessageText
		row[2] = config.DueTime.Format("02/01/2006")
		row[3] = config.DueTime.Format("15:04:05")
		row[4] = config.WhatsappReminderConfig.PhoneNumber
		row[5] = config.WhatsappReminderConfig.MailAddress
		if config.ProcessTime.IsZero() {
			row[6] = ""
		} else {
			row[6] = config.ProcessTime.Format("02/01/2006 15:04:05")
		}

		data = append(data, row)
	}

	return csvWriter.WriteAll(data)
}

func (service *CSVConfigStore) GetConfigs() ([]ConfigEntry, error) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	reader, err := service.openReader()
	if err != nil {
		return nil, err
	}

	result := make([]ConfigEntry, 0)
	csvReader := csv.NewReader(reader)
	// deactivate field length validation
	csvReader.FieldsPerRecord = -1
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	// 'i' starts a 1 to skip the csv header
	for i := 1; i < len(data); i++ {
		creationTime, err := time.ParseInLocation("02/01/2006 15:04:05", getString(data, i, 0), &service.defaultLocation)
		if err != nil {
			log.Printf("could parse time '%v' in %+v", getString(data, i, 0), data[i])
			continue
		}
		processTime := time.Time{}
		if len(getString(data, i, 6)) != 0 {
			processTime, err = time.ParseInLocation("02/01/2006 15:04:05", getString(data, i, 6), &service.defaultLocation)
			if err != nil {
				log.Printf("could parse time '%v' in %+v", getString(data, i, 6), data[i])
				continue
			}
		}
		dueTime, err := time.ParseInLocation("02/01/2006 15:04:05", getString(data, i, 2)+" "+getString(data, i, 3), &service.defaultLocation)
		if err != nil {
			log.Printf("could parse time '%v' in %+v", getString(data, i, 2)+" "+getString(data, i, 3), data[i])
			continue
		}

		item := ConfigEntry{
			CreationTime: creationTime,
			DueTime:      dueTime,
			ProcessTime:  processTime,
			WhatsappReminderConfig: dto.WhatsappReminderConfig{
				MessageText: getString(data, i, 1),
				PhoneNumber: getString(data, i, 4),
				MailAddress: getString(data, i, 5),
			},
		}
		result = append(result, item)
	}

	return result, nil
}

func getString(data [][]string, i int, j int) string {
	if len(data)-1 >= i {
		if len(data[i])-1 >= j {
			return data[i][j]
		}
	}
	return ""
}
