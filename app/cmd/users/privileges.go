package users

import (
	"errors"
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"slices"
	"strconv"
)

var update bool

// PrivilegesCmd represents the users:view command
var PrivilegesCmd = &cobra.Command{
	Use:   "users:privileges",
	Short: "Get a user's privileges details",
	Long: `Get a user's privileges details. For example:

Eg:

$ oversight users:privileges`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("Manage a user's privileges")

		userService, err := services.InitAuthenticationService()
		if err != nil {
			return err
		}

		defer userService.Close()

		users, err := userService.ListUsers()
		if err != nil {
			return err
		}

		tableData := pterm.TableData{
			{"Id", "Client origin", "Username"},
		}

		for i, user := range users {

			i = i + 1

			tableData = append(tableData, []string{strconv.Itoa(i), user.Host, user.Username})
		}

		err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
		if err != nil {
			return err
		}

		useId, _ := pterm.DefaultInteractiveTextInput.Show("Choose a user id")

		id, err := strconv.Atoi(useId)
		if err != nil {
			return err
		}

		if id < 1 || id > len(tableData)-1 {
			return errors.New("invalid user id")
		}

		id -= 1

		options := []string{"Global level", "Database level", "Table level"}

		selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption(options[1]).Show("Choose a level")

		authService, err := services.InitAuthorizationService()
		if err != nil {
			return err
		}

		defer authService.Close()

		switch {
		case options[0] == selectedOption && !update:

			err = showGlobalPrivileges(authService, users, id)
			if err != nil {
				return err
			}
		case options[0] == selectedOption && update:

			err = updateGlobalPrivileges(authService, users, id)
			if err != nil {
				return err
			}
		case options[1] == selectedOption && !update:

			err = showDatabasePrivileges(authService, users, id)
			if err != nil {
				return err
			}
		case options[1] == selectedOption && update:

			err = updateDatabasePrivileges(authService, users, id)
			if err != nil {
				return err
			}
		case options[2] == selectedOption && !update:

			err = showTablePrivileges(authService, users, id)
			if err != nil {
				return err
			}
		case options[2] == selectedOption && update:

			err = updateTablePrivileges(authService, users, id)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	PrivilegesCmd.Flags().BoolVarP(&update, "update", "u", false, "Update a user's privileges")
}

func showGlobalPrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {
	privileges, err := authService.GetGlobalPrivileges(users[id].Username, users[id].Host)
	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"Name", "Granted"},
	}

	for _, priv := range privileges {

		granted := "Yes"

		if priv.Granted == "N" {
			granted = "No"
		}

		tableData = append(tableData, []string{priv.Name, granted})
	}

	color.Green("\nGlobal privileges: %s\n", users[id].Username)
	err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	if err != nil {
		return err
	}

	return nil
}

func updateGlobalPrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {

	var options []string
	var defaults []string

	for _, v := range services.GlobalPrivilegesSet {

		value := maps.Values(v)[0]

		options = append(options, value)
	}

	privs, err := authService.GetGlobalPrivileges(users[id].Username, users[id].Host)
	if err != nil {
		return err
	}

	for _, priv := range privs {

		if priv.Granted == "N" {
			continue
		}

		idx := slices.IndexFunc(services.GlobalPrivilegesSet, func(m map[string]string) bool {

			key := maps.Keys(m)[0]

			return key == priv.Name
		})

		value := maps.Values(services.GlobalPrivilegesSet[idx])[0]

		defaults = append(defaults, value)
	}

	selectedPrivileges, _ := pterm.DefaultInteractiveMultiselect.WithOptions(options).WithMaxHeight(15).WithDefaultOptions(defaults).Show()

	err = authService.UpdateGlobalPrivileges(users[id].Username, users[id].Host, selectedPrivileges)
	if err != nil {
		return err
	}

	color.Green("\nGlobal privileges successfully updated for '%s'\n", users[id].Username)
	return nil
}

func showDatabasePrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {
	databases, err := authService.GetAllDatabases()
	if err != nil {
		return err
	}

	selectedDatabase, _ := pterm.DefaultInteractiveSelect.WithOptions(databases).Show("Choose a database")

	privileges, err := authService.GetDatabasePrivileges(users[id].Username, users[id].Host, selectedDatabase)
	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"Name", "Granted"},
	}

	for _, priv := range privileges {

		granted := "Yes"

		if priv.Granted == "N" {
			granted = "No"
		}

		tableData = append(tableData, []string{priv.Name, granted})
	}

	color.Green("\nDatabase privileges: %s\n", selectedDatabase)
	err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	if err != nil {
		return err
	}

	return nil
}

func updateDatabasePrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {

	var options []string
	var defaults []string

	for _, v := range services.DatabasePrivilegesSet {

		value := maps.Values(v)[0]

		options = append(options, value)
	}

	databases, err := authService.GetAllDatabases()
	if err != nil {
		return err
	}

	selectedDatabase, _ := pterm.DefaultInteractiveSelect.WithOptions(databases).Show("Choose a database")

	privs, err := authService.GetDatabasePrivileges(users[id].Username, users[id].Host, selectedDatabase)
	if err != nil {
		return err
	}

	for _, priv := range privs {

		if priv.Granted == "N" {
			continue
		}

		idx := slices.IndexFunc(services.DatabasePrivilegesSet, func(m map[string]string) bool {

			key := maps.Keys(m)[0]

			return key == priv.Name
		})

		if idx == -1 {
			continue
		}

		value := maps.Values(services.DatabasePrivilegesSet[idx])[0]

		defaults = append(defaults, value)
	}

	selectedPrivileges, _ := pterm.DefaultInteractiveMultiselect.WithOptions(options).WithMaxHeight(15).WithDefaultOptions(defaults).Show()

	err = authService.UpdateDatabasePrivileges(users[id].Username, users[id].Host, selectedDatabase, selectedPrivileges)
	if err != nil {
		return err
	}

	color.Green("\nDatabase privileges successfully updated for '%s' on '%s'\n", users[id].Username, selectedDatabase)
	return nil
}

func updateTablePrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {

	var tableOptions []string
	var tableDefaults []string

	for _, v := range services.TablePrivilegesSet {

		value := maps.Values(v)[0]

		tableOptions = append(tableOptions, value)
	}

	databases, err := authService.GetAllDatabases()
	if err != nil {
		return err
	}

	selectedDatabase, _ := pterm.DefaultInteractiveSelect.WithOptions(databases).Show("Choose a database")

	tables, err := authService.GetAllDatabaseTables(selectedDatabase)
	if err != nil {
		return err
	}

	selectedTable, _ := pterm.DefaultInteractiveSelect.WithOptions(tables).WithMaxHeight(10).Show("Choose a table")

	tablePrivs, err := authService.GetTablePrivileges(users[id].Username, users[id].Host, selectedDatabase, selectedTable)
	if err != nil {
		return err
	}

	for _, priv := range tablePrivs {

		if priv.Granted == "No" {
			continue
		}

		idx := slices.IndexFunc(services.TablePrivilegesSet, func(m map[string]string) bool {

			key := maps.Keys(m)[0]

			return key == priv.Name
		})

		if idx == -1 {
			continue
		}

		value := maps.Values(services.TablePrivilegesSet[idx])[0]

		tableDefaults = append(tableDefaults, value)
	}

	selectedPrivileges, _ := pterm.DefaultInteractiveMultiselect.WithOptions(tableOptions).WithMaxHeight(15).WithDefaultOptions(tableDefaults).Show()

	err = authService.UpdateTablePrivileges(users[id].Username, users[id].Host, selectedDatabase, selectedTable, selectedPrivileges)
	if err != nil {
		return err
	}

	color.Green("\nTable privileges successfully updated for '%s' on '%s.%s'\n", users[id].Username, selectedDatabase, selectedTable)

	return nil
}

func showTablePrivileges(authService *services.AuthorizationService, users []services.UserInfo, id int) error {
	databases, err := authService.GetAllDatabases()
	if err != nil {
		return err
	}

	selectedDatabase, _ := pterm.DefaultInteractiveSelect.WithOptions(databases).Show("Choose a database")

	tables, err := authService.GetAllDatabaseTables(selectedDatabase)
	if err != nil {
		return err
	}

	selectedTable, _ := pterm.DefaultInteractiveSelect.WithOptions(tables).WithMaxHeight(10).Show("Choose a table")

	tablePrivileges, err := authService.GetTablePrivileges(users[id].Username, users[id].Host, selectedDatabase, selectedTable)
	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"Name", "Granted"},
	}

	for _, priv := range tablePrivileges {

		tableData = append(tableData, []string{priv.Name, priv.Granted})
	}

	color.Green("\nTable privileges: %s\n", selectedTable)
	err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	if err != nil {
		return err
	}

	return nil
}
