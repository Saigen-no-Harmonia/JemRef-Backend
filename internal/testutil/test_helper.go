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

// CleanUpTables テスト実施前に全レコードをdeleteする
func CleanUpTables(
	t *testing.T,
	db *sql.DB,
) {
	t.Helper()
	_, err := db.Exec("delete from e_record_field_values")
	require.NoError(t, err)
	_, err = db.Exec("delete from e_record_memos")
	require.NoError(t, err)
	_, err = db.Exec("delete from e_record_urls")
	require.NoError(t, err)
	_, err = db.Exec("delete from e_record_contributions")
	require.NoError(t, err)
	_, err = db.Exec("delete from e_records")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_record_field_values")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_record_contributions")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_record_fields")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_record_types")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_roles")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_contributors")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_records")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_policies")
	require.NoError(t, err)
	_, err = db.Exec("delete from m_users")
	require.NoError(t, err)
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
