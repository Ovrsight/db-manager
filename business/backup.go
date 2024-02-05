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

func Init(database string, storageDriver string) (*BackupManager, error) {
	filename := fmt.Sprintf("%d_%s.sql", time.Now().UnixNano(), database)

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

	mysqlDump := &methods.MysqlDump{
		Database: database,
	}

	if err := mysqlDump.Initialize(); err != nil {
		return nil, err
	}

	return &BackupManager{
		Database:      database,
		Filename:      filename,
		BackupMethod:  mysqlDump,
		StorageDriver: driver,
	}, nil
}

func (manager *BackupManager) Backup() error {

	// generate backup bytes
	data, err := manager.BackupMethod.Generate()
	if err != nil {
		return err
	}

	fmt.Println(string(data), len(data))

	// upload backup bytes

	// backup method cleaner

	return nil
}
