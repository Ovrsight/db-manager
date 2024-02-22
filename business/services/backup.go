package services

import (
	"errors"
	"fmt"
	"github.com/nizigama/ovrsight/business/jobs"
	"github.com/nizigama/ovrsight/business/models"
	"github.com/nizigama/ovrsight/foundation/backup"
	"github.com/nizigama/ovrsight/foundation/storage"
	"gorm.io/gorm"
	"os"
	"time"
)

type BackupService struct {
	BackupMethod  backup.Method
	StorageEngine storage.Engine
	Database      string
	Filename      string
	DB            *gorm.DB
}

func InitBackupService(database, storageEngine string) (*BackupService, error) {

	selectedMethod := os.Getenv("BACKUP_METHOD")

	method := backup.GetBackupMethod(selectedMethod, database)

	filename := fmt.Sprintf("%d_full.sql", time.Now().Unix())

	engine := storage.GetStorageEngine(storageEngine, database, filename)

	if err := method.Initialize(); err != nil {
		return nil, err
	}

	db := models.Init()

	return &BackupService{
		BackupMethod:  method,
		StorageEngine: engine,
		DB:            db,
		Database:      database,
		Filename:      filename,
	}, nil
}

func (bckp *BackupService) Backup() error {

	// get or create database record
	databaseModel := new(models.Database)
	backupModel := new(models.Backup)
	binlogModel := new(models.Binlog)

	err := databaseModel.CreateIfNotExisting(bckp.DB, bckp.Database)
	if err != nil {
		return err
	}

	err = bckp.DB.Transaction(func(tx *gorm.DB) error {

		// create backup record
		backupModel.DatabaseId = int64(databaseModel.ID)
		backupModel.Filename = bckp.Filename
		backupModel.BackupTime = time.Now()
		backupModel.IsActive = true

		err := tx.Model(&models.Backup{}).Where("is_active = ? AND database_id = ?", true, databaseModel.ID).Update("is_active", false).Error
		if err != nil {
			return err
		}

		err = tx.Create(backupModel).Error
		if err != nil {
			return err
		}

		binlogService, err := InitBinlogService(bckp.Database)
		if err != nil {
			return err
		}

		defer binlogService.Close()

		// get master binary and create bin log record
		binlogName, position, err := binlogService.GetMasterLog()
		if err != nil {
			return err
		}

		binlogModel.BackupId = int64(backupModel.ID)
		binlogModel.Filename = fmt.Sprintf("%s_%d", binlogName, time.Now().Unix())
		binlogModel.LogName = binlogName
		binlogModel.Size = position
		binlogModel.Position = position

		err = tx.Create(binlogModel).Error
		if err != nil {
			return err
		}

		backupSuccessful := new(jobs.BackupProcessor).ProcessBackup(bckp.BackupMethod, bckp.StorageEngine)
		if err != nil {
			return err
		}

		if !backupSuccessful {
			return errors.New("failed to process the backup")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (bckp *BackupService) Close() error {

	if bckp.DB != nil {
		db, _ := bckp.DB.DB()
		db.Close()
	}

	return nil
}
