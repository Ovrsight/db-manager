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
	UsingPassword        string
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

type UsernameHostUpdate struct {
	Username        string
	UpdatedUsername string `validate:"required"`
	Host            string
	UpdatedHost     string `validate:"omitempty,ipv4"`
	Localhost       bool
	Everywhere      bool
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

	rows, err := us.DB.Query("SELECT Host,User,authentication_string,max_connections,max_user_connections,plugin as authenticationMethod,account_locked FROM mysql.user")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []UserInfo

	for rows.Next() {

		var user UserInfo
		var usingPassword string

		err = rows.Scan(&user.Host, &user.Username, &usingPassword, &user.SystemMaxConnections, &user.UserMaxConnections, &user.AuthenticationMethod, &user.AccountLocked)
		if err != nil {
			return nil, err
		}

		if usingPassword == "" {
			user.UsingPassword = "No"
		} else {
			user.UsingPassword = "Yes"
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (us *UserService) UpdateUsernameHost(updates UsernameHostUpdate) error {

	validate := validator.New()

	err := validate.Struct(updates)
	if err != nil {
		return err
	}

	switch {
	case updates.Localhost:

		if updates.Username == updates.UpdatedUsername && updates.Host == "localhost" {
			return nil
		}

		query := fmt.Sprintf("RENAME USER '%s'@'%s' TO '%s'@'locahost'", updates.Username, updates.Host, updates.UpdatedUsername)
		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	case updates.Everywhere:

		if updates.Username == updates.UpdatedUsername && updates.Host == "%" {
			return nil
		}

		query := fmt.Sprintf("RENAME USER '%s'@'%s' TO '%s'@'%%'", updates.Username, updates.Host, updates.UpdatedUsername)
		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	default:

		if updates.Username == updates.UpdatedUsername && updates.Host == updates.UpdatedHost {
			return nil
		}

		query := fmt.Sprintf("RENAME USER '%s'@'%s' TO '%s'@'%s'", updates.Username, updates.Host, updates.UpdatedUsername, updates.UpdatedHost)
		_, err := us.DB.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (us *UserService) UpdateUserPassword(username, host, password string) error {

	query := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := us.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) UpdateUserAuthenticationPlugin(username, host, authMethod, password string) error {

	tx, err := us.DB.Begin()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED WITH '%s'", username, host, authMethod)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) UpdateUserLockStatus(username, host string, lock bool) error {

	query := fmt.Sprintf("ALTER USER '%s'@'%s'", username, host)

	if lock {
		query = fmt.Sprintf("%s ACCOUNT LOCK", query)
	} else {
		query = fmt.Sprintf("%s ACCOUNT UNLOCK", query)
	}

	_, err := us.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) DeleteUser(username, host string) error {

	query := fmt.Sprintf("DROP USER '%s'@'%s'", username, host)
	_, err := us.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserService) Close() error {

	if us.DB != nil {
		return us.DB.Close()
	}

	return nil
}
