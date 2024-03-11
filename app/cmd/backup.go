package cmd

import (
	"github.com/fatih/color"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/nizigama/ovrsight/foundation/storage"
	"github.com/spf13/cobra"
	"time"
)

var onlyBinlog bool

// BackupCmd represents the databases.backup command
var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup a database to a storage engine",
	Long: `===============
Database backup
===============
This tool will help you backup a given database to one or multiple storage providers like:

- Filesystem
- DropBox
- Google Drive (coming soon)
- S3 (coming soon)

=========
Arguments
=========
You need to pass to the command two arguments, the first shall always be the name of the database you want to backup
and the second, can be omitted, is the driver used to backup the database.
The second argument can be any of the following: filesystem & dropbox.
If the storage driver is omitted, the filesystem driver will be used by default.

Eg:

With storage driver
$ oversight backup demo_db dropbox

Without storage driver. Filesystem will be used by default
$ oversight backup demo_db

Backup only the binary logs
$ oversight backup demo_db dropbox --binlog`,
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
	ValidArgs: []string{storage.FileSystemType, storage.DropboxType},
	RunE: func(cmd *cobra.Command, args []string) error {
		databaseName := args[0]
		storageEngine := storage.FileSystemType

		if len(args) == 2 {

			storageEngine = args[1]
		}

		binlogService, err := services.InitBinlogService(databaseName)
		if err != nil {
			color.Red("Error occurred while preparing binlog manager => %s", err)
			return err
		}

		if active, err := binlogService.IsActive(); !active {

			if err != nil {
				return err
			}

			err = binlogService.Enable(databaseName)
			if err != nil {
				color.Red("Error occurred while enabling binary logs => %s", err)
				return err
			}
		}

		if !onlyBinlog {
			backupService, err := services.InitBackupService(databaseName, storageEngine)
			if err != nil {
				color.Red("Error occurred while preparing backup mechanism up => %s", err)
				return err
			}

			err = backupService.Backup()
			if err != nil {
				color.Red("Error occurred while backing up => %s", err)
				return err
			}

			color.Green("%s => The '%s' database has been successfully backed up using the %s driver\n", time.Now().Format(time.DateTime), databaseName, storageEngine)

			return nil
		}

		err = binlogService.Backup(storageEngine)
		if err != nil {
			color.Red("Error occurred while backing up binary logs => %s", err)
			return err
		}

		color.Green("%s => The '%s' database's binary logs have been successfully backed up using the %s driver\n", time.Now().Format(time.DateTime), databaseName, storageEngine)

		return nil
	},
}

func init() {
	BackupCmd.Flags().BoolVarP(&onlyBinlog, "binlog", "b", false, "Backup binary logs of the last full backup")
}
