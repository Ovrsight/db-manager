package business

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type binlogConf struct {
	format          string
	lifespanInDays  int64
	sizeInMegabytes int64
}

type BinlogManager struct {
}

func (bm BinlogManager) Enable() error {

	//
	file, err := os.OpenFile("/etc/mysql/mysql.conf.d/oversight-binlog.cnf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	configuration := fmt.Sprintf(`
[mysqld]

log-bin=mysql-bin
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

func (bm BinlogManager) Disable() error {

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

func (bm BinlogManager) IsActive() (bool, error) {

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

// purge binary logs

func (bm BinlogManager) PurgeLogs() error {

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
	// action: GRANT BINLOG_ADMIN ON *.* TO 'user'@'%';
	_, err = db.Exec(fmt.Sprintf("PURGE BINARY LOGS BEFORE '%s 23:59:59'", endOfDay))

	if err != nil {
		return nil
	}

	return nil
}

// close current binary log and open a new one
// list binary logs
// get content of a binary log for a specific database
// get content of (a) binary log(s) from/until a certain point in time
// !!! REMEMBER TO USE --disable-log-bin WHEN READING BINARY LOG DATA TO AVOID AN ENDLESS LOOP OF LOGS !!!
