package users

import (
	"errors"
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strconv"
)

// ViewPrivilegesCmd represents the users:view command
var ViewPrivilegesCmd = &cobra.Command{
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
			Println("View a user's privileges")

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

		switch selectedOption {
		case options[0]:

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

			color.Green("\nGlobal privileges\n")
			err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
			if err != nil {
				return err
			}
		case options[1]:

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
		case options[2]:

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

			tablePrivileges, columnPrivileges, err := authService.GetTablePrivileges(users[id].Username, users[id].Host, selectedDatabase, selectedTable)
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

			tableData = pterm.TableData{
				{"Name", "Granted"},
			}

			for _, priv := range columnPrivileges {

				tableData = append(tableData, []string{priv.Name, priv.Granted})
			}

			color.Green("\nColumn privileges: %s\n", selectedTable)
			err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
			if err != nil {
				return err
			}

		}

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users:viewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users:viewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
