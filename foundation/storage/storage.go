package storage

type Engine interface {
	Save(receiver <-chan []byte) error
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
