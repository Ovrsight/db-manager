package cmd

import (
	"github.com/nizigama/ovrsight/business/services"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"strconv"
)

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

		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
