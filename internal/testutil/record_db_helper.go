package testutil

// 書誌情報(e_records系テーブル)のテスト用DB検証ヘルパ。
//
// m_users向けのヘルパ(db_helper.go)はドメインモデルを返すが、こちらは
// DBの生値をそのまま保持する専用の行構造体を返す。書誌情報はNULL許容列が多く、
// またTYPE_ID/ROLE_ID/READ_STATUSはドメインの値ではなくID体系で格納されるため、
// 「DBに何が入ったか」(NULLか空文字か、ID変換が正しいか)を検証するには
// ドメインモデルに詰め替えないほうが正確に確認できる。

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RecordRow e_recordsの1行
type RecordRow struct {
	Id              string
	UserId          int64
	TypeId          string
	MainTitle       string
	SubTitle        sql.NullString
	PageRange       sql.NullString
	Volume          sql.NullString
	Issue           sql.NullString
	Publisher       sql.NullString
	PublicationDate sql.NullTime
	DatePrecision   sql.NullInt64
	ReadStatus      sql.NullInt64
	ContainerId     sql.NullString
	MRecordId       sql.NullString
	DelFlg          int
	InsPg           string
	InsId           string
	InsDt           time.Time
	UpdPg           string
	UpdId           string
	UpdDt           time.Time
}

// ContributorRow e_contributorsの1行
type ContributorRow struct {
	Id             int64
	RawName        string
	UserId         int64
	MContributorId sql.NullString
	DelFlg         int
	InsPg          string
	InsId          string
	InsDt          time.Time
	UpdPg          string
	UpdId          string
	UpdDt          time.Time
}

// ContributionRow e_record_contributionsの1行
type ContributionRow struct {
	Id            int64
	RecordId      string
	ContributorId int64
	RoleId        string
	DisplayNumber int
	DelFlg        int
	InsPg         string
	InsId         string
	InsDt         time.Time
	UpdPg         string
	UpdId         string
	UpdDt         time.Time
}

// UrlRow e_record_urlsの1行
type UrlRow struct {
	Id            int64
	RecordId      string
	DisplayNumber int
	RawLabel      sql.NullString
	RawUrl        string
	UserId        int64
	DelFlg        int
	InsPg         string
	InsId         string
	InsDt         time.Time
	UpdPg         string
	UpdId         string
	UpdDt         time.Time
}

// MemoRow e_record_memosの1行
type MemoRow struct {
	Id            int64
	RecordId      string
	DisplayNumber int
	Title         sql.NullString
	Content       string
	UserId        int64
	ShareLevel    int
	DelFlg        int
	InsPg         string
	InsId         string
	InsDt         time.Time
	UpdPg         string
	UpdId         string
	UpdDt         time.Time
}

// FieldValueRow e_record_field_valuesの1行
type FieldValueRow struct {
	Id            int64
	RecordId      string
	RecordFieldId string
	DisplayNumber int
	RawValue      string
	UserId        int64
	DelFlg        int
	InsPg         string
	InsId         string
	InsDt         time.Time
	UpdPg         string
	UpdId         string
	UpdDt         time.Time
}

// SelectRecordByPk ユーザ入力書誌をPKで指定して取得する（テスト用）
func SelectRecordByPk(t *testing.T, db *sql.DB, recordId string) (*RecordRow, error) {
	t.Helper()

	var r RecordRow
	err := db.QueryRow(
		`
		SELECT
			ID,
			USER_ID,
			TYPE_ID,
			MAIN_TITLE,
			SUB_TITLE,
			PAGE_RANGE,
			VOLUME,
			ISSUE,
			PUBLISHER,
			PUBLICATION_DATE,
			DATE_PRECISION,
			READ_STATUS,
			CONTAINER_ID,
			M_RECORD_ID,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_records
		WHERE ID = ?
		`,
		recordId,
	).Scan(
		&r.Id,
		&r.UserId,
		&r.TypeId,
		&r.MainTitle,
		&r.SubTitle,
		&r.PageRange,
		&r.Volume,
		&r.Issue,
		&r.Publisher,
		&r.PublicationDate,
		&r.DatePrecision,
		&r.ReadStatus,
		&r.ContainerId,
		&r.MRecordId,
		&r.DelFlg,
		&r.InsPg,
		&r.InsId,
		&r.InsDt,
		&r.UpdPg,
		&r.UpdId,
		&r.UpdDt,
	)

	return &r, err
}

// SelectContributorsByUserId ユーザ入力貢献者をユーザIDで取得する（テスト用。ID昇順）
func SelectContributorsByUserId(t *testing.T, db *sql.DB, userId int64) []ContributorRow {
	t.Helper()

	rows, err := db.Query(
		`
		SELECT
			ID,
			RAW_NAME,
			USER_ID,
			M_CONTRIBUTOR_ID,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_contributors
		WHERE USER_ID = ?
		ORDER BY ID
		`,
		userId,
	)
	require.NoError(t, err)
	defer rows.Close()

	var contributors []ContributorRow
	for rows.Next() {
		var c ContributorRow
		err = rows.Scan(
			&c.Id,
			&c.RawName,
			&c.UserId,
			&c.MContributorId,
			&c.DelFlg,
			&c.InsPg,
			&c.InsId,
			&c.InsDt,
			&c.UpdPg,
			&c.UpdId,
			&c.UpdDt,
		)
		require.NoError(t, err)
		contributors = append(contributors, c)
	}
	require.NoError(t, rows.Err())

	return contributors
}

// SelectContributionsByRecordId ユーザ入力書誌貢献を書誌IDで取得する（テスト用。表示順昇順）
func SelectContributionsByRecordId(t *testing.T, db *sql.DB, recordId string) []ContributionRow {
	t.Helper()

	rows, err := db.Query(
		`
		SELECT
			ID,
			E_RECORD_ID,
			E_CONTRIBUTOR_ID,
			ROLE_ID,
			DISPLAY_NUMBER,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_record_contributions
		WHERE E_RECORD_ID = ?
		ORDER BY DISPLAY_NUMBER
		`,
		recordId,
	)
	require.NoError(t, err)
	defer rows.Close()

	var contributions []ContributionRow
	for rows.Next() {
		var c ContributionRow
		err = rows.Scan(
			&c.Id,
			&c.RecordId,
			&c.ContributorId,
			&c.RoleId,
			&c.DisplayNumber,
			&c.DelFlg,
			&c.InsPg,
			&c.InsId,
			&c.InsDt,
			&c.UpdPg,
			&c.UpdId,
			&c.UpdDt,
		)
		require.NoError(t, err)
		contributions = append(contributions, c)
	}
	require.NoError(t, rows.Err())

	return contributions
}

// SelectUrlsByRecordId ユーザ入力URLを書誌IDで取得する（テスト用。表示順昇順）
func SelectUrlsByRecordId(t *testing.T, db *sql.DB, recordId string) []UrlRow {
	t.Helper()

	rows, err := db.Query(
		`
		SELECT
			ID,
			E_RECORD_ID,
			DISPLAY_NUMBER,
			RAW_LABEL,
			RAW_URL,
			USER_ID,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_record_urls
		WHERE E_RECORD_ID = ?
		ORDER BY DISPLAY_NUMBER
		`,
		recordId,
	)
	require.NoError(t, err)
	defer rows.Close()

	var urls []UrlRow
	for rows.Next() {
		var u UrlRow
		err = rows.Scan(
			&u.Id,
			&u.RecordId,
			&u.DisplayNumber,
			&u.RawLabel,
			&u.RawUrl,
			&u.UserId,
			&u.DelFlg,
			&u.InsPg,
			&u.InsId,
			&u.InsDt,
			&u.UpdPg,
			&u.UpdId,
			&u.UpdDt,
		)
		require.NoError(t, err)
		urls = append(urls, u)
	}
	require.NoError(t, rows.Err())

	return urls
}

// SelectMemosByRecordId ユーザ入力メモを書誌IDで取得する（テスト用。表示順昇順）
func SelectMemosByRecordId(t *testing.T, db *sql.DB, recordId string) []MemoRow {
	t.Helper()

	rows, err := db.Query(
		`
		SELECT
			ID,
			E_RECORD_ID,
			DISPLAY_NUMBER,
			TITLE,
			CONTENT,
			USER_ID,
			SHARE_LEVEL,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_record_memos
		WHERE E_RECORD_ID = ?
		ORDER BY DISPLAY_NUMBER
		`,
		recordId,
	)
	require.NoError(t, err)
	defer rows.Close()

	var memos []MemoRow
	for rows.Next() {
		var m MemoRow
		err = rows.Scan(
			&m.Id,
			&m.RecordId,
			&m.DisplayNumber,
			&m.Title,
			&m.Content,
			&m.UserId,
			&m.ShareLevel,
			&m.DelFlg,
			&m.InsPg,
			&m.InsId,
			&m.InsDt,
			&m.UpdPg,
			&m.UpdId,
			&m.UpdDt,
		)
		require.NoError(t, err)
		memos = append(memos, m)
	}
	require.NoError(t, rows.Err())

	return memos
}

// SelectFieldValuesByRecordId ユーザ入力項目値を書誌IDで取得する（テスト用。表示順昇順）
func SelectFieldValuesByRecordId(t *testing.T, db *sql.DB, recordId string) []FieldValueRow {
	t.Helper()

	rows, err := db.Query(
		`
		SELECT
			ID,
			E_RECORD_ID,
			RECORD_FIELD_ID,
			DISPLAY_NUMBER,
			RAW_VALUE,
			USER_ID,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM e_record_field_values
		WHERE E_RECORD_ID = ?
		ORDER BY DISPLAY_NUMBER
		`,
		recordId,
	)
	require.NoError(t, err)
	defer rows.Close()

	var fieldValues []FieldValueRow
	for rows.Next() {
		var f FieldValueRow
		err = rows.Scan(
			&f.Id,
			&f.RecordId,
			&f.RecordFieldId,
			&f.DisplayNumber,
			&f.RawValue,
			&f.UserId,
			&f.DelFlg,
			&f.InsPg,
			&f.InsId,
			&f.InsDt,
			&f.UpdPg,
			&f.UpdId,
			&f.UpdDt,
		)
		require.NoError(t, err)
		fieldValues = append(fieldValues, f)
	}
	require.NoError(t, rows.Err())

	return fieldValues
}

// CountRecords ユーザ入力書誌の件数を取得する（テスト用）
func CountRecords(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_records")
}

// CountContributors ユーザ入力貢献者の件数を取得する（テスト用）
func CountContributors(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_contributors")
}

// CountContributions ユーザ入力書誌貢献の件数を取得する（テスト用）
func CountContributions(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_record_contributions")
}

// CountUrls ユーザ入力URLの件数を取得する（テスト用）
func CountUrls(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_record_urls")
}

// CountMemos ユーザ入力メモの件数を取得する（テスト用）
func CountMemos(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_record_memos")
}

// CountFieldValues ユーザ入力項目値の件数を取得する（テスト用）
func CountFieldValues(t *testing.T, db *sql.DB) int {
	t.Helper()
	return countTable(t, db, "e_record_field_values")
}

// countTable 指定テーブルの件数を取得する
func countTable(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var count int
	err := db.QueryRow("select COUNT(*) from " + table).Scan(&count)
	require.NoError(t, err)
	return count
}
