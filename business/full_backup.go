package business

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/methods"
	"github.com/nizigama/ovrsight/foundation/models"
	"github.com/nizigama/ovrsight/foundation/storage"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type FullBackupManager struct {
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

func InitBackupManager(database string, storageDriver string) (*FullBackupManager, error) {
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

	return &FullBackupManager{
		Database:      database,
		Filename:      filename,
		BackupMethod:  mysqlDump,
		StorageDriver: driver,
	}, nil
}

func (manager *FullBackupManager) getMasterLog() (string, int, error) {

	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", 0, err
	}

	err = db.Ping()
	if err != nil {
		return "", 0, err
	}

	var filename string
	var position int
	var doDb string
	var ignoreDb string
	var gtidSet string

	err = db.QueryRow("SHOW MASTER STATUS").Scan(&filename, &position, &doDb, &ignoreDb, &gtidSet)
	if err != nil {
		return "", 0, err
	}

	return filename, position, nil
}

func (manager *FullBackupManager) Backup() error {

	storageEngine := storage.GetStorageEngine(manager.StorageDriver, manager.Filename, manager.Database)

	wg := sync.WaitGroup{}

	wg.Add(2)
	dataChan := make(chan []byte)
	var backupSuccessful bool
	var backupId int64
	var binlogId int64

	models.Init()

	// get or create database record
	database := new(models.Database)

	err := database.FindOrCreate(manager.Database)
	if err != nil {
		return err
	}

	err = models.Db.Transaction(func(tx *gorm.DB) error {

		// create backup record
		backup := models.Backup{
			DatabaseId: int64(database.ID),
			Filename:   manager.Filename,
			BackupTime: time.Now(),
			IsActive:   true,
		}

		err := tx.Model(&models.Backup{}).Where("is_active = ? AND database_id = ?", true, database.ID).Update("is_active", false).Error
		if err != nil {
			return err
		}

		err = tx.Create(&backup).Error
		if err != nil {
			return err
		}
		backupId = int64(backup.ID)

		// get master binary and create bin log record
		binlogName, position, err1 := manager.getMasterLog()
		if err1 != nil {
			return err1
		}

		binlog := models.Binlog{
			BackupId: int64(backup.ID),
			Filename: binlogName,
			Size:     int64(position),
			Position: int64(position),
		}

		err = tx.Create(&binlog).Error
		if err != nil {
			return err
		}
		binlogId = int64(binlog.ID)

		return nil
	})
	if err != nil {
		return err
	}

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
		backupSuccessful = true
	}(storageEngine)

	wg.Wait()

	if !backupSuccessful {

		tx := models.Db.Begin()

		fmt.Println("binlog and backup ids", binlogId, backupId)

		err = tx.Delete(&models.Binlog{}, binlogId).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Delete(&models.Backup{}, backupId).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
	}

	return nil
}
