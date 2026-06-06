package testutil

import (
	"database/sql"
	"os"
)

// loadFixture テストデータの読み込み
func LoadFixture(
	db *sql.DB,
	path string,
) error {

	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return err
	}

	return nil
}

// TruncateTable テスト実施前に対象テーブルをtruncateする
func TruncateTable(
	db *sql.DB,
	tableName string,
) error {

	_, err := db.Exec(
		"TRUNCATE TABLE " + tableName,
	)

	return err
}
