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

func (dbx *Dropbox) start() (string, error) {

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload_session/start", nil)

	if err != nil {
		return "", err
	}

	token := os.Getenv("DROPBOX_ACCESS_TOKEN")

	params := map[string]interface{}{
		"close":        false,
		"session_type": "concurrent",
	}

	parameters, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(parameters))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	data, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {

		resErr := uploadError{}

		err = json.Unmarshal(data, &resErr)
		if err != nil {
			return "", errors.New(string(data))
		}

		return "", errors.New(resErr.Error.Tag)
	}

	successRes := map[string]string{}
	err = json.Unmarshal(data, &successRes)
	if err != nil {
		return "", err
	}

	sessionID := successRes["session_id"]

	return sessionID, nil
}

func (dbx *Dropbox) append(sessionID string, offset int64, data []byte) error {

	dropboxBuf := bytes.Buffer{}

	_, err := dropboxBuf.Write(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload_session/append_v2", &dropboxBuf)

	if err != nil {
		return err
	}

	token := os.Getenv("DROPBOX_ACCESS_TOKEN")

	params := map[string]interface{}{
		"close": false,
		"cursor": map[string]interface{}{
			"offset":     offset,
			"session_id": sessionID,
		},
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

		resData, _ := io.ReadAll(res.Body)

		resErr := uploadError{}

		err = json.Unmarshal(data, &resErr)
		if err != nil {
			return errors.New(string(resData))
		}

		return errors.New(resErr.Error.Tag)
	}

	return nil
}

func (dbx *Dropbox) finish(sessionID string) error {

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload_session/finish", nil)

	if err != nil {
		return err
	}

	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	path := os.Getenv("DROPBOX_PATH")

	params := map[string]interface{}{
		"commit": map[string]interface{}{
			"autorename":      false,
			"mode":            "add",
			"mute":            false,
			"path":            fmt.Sprintf("%s/%s", path, dbx.Filename),
			"strict_conflict": false,
		},
		"cursor": map[string]interface{}{
			"offset":     0,
			"session_id": sessionID,
		},
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

		resData, _ := io.ReadAll(res.Body)

		resErr := uploadError{}

		err = json.Unmarshal(resData, &resErr)
		if err != nil {
			return errors.New(string(resData))
		}

		return errors.New(resErr.Error.Tag)
	}

	return nil
}

func (dbx *Dropbox) Save(receiver <-chan []byte) error {

	//sessionID, err := dbx.start()
	//if err != nil {
	//	return err
	//}

	//payloadSize := 1 * 1024 * 1024
	payloadSize := 3
	var offset int64

	dropboxBuf := bytes.Buffer{}
	reader := bytes.Buffer{}

	for {

		fmt.Println("Buffer size:", dropboxBuf.Len())

		data, open := <-receiver

		_, err := dropboxBuf.Write(data)
		if err != nil {
			return err
		}

		payloadData := dropboxBuf.Next(payloadSize)

		fmt.Println(offset, string(payloadData))

		reader.Write(payloadData)

		//err = dbx.append(sessionID, offset, dropboxBuf.Next(payloadSize))
		//if err != nil {
		//	return err
		//}

		offset += int64(payloadSize)

		if !open {
			payloadData = dropboxBuf.Next(payloadSize)

			fmt.Println(offset, string(payloadData))

			reader.Write(payloadData)
			break
		}
	}

	fmt.Println("Buffer size:", dropboxBuf.Len())

	fmt.Println("read data:", string(reader.Bytes()))

	//err = dbx.finish(sessionID)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (dbx *DropboxMock) Save(receiver <-chan []byte) error {

	args := dbx.Called(<-receiver)
	return args.Error(0)
}
