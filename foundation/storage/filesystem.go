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

func (fs *FileSystem) Save(receiver <-chan []byte) error {

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

	for content := range receiver {

		_, err = file.Write(content)
		if err != nil {

			os.Remove(fs.Filename)
			return err
		}
	}

	return nil
}
