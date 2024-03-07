package users

import (
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strconv"
)

// ListCmd represents the users:list command
var ListCmd = &cobra.Command{
	Use:   "users:list",
	Short: "List all mysql users",
	Long: `List all mysql users. For example:

Eg:

$ oversight users:list`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("MySQL users list")

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
			{"Client origin", "Username", "Using password", "System allowed connections", "User allowed connections", "Auth method", "Locked"},
		}

		for _, user := range users {

			systemMaxConnections := strconv.Itoa(user.SystemMaxConnections)
			userMaxConnections := strconv.Itoa(user.UserMaxConnections)

			locked := "No"

			if user.AccountLocked == "Y" {
				locked = "Yes"
			}

			tableData = append(tableData, []string{user.Host, user.Username, user.UsingPassword, systemMaxConnections, userMaxConnections, user.AuthenticationMethod, locked})
		}

		err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users:listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users:listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
