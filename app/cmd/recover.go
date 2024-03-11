package cmd

import (
	"errors"
	"github.com/nizigama/ovrsight/business/services"
	"github.com/nizigama/ovrsight/foundation/storage"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// RecoverCmd represents the recover command
var RecoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover a database from a storage engine",
	Long: `================
Database recover
================

This tool will help you recover a backed up database from a storage engine using point in time recover technology.

=========
Arguments
=========
You need to pass to the command two arguments, the first shall always be the name of the database you want to recover
and the second is the datetime that will be used as the point in time to not exceed while recovering.

The second argument can be any of the following:
- unix timestamp: 1708627278 
or 
- datetime of the following format: 2024-01-01 00:00:00

Eg:

With timestamp
$ oversight recover demo_db 1708627278

With formatted datetime
$ oversight recover demo_db "2024-01-01 00:00:00"

With specific storage engine
$ oversight recover demo_db "2024-01-01 00:00:00" dropbox
`,
	Args: func(cmd *cobra.Command, args []string) error {

		if err := cobra.MinimumNArgs(2)(cmd, args); err != nil {
			return err
		}

		if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
			return err
		}

		if len(args) == 1 {
			return nil
		}

		if err := cobra.OnlyValidArgs(cmd, args[2:]); err != nil {
			return err
		}

		return nil
	},
	ValidArgs: []string{storage.FileSystemType, storage.DropboxType},
	RunE: func(cmd *cobra.Command, args []string) error {

		databaseName := args[0]
		timeValue := args[1]
		storageEngine := storage.FileSystemType

		if len(args) == 3 {

			storageEngine = args[2]
		}

		var pointInTime time.Time

		timestamp, err := strconv.Atoi(timeValue)
		if err == nil {
			pointInTime = time.Unix(int64(timestamp), 0)
			goto recoverService
		}

		pointInTime, err = time.Parse(time.DateTime, timeValue)
		if err != nil {
			return errors.New("invalid point in time")
		}

	recoverService:
		service, err := services.InitRecoveryService(databaseName, storageEngine, pointInTime)
		if err != nil {
			return err
		}

		err = service.Recover()
		if err != nil {
			return err
		}

		return nil
	},
}
