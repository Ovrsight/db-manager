package databases_test

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nizigama/ovrsight/business/databases"
	backupDB "github.com/nizigama/ovrsight/foundation/databases"
	"github.com/nizigama/ovrsight/foundation/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/qustavo/dotsql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
)

var _ = Describe("Backup", func() {

	var backer *databases.Backer
	var filesystemDvr *storage.FileSystemDriverMock
	var dpbxDvr *storage.DropboxDriverMock
	var gglDvr *storage.GoogleDriveDriverMock
	var mysqlDumper *backupDB.MysqlDumperMock
	var mysqlContainer *mysql.MySQLContainer
	var err error
	var ctx context.Context

	BeforeEach(func() {

		filesystemDvr = &storage.FileSystemDriverMock{}
		dpbxDvr = &storage.DropboxDriverMock{}
		gglDvr = &storage.GoogleDriveDriverMock{}
		mysqlDumper = &backupDB.MysqlDumperMock{}

		ctx = context.Background()

		mysqlContainer, err = mysql.RunContainer(ctx,
			testcontainers.WithImage("mysql:latest"),
			mysql.WithDatabase("test_db"),
			mysql.WithUsername("user"),
			mysql.WithPassword("password"),
			testcontainers.CustomizeRequest(
				testcontainers.GenericContainerRequest{
					ContainerRequest: testcontainers.ContainerRequest{
						ExposedPorts: []string{"3306/tcp"},
						WaitingFor:   wait.ForLog("socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server - GPL"),
					},
				},
			),
		)

		if err != nil {
			panic(err)
		}

		p, _ := mysqlContainer.MappedPort(ctx, "3306/tcp")
		port := p.Int()

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", "user", "password", "127.0.0.1", port, "test_db")

		dbConn, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}

		dot, err := dotsql.LoadFromFile("../../scripts/queries.sql")
		if err != nil {
			panic(err)
		}

		_, err = dot.WithData(map[string]string{
			"dbname": "test_db",
		}).Exec(dbConn, "create-database")
		if err != nil {
			panic(err)
		}

		os.Setenv("DB_HOST", "127.0.0.1")
		os.Setenv("DB_PORT", fmt.Sprintf("%d", port))
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
	})

	AfterEach(func() {

		if err := mysqlContainer.Terminate(ctx); err != nil {
			panic(err)
		}
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

			err = backer.Backup()

			Expect(err).To(BeNil())
		})
	})

	Context("with dropbox driver", func() {
		It("can backup a database", func() {

			dpbxDvr.On("Upload", []byte{}).Return(nil)

			//backer = &databases.Backer{
			//	StorageDriver:       dpbxDvr,
			//	DatabaseName: "test_db",
			//	Filename:     "file.sql",
			//}

			//err := backer.Backup()
			//
			//Expect(err).To(BeNil())
		})
	})

	Context("with google drive driver", func() {
		It("can backup a database", func() {

			gglDvr.On("Upload", []byte{}).Return(nil)

			//backer = &databases.Backer{
			//	StorageDriver:       gglDvr,
			//	DatabaseName: "test_db",
			//	Filename:     "file.sql",
			//}

			//err := backer.Backup()
			//
			//Expect(err).To(BeNil())
		})
	})
})
