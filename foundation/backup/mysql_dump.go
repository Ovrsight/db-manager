package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"os/exec"
	"strconv"
)

type MysqlDump struct {
	Database    string
	host        string
	port        int
	user        string
	password    string
	programPath string
}

type MysqlDumpMock struct {
	mock.Mock
}

func (md *MysqlDump) Initialize() error {

	// mysqldump is available
	programPath, err := exec.LookPath("mysqldump")
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, md.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	md.user = user
	md.password = password
	md.host = host
	md.port = port
	md.programPath = programPath

	return nil
}

func (md *MysqlDump) Generate(sender chan<- []byte, failureChan chan struct{}) error {

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cmd := exec.CommandContext(
		ctx,
		fmt.Sprintf("%s", md.programPath),
		fmt.Sprintf("-u%s", md.user),
		fmt.Sprintf("-p%s", md.password),
		fmt.Sprintf("--host=%s", md.host),
		fmt.Sprintf("--port=%d", md.port),
		// the following option will allow CRUD operations to continue while mysqldump is working
		// it also creates a snapshot of the database prior to backing up to make sure the export keeps consistency and integrity
		fmt.Sprintf("--single-transaction"),
		md.Database,
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
			return nil
		}
	}

	err = cmd.Wait()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		return errors.New(string(execErr.Stderr))
	}

	return nil
}

func (md *MysqlDump) Clean(sender chan<- []byte) error {

	close(sender)

	return nil
}
