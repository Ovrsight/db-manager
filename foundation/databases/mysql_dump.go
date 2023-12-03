package databases

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"log"
	"os/exec"
)

type MysqlDumper struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type MysqlDumperMock struct {
	mock.Mock
}

func (md *MysqlDumper) Generate() ([]byte, error) {

	program, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(fmt.Sprintf("%s", program), fmt.Sprintf("-u%s", md.User), fmt.Sprintf("-p%s", md.Password), md.Database)

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return out, nil
}

func (mdm *MysqlDumperMock) Generate() ([]byte, error) {

	args := mdm.Called()
	return args.Get(0).([]byte), args.Error(1)
}
