package services

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"github.com/nizigama/ovrsight/business/jobs"
	"github.com/nizigama/ovrsight/business/models"
	"github.com/nizigama/ovrsight/foundation/backup"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"github.com/nizigama/ovrsight/foundation/storage"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"slices"
	"time"
)

type BinlogService struct {
	Rdbms    *sql.DB
	DB       *gorm.DB
	Database string
}

type Binlog struct {
	Name      string
	Size      int64
	encrypted string
}

func InitBinlogService(database string) (*BinlogService, error) {

	service := BinlogService{
		Database: database,
	}

	selectedRdbms := os.Getenv("RDBMS")

	dbms := rdbms.GetRdbms(selectedRdbms)

	db, err := dbms.OpenConnection()
	if err != nil {
		return nil, err
	}

	service.Rdbms = db
	service.DB = models.Init()

	// check binlog program existence
	_, err = exec.LookPath("mysqlbinlog")
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (bs *BinlogService) Backup(storageEngine string) error {

	var database models.Database

	tx := bs.DB.First(&database, "name = ?", bs.Database)
	if tx.Error != nil {

		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	//- Get the active backup
	var bck models.Backup
	tx = bs.DB.First(&bck, "is_active = ? AND database_id = ?", true, database.ID)
	if tx.Error != nil {

		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	//- Get all binary logs using `show binary logs`
	rows, err := bs.Rdbms.Query("show binary logs")
	if err != nil {
		return err
	}

	defer rows.Close()

	var logs []Binlog

	for rows.Next() {
		log := Binlog{}
		err = rows.Scan(&log.Name, &log.Size, &log.encrypted)
		if err != nil {
			return err
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// get the last binlog of the backup
	var savedLogs []models.Binlog
	tx = bs.DB.Find(&savedLogs, "backup_id = ?", bck.ID)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	slices.SortStableFunc(savedLogs, func(a, b models.Binlog) int {
		return cmp.Compare(a.Filename, b.Filename)
	})

	firstLog := savedLogs[0]
	var needToBeBackedUp []models.Binlog

	for _, v := range logs {

		idx, found := slices.BinarySearchFunc(savedLogs, Binlog{Name: v.Name}, func(a models.Binlog, b Binlog) int {
			return cmp.Compare(a.Filename, b.Name)
		})

		// it means that the binary log has got new changes saved
		if found && v.Size > savedLogs[idx].Position {

			savedLogs[idx].BackedUp = false
			savedLogs[idx].Size = v.Size
			needToBeBackedUp = append(needToBeBackedUp, savedLogs[idx])

			bs.DB.Model(savedLogs[idx]).Updates(models.Binlog{BackedUp: false, Size: v.Size})

			continue
		}

		// it means this is a new binary log file
		if !found && v.Name > firstLog.Filename {

			needToBeBackedUp = append(needToBeBackedUp, models.Binlog{
				BackupId: int64(bck.ID),
				Filename: v.Name,
				Size:     v.Size,
				Position: 0,
				BackedUp: false,
			})

			binlog := models.Binlog{
				BackupId: int64(bck.ID),
				Filename: fmt.Sprintf("%s_%d", v.Name, time.Now().Unix()),
				LogName:  v.Name,
				Size:     v.Size,
				Position: 0,
			}

			bs.DB.Create(&binlog)

			continue
		}
	}

	err = bs.ProcessBinLogs(storageEngine)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BinlogService) GetMasterLog() (string, int64, error) {

	var filename string
	var position int64
	var doDb string
	var ignoreDb string
	var gtidSet string

	err := bs.Rdbms.QueryRow("SHOW MASTER STATUS").Scan(&filename, &position, &doDb, &ignoreDb, &gtidSet)
	if err != nil {
		return "", 0, err
	}

	return filename, position, nil
}

func (bs *BinlogService) Close() error {

	if bs.Rdbms != nil {

		bs.Rdbms.Close()
	}

	if bs.DB != nil {
		db, _ := bs.DB.DB()
		db.Close()
	}

	return nil
}

func (bs *BinlogService) ProcessBinLogs(storageEngine string) error {

	var unprocessedLogs []models.Binlog
	tx := bs.DB.Find(&unprocessedLogs, "backed_up = ?", false)
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	for _, log := range unprocessedLogs {

		method := &backup.MysqlBinlog{
			Database:         bs.Database,
			StartingPosition: log.Position,
			Filename:         log.Filename,
			LogName:          log.LogName,
		}

		if err := method.Initialize(); err != nil {
			return err
		}

		engine := storage.GetStorageEngine(storageEngine, bs.Database, log.Filename)

		backupSuccessful := new(jobs.BackupProcessor).ProcessBackup(method, engine)

		if !backupSuccessful {
			return errors.New("failed to process the backup")
		}

		bs.DB.Model(&log).Updates(models.Binlog{BackedUp: true})
	}

	return nil
}
