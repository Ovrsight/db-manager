package backup

import (
	"context"
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
	LogName          string
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

func (mb *MysqlBinlog) Generate(sender chan<- []byte, failureChan chan struct{}) error {

	binlogPath := fmt.Sprintf("/var/lib/mysql/%s", mb.LogName)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cmd := exec.CommandContext(
		ctx,
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

	savingBackupFailed := false
	completed := make(chan struct{}, 1)

	go func(flag *bool) {

		select {
		case _ = <-failureChan:
			*flag = true
			break
		case _ = <-completed:
			break
		}
	}(&savingBackupFailed)
	defer func() {
		completed <- struct{}{}
	}()

	for {

		content := make([]byte, bufferSize) // reading 5MB

		read, err := outPipe.Read(content)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if !savingBackupFailed {
			sender <- content[:read]
		}

		if savingBackupFailed {
			ctx.Done()
			break
		}
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
