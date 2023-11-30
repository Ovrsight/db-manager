package databases

import (
	"fmt"
	"github.com/nizigama/ovrsight/business/databases"

	"github.com/spf13/cobra"
)

// BackupCmd represents the databases.backup command
var BackupCmd = &cobra.Command{
	Use:   "databases:backup",
	Short: "Select a database to backup using one or multiple storage systems",
	Long: `===============
Database backup
===============
This tool will help you backup a given database to one or multiple storage providers like:

- DropBox
- Google Drive
- S3 (coming soon)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("databases:backup run")

		return databases.Backup(args[0], args[1])
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// databases.backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// BackupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
