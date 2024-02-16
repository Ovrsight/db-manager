package storage

type Engine interface {
	Save(receiver <-chan []byte) error
}

type EngineType string

const (
	FileSystemType EngineType = "filesystem"
	DropboxType    EngineType = "dropbox"
)

func GetStorageEngine(engineType EngineType, fileName, database string) Engine {
	switch engineType {
	case FileSystemType:
		return &FileSystem{
			Filename: fileName,
			Database: database,
		}
	case DropboxType:
		return &Dropbox{
			Filename: fileName,
			Database: database,
		}
	default:
		return &FileSystem{
			Filename: fileName,
			Database: database,
		}
	}
}
