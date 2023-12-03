package storage

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
)

type FileSystemDriver struct {
	Filename string
}

type FileSystemDriverMock struct {
	mock.Mock
}

func (fs *FileSystemDriver) Upload(data []byte) error {

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

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileSystemDriverMock) Upload(data []byte) error {

	args := fs.Called(data)
	return args.Error(0)
}
