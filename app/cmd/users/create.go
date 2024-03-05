package users

import (
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strings"
)

// CreateCmd represents the users:create command
var CreateCmd = &cobra.Command{
	Use:   "users:create",
	Short: "Create a new mysql user",
	Long: `Create a new mysql user. For example:

Eg:

$ oversight users:create`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("Create a new MySQL user")

		hostTypeOptions := []string{"localhost", "Everywhere", "Specific IPs"}
		authenticationOptions := []string{"auth_socket", "mysql_native_password", "caching_sha2_password"}

		username, _ := pterm.DefaultInteractiveTextInput.Show("Username")

		selectedHostType, _ := pterm.DefaultInteractiveSelect.WithOptions(hostTypeOptions).WithDefaultOption(hostTypeOptions[0]).Show("Choose where the user can connect from")

		var selectedHost []string

		switch selectedHostType {
		case hostTypeOptions[0]:
			selectedHost = append(selectedHost, hostTypeOptions[0])
		case hostTypeOptions[1]:
			selectedHost = append(selectedHost, "%")
		case hostTypeOptions[2]:
			hosts, _ := pterm.DefaultInteractiveTextInput.Show("IPs separated with commas")

			selectedHost = strings.Split(hosts, ",")
		}

		selectedAuthMethod, _ := pterm.DefaultInteractiveSelect.WithOptions(authenticationOptions).WithDefaultOption(authenticationOptions[2]).Show("Choose the user's authentication method")

		password, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Password")

		userService, err := services.InitUserService()
		if err != nil {
			return err
		}

		defer userService.Close()

		err = userService.CreateUser(username, selectedAuthMethod, password, selectedHost...)
		if err != nil {
			return err
		}

		color.Green("New user successfully created")

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users:createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users:createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
