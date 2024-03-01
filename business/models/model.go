package models

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

type Database struct {
	gorm.Model
	Name             string
	FirstBackupTime  time.Time
	LatestBackupTime time.Time
	EnableLogging    bool
}

type Backup struct {
	gorm.Model
	DatabaseId int64
	Filename   string
	BackupTime time.Time
	Size       int64
	IsActive   bool
}

type Binlog struct {
	gorm.Model
	BackupId int64
	Filename string
	LogName  string
	Size     int64
	BackedUp bool
}

func Init() *gorm.DB {
	dbFile := os.Getenv("SYSTEM_DB_FILE")

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	err = db.AutoMigrate(&Database{}, &Backup{}, &Binlog{})
	if err != nil {
		log.Fatalln(err)
	}

	return db
}
