package cmd

import (
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strconv"
)

var update bool

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure MySQL server",
	Long: `Configure MySQL server. For example:

Eg:

$ oversight config`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pterm.DefaultHeader.
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
			WithFullWidth(true).
			Println("Manage server's configurations")

		service, err := services.InitConfigurationService()
		if err != nil {
			return err
		}

		configs, err := service.GetConfigurations()
		if err != nil {
			return err
		}

		tableData := pterm.TableData{
			{"Configuration", "Value"},
			{"Max connections", strconv.Itoa(configs.MaxConnections)},
			{"Allows remote connections", strconv.FormatBool(configs.AllowsRemoteConnections)},
			{"Server port", strconv.Itoa(configs.ServerPort)},
			{"Log slow queries", strconv.FormatBool(configs.LogsSlowQueries)},
			{"General logging", strconv.FormatBool(configs.GeneralLogging)},
			{"Long query time", strconv.Itoa(configs.LongQueryTime)},
		}

		err = pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
		if err != nil {
			return err
		}

		if !update {
			return nil
		}

		options := []string{"Max connections", "Allow remote connections", "Server port", "Log slow queries", "General logging", "Long query time"}

		selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithMaxHeight(10).Show("Choose a configuration to update")

		switch selectedOption {
		case options[0]:
			selectedMax, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue(strconv.Itoa(configs.MaxConnections)).Show("Enter new max number")

			maxConns, err := strconv.Atoi(selectedMax)
			if err != nil {
				return err
			}

			configs.MaxConnections = maxConns
		case options[1]:
			configs.AllowsRemoteConnections, _ = pterm.DefaultInteractiveConfirm.WithDefaultValue(configs.AllowsRemoteConnections).Show("Allow remote connections")

		case options[2]:
			selectedPort, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue(strconv.Itoa(configs.ServerPort)).Show("Enter new port number")

			port, err := strconv.Atoi(selectedPort)
			if err != nil {
				return err
			}

			configs.ServerPort = port
		case options[3]:
			configs.LogsSlowQueries, _ = pterm.DefaultInteractiveConfirm.WithDefaultValue(configs.LogsSlowQueries).Show("Log slow queries")
		case options[4]:
			configs.GeneralLogging, _ = pterm.DefaultInteractiveConfirm.WithDefaultValue(configs.GeneralLogging).Show("Log all queries")
		case options[5]:
			selectedDuration, _ := pterm.DefaultInteractiveTextInput.WithDefaultValue(strconv.Itoa(configs.LongQueryTime)).Show("Enter duration of a long slow query")

			duration, err := strconv.Atoi(selectedDuration)
			if err != nil {
				return err
			}

			configs.LongQueryTime = duration
		}

		err = service.UpdateConfiguration(configs)
		if err != nil {
			return err
		}

		color.Green("\nServer configurations successfully updated\n")

		return nil
	},
}

func init() {
	ConfigCmd.Flags().BoolVarP(&update, "update", "u", false, "Update the server's configurations")
}
