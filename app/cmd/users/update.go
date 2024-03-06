package users

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

// UpdateCmd represents the users:update command
var UpdateCmd = &cobra.Command{
	Use:   "users:update",
	Short: "Update a mysql user",
	Long: `Update a mysql user. For example:

Eg:

$ oversight users:update`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("Update a MySQL user")

		userService, err := services.InitUserService()
		if err != nil {
			return err
		}

		defer userService.Close()

		users, err := userService.ListUsers()
		if err != nil {
			return err
		}

		tableData := pterm.TableData{
			{"Id", "Client origin", "Username", "Authentication plugin", "Using password", "Locked"},
		}

		for i, user := range users {

			locked := "No"

			if user.AccountLocked == "Y" {
				locked = "Yes"
			}

			i = i + 1

			tableData = append(tableData, []string{strconv.Itoa(i), user.Host, user.Username, user.AuthenticationMethod, user.UsingPassword, locked})
		}

		err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
		if err != nil {
			return err
		}

		options := []string{"Username or/and Host", "Password", "Authentication method", "Lock status"}

		useId, _ := pterm.DefaultInteractiveTextInput.Show("Choose a user id")

		id, err := strconv.Atoi(useId)
		if err != nil {
			return err
		}

		if id < 1 || id > len(tableData)-1 {
			return errors.New("invalid user id")
		}

		selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption(options[0]).Show("Choose what to update")
		hostTypeOptions := []string{"Localhost", "Everywhere", "Specific IP"}

		switch selectedOption {
		case options[0]:
			// update username-host

			defaultHost := ""

			switch tableData[id][1] {
			case "localhost":
				defaultHost = "localhost"
			case "%":
				defaultHost = "Everywhere"
			default:
				defaultHost = "Specific IP"
			}

			username, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue(tableData[id][2]).Show("Enter updated username")

			selectedHostType, _ := pterm.DefaultInteractiveSelect.WithOptions(hostTypeOptions).WithDefaultOption(defaultHost).Show("Choose where the user can connect from")

			var selectedHost string
			var localhost bool
			var everywhere bool

			switch selectedHostType {
			case hostTypeOptions[0]:
				localhost = true
			case hostTypeOptions[1]:
				everywhere = true
			case hostTypeOptions[2]:
				selectedHost, _ = pterm.DefaultInteractiveTextInput.WithDefaultValue(tableData[id][1]).Show("IP address")
			}

			updates := services.UsernameHostUpdate{
				Username:        tableData[id][2],
				UpdatedUsername: username,
				Host:            tableData[id][1],
				UpdatedHost:     strings.TrimSpace(selectedHost),
				Localhost:       localhost,
				Everywhere:      everywhere,
			}

			err = userService.UpdateUsernameHost(updates)
			if err != nil {
				return err
			}

			color.Green("Username and/or host successfully updated")

		case options[1]:
			// update password

			newPassword, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter new password")

			err = userService.UpdateUserPassword(tableData[id][2], tableData[id][1], newPassword)
			if err != nil {
				return err
			}

			color.Green("Password successfully updated")
		case options[2]:
			// update authentication method

			authenticationOptions := []string{"auth_socket", "mysql_native_password", "caching_sha2_password"}

			selectedAuthMethod, _ := pterm.DefaultInteractiveSelect.WithOptions(authenticationOptions).WithDefaultOption(tableData[id][3]).Show("Choose the new authentication method")

			password, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter user's password")

			err = userService.UpdateUserAuthenticationPlugin(tableData[id][2], tableData[id][1], selectedAuthMethod, password)
			if err != nil {
				return err
			}

			color.Green("Authentication plugin successfully updated")
		case options[3]:
			// update lock status

			message := "Unlock account"

			if tableData[id][5] == "No" {
				message = "Lock account"
			}

			fmt.Println(tableData[id][5])

			result, _ := pterm.DefaultInteractiveConfirm.Show(message)

			lock := tableData[id][5] == "No" && result

			err = userService.UpdateUserLockStatus(tableData[id][2], tableData[id][1], lock)
			if err != nil {
				return err
			}

			if tableData[id][5] == "No" {
				color.Green("Account successfully locked")
			} else {
				color.Green("Account successfully unlocked")
			}
		}

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users:updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users:updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
