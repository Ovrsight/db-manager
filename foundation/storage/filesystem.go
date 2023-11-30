package storage

import (
	"fmt"
	"os"
	"strings"
)

type FileSystemDriver struct {
	Filename string
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
