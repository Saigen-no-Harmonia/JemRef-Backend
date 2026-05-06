package mock

import "jemref_go/internal/domain"

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

func (m *UserRepositoryMock) Create(u *domain.User) error {
	// mockなので保存はしない
	u.PublicUserId = m.PublicUserId
	u.FirebaseUserId = m.FirebaseUserId
	u.Email = m.Email
	return nil
}
