package services

import (
	"database/sql"
	"fmt"
	"github.com/go-playground/validator/v10"
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

type NewUser struct {
	Username   string `validate:"required"`
	AuthMethod string `validate:"required"`
	Password   string
	Hosts      []string `validate:"dive,required,ipv4"`
	Localhost  bool
	Everywhere bool
	Locked     bool
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

func (us *UserService) CreateUser(user NewUser) error {

	validate := validator.New()

	err := validate.Struct(user)
	if err != nil {
		return err
	}

	switch {
	case user.Localhost:
		query := fmt.Sprintf("CREATE USER '%s'@'localhost' IDENTIFIED WITH %s BY '%s'", user.Username, user.AuthMethod, user.Password)

		if user.Locked {
			query = fmt.Sprintf("%s ACCOUNT LOCK", query)
		}

		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	case user.Everywhere:
		query := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED WITH %s BY '%s'", user.Username, user.AuthMethod, user.Password)

		if user.Locked {
			query = fmt.Sprintf("%s ACCOUNT LOCK", query)
		}

		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	default:
		for _, h := range user.Hosts {

			query := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED WITH %s BY '%s'", user.Username, h, user.AuthMethod, user.Password)

			if user.Locked {
				query = fmt.Sprintf("%s ACCOUNT LOCK", query)
			}

			_, err := us.DB.Exec(query)
			if err != nil {
				return err
			}
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
