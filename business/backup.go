package business

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/databases"
	"github.com/nizigama/ovrsight/foundation/storage"
	"os"
	"strconv"
	"time"
)

type dbConfig struct {
	host     string
	port     int
	user     string
	password string
	database string
	dsn      string
}

type Backer struct {
	StorageDriver storage.Storage
	BackupMethod  databases.BackupMethod
	DatabaseName  string
	Filename      string
}

var (
	config dbConfig
)

func NewBacker(storageDriver, databaseName string) (*Backer, error) {

	filename := fmt.Sprintf("%d_%s.sql", time.Now().UnixNano(), databaseName)

	driver, err := getDriver(storage.StorageDriverType(storageDriver), filename)
	if err != nil {
		return nil, err
	}

	cfg := configure(databaseName)

	method := &databases.MysqlDumper{
		Host:     cfg.host,
		Port:     cfg.port,
		User:     cfg.user,
		Password: cfg.password,
		Database: cfg.database,
	}

	//method := &databases.XtraBackup{}

	return &Backer{
		StorageDriver: driver,
		BackupMethod:  method,
		DatabaseName:  databaseName,
		Filename:      filename,
	}, nil
}

func (bckr *Backer) Backup() error {

	cfg := configure(bckr.DatabaseName)

	if err := ping(cfg); err != nil {
		return err
	}

	bckpData, err := bckr.BackupMethod.Generate()
	if err != nil {
		return err
	}

	err = bckr.StorageDriver.Upload(bckpData)
	if err != nil {
		return err
	}

	return nil
}

func configure(databaseName string) dbConfig {
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, _ := strconv.Atoi(p)

	config = dbConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: databaseName,
		dsn:      fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, databaseName),
	}

	return config
}

func ping(cfg dbConfig) error {

	db, err := sql.Open("mysql", cfg.dsn)
	if err != nil {
		return err
	}

	return db.Ping()
}

func getDriver(driverType storage.StorageDriverType, fileName string) (storage.Storage, error) {

	var storageDriver storage.Storage

	switch driverType {
	case storage.FileSystemType:
		storageDriver = &storage.FileSystemDriver{
			Filename: fileName,
		}

		return storageDriver, nil
	case storage.DropboxType:
		storageDriver := &storage.DropboxDriver{
			Filename: fileName,
		}

		return storageDriver, nil
	case storage.GoogleDriveType:
		storageDriver := &storage.GoogleDriveDriver{
			Filename: fileName,
		}

		return storageDriver, nil
	default:
		storageDriver = &storage.FileSystemDriver{
			Filename: fileName,
		}

		return storageDriver, nil
	}
}
