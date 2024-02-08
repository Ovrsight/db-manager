package storage

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
)

type FileSystem struct {
	Filename string
}

type FileSystemMock struct {
	mock.Mock
}

func (fs *FileSystem) Save(receiver <-chan []byte) error {

	filesystemPath := os.Getenv("FILESYSTEM_PATH")

	steps := strings.Split(filesystemPath, "/")

	if len(steps) > 1 {

		path := strings.Join(steps, "/")

		err := os.MkdirAll(path, 0555)
		if err != nil {
			return err
		}

		fs.Filename = fmt.Sprintf("%s/%s", path, fs.Filename)
	}

	fs.Filename = strings.ReplaceAll(fs.Filename, "//", "/")

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

func (fs *FileSystemMock) Save(receiver <-chan []byte) error {

	args := fs.Called(<-receiver)
	return args.Error(0)
}
