package cmd

import (
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business"
	"github.com/spf13/cobra"
	"time"
)

// BackupCmd represents the databases.backup command
var BackupCmd = &cobra.Command{
	Use:   "backup",
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
$ oversight backup demo_db dropbox

Without storage driver. Filesystem will be used by default
$ oversight backup demo_db`,
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
	ValidArgs: business.GetSupportedStorageDrivers(),
	RunE: func(cmd *cobra.Command, args []string) error {
		databaseName := args[0]
		storageDriver := business.GetDefaultStorageDriver()

		if len(args) == 2 {

			storageDriver = args[1]
		}

		backupManager, err := business.Init(databaseName, storageDriver)
		if err != nil {
			color.Red("Error occurred while preparing backup method up => %s", err)
			return err
		}

		err = backupManager.Backup()
		if err != nil {
			color.Red("Error occurred while backing up => %s", err)
			return err
		}

		color.Green("%s => The '%s' database has been successfully backed up using the %s driver\n", time.Now().Format(time.DateTime), databaseName, storageDriver)

		return nil
	},
}
