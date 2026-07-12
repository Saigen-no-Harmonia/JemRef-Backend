package usecase

import (
	"context"
	"errors"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
)

type AuthUsecaseImpl struct {
	userRepo     repository.UserRepository
	firebaseRepo repository.FirebaseRepository
}

func NewAuthUsecaseImpl(ur repository.UserRepository, fr repository.FirebaseRepository) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{
		userRepo:     ur,
		firebaseRepo: fr,
	}
}

// Authenticate ユーザ情報を認証する。Ph0ではDBからユーザ情報を取得するだけ。
func (a *AuthUsecaseImpl) Authenticate(
	ctx context.Context,
	firebaseUid string,
) (*dto.AuthUserOutput, error) {

	// ユーザ情報を取得
	u, err := a.userRepo.SelectByFirebaseUid(ctx, firebaseUid)

	// ユーザ情報が取得できなかった場合
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 退会済みの場合
	if u.DeletedAt != nil {
		return nil, ErrUserDeleted
	}

	// ユーザ情報を返却
	return &dto.AuthUserOutput{
		InternalUserId: u.InternalUserId,
		PublicUserId:   u.PublicUserId,
		FirebaseUserId: u.FirebaseUserId,
	}, nil
}

// DeleteUser FirebaseUserを削除する
func (a *AuthUsecaseImpl) DeleteUser(
	ctx context.Context,
	firebaseUid string,
) error {
	err := a.firebaseRepo.DeleteUser(ctx, firebaseUid)

	if err != nil {
		return err
	}

	return nil
}
