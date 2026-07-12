package usecase

// ユーザUsecase実装

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

type UserUsecaseImpl struct {
	userRepo     repository.UserRepository
	firebaseRepo repository.FirebaseRepository
	idGen        id.Generator
	generalRepo  repository.GeneralRepository
	// txManager repository.TxManager
}

func NewUserUsecase(ur repository.UserRepository, gr repository.GeneralRepository, fr repository.FirebaseRepository, g id.Generator) UserUsecase {
	return &UserUsecaseImpl{
		userRepo:     ur,
		generalRepo:  gr,
		firebaseRepo: fr,
		idGen:        g,
	}
}

// Create ユーザを作成し、公開用UIDを返却する
func (u *UserUsecaseImpl) Create(ctx context.Context, cu dto.CreateUserInput) (*dto.CreateUserOutput, error) {
	log.Printf("ユーザ情報登録usecase 処理開始: firebase uid: %s", cu.FirebaseUserId)

	publicUserId := u.idGen.Generate()

	// 規約マスタチェック
	t, err := u.generalRepo.SelectLatestById(
		ctx,
		policy.PolicyIdTermsOfService,
	)
	if err != nil {
		log.Println("ユーザ規約が取得できませんでした")
		return nil, ErrPolicyNotFound
	}

	p, err := u.generalRepo.SelectLatestById(
		ctx,
		policy.PolicyIdPrivacyPolicy,
	)
	if err != nil {
		log.Println("プライバシーポリシーが取得できませんでした")
		return nil, ErrPolicyNotFound
	}

	sysDate := time.Now()
	termsVersion := t.Version
	privacyPolicyVersion := p.Version

	user := &user.User{
		PublicUserId:          publicUserId,
		FirebaseUserId:        cu.FirebaseUserId,
		Email:                 cu.Email,
		TermsAgreedDt:         &sysDate,
		TermsVersion:          termsVersion,
		PrivacyPolicyAgreedDt: &sysDate,
		PrivacyPolicyVersion:  privacyPolicyVersion,
	}

	err = u.userRepo.Create(ctx, user)

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return nil, ErrUserAlreadyExists
	}

	if err != nil {
		return nil, err
	}

	log.Printf("ユーザ情報登録usecase 処理完了: firebase uid: %s", cu.FirebaseUserId)
	return &dto.CreateUserOutput{
		PublicUserId: publicUserId,
	}, nil

}

// Delete DBのユーザと Firebaseユーザを削除する
func (u *UserUsecaseImpl) Delete(ctx context.Context, du dto.DeleteUserInput) error {

	internalUid := du.InternalUserid
	firebaseUid := du.FirebaseUserId
	log.Printf("ユーザ削除usecase 処理開始: internal uid: %d", internalUid)

	user, err := u.userRepo.SelectByInternalUid(ctx, internalUid)
	// ユーザデータが存在しない場合
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("削除対象ユーザが存在しません。internal uid: %d", internalUid)
			return ErrUserNotFound
		}
		// その他のDBエラー
		log.Printf("database error: %s", err)
		return err
	}

	// マスタ整合性チェック
	if user.FirebaseUserId != du.FirebaseUserId {
		log.Printf("ユーザデータに不整合があります。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		return ErrInvalidUser
	}

	if user.DeletedAt == nil {
		if err := u.userRepo.Delete(ctx, internalUid); err != nil {
			log.Print(err)
			log.Printf("DBのユーザデータ削除に失敗しました。internal uid: %d", internalUid)
			return ErrUserDeleteFailed
		}
		// すでに論理削除済みならDB更新のみスキップ
	} else {
		log.Printf("すでに退会処理済みのため、FirebaseUser削除のみ実施します。firebase uid: %s", firebaseUid)
	}

	// firebaseユーザ削除
	if err = u.firebaseRepo.DeleteUser(ctx, firebaseUid); err != nil {
		log.Printf("firebaseユーザの削除に失敗しました。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		// 失敗時は、共通認証で自動リカバリ（FirebaseUser削除）する設計のため、ログ出力のみで終了
		return nil
	}

	log.Printf("ユーザ削除usecase 処理完了: internal uid: %d", internalUid)
	return nil
}

// Login ユーザログイン処理 Ph0では公開用UIDを返却するだけ
func (uu *UserUsecaseImpl) Login(ctx context.Context, ui dto.UserLoginInput) (*dto.UserLoginOutput, error) {
	uid := ui.InternalUserId
	log.Printf("ユーザログインusecase 処理開始: internal uid: %d", uid)

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	log.Printf("ユーザログインusecase 処理完了: internal uid: %d", uid)
	return &dto.UserLoginOutput{PublicUserId: u.PublicUserId}, err
}
