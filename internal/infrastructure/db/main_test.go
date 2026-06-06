package db

// Repositoryテストのtest_main

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	tcMysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

var testDb *sql.DB

func TestMain(m *testing.M) {

	ctx := context.Background()

	container, err := tcMysql.Run(
		ctx,
		"mysql:8.0",
		tcMysql.WithDatabase("testdb"),
		tcMysql.WithUsername("test"),
		tcMysql.WithPassword("test"),
	)
	if err != nil {
		panic(err)
	}

	defer container.Terminate(ctx)

	dsn, err := container.ConnectionString(
		ctx,
		"parseTime=true",
		"multiStatements=true",
	)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	testDb = db

	if err = executeDdl(); err != nil {
		panic(err)
	}

	code := m.Run()

	db.Close()

	os.Exit(code)
}

// executeDdl テスト用DB作成のため、DDLを実行
func executeDdl() error {
	sqlBytes, err := os.ReadFile("../../../sql/ddl.sql")
	if err != nil {
		return err
	}

	_, err = testDb.Exec(string(sqlBytes))
	return err
}
