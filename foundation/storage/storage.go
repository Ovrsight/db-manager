package storage

type Storage interface {
	Upload([]byte) error
}

type DriverType string

const (
	FileSystemType  DriverType = "filesystem"
	DropboxType     DriverType = "dropbox"
	GoogleDriveType DriverType = "google_drive"
)
