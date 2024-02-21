package business

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"github.com/nizigama/ovrsight/foundation/models"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"time"
)

type binlogConf struct {
	format          string
	lifespanInDays  int64
	sizeInMegabytes int64
}

type Binlog struct {
	Name      string
	Size      int64
	encrypted string
}

type BinlogBackupManager struct {
	DB       *sql.DB
	Database string
}

func InitBinlogBackupManager(database string) (*BinlogBackupManager, error) {

	manager := BinlogBackupManager{
		Database: database,
	}

	// get database connection
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	manager.DB = db
	// check binlog program existence
	return &manager, nil
}

func (bm *BinlogBackupManager) Backup() error {

	models.Init()

	var database models.Database

	tx := models.Db.First(&database, "name = ?", bm.Database)
	if tx.Error != nil {

		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	//- Get the active backup
	var bck models.Backup
	tx = models.Db.First(&bck, "is_active = ? AND database_id = ?", true, database.ID)
	if tx.Error != nil {

		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return tx.Error
		}

		return nil
	}

	//- Get all binary logs using `show binary logs`
	rows, err := bm.DB.Query("show binary logs")
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
	tx = models.Db.Find(&savedLogs, "backup_id = ?", bck.ID)
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

			models.Db.Model(savedLogs[idx]).Updates(models.Binlog{BackedUp: false, Size: v.Size})

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
				Filename: v.Name,
				Size:     v.Size,
				Position: 0,
			}

			models.Db.Create(&binlog)

			continue
		}
	}

	// TODO: pass the relay to the job in charge of reading and uploading to storage
	// TODO: use golang's ticker and timer to run tasks and jobs on a schedule instead of using cron jobs,
	// TODO: oversight will run as a linux service
	// TODO: create services, models & jobs folders to contain their specific files
	// TODO: newly created models will have to be added in the `AutoMigrate` call within the models `init` function

	//- For every binary log record that has a `BackedUp` column set to false
	//	- Read its data starting from the `Position` point and back it up in the storage engine
	//	- Set `BackedUp` column to true on successful backup

	return nil
}

func (bm *BinlogBackupManager) Enable() error {

	//
	file, err := os.OpenFile("/etc/mysql/mysql.conf.d/oversight-binlog.cnf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	configuration := fmt.Sprintf(`
[mysqld]

log-bin=oversight-bin
binlog_format=%s
expire_logs_days=%d
max_binlog_size=%dM
`, "ROW", 1, 10)

	_, err = file.WriteString(configuration)
	if err != nil {
		return err
	}

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
		fmt.Println(string(data))
		return err
	}

	programPath, err = exec.LookPath("/etc/init.d/mysql")
	if err != nil {
		return err
	}

	cmd = exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("restart"),
	)

	data, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(data))
		return err
	}

	return nil
}

func (bm *BinlogBackupManager) Disable() error {

	//
	file, err := os.OpenFile("/etc/mysql/mysql.conf.d/oversight-binlog.cnf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	configuration := fmt.Sprintf(`
[mysqld]

disable-log-bin
`)

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
		fmt.Println(string(data))
		return err
	}

	programPath, err = exec.LookPath("/etc/init.d/mysql")
	if err != nil {
		return err
	}

	cmd = exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("restart"),
	)

	data, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(data))
		return err
	}

	return nil
}

// check binlog status

func (bm *BinlogBackupManager) IsActive() (bool, error) {

	// connection to database server is possible
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return false, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err
	}

	err = db.Ping()
	if err != nil {
		return false, err
	}

	row := db.QueryRow("show variables like 'log_bin'")

	if row.Err() != nil {
		return false, nil
	}

	var option, value string

	err = row.Scan(&option, &value)
	if err != nil {
		return false, err
	}

	return value == "ON", nil
}

func (bm *BinlogBackupManager) PurgeLogs() error {

	// connection to database server is possible
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	endOfDay := time.Now().Format(time.DateOnly)

	// use needs `BINLOG_ADMIN` privilege
	// action: GRANT BINLOG_ADMIN ON *.* TO 'user'@'%'; FLUSH PRIVILEGES;
	_, err = db.Exec(fmt.Sprintf("PURGE BINARY LOGS BEFORE '%s 23:59:59'", endOfDay))

	if err != nil {
		return nil
	}

	return nil
}

// close current binary log and open a new one

func (bm *BinlogBackupManager) FlushLogs() error {

	// connection to database server is possible
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	// use needs `RELOAD` privilege
	// action: GRANT RELOAD ON *.* TO 'user'@'%';FLUSH PRIVILEGES;
	_, err = db.Exec("FLUSH BINARY LOGS")

	if err != nil {
		return nil
	}

	return nil
}

// list binary logs

func (bm *BinlogBackupManager) ListLogs() ([]Binlog, error) {

	// connection to database server is possible
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// use needs `REPLICATION CLIENT` privilege
	// action: GRANT REPLICATION CLIENT ON *.* TO 'user'@'%';FLUSH PRIVILEGES;
	rows, err := db.Query("SHOW BINARY LOGS")

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

// get content of a binary log for a specific database

func (bm *BinlogBackupManager) GetAllDatabaseChanges(database, binlogPath string) ([]byte, error) {

	programPath, err := exec.LookPath("mysqlbinlog")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("--database"),
		fmt.Sprintf(database),
		fmt.Sprintf("--disable-log-bin"),
		fmt.Sprintf(binlogPath),
	)

	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// get content of a binary log from/until a certain point in time

func (bm *BinlogBackupManager) GetDatabaseChangesWithinRange(database, binlogPath string, from, until time.Time) ([]byte, error) {

	programPath, err := exec.LookPath("mysqlbinlog")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		fmt.Sprintf("%s", programPath),
		fmt.Sprintf("--database"),
		fmt.Sprintf(database),
		fmt.Sprintf("--disable-log-bin"),
		fmt.Sprintf("--start-datetime=%s", from.Format(time.DateTime)),
		fmt.Sprintf("--stop-datetime=%s", until.Format(time.DateTime)),
		fmt.Sprintf(binlogPath),
	)

	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return data, nil
}
