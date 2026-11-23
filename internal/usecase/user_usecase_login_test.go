package usecase

import (
	"context"
	"testing"

	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------Login(context.Context, dto.UserLoginInput)のテスト-------------

// runUserLogin Loginを実行し、ユーザの取得だけが1回行われたことを検証する
func runUserLogin(t *testing.T, userRepo *mock.MockUserRepository) (*dto.UserLoginOutput, error) {
	t.Helper()

	u := NewUserUsecaseImpl(userRepo, nil, nil, nil)
	output, err := u.Login(context.Background(), dto.UserLoginInput{InternalUserId: userTestUid})

	assert.Equal(t, 1, userRepo.SelectByInternalUidCalled)
	assert.Equal(t, 0, userRepo.DeleteCalled)
	assert.Equal(t, 0, userRepo.CreateCalled)
	assert.Equal(t, 0, userRepo.SelectByFirebaseUidCalled)

	return output, err
}

func TestUserUsecaseImpl_Login_正常(t *testing.T) {
	t.Run("正常_ログイン", func(t *testing.T) {
		userRepo := newSelectUserRepoMock(&user.User{PublicUserId: "_public_uid_"}, nil)

		output, err := runUserLogin(t, userRepo)

		require.NoError(t, err)
		assert.Equal(t, userTestUid, userRepo.LastSelectInternalUid)
		assert.Equal(t, "_public_uid_", output.PublicUserId)
	})
}

func TestUserUsecaseImpl_Login_異常(t *testing.T) {

	tests := []struct {
		name              string
		selectErr         error
		expectedErrorBody error
	}{
		{
			name:              "異常_該当ユーザなし",
			selectErr:         repository.ErrNotFound,
			expectedErrorBody: ErrUserNotFound,
		},
		{
			name:              "異常_予期せぬDBエラー",
			selectErr:         repository.ErrTestUnexpected,
			expectedErrorBody: repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runUserLogin(t, newSelectUserRepoMock(nil, tt.selectErr))

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErrorBody)
			assert.Nil(t, output)
		})
	}
}
