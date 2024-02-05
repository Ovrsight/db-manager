package methods

type BackupMethod interface {
	Initialize() error         // do some setup or checks
	Generate() ([]byte, error) // get the backup data
	Clean() error              // do some cleaning after backup
}
