package databases_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/qustavo/dotsql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatabases(t *testing.T) {
	RegisterFailHandler(Fail)

	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
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

	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	RunSpecs(t, "Databases Suite")
}
