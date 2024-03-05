package services

import (
	"database/sql"
	"fmt"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"os"
)

type UserService struct {
	DB *sql.DB
}

type UserInfo struct {
	Host                 string
	Username             string
	SystemMaxConnections int
	UserMaxConnections   int
	AuthenticationMethod string
	AccountLocked        string
}

func InitUserService() (*UserService, error) {

	selectedRdbms := os.Getenv("RDBMS")

	dbms := rdbms.GetRdbms(selectedRdbms)

	db, err := dbms.OpenConnection()
	if err != nil {
		return nil, err
	}

	service := UserService{
		DB: db,
	}

	return &service, nil
}

func (us *UserService) CreateUser(username, authMethod, password string, hosts ...string) error {

	for _, h := range hosts {

		query := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED WITH %s BY '%s'", username, h, authMethod, password)
		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (us *UserService) ListUsers() ([]UserInfo, error) {

	rows, err := us.DB.Query("SELECT Host,User,max_connections,max_user_connections,plugin as authenticationMethod,account_locked FROM mysql.user")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []UserInfo

	for rows.Next() {

		var user UserInfo

		err = rows.Scan(&user.Host, &user.Username, &user.SystemMaxConnections, &user.UserMaxConnections, &user.AuthenticationMethod, &user.AccountLocked)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) Close() error {

	if us.DB != nil {
		return us.DB.Close()
	}

	return nil
}
