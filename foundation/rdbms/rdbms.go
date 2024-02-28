package rdbms

import (
	"database/sql"
	"log"
	"strings"
)

type Rdbms interface {
	OpenConnection() (*sql.DB, error)
	Close() error
	Restore(backupFile, databaseName string) error
}

const (
	MysqlServer string = "mysql"
)

func GetRdbms(rdbmsName string) Rdbms {
	switch rdbmsName {
	case strings.ToLower(MysqlServer):
		return new(Mysql)
	default:
		log.Fatalln("unsupported rdbms")
	}
	return nil
}
