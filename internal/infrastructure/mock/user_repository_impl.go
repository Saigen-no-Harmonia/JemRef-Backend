package mock

import (
	"context"
	"jemref_go/internal/domain/user"
)

type UserRepositoryMock struct {
	PublicUserId   string
	FirebaseUserId string
	Email          string
}

func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{
		PublicUserId:   "sample_public_uid",
		FirebaseUserId: "sample_firebase_uid",
		Email:          "sample@example.com",
	}
}

func (m *UserRepositoryMock) Create(ctx context.Context, u *user.User) error {
	// mockなので保存はしない
	u.PublicUserId = m.PublicUserId
	u.FirebaseUserId = m.FirebaseUserId
	u.Email = m.Email
	return nil
}
