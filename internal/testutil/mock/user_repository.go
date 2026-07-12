package mock

import (
	"context"
	"jemref_go/internal/domain/user"
)

type MockUserRepository struct {
	CreateFunc                func(ctx context.Context, u *user.User) error
	DeleteFunc                func(ctx context.Context, uid int64) error
	SelectByInternalUidFunc   func(ctx context.Context, uid int64) (*user.User, error)
	SelectByFirebaseUidFunc   func(ctx context.Context, firebaseUid string) (*user.User, error)
	LastInsertUser            *user.User
	LastDeleteUserId          int64
	LastSelectInternalUid     int64
	LastSelectFirebaseUid     string
	Called                    int
	CreateCalled              int
	DeleteCalled              int
	SelectByInternalUidCalled int
	SelectByFirebaseUidCalled int
}

func (m *MockUserRepository) Create(
	ctx context.Context,
	u *user.User,
) error {
	m.Called++
	m.CreateCalled++
	m.LastInsertUser = u
	return m.CreateFunc(ctx, u)
}

func (m *MockUserRepository) Delete(
	ctx context.Context,
	uid int64,
) error {
	m.Called++
	m.DeleteCalled++
	m.LastDeleteUserId = uid
	return m.DeleteFunc(ctx, uid)
}

func (m *MockUserRepository) SelectByInternalUid(
	ctx context.Context,
	uid int64,
) (*user.User, error) {
	m.Called++
	m.SelectByInternalUidCalled++
	m.LastSelectInternalUid = uid
	return m.SelectByInternalUidFunc(ctx, uid)
}

func (m *MockUserRepository) SelectByFirebaseUid(
	ctx context.Context,
	firebaseUid string,
) (*user.User, error) {
	m.Called++
	m.SelectByFirebaseUidCalled++
	m.LastSelectFirebaseUid = firebaseUid
	return m.SelectByFirebaseUidFunc(ctx, firebaseUid)
}
