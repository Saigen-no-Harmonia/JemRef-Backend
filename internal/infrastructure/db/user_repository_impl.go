package db

// ユーザ情報Repository実装

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"log"

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
			log.Print("SelectByInternalUid: 件数なし")
			return nil, repository.ErrNotFound
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
			return nil, repository.ErrNotFound
		}

		return nil, err
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
			log.Print("Create:ユーザ登録時に重複が生じました。")
			return repository.ErrDuplicateEntry
		}
		log.Print("Create:ユーザ登録時にDBエラーが生じました。")
		return err
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
		return err
	}

	// 更新対象0件はエラーとする（仕様上、意図しないエラーに限られるので）
	cnt, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return repository.ErrNotFound
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
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, err
}
