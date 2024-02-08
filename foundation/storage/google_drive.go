package storage

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"os"
)

type GoogleDrive struct {
	Filename string
}

type GoogleDriveMock struct {
	mock.Mock
}

func (gglD *GoogleDrive) Save(receiver <-chan []byte) error {

	jsonData := os.Getenv("GOOGLE_DRIVE_API_CREDENTIALS_JSON")
	jsonCreds := []byte(jsonData)

	driveService, err := drive.NewService(context.Background(), option.WithCredentialsJSON(jsonCreds))
	if err != nil {
		return err
	}

	folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID")
	driveFile := &drive.File{
		Name:    gglD.Filename,
		Parents: []string{folderID},
	}

	googleDriveBuf := bytes.Buffer{}

	for data := range receiver {
		_, err := googleDriveBuf.Write(data)
		if err != nil {
			return err
		}
	}

	f, err := driveService.Files.Create(driveFile).Media(&googleDriveBuf).Do()
	if err != nil {
		return err
	}

	if f.ServerResponse.HTTPStatusCode != 200 {
		return errors.New("failed uploading backup to google drive folder")
	}

	return nil
}

func (gglD *GoogleDriveMock) Save(receiver <-chan []byte) error {

	args := gglD.Called(<-receiver)
	return args.Error(0)
}
