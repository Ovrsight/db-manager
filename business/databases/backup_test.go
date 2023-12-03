package databases_test

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/business/databases"
	backupDB "github.com/nizigama/ovrsight/foundation/databases"
	"github.com/nizigama/ovrsight/foundation/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {

	var backer *databases.Backer
	var filesystemDvr *storage.FileSystemDriverMock
	var dpbxDvr *storage.DropboxDriverMock
	var gglDvr *storage.GoogleDriveDriverMock
	var mysqlDumper *backupDB.MysqlDumperMock

	BeforeEach(func() {

		filesystemDvr = &storage.FileSystemDriverMock{}
		dpbxDvr = &storage.DropboxDriverMock{}
		gglDvr = &storage.GoogleDriveDriverMock{}
		mysqlDumper = &backupDB.MysqlDumperMock{}
	})

	Context("with filesystem driver", func() {
		It("can backup a database", func() {

			filesystemDvr.On("Upload", []byte{}).Return(nil)
			mysqlDumper.On("Generate").Return([]byte{}, nil)

			backer = &databases.Backer{
				StorageDriver: filesystemDvr,
				BackupMethod:  mysqlDumper,
				DatabaseName:  "test_db",
				Filename:      "file.sql",
			}

			err := backer.Backup()

			Expect(err).To(BeNil())
		})
	})

	Context("with dropbox driver", func() {
		It("can backup a database", func() {

			dpbxDvr.On("Upload", []byte{}).Return(nil)
			mysqlDumper.On("Generate").Return([]byte{}, nil)

			backer = &databases.Backer{
				StorageDriver: dpbxDvr,
				BackupMethod:  mysqlDumper,
				DatabaseName:  "test_db",
				Filename:      "file.sql",
			}

			err := backer.Backup()

			Expect(err).To(BeNil())
		})
	})

	Context("with google drive driver", func() {
		It("can backup a database", func() {

			gglDvr.On("Upload", []byte{}).Return(nil)
			mysqlDumper.On("Generate").Return([]byte{}, nil)

			backer = &databases.Backer{
				StorageDriver: gglDvr,
				BackupMethod:  mysqlDumper,
				DatabaseName:  "test_db",
				Filename:      "file.sql",
			}

			err := backer.Backup()

			Expect(err).To(BeNil())
		})
	})
})
