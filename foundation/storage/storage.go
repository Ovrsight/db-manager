package storage

type Engine interface {
	Save(receiver <-chan []byte) error
}

type EngineType string

const (
	FileSystemType  EngineType = "filesystem"
	DropboxType     EngineType = "dropbox"
	GoogleDriveType EngineType = "google_drive"
)

func GetStorageEngine(engineType EngineType, fileName string) Engine {
	switch engineType {
	case FileSystemType:
		return &FileSystem{
			Filename: fileName,
		}
	case GoogleDriveType:
		return &GoogleDrive{
			Filename: fileName,
		}
	case DropboxType:
		return &Dropbox{
			Filename: fileName,
		}
	default:
		return &FileSystem{
			Filename: fileName,
		}
	}
}
