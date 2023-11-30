package storage

type Storage interface {
	Upload([]byte) error
}

type StorageDriverType string

const (
	FileSystemType  StorageDriverType = "filesystem"
	DropboxType     StorageDriverType = "dropbox"
	GoogleDriveType StorageDriverType = "googledrive"
)
