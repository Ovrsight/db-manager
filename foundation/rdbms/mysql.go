package rdbms

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"os/exec"
	"strconv"
)

type Mysql struct {
	host     string
	port     int
	user     string
	password string
	dsn      string
	Conn     *sql.DB
}

func (m *Mysql) OpenConnection() (*sql.DB, error) {

	var err error

	// get database connection
	m.host = os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	m.user = os.Getenv("DB_USER")
	m.password = os.Getenv("DB_PASSWORD")

	m.port, err = strconv.Atoi(p)
	if err != nil {
		return nil, err
	}

	m.dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", m.user, m.password, m.host, m.port)

	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	m.Conn = db

	return db, nil
}

func (m *Mysql) Restore(backupFile, databaseName string) error {

	mysqlProgram, err := exec.LookPath("mysql")
	if err != nil {
		return err
	}

	cmd := exec.Command(
		mysqlProgram,
		"-u",
		m.user,
		fmt.Sprintf("-p%s", m.password),
		"-e",
		fmt.Sprintf("SET autocommit=0; source %s; COMMIT;", backupFile),
		databaseName,
	)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (m *Mysql) Close() error {

	if m.Conn == nil {
		return nil
	}

	return m.Conn.Close()
}
