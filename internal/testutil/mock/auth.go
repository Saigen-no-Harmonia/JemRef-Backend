package mock

import (
	"context"
	usecaseDto "jemref_go/internal/usecase/dto"
)

type MockAuthUsecase struct {
	AuthenticateFunc      func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error)
	DeleteUserFunc        func(c context.Context, firebaseUid string) error
	LastAuthFirebaseUid   string
	LastDeleteFirebaseUid string
	AuthenticateCalled    int
	DeleteUserCalled      int
}

func (m *MockAuthUsecase) Authenticate(
	ctx context.Context,
	firebaseUid string,
) (*usecaseDto.AuthUserOutput, error) {
	m.AuthenticateCalled++
	m.LastAuthFirebaseUid = firebaseUid
	return m.AuthenticateFunc(ctx, firebaseUid)
}

func (m *MockAuthUsecase) DeleteUser(
	ctx context.Context,
	firebaseUid string,
) error {
	m.DeleteUserCalled++
	m.LastDeleteFirebaseUid = firebaseUid
	return m.DeleteUserFunc(ctx, firebaseUid)
}
