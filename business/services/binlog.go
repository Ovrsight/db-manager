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
	"log"
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

		err := new(jobs.BackupProcessor).ProcessBackup(method, engine, nil)

		if err != nil {
			return err
		}

		bs.DB.Model(&log).Updates(models.Binlog{BackedUp: true})
	}

	return nil
}

func (bs *BinlogService) Enable(strictlyThese ...string) error {

	if len(strictlyThese) > 0 {
		err := bs.DB.Model(&models.Database{}).Where(map[string]interface{}{"name": strictlyThese}).Update("enable_logging", true).Error
		if err != nil {
			return err
		}
	} else {
		err := bs.DB.Model(&models.Database{}).Update("enable_logging", true).Error
		if err != nil {
			return err
		}
	}

	err := bs.updateConfiguration()
	if err != nil {
		return err
	}

	return nil
}

func (bs *BinlogService) Disable(strictlyThese ...string) error {

	if len(strictlyThese) > 0 {
		err := bs.DB.Model(&models.Database{}).Where(map[string]interface{}{"name": strictlyThese}).Update("enable_logging", false).Error
		if err != nil {
			return err
		}
	} else {
		err := bs.DB.Model(&models.Database{}).Update("enable_logging", false).Error
		if err != nil {
			return err
		}
	}

	err := bs.updateConfiguration()
	if err != nil {
		return err
	}

	return nil
}

func (bs *BinlogService) IsActive() (bool, error) {

	row := bs.Rdbms.QueryRow("show variables like 'log_bin'")

	if row.Err() != nil {
		return false, nil
	}

	var option, value string

	err := row.Scan(&option, &value)
	if err != nil {
		return false, err
	}

	return value == "ON", nil
}

func (bs *BinlogService) PurgeLogs() error {

	endOfDay := time.Now().Format(time.DateOnly)

	// user needs `BINLOG_ADMIN` privilege
	// action: GRANT BINLOG_ADMIN ON *.* TO 'user'@'%'; FLUSH PRIVILEGES;
	_, err := bs.Rdbms.Exec(fmt.Sprintf("PURGE BINARY LOGS BEFORE '%s 23:59:59'", endOfDay))

	if err != nil {
		return nil
	}

	return nil
}

func (bs *BinlogService) FlushLogs() error {

	// user needs `RELOAD` privilege
	// action: GRANT RELOAD ON *.* TO 'user'@'%';FLUSH PRIVILEGES;
	_, err := bs.Rdbms.Exec("FLUSH BINARY LOGS")

	if err != nil {
		return nil
	}

	return nil
}

func (bs *BinlogService) ListLogs() ([]Binlog, error) {

	// user needs `REPLICATION CLIENT` privilege
	// action: GRANT REPLICATION CLIENT ON *.* TO 'user'@'%';FLUSH PRIVILEGES;
	rows, err := bs.Rdbms.Query("SHOW BINARY LOGS")

	if err != nil {
		return nil, nil
	}

	var logs []Binlog

	for rows.Next() {
		log := Binlog{}

		err = rows.Scan(&log.Name, &log.Size, &log.encrypted)
		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
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

func (bs *BinlogService) updateConfiguration() error {

	file, err := os.OpenFile("/etc/mysql/mysql.conf.d/oversight-binlog.cnf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	var databases []models.Database

	err = bs.DB.Find(&databases).Error
	if err != nil {
		return err
	}

	var activeDatabases []models.Database
	var inactiveDatabases []models.Database

	for _, v := range databases {

		if v.EnableLogging {
			activeDatabases = append(activeDatabases, v)
			continue
		}

		inactiveDatabases = append(inactiveDatabases, v)
	}

	var configuration string

	if len(databases) == len(inactiveDatabases) {
		configuration = fmt.Sprintf(`
[mysqld]

disable-log-bin
`)
	} else {

		configuration = fmt.Sprintf(`
[mysqld]

log-bin=oversight-bin
binlog_format=%s
binlog_expire_logs_seconds=%d
max_binlog_size=%dM
`, "ROW", 60*60*24, 10)

		for _, v := range inactiveDatabases {
			configuration += fmt.Sprintf("\nbinlog-ignore-db=%s", v.Name)
		}
	}

	_, err = file.WriteString(configuration)
	if err != nil {
		return err
	}

	// restart mysql
	programPath, err := exec.LookPath("mysqld")
	if err != nil {
		return err
	}

	cmd := exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("--validate-config"),
	)

	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(data))
		return err
	}

	programPath, err = exec.LookPath("/etc/init.d/mysql")
	if err != nil {
		return err
	}

	cmd = exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("reload"),
	)

	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(data))
		return err
	}

	return nil
}
