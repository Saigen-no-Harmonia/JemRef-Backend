package testutil

import (
	"bytes"
	"context"
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

// cleanUpTargetTables CleanUpTablesの削除対象テーブル(参照元->参照先の順)
var cleanUpTargetTables = []string{
	"e_record_field_values",
	"e_record_memos",
	"e_record_urls",
	"e_record_contributions",
	"e_contributors",
	"e_records",
	"m_record_field_values",
	"m_record_contributions",
	"m_record_fields",
	"m_record_types",
	"m_roles",
	"m_contributors",
	"m_records",
	"m_policies",
	"m_users",
}

// CleanUpTables テスト実施前に全レコードをdeleteする
//
// e_recordsはCONTAINER_IDで自己参照するため、単一のdelete文であっても
// 行の削除順序によってはFK違反になりうる。そのためFK制約を一時解除して削除する。
// SET FOREIGN_KEY_CHECKSはセッション単位の設定なので、コネクションプール越しの
// db.Execではなく単一コネクション上で一連の処理を実行する
func CleanUpTables(
	t *testing.T,
	db *sql.DB,
) {
	t.Helper()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0")
	require.NoError(t, err)

	for _, table := range cleanUpTargetTables {
		_, err = conn.ExecContext(ctx, "delete from "+table)
		require.NoError(t, err, "delete from %s", table)
	}

	_, err = conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
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
