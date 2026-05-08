package usecase

import (
	"context"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"time"
)

type UserUsecase struct {
	userRepo repository.UserRepository
	idGen    id.Generator
	// termsRepo repository.TermsRepository
	// txManager repository.TxManager
}

func NewUserUsecase(r repository.UserRepository, g id.Generator) *UserUsecase {
	return &UserUsecase{
		userRepo: r,
		idGen:    g,
	}
}

func (u *UserUsecase) CreateUser(ctx context.Context, cu CreateUserInput) (*CreateUserOutput, error) {

	// 公開用ユーザIDを生成
	publicUserId := u.idGen.Generate()

	// 規約関係は仮の値
	termsAgreedDt := time.Now()
	termsVersion := "1.0"
	privacyPolicyAgreedDt := time.Now()
	privacyPolicyVersion := "1.0"

	// 引数をマッピング
	user := &user.User{
		PublicUserId:          publicUserId,
		FirebaseUserId:        cu.FirebaseUserId,
		Email:                 cu.Email,
		TermsAgreedDt:         termsAgreedDt,
		TermsVersion:          termsVersion,
		PrivacyPolicyAgreedDt: privacyPolicyAgreedDt,
		PrivacyPolicyVersion:  privacyPolicyVersion,
	}

	// DBに登録
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// レスポンス
	return &CreateUserOutput{
		PublicUserId: publicUserId,
	}, nil
}
