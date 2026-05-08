package db

// ユーザ情報Repository実装

import (
	"context"
	"database/sql"
	domain "jemref_go/internal/domain/user"

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

	_, err := r.db.ExecContext(
		ctx,
		createUserQuery,
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
