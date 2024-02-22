package rdbms

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strconv"
)

type Mysql struct {
}

func (m *Mysql) OpenConnection() (*sql.DB, error) {

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

	return db, nil
}
