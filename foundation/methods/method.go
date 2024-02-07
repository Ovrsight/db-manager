package methods

type BackupMethod interface {
	Initialize() error                   // do some setup or checks
	Generate(sender chan<- []byte) error // get the backup data
	Clean(sender chan<- []byte) error    // do some cleaning after backup
}
