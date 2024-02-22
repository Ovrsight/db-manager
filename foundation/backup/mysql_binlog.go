package backup

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type MysqlBinlog struct {
	Database         string
	StartingPosition int64
	Filename         string
	host             string
	port             int
	user             string
	password         string
	programPath      string
}

func (mb *MysqlBinlog) Initialize() error {

	// mysqldump is available
	programPath, err := exec.LookPath("mysqlbinlog")
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, mb.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	mb.user = user
	mb.password = password
	mb.host = host
	mb.port = port
	mb.programPath = programPath

	return nil
}

func (mb *MysqlBinlog) Generate(sender chan<- []byte) error {

	binlogPath := fmt.Sprintf("/var/lib/mysql/%s", mb.Filename)

	cmd := exec.Command(
		fmt.Sprintf("%s", mb.programPath),
		fmt.Sprintf("--database"),
		fmt.Sprintf(mb.Database),
		fmt.Sprintf("--disable-log-bin"),
		fmt.Sprintf("--start-position=%d", mb.StartingPosition),
		fmt.Sprintf(binlogPath),
	)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	defer outPipe.Close()

	err = cmd.Start()
	if err != nil {
		return err
	}

	for {

		content := make([]byte, 5000000) // reading 5MB

		read, err := outPipe.Read(content)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		sender <- content[:read]
	}

	err = cmd.Wait()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return nil
}

func (mb *MysqlBinlog) Clean(sender chan<- []byte) error {

	close(sender)

	return nil
}
