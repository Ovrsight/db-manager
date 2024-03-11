package services

import (
	"database/sql"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"os"
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

	cnf.LongQueryTime, _ = strconv.Atoi(value)

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
