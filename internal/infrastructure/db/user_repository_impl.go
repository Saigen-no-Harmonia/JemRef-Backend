package db

// ユーザ情報Repository実装

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"

	"github.com/go-sql-driver/mysql"
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

// SelectByInternalUid  内部用ユーザIDからユーザ情報を取得する
func (r *UserRepositoryImpl) SelectByInternalUid(
	ctx context.Context,
	internalUid int64,
) (*user.User, error) {

	row := r.db.QueryRowContext(
		ctx,
		selectUserByInternalUidQuery,
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
			return nil, fmt.Errorf("ユーザが存在しません。:%w", repository.ErrNotFound)
		}
		return nil, fmt.Errorf("select user by internal uid=%d :%w", internalUid, err)
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
			return nil, fmt.Errorf("ユーザが存在しません。:%w", repository.ErrNotFound)
		}

		return nil, fmt.Errorf("select user by firebase uid=%s :%w", firebaseUid, err)
	}

	return &u, nil
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
		var mysqlErr *mysql.MySQLError
		// キー重複エラー
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("登録ずみユーザです。:%w", repository.ErrDuplicateEntry)
		}
		return fmt.Errorf("create user firebase uid=%s, email=%s :%w", user.FirebaseUserId, user.Email, err)
	}

	return nil
}

// Delete ユーザ１件削除
func (r *UserRepositoryImpl) Delete(
	ctx context.Context,
	internalUid int64,
) error {
	result, err := r.db.ExecContext(
		ctx,
		deleteUserQuery,
		internalUid,
	)

	if err != nil {
		return fmt.Errorf("delete user by internal uid=%d :%w", internalUid, err)
	}

	// 更新対象0件はエラーとする（仕様上、意図しないエラーに限られるので）
	cnt, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user by internal uid=%d :%w", internalUid, err)
	}

	if cnt == 0 {
		return fmt.Errorf("削除対象ユーザが存在しません:%w", repository.ErrNotFound)
	}

	return nil
}

// UpdateUserAgreement ユーザ規約同意状況を、引数のユーザ情報で更新する
func (r *UserRepositoryImpl) UpdateUserAgreement(ctx context.Context, u *user.User) (int64, error) {
	result, err := r.db.ExecContext(
		ctx,
		updateUserAgreementQuery,
		u.TermsAgreedDt,
		u.TermsVersion,
		u.PrivacyPolicyAgreedDt,
		u.PrivacyPolicyVersion,
		u.InternalUserId,
	)
	if err != nil {
		return 0, fmt.Errorf("update user agreement by internal uid=%d :%w", u.InternalUserId, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("update user agreement by internal uid=%d :%w", u.InternalUserId, err)
	}

	return rows, nil
}
