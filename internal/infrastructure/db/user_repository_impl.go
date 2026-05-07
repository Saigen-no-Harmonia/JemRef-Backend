package db

// ユーザ情報Repository実装

import (
	"context"
	"database/sql"
	"jemref_go/internal/domain"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepositoryImpl(db *sql.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

// ユーザ作成
func (r *UserRepositoryImpl) Create(
	ctx context.Context,
	user *domain.User,
) error {

	query := `
		insert into m_users (
			PUBLIC_ID,
			FIREBASE_ID,
			EMAIL,
			TERMS_AGREED_DT,
			TERMS_VERSION,
			PRIVACY_POLICY_AGREED_DT,
			PRIVACY_POLICY_VERSION,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		)
		values (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			'system',
			'MEM-API-002',
			'system',
			'MEM-API-002'
		)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.PublicUserId,
		user.FirebaseUserId,
		user.Email,
		user.TermsAgreedDt,
		user.TermsVersion,
		user.PrivacyPolicyAgreedDt,
		user.PrivacyPolicyVersion,
	)

	if err != nil {
		return err
	}

	return nil
}
