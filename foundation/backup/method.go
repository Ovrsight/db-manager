package backup

import "log"

type Method interface {
	Initialize() error                   // do some setup or checks
	Generate(sender chan<- []byte) error // get the backup data
	Clean(sender chan<- []byte) error    // do some cleaning after backup
}

const bufferSize int64 = 5000000

func GetBackupMethod(methodName, database string) Method {

	switch methodName {
	case "mysqldump":
		return &MysqlDump{
			Database: database,
		}
	default:
		log.Fatalln("method not yet supported")
	}

	return nil
}
