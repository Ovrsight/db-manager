package services

import (
	"database/sql"
	"fmt"
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type ConfigurationService struct {
	Rdbms *sql.DB
}

type Config struct {
	MaxConnections          int
	AllowsRemoteConnections bool
	ServerPort              int
	LogsSlowQueries         bool
	LongQueryTime           int
	GeneralLogging          bool
}

func InitConfigurationService() (*ConfigurationService, error) {
	service := ConfigurationService{}

	selectedRdbms := os.Getenv("RDBMS")

	dbms := rdbms.GetRdbms(selectedRdbms)

	db, err := dbms.OpenConnection()
	if err != nil {
		return nil, err
	}

	service.Rdbms = db

	return &service, nil
}

func (cs *ConfigurationService) GetConfigurations() (Config, error) {

	cnf := Config{}

	var (
		name  string
		value string
	)

	row := cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'max_connections'")

	err := row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	cnf.MaxConnections, _ = strconv.Atoi(value)

	row = cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'bind_address'")

	err = row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	if value == "0.0.0.0" {
		cnf.AllowsRemoteConnections = true
	}

	row = cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'port'")

	err = row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	cnf.ServerPort, _ = strconv.Atoi(value)

	row = cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'long_query_time'")

	err = row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	queryTime, _ := strconv.ParseFloat(value, 64)

	cnf.LongQueryTime = int(queryTime)

	row = cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'general_log'")

	err = row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	if value == "ON" {
		cnf.GeneralLogging = true
	}

	row = cs.Rdbms.QueryRow("SHOW VARIABLES LIKE 'slow_query_log'")

	err = row.Scan(&name, &value)
	if err != nil {
		return Config{}, nil
	}

	if value == "ON" {
		cnf.LogsSlowQueries = true
	}

	return cnf, nil
}

func (cs *ConfigurationService) UpdateConfiguration(config Config) error {

	file, err := os.OpenFile("/etc/mysql/mysql.conf.d/oversight.cnf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	bindAddress := "0.0.0.0"
	logSlowQueries := 0
	logEverything := 0

	if !config.AllowsRemoteConnections {
		bindAddress = "127.0.0.1"
	}

	if config.LogsSlowQueries {
		logSlowQueries = 1
	}

	if config.GeneralLogging {
		logEverything = 1
	}

	configuration := fmt.Sprintf(`
[mysqld]

max_connections = %d
bind-address = %s
port          = %d
slow-query-log = %d
general-log = %d
log-output = TABLE
long_query_time = %d
log_error = /var/log/mysql/oversight-error.log
`, config.MaxConnections, bindAddress, config.ServerPort, logSlowQueries, logEverything, config.LongQueryTime)

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
		fmt.Sprintf("restart"),
	)

	color.HiBlue("\nRestarting mysql server...\n")
	data, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(data))
		return err
	}

	return nil
}
