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

func (fs *FileSystem) getFilePath() string {
	filesystemPath := os.Getenv("FILESYSTEM_PATH")

	wd, _ := os.Getwd()

	filesystemPath = fmt.Sprintf("%s/%s", wd, filesystemPath)

	filesystemPath = strings.ReplaceAll(filesystemPath, "//", "/")

	filesystemPath = strings.TrimSuffix(filesystemPath, "/")

	steps := strings.Split(filesystemPath, "/")

	steps = append(steps, fs.Database)

	return strings.Join(steps, "/")
}

func (fs *FileSystem) DeleteRetrievals(filesLocations ...string) error {

	// should be deleting the files used to recover a database, but since it used backup files and not downloaded ones they can't be deleted

	return nil
}

func (fs *FileSystem) Retrieve(filesNames ...string) (locations []string, err error) {

	path := fs.getFilePath()

	for _, fileName := range filesNames {
		location := fmt.Sprintf("%s/%s", path, fileName)

		var f *os.File

		f, err = os.Open(location)
		if err != nil {
			locations = nil
			return
		}

		f.Close()

		locations = append(locations, location)
	}

	return
}

func (fs *FileSystem) Save(receiver <-chan []byte, failureChan chan struct{}) (int, error) {

	path := fs.getFilePath()
	writtenSize := 0

	err := os.MkdirAll(path, 0555)
	if err != nil {
		return 0, err
	}

	fs.Filename = fmt.Sprintf("%s/%s", path, fs.Filename)

	file, err := os.Create(fs.Filename)
	if err != nil {
		return 0, err
	}

	defer file.Close()

	waitingForData := true

	for waitingForData {

		select {
		case content, moreComing := <-receiver:

			written, err := file.Write(content)
			writtenSize += written
			if err != nil {

				os.Remove(fs.Filename)
				return 0, err
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

	return writtenSize, nil
}
