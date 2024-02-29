package services

import (
	"errors"
	"fmt"
	"github.com/nizigama/ovrsight/business/models"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"github.com/nizigama/ovrsight/foundation/storage"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"time"
)

type RecoveryService struct {
	StorageEngine storage.Engine
	Database      string
	PointInTime   time.Time
	DB            *gorm.DB
	Filename      string
	Rdbms         rdbms.Rdbms
	binlogService *BinlogService
}

func InitRecoveryService(db, storageEngine string, moment time.Time) (*RecoveryService, error) {

	dbConn := models.Init()

	selectedRdbms := os.Getenv("RDBMS")

	dbms := rdbms.GetRdbms(selectedRdbms)

	_, err := dbms.OpenConnection()
	if err != nil {
		return nil, err
	}

	engine := storage.GetStorageEngine(storageEngine, db, "")

	logService, err := InitBinlogService(db)
	if err != nil {
		return nil, err
	}

	service := RecoveryService{
		Database:      db,
		PointInTime:   moment,
		DB:            dbConn,
		Rdbms:         dbms,
		StorageEngine: engine,
		binlogService: logService,
	}

	return &service, nil
}

func (rs *RecoveryService) Recover() error {

	//- check if database backup is available and also if the point in time is not before the `FirstBackupTime` value
	database := models.Database{}
	err := rs.DB.Where("name = ?", rs.Database).First(&database).Error
	if err != nil {
		return err
	}

	if rs.PointInTime.Before(database.FirstBackupTime) {
		return errors.New("select point in time is before the first backup record")
	}

	//- get a backup whose `BackupTime` is right before the given point in time
	backup := models.Backup{}
	err = rs.DB.Where("database_id = ? AND backup_time <= ?", database.ID, rs.PointInTime).Last(&backup).Error
	if err != nil {
		return err
	}

	rs.Filename = backup.Filename

	//- check if there's enough space on the server's disk to download the backup
	dfCommand, err := exec.LookPath("df")
	if err != nil {
		return err
	}

	cmd := exec.Command(
		dfCommand,
		"--output=size,avail,used",
		"/",
	)

	outputData, err := cmd.Output()
	if err != nil {
		return err
	}

	var totalSize int64
	var availableSize int64
	var usedSize int64

	_, err = fmt.Sscanf(string(outputData), "1K-blocks\tAvail\t\tUsed\n%d%d%d", &totalSize, &availableSize, &usedSize)
	if err != nil {
		return err
	}

	// the size returned by the df command is in KB, this is to convert it into bytes
	availableSize = availableSize * 1000

	if backup.Size >= availableSize {
		return errors.New("no space left on the disk to download the backup")
	}

	//- get the total size of all the binary log files of the backup and check if there's enough space on the server's disk to download them locally

	var logs []models.Binlog
	err = rs.DB.Model(&models.Binlog{}).Where("backup_id = ?", backup.ID).Find(&logs).Error
	if err != nil {
		return err
	}

	var logsTotalSize int64

	for _, v := range logs {
		logsTotalSize += v.Size
	}

	if logsTotalSize >= availableSize {
		return errors.New("no space left on the disk to download binary log files")
	}

	if (logsTotalSize + backup.Size) >= availableSize {
		return errors.New("no space left on the disk to download all the needed backup files")
	}

	//- download backup file
	filesLocations, err := rs.StorageEngine.Retrieve(rs.Filename)
	if err != nil {
		return err
	}

	//- import backup using mysql
	err = rs.binlogService.Disable(rs.Database)
	if err != nil {
		return err
	}

	err = rs.Rdbms.Restore(filesLocations[0], rs.Database)
	if err != nil {
		return err
	}

	err = rs.binlogService.Enable(rs.Database)
	if err != nil {
		return err
	}

	err = rs.StorageEngine.DeleteRetrievals(filesLocations[0])
	if err != nil {
		return err
	}

	var logsNames []string

	for _, v := range logs {
		logsNames = append(logsNames, v.Filename)
	}

	filesLocations, err = rs.StorageEngine.Retrieve(logsNames...)
	if err != nil {
		return err
	}

	err = rs.binlogService.ApplyLogChanges(rs.Database, rs.PointInTime, filesLocations...)
	if err != nil {
		return err
	}

	err = rs.StorageEngine.DeleteRetrievals(filesLocations...)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RecoveryService) Close() error {

	rs.binlogService.Close()

	if rs.DB != nil {
		db, _ := rs.DB.DB()
		db.Close()
	}

	rs.Rdbms.Close()

	return nil
}
