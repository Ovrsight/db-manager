package users

import (
	"errors"
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteCmd represents the users:delete command
var DeleteCmd = &cobra.Command{
	Use:   "users:delete",
	Short: "Delete a mysql user",
	Long: `Delete a mysql user. For example:

Eg:

$ oversight users:delete`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("Delete a MySQL user")

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

		confirm, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to delete the user")

		if !confirm {
			color.Yellow("Deletion cancelled")
			return nil
		}

		err = userService.DeleteUser(tableData[id][2], tableData[id][1])
		if err != nil {
			return err
		}

		color.Green("User successfully cancelled")

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users:deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users:deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
