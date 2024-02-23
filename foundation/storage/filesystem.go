package storage

import (
	"fmt"
	"os"
	"strings"
)

type FileSystem struct {
	Filename string
	Database string
}

func (fs *FileSystem) Save(receiver <-chan []byte, failureChan chan struct{}) error {

	filesystemPath := os.Getenv("FILESYSTEM_PATH")

	if filesystemPath == "" || filesystemPath == "/" {
		filesystemPath = "./"
	}

	steps := strings.Split(filesystemPath, "/")

	steps = append(steps, fs.Database)

	path := strings.Join(steps, "/")

	path = strings.ReplaceAll(path, "//", "/")

	err := os.MkdirAll(path, 0555)
	if err != nil {
		return err
	}

	fs.Filename = fmt.Sprintf("%s/%s", path, fs.Filename)

	file, err := os.Create(fs.Filename)
	if err != nil {
		return err
	}

	defer file.Close()

	waitingForData := true

	for waitingForData {

		select {
		case content, moreComing := <-receiver:

			_, err = file.Write(content)
			if err != nil {

				os.Remove(fs.Filename)
				return err
			}

			if !moreComing {
				waitingForData = false
				break
			}
		case _ = <-failureChan:
			os.Remove(fs.Filename)

			waitingForData = false
			break
		}

	}

	return nil
}
