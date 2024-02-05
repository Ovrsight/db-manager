package methods

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type MysqlDump struct {
	host        string
	port        int
	user        string
	password    string
	database    string
	programPath string
}

type MysqlDumpMock struct {
	mock.Mock
}

func (md *MysqlDump) Initialize(database string) error {

	// mysqldump is available
	_, err := exec.LookPath("mysqldump")
	if err != nil {
		return err
	}

	// connection to database server is possible
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (md *MysqlDump) Generate() ([]byte, error) {

	program, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: add host and port
	cmd := exec.Command(fmt.Sprintf("%s", program), fmt.Sprintf("-u%s", md.user), fmt.Sprintf("-p%s", md.password), md.database)

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return out, nil
}

func (md *MysqlDump) Clean() error {

	// nothing for this method

	return nil
}

func (mdm *MysqlDumpMock) Generate() ([]byte, error) {

	args := mdm.Called()
	return args.Get(0).([]byte), args.Error(1)
}
