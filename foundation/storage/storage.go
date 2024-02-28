package storage

type Engine interface {
	Save(receiver <-chan []byte, failureChan chan struct{}) (int, error)
	Retrieve(fileName string) (location string, err error)
}

const (
	FileSystemType string = "filesystem"
	DropboxType    string = "dropbox"
)

func GetStorageEngine(engineType, database, filename string) Engine {

	switch engineType {
	case FileSystemType:
		return &FileSystem{
			Database: database,
			Filename: filename,
		}
	case DropboxType:
		return &Dropbox{
			Database: database,
			Filename: filename,
		}
	default:
		return &FileSystem{
			Database: database,
			Filename: filename,
		}
	}
}
