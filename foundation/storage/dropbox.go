package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Dropbox struct {
	Filename string
	Database string
}

type uploadError struct {
	ErrorSummary string `json:"error_summary"`
	Error        struct {
		Tag string `json:".tag"`
	} `json:"error"`
}

const (
	downloadChunkSize int = 1 * 1024 * 1024
)

func (dbx *Dropbox) getFilePath() string {
	dropboxPath := os.Getenv("DROPBOX_PATH")

	dropboxPath = strings.ReplaceAll(dropboxPath, "//", "/")

	dropboxPath = strings.TrimSuffix(dropboxPath, "/")

	steps := strings.Split(dropboxPath, "/")

	steps = append(steps, dbx.Database)

	return strings.Join(steps, "/")
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
	path := dbx.getFilePath()

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

func (dbx *Dropbox) getFileMetadata(fileName string) (map[string]interface{}, error) {
	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	path := dbx.getFilePath()

	payload := map[string]interface{}{
		"include_deleted":                     false,
		"include_has_explicit_shared_members": false,
		"include_media_info":                  false,
		"path":                                fmt.Sprintf("%s/%s", path, fileName),
	}

	payloadBx, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(payloadBx)

	// get file details
	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/get_metadata", buf)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {

		return nil, errors.New("failed reading file's metadata")
	}

	resData, _ := io.ReadAll(res.Body)
	defer res.Body.Close()

	response := map[string]interface{}{}

	err = json.Unmarshal(resData, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (dbx *Dropbox) downloadFile(fileName string) (string, error) {

	token := os.Getenv("DROPBOX_ACCESS_TOKEN")
	path := dbx.getFilePath()

	fileInfo, err := dbx.getFileMetadata(fileName)
	if err != nil {
		return "", err
	}

	totalSize := fileInfo["size"].(float64)

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)

	if err != nil {
		return "", err
	}

	params := map[string]string{
		"path": fmt.Sprintf("%s/%s", path, fileName),
	}

	parameters, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(parameters))

	readBytes := 0
	chunkStart := 0
	chunkEnd := downloadChunkSize

	wd, _ := os.Getwd()

	tmpPath := fmt.Sprintf("%s/tmp/%s", wd, dbx.Database)

	err = os.MkdirAll(tmpPath, 0555)
	if err != nil {
		return "", err
	}

	tmpFilePath := fmt.Sprintf("%s/%s", tmpPath, fileName)

	f, err := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}

	defer f.Close()

	for readBytes < int(totalSize) {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunkStart, chunkEnd))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		if res.StatusCode != http.StatusPartialContent {
			return "", errors.New(fmt.Sprintf("failed fetching file with http status code: %d", res.StatusCode))
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		_, err = f.Write(data)
		if err != nil {
			return "", err
		}

		res.Body.Close()

		length, _ := strconv.Atoi(res.Header.Get("Content-Length"))

		readBytes += length

		chunkStart = chunkEnd + 1
		chunkEnd += downloadChunkSize
	}

	return tmpFilePath, nil
}

func (dbx *Dropbox) Save(receiver <-chan []byte, failureChan chan struct{}) (int, error) {

	sessionID, err := dbx.start()
	if err != nil {
		return 0, err
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
	uploadedSize := 0

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
			return 0, err
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

		uploadedSize += dropboxBuf.Len()
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

		uploadedSize = dropboxBuf.Len()

		err = dbx.append(sessionID, offset, dropboxBuf.Bytes(), true)
		if err != nil {
			return 0, err
		}
	}

	wg.Wait()

	for dropboxBuf.Len() > 0 && generatingBackupFailed {

		uploadedSize += dropboxBuf.Len()
		payloadData := dropboxBuf.Next(payloadSize)

		err = dbx.append(sessionID, offset, payloadData, len(payloadData) < payloadSize)
		if err != nil {
			return 0, err
		}
	}

	if !generatingBackupFailed {

		err = dbx.finish(sessionID)
		if err != nil {
			return 0, err
		}
	}

	return uploadedSize, nil
}

func (dbx *Dropbox) Retrieve(filesNames ...string) (locations []string, err error) {

	for _, fn := range filesNames {

		var fl string

		fl, err = dbx.downloadFile(fn)

		if err != nil {
			locations = nil
			return
		}

		locations = append(locations, fl)
	}

	return
}

func (dbx *Dropbox) DeleteRetrievals(filesLocations ...string) error {

	for _, fn := range filesLocations {

		err := os.Remove(fn)
		if err != nil {
			return err
		}
	}

	return nil
}
