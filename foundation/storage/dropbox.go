package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"os"
)

type Dropbox struct {
	Filename string
}

type DropboxMock struct {
	mock.Mock
}

type uploadError struct {
	ErrorSummary string `json:"error_summary"`
	Error        struct {
		Tag string `json:".tag"`
	} `json:"error"`
}

func (dbx *Dropbox) Save(receiver <-chan []byte) error {

	dropboxBuf := bytes.Buffer{}

	for data := range receiver {
		_, err := dropboxBuf.Write(data)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", &dropboxBuf)
	if err != nil {
		return err
	}

	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	path := os.Getenv("DROPBOX_PATH")
	params := map[string]interface{}{
		"autorename":      false,
		"mode":            "add",
		"mute":            false,
		"path":            fmt.Sprintf("%s/%s", path, dbx.Filename),
		"strict_conflict": false,
	}
	parameters, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(parameters))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		data, _ := io.ReadAll(res.Body)

		resErr := uploadError{}

		err = json.Unmarshal(data, &resErr)
		if err != nil {
			return errors.New(string(data))
		}

		return errors.New(resErr.Error.Tag)
	}

	return nil
}

func (dbx *DropboxMock) Save(receiver <-chan []byte) error {

	args := dbx.Called(<-receiver)
	return args.Error(0)
}
