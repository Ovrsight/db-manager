package databases

type BackupMethod interface {
	Generate() ([]byte, error)
}

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}
