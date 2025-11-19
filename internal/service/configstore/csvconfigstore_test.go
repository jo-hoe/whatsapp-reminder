package configstore

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jo-hoe/whatsapp-reminder/internal/dto"
)

const testFileName = "testdata/test.csv"

func TestCSVConfigStore_GetConfigs(t *testing.T) {
	expected := getTestConfig(t)

	configStore := NewCSVConfigStore(openTestReader, nil, *getDefaultTestLocation(t))

	actual, err := configStore.GetConfigs()

	if err != nil {
		t.Errorf("error found %+v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("CSVConfigStore.GetConfigs() = %v, want %v", actual, expected)
	}
}

func TestCSVConfigStore_OverwriteConfigs(t *testing.T) {
	file, err := os.CreateTemp("", "tempfile-")
	if err != nil {
		t.Errorf("found error %+v", err)
	}

	openWriter := func() (writer io.Writer, err error) {
		return file, err
	}

	configStore := NewCSVConfigStore(nil, openWriter, *getDefaultTestLocation(t))

	err = configStore.OverwriteConfigs(getTestConfig(t))
	if err != nil {
		t.Errorf("found error %+v", err)
	}

	actual := getFileContent(t, file.Name())
	expected := getFileContent(t, testFileName)

	// remove OS dependent EOL character
	actual = strings.ReplaceAll(actual, "\r", "")
	expected = strings.ReplaceAll(expected, "\r", "")

	if actual != expected {
		t.Errorf("actual:\n%s\nnot equal to expected:\n%s", actual, expected)
	}
}

func openTestReader() (reader io.Reader, err error) {
	return os.Open(testFileName)
}

func getDefaultTestLocation(t *testing.T) *time.Location {
	location, err := time.LoadLocation("Europe/Berlin")

	if err != nil {
		t.Errorf("found error %+v", err)
	}

	return location
}

func getTestConfig(t *testing.T) []ConfigEntry {
	return []ConfigEntry{
		{
			WhatsappReminderConfig: dto.WhatsappReminderConfig{
				PhoneNumber: "01234567890",
				MessageText: "Test 1",
				MailAddress: "test@mail.de",
			},
			CreationTime: time.Date(2022, 07, 20, 13, 13, 13, 0, getDefaultTestLocation(t)),
			DueTime:      time.Date(2022, 07, 22, 15, 15, 15, 0, getDefaultTestLocation(t)),
			ProcessTime:  time.Date(2022, 07, 24, 17, 17, 17, 0, getDefaultTestLocation(t)),
		}, {
			WhatsappReminderConfig: dto.WhatsappReminderConfig{
				PhoneNumber: "01234567890",
				MessageText: "Test 2",
				MailAddress: "test@mail.de",
			},
			CreationTime: time.Date(2022, 07, 21, 14, 14, 14, 0, getDefaultTestLocation(t)),
			DueTime:      time.Date(2022, 07, 23, 16, 16, 16, 0, getDefaultTestLocation(t)),
			ProcessTime:  time.Time{},
		},
	}
}

func getFileContent(t *testing.T, filepath string) string {
	readfile, err := os.Open(filepath)
	if err != nil {
		t.Errorf("found error %+v", err)
	}
	defer func() {
		cerr := readfile.Close()
		if err == nil {
			err = cerr
		}
	}()
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(readfile)
	if err != nil {
		t.Errorf("found error %+v", err)
	}

	return buffer.String()
}
