package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Dropbox struct {
	Filename string
	Database string
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

func (dbx *Dropbox) append(sessionID string, offset int64, data []byte, final bool) error {

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
		"close": final,
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

	defer res.Body.Close()

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
			"path":            fmt.Sprintf("%s/%s/%s", path, dbx.Database, dbx.Filename),
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

func (dbx *Dropbox) Save(receiver <-chan []byte, failureChan chan struct{}) error {

	sessionID, err := dbx.start()
	if err != nil {
		return err
	}

	concurrency := os.Getenv("DROPBOX_CONCURRENT_REQUESTS")

	concurrentTasks, err := strconv.Atoi(concurrency)
	if err != nil {
		concurrentTasks = 5
	}

	payloadSize := 4 * 1024 * 1024
	var offset int64
	singleChunk := true
	tasksToComplete := make(chan struct{}, concurrentTasks)
	completed := make(chan struct{}, 1)
	wg := sync.WaitGroup{}

	dropboxBuf := bytes.Buffer{}

	generatingBackupFailed := false

	go func(flag *bool) {

		select {
		case _ = <-failureChan:
			*flag = true
		case _ = <-completed:
			break
		}
	}(&generatingBackupFailed)
	defer func() {
		completed <- struct{}{}
	}()

	for {

		data, open := <-receiver

		_, err := dropboxBuf.Write(data)
		if err != nil {
			return err
		}

		if dropboxBuf.Len() < payloadSize && offset == 0 {

			if !open {
				//	finished getting data from backup method
				break
			}

			// grow the buffer to reach the minimum payload size
			continue
		}

		singleChunk = false

		if dropboxBuf.Len() < payloadSize && offset > 0 {

			if open {
				continue
			}
		}

		payloadData := dropboxBuf.Next(payloadSize)

		taskID := uuid.New().String()
		tasksToComplete <- struct{}{}

		lastChunk := len(payloadData) < payloadSize

		wg.Add(1)

		go func(id string, data []byte, offsetPoint int64, lastChunk bool) {

			defer wg.Done()

			log.Println("Processing task:", id)

			if generatingBackupFailed {
				return
			}

			err = dbx.append(sessionID, offsetPoint, payloadData, lastChunk)
			if err != nil {
				log.Println("Failed task:", id)
				log.Println(err)
				return
			}

			<-tasksToComplete
			log.Println("Completed task:", id)
		}(taskID, payloadData[:], offset, lastChunk)

		offset += int64(len(payloadData))

		if !open {
			break
		}
	}

	if singleChunk && !generatingBackupFailed {
		err = dbx.append(sessionID, offset, dropboxBuf.Bytes(), true)
		if err != nil {
			return err
		}
	}

	wg.Wait()

	for dropboxBuf.Len() > 0 && generatingBackupFailed {

		payloadData := dropboxBuf.Next(payloadSize)

		err = dbx.append(sessionID, offset, payloadData, len(payloadData) < payloadSize)
		if err != nil {
			return err
		}
	}

	if !generatingBackupFailed {

		err = dbx.finish(sessionID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dbx *DropboxMock) Save(receiver <-chan []byte) error {

	args := dbx.Called(<-receiver)
	return args.Error(0)
}
