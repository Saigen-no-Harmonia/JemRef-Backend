package mock

import "context"

type MockFirebaseRepository struct {
	DeleteUserFunc     func(ctx context.Context, uid string) error
	DeletedFirebaseUid string
	Called             int
	DeleteUserCalled   int
}

func (m *MockFirebaseRepository) DeleteUser(
	ctx context.Context,
	uid string,
) error {
	m.Called++
	m.DeleteUserCalled++
	m.DeletedFirebaseUid = uid
	return m.DeleteUserFunc(ctx, uid)
}
