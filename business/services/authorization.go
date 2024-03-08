package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/nizigama/ovrsight/foundation/rdbms"
	"os"
	"slices"
	"strings"
)

type AuthorizationService struct {
	DB *sql.DB
}

type Privilege struct {
	Name    string
	Granted string
}

var GlobalPrivilegesSet = []map[string]string{
	{"Alter_priv": "ALTER"},
	{"Alter_routine_priv": "ALTER ROUTINE"},
	{"Create_priv": "CREATE"},
	{"Create_role_priv": "CREATE ROLE"},
	{"Create_routine_priv": "CREATE ROUTINE"},
	{"Create_tablespace_priv": "CREATE TABLESPACE"},
	{"Create_tmp_table_priv": "CREATE TEMPORARY TABLES"},
	{"Create_user_priv": "CREATE USER"},
	{"Create_view_priv": "CREATE VIEW"},
	{"Delete_priv": "DELETE"},
	{"Drop_priv": "DROP"},
	{"Drop_role_priv": "DROP ROLE"},
	{"Event_priv": "EVENT"},
	{"Execute_priv": "EXECUTE"},
	{"File_priv": "FILE"},
	{"Grant_priv": "GRANT OPTION"},
	{"Index_priv": "INDEX"},
	{"Insert_priv": "INSERT"},
	{"Lock_tables_priv": "LOCK TABLES"},
	{"Process_priv": "PROCESS"},
	{"References_priv": "REFERENCES"},
	{"Reload_priv": "RELOAD"},
	{"Repl_client_priv": "REPLICATION CLIENT"},
	{"Repl_slave_priv": "REPLICATION SLAVE"},
	{"Select_priv": "SELECT"},
	{"Show_db_priv": "SHOW DATABASES"},
	{"Show_view_priv": "SHOW VIEW"},
	{"Shutdown_priv": "SHUTDOWN"},
	{"Super_priv": "SUPER"},
	{"Trigger_priv": "TRIGGER"},
	{"Update_priv": "UPDATE"},
}
var DatabasePrivilegesSet = []map[string]string{
	{"Select_priv": "Select"},
	{"Insert_priv": "Insert"},
	{"Update_priv": "Update"},
	{"Delete_priv": "Delete"},
	{"Create_priv": "Create"},
	{"Drop_priv": "Drop"},
	{"Grant_priv": "Grant"},
	{"References_priv": "References"},
	{"Index_priv": "Index"},
	{"Alter_priv": "Alter"},
	{"Create_tmp_table_priv": "Create Temporary Table"},
	{"Lock_tables_priv": "Lock Tables"},
	{"Create_view_priv": "Create View"},
	{"Execute_priv": "Execute"},
	{"Event_priv": "Event"},
	{"Trigger_priv": "Trigger"},
	{"Show_view_priv": "Show View"},
	{"Create_routine_priv": "Create Routine"},
	{"Alter_routine_priv": "Alter Routine"},
}
var tablePrivilegesSet = [13]string{"Select", "Insert", "Update", "Delete", "Create", "Drop", "Grant", "References", "Index", "Alter", "Create View", "Show view", "Trigger"}

func InitAuthorizationService() (*AuthorizationService, error) {
	selectedRdbms := os.Getenv("RDBMS")

	dbms := rdbms.GetRdbms(selectedRdbms)

	db, err := dbms.OpenConnection()
	if err != nil {
		return nil, err
	}

	service := AuthorizationService{
		DB: db,
	}

	return &service, nil
}

func (as *AuthorizationService) GetGlobalPrivileges(username, host string) ([]Privilege, error) {

	row := as.DB.QueryRow(`SELECT 
    									Select_priv,Insert_priv,Update_priv,Delete_priv,Create_priv,
    									Drop_priv,Reload_priv,Shutdown_priv,Process_priv,File_priv,Grant_priv,
    									References_priv,Index_priv,Alter_priv,Show_db_priv,Super_priv,Create_tmp_table_priv,
    									Lock_tables_priv,Execute_priv,Repl_slave_priv,Repl_client_priv,Create_view_priv,
    									Show_view_priv,Create_routine_priv,Alter_routine_priv,Create_user_priv,Event_priv,
    									Trigger_priv,Create_tablespace_priv,Create_role_priv,Drop_role_priv
							FROM mysql.user
							WHERE User = ? AND Host = ?`, username, host)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var (
		selectPriv           string
		insertPriv           string
		updatePriv           string
		deletePriv           string
		createPriv           string
		dropPriv             string
		reloadPriv           string
		shutdownPriv         string
		processPriv          string
		filePriv             string
		grantPriv            string
		referencesPriv       string
		indexPriv            string
		alterPriv            string
		showPriv             string
		superPriv            string
		createTmpTablePriv   string
		lockTablesPriv       string
		executePriv          string
		replSlacePriv        string
		replClientPriv       string
		createViewPriv       string
		showViewPriv         string
		createRoutinePriv    string
		alterRoutinePriv     string
		createUserPriv       string
		eventPriv            string
		triggerPriv          string
		createTableSpacePriv string
		createRolePriv       string
		dropRolePriv         string
	)

	err := row.Scan(
		&selectPriv, &insertPriv, &updatePriv, &deletePriv,
		&createPriv, &dropPriv, &reloadPriv, &shutdownPriv,
		&processPriv, &filePriv, &grantPriv, &referencesPriv,
		&indexPriv, &alterPriv, &showPriv, &superPriv, &createTmpTablePriv,
		&lockTablesPriv, &executePriv, &replSlacePriv, &replClientPriv,
		&createViewPriv, &showViewPriv, &createRoutinePriv, &alterRoutinePriv,
		&createUserPriv, &eventPriv, &triggerPriv, &createTableSpacePriv,
		&createRolePriv, &dropRolePriv,
	)
	if err != nil {
		return nil, err
	}

	var privileges []Privilege

	privileges = append(
		privileges,
		Privilege{Name: "Select_priv", Granted: selectPriv},
		Privilege{Name: "Insert_priv", Granted: insertPriv},
		Privilege{Name: "Update_priv", Granted: updatePriv},
		Privilege{Name: "Delete_priv", Granted: deletePriv},
		Privilege{Name: "Create_priv", Granted: createPriv},
		Privilege{Name: "Drop_priv", Granted: dropPriv},
		Privilege{Name: "Reload_priv", Granted: reloadPriv},
		Privilege{Name: "Shutdown_priv", Granted: shutdownPriv},
		Privilege{Name: "Process_priv", Granted: processPriv},
		Privilege{Name: "File_priv", Granted: filePriv},
		Privilege{Name: "Grant_priv", Granted: grantPriv},
		Privilege{Name: "References_priv", Granted: referencesPriv},
		Privilege{Name: "Index_priv", Granted: indexPriv},
		Privilege{Name: "Alter_priv", Granted: alterPriv},
		Privilege{Name: "Show_db_priv", Granted: showPriv},
		Privilege{Name: "Super_priv", Granted: superPriv},
		Privilege{Name: "Create_tmp_table_priv", Granted: createTmpTablePriv},
		Privilege{Name: "Lock_tables_priv", Granted: lockTablesPriv},
		Privilege{Name: "Execute_priv", Granted: executePriv},
		Privilege{Name: "Repl_slave_priv", Granted: replSlacePriv},
		Privilege{Name: "Repl_client_priv", Granted: replClientPriv},
		Privilege{Name: "Create_view_priv", Granted: createViewPriv},
		Privilege{Name: "Show_view_priv", Granted: showViewPriv},
		Privilege{Name: "Create_routine_priv", Granted: createRoutinePriv},
		Privilege{Name: "Alter_routine_priv", Granted: alterRoutinePriv},
		Privilege{Name: "Create_user_priv", Granted: createUserPriv},
		Privilege{Name: "Event_priv", Granted: eventPriv},
		Privilege{Name: "Trigger_priv", Granted: triggerPriv},
		Privilege{Name: "Create_tablespace_priv", Granted: createTableSpacePriv},
		Privilege{Name: "Create_role_priv", Granted: createRolePriv},
		Privilege{Name: "Drop_role_priv", Granted: dropRolePriv},
	)

	return privileges, nil
}

func (as *AuthorizationService) GetDatabasePrivileges(username, host, database string) ([]Privilege, error) {

	row := as.DB.QueryRow(`SELECT 
    									Select_priv,Insert_priv,Update_priv,Delete_priv,Create_priv,
    									Drop_priv,Grant_priv, References_priv,Index_priv,Alter_priv,
    									Create_tmp_table_priv,Lock_tables_priv,Create_view_priv,
    									Execute_priv,Event_priv,Trigger_priv,Show_view_priv,Create_routine_priv,Alter_routine_priv
							FROM mysql.db
							WHERE User = ? AND Host = ? AND Db = ?`, username, host, database)

	if err := row.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	selectPriv := "N"
	insertPriv := "N"
	updatePriv := "N"
	deletePriv := "N"
	createPriv := "N"
	dropPriv := "N"
	grantPriv := "N"
	referencesPriv := "N"
	indexPriv := "N"
	alterPriv := "N"
	createTmpTablePriv := "N"
	lockTablesPriv := "N"
	executePriv := "N"
	createViewPriv := "N"
	showViewPriv := "N"
	createRoutinePriv := "N"
	alterRoutinePriv := "N"
	eventPriv := "N"
	triggerPriv := "N"

	err := row.Scan(
		&selectPriv, &insertPriv, &updatePriv, &deletePriv,
		&createPriv, &dropPriv, &grantPriv, &referencesPriv,
		&indexPriv, &alterPriv, &createTmpTablePriv, &lockTablesPriv,
		&createViewPriv, &executePriv, &eventPriv, &triggerPriv,
		&showViewPriv, &createRoutinePriv, &alterRoutinePriv,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var privileges []Privilege

	privileges = append(
		privileges,
		Privilege{Name: "Select_priv", Granted: selectPriv},
		Privilege{Name: "Insert_priv", Granted: insertPriv},
		Privilege{Name: "Update_priv", Granted: updatePriv},
		Privilege{Name: "Delete_priv", Granted: deletePriv},
		Privilege{Name: "Create_priv", Granted: createPriv},
		Privilege{Name: "Drop_priv", Granted: dropPriv},
		Privilege{Name: "Grant_priv", Granted: grantPriv},
		Privilege{Name: "References_priv", Granted: referencesPriv},
		Privilege{Name: "Index_priv", Granted: indexPriv},
		Privilege{Name: "Alter_priv", Granted: alterPriv},
		Privilege{Name: "Create_tmp_table_priv", Granted: createTmpTablePriv},
		Privilege{Name: "Lock_tables_priv", Granted: lockTablesPriv},
		Privilege{Name: "Create_view_priv", Granted: createViewPriv},
		Privilege{Name: "Execute_priv", Granted: executePriv},
		Privilege{Name: "Event_priv", Granted: eventPriv},
		Privilege{Name: "Trigger_priv", Granted: triggerPriv},
		Privilege{Name: "Show_view_priv", Granted: showViewPriv},
		Privilege{Name: "Create_routine_priv", Granted: createRoutinePriv},
		Privilege{Name: "Alter_routine_priv", Granted: alterRoutinePriv},
	)

	return privileges, nil
}

func (as *AuthorizationService) GetTablePrivileges(username, host, database, table string) ([]Privilege, []Privilege, error) {

	row := as.DB.QueryRow(`SELECT	Table_priv,Column_priv
										FROM mysql.tables_priv
										WHERE User = ? AND Host = ? AND Db = ? AND Table_name = ?`, username, host, database, table)

	if err := row.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	var tablePriv string
	var columnPriv string

	err := row.Scan(&tablePriv, &columnPriv)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	tablePrivileges := append(
		[]Privilege{},
		Privilege{Name: "Select", Granted: "No"},
		Privilege{Name: "Insert", Granted: "No"},
		Privilege{Name: "Update", Granted: "No"},
		Privilege{Name: "Delete", Granted: "No"},
		Privilege{Name: "Create", Granted: "No"},
		Privilege{Name: "Drop", Granted: "No"},
		Privilege{Name: "Grant", Granted: "No"},
		Privilege{Name: "References", Granted: "No"},
		Privilege{Name: "Index", Granted: "No"},
		Privilege{Name: "Alter", Granted: "No"},
		Privilege{Name: "Create View", Granted: "No"},
		Privilege{Name: "Show View", Granted: "No"},
		Privilege{Name: "Trigger", Granted: "No"},
	)

	columnPrivileges := append(
		[]Privilege{},
		Privilege{Name: "Select", Granted: "No"},
		Privilege{Name: "Insert", Granted: "No"},
		Privilege{Name: "Update", Granted: "No"},
		Privilege{Name: "References", Granted: "No"},
	)

	tblPrivs := strings.Split(tablePriv, ",")
	clsPrivs := strings.Split(columnPriv, ",")

	for _, v := range tblPrivs {

		idx := slices.IndexFunc(tablePrivileges, func(privilege Privilege) bool {
			return privilege.Name == v
		})

		if idx != -1 {
			tablePrivileges[idx].Granted = "Yes"
		}

	}

	for _, v := range clsPrivs {

		idx := slices.IndexFunc(columnPrivileges, func(privilege Privilege) bool {
			return privilege.Name == v
		})

		if idx != -1 {
			columnPrivileges[idx].Granted = "Yes"
		}

	}

	return tablePrivileges, columnPrivileges, nil
}

func (as *AuthorizationService) GetAllDatabases() ([]string, error) {

	query := `show databases`

	rows, err := as.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var databases []string

	for rows.Next() {
		var db string

		err = rows.Scan(&db)
		if err != nil {
			return nil, err
		}

		databases = append(databases, db)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return databases, nil

}

func (as *AuthorizationService) GetAllDatabaseTables(database string) ([]string, error) {

	_, err := as.DB.Exec(fmt.Sprintf(`use %s`, database))
	if err != nil {
		return nil, err
	}

	rows, err := as.DB.Query(`show tables`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tables []string

	for rows.Next() {
		var tbl string

		err = rows.Scan(&tbl)
		if err != nil {
			return nil, err
		}

		tables = append(tables, tbl)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil

}

func (as *AuthorizationService) UpdateGlobalPrivileges(username, host string, privileges []string) error {

	tx, err := as.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("REVOKE ALL PRIVILEGES ON *.* FROM '%s'@'%s'", username, host))
	if err != nil {
		tx.Rollback()
		return err
	}

	selectedPrivileges := strings.Join(privileges, ",")

	query := fmt.Sprintf("GRANT %s ON *.* TO '%s'@'%s'", selectedPrivileges, username, host)

	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("FLUSH PRIVILEGES")
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (as *AuthorizationService) UpdateDatabasePrivileges(username, host, database string, privileges []string) error {

	row := as.DB.QueryRow(`SELECT Select_priv
										FROM mysql.db
										WHERE User = ? AND Host = ? AND Db = ?`, username, host, database)
	err := row.Err()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	alreadyHasPrivileges := errors.Is(err, sql.ErrNoRows)

	tx, err := as.DB.Begin()
	if err != nil {
		return err
	}

	if alreadyHasPrivileges {
		_, err = tx.Exec(fmt.Sprintf("REVOKE ALL PRIVILEGES ON %s.* FROM '%s'@'%s'", database, username, host))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	selectedPrivileges := strings.Join(privileges, ",")

	query := fmt.Sprintf("GRANT %s ON %s.* TO '%s'@'%s'", selectedPrivileges, database, username, host)

	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("FLUSH PRIVILEGES")
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (as *AuthorizationService) Close() error {

	if as.DB != nil {
		return as.DB.Close()
	}

	return nil
}
