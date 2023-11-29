package databases

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

type dbConfig struct {
	host     string
	port     int
	user     string
	password string
	database string
}

var (
	config dbConfig
)

func Backup(storageDriver string) error {

	if err := ping(); err != nil {
		return err
	}

	bckpData, err := dumpData()
	if err != nil {
		return err
	}

	switch storageDriver {
	case "filesystem":
		file, err := os.Create("backup.sql")
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = file.Write(bckpData)
		if err != nil {
			return err
		}
	case "dropbox":
		dropboxBuf := bytes.Buffer{}
		_, err = dropboxBuf.Write(bckpData)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", &dropboxBuf)
		if err != nil {
			return err
		}

		token := os.Getenv("DROPBOX_ACCESS_TOKEN")
		params := map[string]interface{}{
			"autorename":      false,
			"mode":            "add",
			"mute":            false,
			"path":            "/database/backups/oversight.backup.sql",
			"strict_conflict": false,
		}
		parameters, err := json.Marshal(params)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Dropbox-API-Arg", string(parameters))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 200 {
			data, _ := io.ReadAll(res.Body)
			return errors.New(string(data))
		}
	case "googledrive":
		jsonData := os.Getenv("GOOGLE_DRIVE_API_CREDENTIALS_JSON")
		jsonCreds := []byte(jsonData)

		driveService, err := drive.NewService(context.Background(), option.WithCredentialsJSON(jsonCreds))
		if err != nil {
			return err
		}

		driveFile := &drive.File{
			Name:    "backup.sql",
			Parents: []string{"1auxyWbwVpzVZ_XJdi9ySsl8qD4MS95pO"},
		}

		googleDriveBuf := bytes.Buffer{}
		_, err = googleDriveBuf.Write(bckpData)
		if err != nil {
			return err
		}

		f, err := driveService.Files.Create(driveFile).Media(&googleDriveBuf).Do()
		if err != nil {
			return err
		}

		if f.ServerResponse.HTTPStatusCode != 200 {
			return errors.New("failed uploading backup to google drive folder")
		}
	default:
		return errors.New("invalid storage driver")
	}

	return nil
}

func configure() (dbConfig, string) {
	host := os.Getenv("DB_HOST")
	p := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")

	port, _ := strconv.Atoi(p)

	config = dbConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
	}

	return config, fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.user, config.password, config.host, config.port, config.database)
}

func ping() error {

	_, dsn := configure()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	return db.Ping()
}

func dumpData() ([]byte, error) {

	program, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatal(err)
	}

	cfg, _ := configure()

	cmd := exec.Command(fmt.Sprintf("%s", program), fmt.Sprintf("-u%s", cfg.user), fmt.Sprintf("-p%s", cfg.password), cfg.database)

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return out, nil
}
