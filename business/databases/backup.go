package databases

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/foundation/storage"
	"log"
	"os"
	"os/exec"
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

var (
	config dbConfig
)

func Backup(databaseName, storageDriver string) error {

	cfg := configure(databaseName)

	if err := ping(cfg); err != nil {
		return err
	}

	bckpData, err := dumpData(cfg)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%d_%s.sql", time.Now().UnixNano(), databaseName)

	driver, err := getDriver(storage.StorageDriverType(storageDriver), fileName)
	if err != nil {
		return err
	}

	err = driver.Upload(bckpData)
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

func dumpData(cfg dbConfig) ([]byte, error) {

	program, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(fmt.Sprintf("%s", program), fmt.Sprintf("-u%s", cfg.user), fmt.Sprintf("-p%s", cfg.password), cfg.database)

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return out, nil
}
