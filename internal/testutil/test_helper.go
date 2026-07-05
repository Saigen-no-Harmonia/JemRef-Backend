package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// NewJsonReader handlerテスト用JSONReader作成メソッド
func NewJsonReader(
	t *testing.T,
	body any,
) *bytes.Reader {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	return bytes.NewReader(jsonBody)
}

// AssertResponse HandlerテストのレスポンスJSONを検証する
func AssertResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, expectBody T) {
	t.Helper()

	var actual T
	err := json.Unmarshal(
		rec.Body.Bytes(),
		&actual,
	)
	assert.NoError(t, err)
	assert.Equal(t, expectBody, actual)
}
