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
)

type UserUsecase struct {
	userRepo     repository.UserRepository
	firebaseRepo repository.FirebaseRepository
	idGen        id.Generator
	generalRepo  repository.GeneralRepository
	// txManager repository.TxManager
}

func NewUserUsecase(ur repository.UserRepository, gr repository.GeneralRepository, fr repository.FirebaseRepository, g id.Generator) *UserUsecase {
	return &UserUsecase{
		userRepo:     ur,
		generalRepo:  gr,
		firebaseRepo: fr,
		idGen:        g,
	}
}

// ユーザ情報登録Usecase
// CreateUser ユーザを作成し、公開用UIDを返却する
func (u *UserUsecase) CreateUser(ctx context.Context, cu dto.CreateUserInput) (*dto.CreateUserOutput, error) {
	log.Printf("ユーザ情報登録usecase 処理開始: firebase uid: %s", cu.FirebaseUserId)

	// 公開用ユーザIDを生成
	publicUserId := u.idGen.Generate()

	// 規約関係
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
		log.Println("ユーザ規約が取得できませんでした")
		return nil, ErrPolicyNotFound
	}

	termsAgreedDt := time.Now()
	termsVersion := t.Version
	privacyPolicyAgreedDt := time.Now()
	privacyPolicyVersion := p.Version

	// 引数をマッピング
	user := &user.User{
		PublicUserId:          publicUserId,
		FirebaseUserId:        cu.FirebaseUserId,
		Email:                 cu.Email,
		TermsAgreedDt:         &termsAgreedDt,
		TermsVersion:          termsVersion,
		PrivacyPolicyAgreedDt: &privacyPolicyAgreedDt,
		PrivacyPolicyVersion:  privacyPolicyVersion,
	}

	// DBに登録
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// レスポンス
	log.Printf("ユーザ情報登録usecase 処理完了: firebase uid: %s", cu.FirebaseUserId)
	return &dto.CreateUserOutput{
		PublicUserId: publicUserId,
	}, nil

}

// DeleteUser DBのユーザと Firebaseユーザを削除する
func (u *UserUsecase) DeleteUser(ctx context.Context, ud dto.DeleteUserInput) error {

	internalUid := ud.InternalUserid
	firebaseUid := ud.FirebaseUserId
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

	// ユーザデータ不整合があった場合
	if user.FirebaseUserId != ud.FirebaseUserId {
		log.Printf("ユーザデータに不整合があります。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		return ErrInvalidUser
	}

	// すでに論理削除済みなら何もしない
	if user.DeletedAt == nil {
		if err := u.userRepo.Delete(ctx, internalUid); err != nil {
			log.Printf("退会済みのため、FIrebaseユーザを削除します。internalUid: %d", internalUid)
			return err
		}
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
func (uu *UserUsecase) Login(ctx context.Context, ui dto.UserLoginInput) (*dto.UserLoginOutput, error) {
	uid := ui.InternalUserId
	log.Printf("ユーザログインusecase 処理開始: internal uid: %d", uid)

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	log.Printf("ユーザログインusecase 処理完了: internal uid: %d", uid)
	return &dto.UserLoginOutput{PublicUserId: u.PublicUserId}, err
}
