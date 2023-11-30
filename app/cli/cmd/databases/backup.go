package databases

import (
	"github.com/nizigama/ovrsight/business/databases"
	"log"

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
- S3 (coming soon)

=========
Arguments
=========
You need to pass to the command two arguments, the first shall always be the name of the database want to backup
and the second, can be omitted, is the driver used to backup the database.
The second argument can be any of the following: filesystem, dropbox, googledrive.
If the storage driver is omitted, the filesystem driver will be used by default.

Eg:

With storage driver
$ oversight databases:backup demo_db dropbox

Without storage driver. Filesystem will used by default
$ oversight databases:backup demo_db`,
	Args: func(cmd *cobra.Command, args []string) error {

		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
			return err
		}

		if len(args) == 1 {
			return nil
		}

		if err := cobra.OnlyValidArgs(cmd, args[1:]); err != nil {
			return err
		}

		return nil
	},
	ValidArgs: []string{"filesystem", "dropbox", "googledrive"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var databaseName string
		storageDriver := "filesystem"

		databaseName = args[0]

		if len(args) == 2 {
			storageDriver = args[1]
		}

		err := databases.Backup(databaseName, storageDriver)
		if err != nil {
			return err
		}

		log.Printf("The '%s' database has been successfully backed up using the %s driver\n", databaseName, storageDriver)

		return nil
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
