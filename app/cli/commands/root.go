/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"bytes"
	"context"
	"fmt"
	"github.com/nizigama/ovrsight/business/databases"
	"google.golang.org/api/drive/v3"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cobra",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		if err := databases.Ping(); err != nil {
			log.Fatalln(err)
		}

		backup, err := databases.Execute("oversight")
		if err != nil {
			log.Fatalln(err)
		}

		// filesystem,
		//file, err := os.Create("backup.sql")
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//defer file.Close()
		//_, err = file.Write(backup)
		//if err != nil {
		//	log.Fatalln(err)
		//}

		// dropbox,
		//buf := bytes.Buffer{}
		//_, err = buf.Write(backup)
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//
		//req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", &buf)
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//
		//token := os.Getenv("DROPBOX_ACCESS_TOKEN")
		//params := map[string]interface{}{
		//	"autorename":      false,
		//	"mode":            "add",
		//	"mute":            false,
		//	"path":            "/database/backups/oversight.backup.sql",
		//	"strict_conflict": false,
		//}
		//parameters, err := json.Marshal(params)
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//
		//req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		//req.Header.Set("Content-Type", "application/octet-stream")
		//req.Header.Set("Dropbox-API-Arg", string(parameters))
		//
		//res, err := http.DefaultClient.Do(req)
		//if err != nil {
		//	log.Fatalln(err)
		//}
		//
		//if res.StatusCode != 200 {
		//	data, _ := io.ReadAll(res.Body)
		//	log.Fatalln(err, string(data))
		//}

		// google drive
		jsonData := os.Getenv("GOOGLE_DRIVE_API_CREDENTIALS_JSON")
		jsonCreds := []byte(jsonData)

		driveService, err := drive.NewService(context.Background(), option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			log.Fatalln(err)
		}

		driveFile := &drive.File{
			Name:    "backup.sql",
			Parents: []string{"1auxyWbwVpzVZ_XJdi9ySsl8qD4MS95pO"},
		}

		buf := bytes.Buffer{}
		_, err = buf.Write(backup)
		if err != nil {
			log.Fatalln(err)
		}
		f, err := driveService.Files.Create(driveFile).Media(&buf).Do()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Backup completed successfully", f.ServerResponse.HTTPStatusCode)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
