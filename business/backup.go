package business

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/methods"
	"github.com/nizigama/ovrsight/foundation/storage"
	"time"
)

type BackupManager struct {
	Database      string
	Filename      string
	BackupMethod  methods.BackupMethod
	StorageDriver storage.DriverType
}

func GetSupportedStorageDrivers() []string {
	return []string{
		string(storage.FileSystemType),
		string(storage.DropboxType),
		string(storage.GoogleDriveType),
	}
}

func GetDefaultStorageDriver() string {

	return string(storage.FileSystemType)
}

func Init(databaseName string, storageDriver string) *BackupManager {
	filename := fmt.Sprintf("%d_%s.sql", time.Now().UnixNano(), databaseName)

	var driver storage.DriverType

	switch storageDriver {
	case "filesystem":
		driver = storage.FileSystemType
	case "dropbox":
		driver = storage.DropboxType
	case "google_drive":
		driver = storage.GoogleDriveType
	default:
		driver = storage.FileSystemType
	}

	// initialize backup method

	return &BackupManager{
		Database:      databaseName,
		Filename:      filename,
		BackupMethod:  nil,
		StorageDriver: driver,
	}
}

func (manager *BackupManager) Backup() error {

	// generate backup bytes

	// upload backup bytes

	// backup method cleaner

	return nil
}
