package usecase

import (
	"jemref_go/internal/domain"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/interface/repository"
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

func (u *UserUsecase) CreateUser(cu CreateUserInput) (*CreateUserOutput, error) {

	// 公開用ユーザIDを生成
	publicUid := u.idGen.Generate()

	// 引数をマッピング
	user := &domain.User{
		PublicUserId:   publicUid,
		FirebaseUserId: cu.FirebaseUserId,
		Email:          cu.Email,
	}

	// DBに登録
	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	// レスポンス
	return &CreateUserOutput{
		PublicUserId: publicUid,
	}, nil
}
