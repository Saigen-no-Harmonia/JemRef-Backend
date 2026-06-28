package db

// ユーザ情報Repository実装

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"

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

// Create ユーザ作成
func (r *UserRepositoryImpl) Create(
	ctx context.Context,
	user *user.User,
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

// Delete ユーザ削除
func (r *UserRepositoryImpl) Delete(
	ctx context.Context,
	uid uint64,
) error {
	_, err := r.db.ExecContext(
		ctx,
		deleteUserQuery,
		uid,
	)

	if err != nil {
		return err
	}

	return nil
}

// SelectByInternalUid  ユーザIDからユーザ情報を取得する
func (r *UserRepositoryImpl) SelectByInternalUid(
	ctx context.Context,
	internalUid uint64,
) (*user.User, error) {

	row := r.db.QueryRowContext(
		ctx,
		selectUserByUidQuery,
		internalUid,
	)

	var u user.User
	err := row.Scan(
		&u.InternalUserId,
		&u.PublicUserId,
		&u.FirebaseUserId,
		&u.Email,
		&u.TermsAgreedDt,
		&u.TermsVersion,
		&u.PrivacyPolicyAgreedDt,
		&u.PrivacyPolicyVersion,
		&u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, err
	}

	return &u, nil
}

// SelectByFirebaseUid  FirebaseユーザIDからユーザ情報を取得する
func (r *UserRepositoryImpl) SelectByFirebaseUid(
	ctx context.Context,
	firebaseUid string,
) (*user.User, error) {

	row := r.db.QueryRowContext(
		ctx,
		selectUserByUFirebaseUidQuery,
		firebaseUid,
	)

	var u user.User
	err := row.Scan(
		&u.InternalUserId,
		&u.PublicUserId,
		&u.FirebaseUserId,
		&u.Email,
		&u.TermsAgreedDt,
		&u.TermsVersion,
		&u.PrivacyPolicyAgreedDt,
		&u.PrivacyPolicyVersion,
		&u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, err
	}

	return &u, nil
}
