package databases

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type dbConfig struct {
	host     string
	port     int
	user     string
	password string
	database string
}

var (
	config dbConfig
)

func configure() string {
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")

	port, _ := strconv.Atoi(p)

	config = dbConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.user, config.password, config.host, config.port, config.database)
}

func Ping() error {

	dsn := configure()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	return db.Ping()
}

func Execute(dbName string) ([]byte, error) {

	program, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(fmt.Sprintf("%s", program), fmt.Sprintf("-u%s", "root"), fmt.Sprintf("-p%s", ""), "oversight")

	out, err := cmd.Output()
	if err != nil {
		cerr := &exec.ExitError{}
		errors.As(err, &cerr)
		log.Fatalln(string(cerr.Stderr))
	}

	return out, nil
}
