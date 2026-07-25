package testutil

import (
	"database/sql"
	"jemref_go/internal/domain/user"
	"testing"

	"github.com/stretchr/testify/require"
)

// SelectUserByPk ユーザマスタのレコードをPKで指定して取得する（テスト用）
func SelectUserByPk(t *testing.T, db *sql.DB, uid int64) (*user.User, error) {
	t.Helper()
	var actual user.User
	err := db.QueryRow(
		`
		SELECT
			ID,
			FIREBASE_ID,
			PUBLIC_ID,
			EMAIL,
			TERMS_AGREED_DT,
			TERMS_VERSION,
			PRIVACY_POLICY_AGREED_DT,
			PRIVACY_POLICY_VERSION,
			DELETED_AT,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM m_users
		WHERE ID = ?
		`,
		uid,
	).Scan(
		&actual.InternalUserId,
		&actual.FirebaseUserId,
		&actual.PublicUserId,
		&actual.Email,
		&actual.TermsAgreedDt,
		&actual.TermsVersion,
		&actual.PrivacyPolicyAgreedDt,
		&actual.PrivacyPolicyVersion,
		&actual.DeletedAt,
		&actual.DelFlg,
		&actual.InsPg,
		&actual.InsId,
		&actual.InsDt,
		&actual.UpdPg,
		&actual.UpdId,
		&actual.UpdDt,
	)

	return &actual, err
}

// SelectUserByPk ユーザマスタのレコードをPKで指定して取得する（テスト用）
func SelectUserByFirebaseUid(t *testing.T, db *sql.DB, firebaseUid string) (*user.User, error) {
	t.Helper()
	var actual user.User
	err := db.QueryRow(
		`
		SELECT
			ID,
			FIREBASE_ID,
			PUBLIC_ID,
			EMAIL,
			TERMS_AGREED_DT,
			TERMS_VERSION,
			PRIVACY_POLICY_AGREED_DT,
			PRIVACY_POLICY_VERSION,
			DELETED_AT,
			DEL_FLG,
			INS_PG,
			INS_ID,
			INS_DT,
			UPD_PG,
			UPD_ID,
			UPD_DT
		FROM m_users
		WHERE FIREBASE_ID = ?
		`,
		firebaseUid,
	).Scan(
		&actual.InternalUserId,
		&actual.FirebaseUserId,
		&actual.PublicUserId,
		&actual.Email,
		&actual.TermsAgreedDt,
		&actual.TermsVersion,
		&actual.PrivacyPolicyAgreedDt,
		&actual.PrivacyPolicyVersion,
		&actual.DeletedAt,
		&actual.DelFlg,
		&actual.InsPg,
		&actual.InsId,
		&actual.InsDt,
		&actual.UpdPg,
		&actual.UpdId,
		&actual.UpdDt,
	)

	return &actual, err
}

// CountUser ユーザマスタのレコード数を返却する（テスト用）
func CountUsers(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	err := db.QueryRow("select COUNT(*) from m_users").Scan(&count)
	require.NoError(t, err)
	return count
}
