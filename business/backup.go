package business

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/methods"
	"github.com/nizigama/ovrsight/foundation/models"
	"github.com/nizigama/ovrsight/foundation/storage"
	"log"
	"os"
	"strconv"
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

func (manager *BackupManager) getMasterLog() (string, int, error) {

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

func (manager *BackupManager) Backup() error {

	storageEngine := storage.GetStorageEngine(manager.StorageDriver, manager.Filename, manager.Database)

	wg := sync.WaitGroup{}

	wg.Add(2)
	dataChan := make(chan []byte)
	var backupSuccessful bool

	models.Init()

	// get or create database record
	database := new(models.Database)
	err := database.FindOrCreate(manager.Database)

	if err != nil {
		return err
	}

	// create backup record
	backup := models.Backup{
		DatabaseId: int64(database.ID),
		Filename:   manager.Filename,
		BackupTime: time.Now(),
		IsActive:   true,
	}

	res := models.Db.Create(&backup)
	if res.Error != nil {
		return res.Error
	}

	// get master binary and create bin log record
	binlogName, position, err := manager.getMasterLog()
	if err != nil {
		return err
	}

	binlog := models.Binlog{
		BackupId:  int64(backup.ID),
		Filename:  binlogName,
		Size:      0,
		Position:  int64(position),
		StartTime: time.Now(),
	}

	res = models.Db.Create(&binlog)
	if res.Error != nil {
		return res.Error
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

		res = tx.Delete(&binlog)
		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}

		res := res.Delete(&backup)
		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}

		tx.Commit()
	}

	return nil
}
