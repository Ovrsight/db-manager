package business

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/methods"
	"github.com/nizigama/ovrsight/foundation/storage"
	"log"
	"sync"
	"time"
)

type BackupManager struct {
	Database      string
	Filename      string
	BackupMethod  methods.BackupMethod
	StorageDriver storage.EngineType
}

func GetSupportedStorageDrivers() []string {
	return []string{
		string(storage.FileSystemType),
		string(storage.DropboxType),
	}
}

func GetDefaultStorageDriver() string {

	return string(storage.FileSystemType)
}

func Init(database string, storageDriver string) (*BackupManager, error) {
	filename := fmt.Sprintf("%d_full.sql", time.Now().Unix())

	var driver storage.EngineType

	switch storageDriver {
	case "filesystem":
		driver = storage.FileSystemType
	case "dropbox":
		driver = storage.DropboxType
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

	storageEngine := storage.GetStorageEngine(manager.StorageDriver, manager.Filename, manager.Database)

	wg := sync.WaitGroup{}

	wg.Add(2)
	dataChan := make(chan []byte)

	go func() {
		// backup method cleaner
		defer manager.BackupMethod.Clean(dataChan)
		defer wg.Done()

		// generate backup bytes
		err := manager.BackupMethod.Generate(dataChan)

		if err != nil {
			log.Fatalln("Backup failure:", err)
		}
	}()

	go func(se storage.Engine) {

		defer wg.Done()

		// upload backup bytes
		err := se.Save(dataChan)

		if err != nil {
			log.Fatalln("Storage failure:", err)
		}
	}(storageEngine)

	wg.Wait()

	return nil
}
