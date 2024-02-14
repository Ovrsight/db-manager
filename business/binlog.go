package business

import (
	"fmt"
	"os"
	"os/exec"
)

type binlogConf struct {
	format          string
	lifespanInDays  int64
	sizeInMegabytes int64
}

type BinlogManager struct {
}

// enable binlog

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

// disable binlog

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
// purge binary logs
// close current binary log and open a new one
// list binary logs
// get content of a binary log for a specific database
// get content of (a) binary log(s) from/until a certain point in time
// !!! REMEMBER TO USE --disable-log-bin WHEN READING BINARY LOG DATA TO AVOID AN ENDLESS LOOP OF LOGS !!!
