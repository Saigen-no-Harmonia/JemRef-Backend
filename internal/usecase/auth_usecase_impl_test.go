package usecase

import (
	"context"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthUsecaseImpl_Authenticate(t *testing.T) {
	tests := []struct {
		name                     string
		mockUserRepo             *mock.MockUserRepository
		inputUid                 string
		expectDto                dto.AuthUserOutput
		selectByFirebaseUidCalls int
		expectedError            bool
		expectedErrorOutput      error
	}{
		{
			name: "Authenticate_正常",
			mockUserRepo: &mock.MockUserRepository{
				SelectByFirebaseUidFunc: func(ctx context.Context, firebaseUid string) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(2001),
						PublicUserId:   "_public_uid_",
						FirebaseUserId: "_firebase_uid",
					}, nil
				},
			},
			inputUid: "_firebase_uid_",
			expectDto: dto.AuthUserOutput{
				InternalUserId: int64(2001),
				PublicUserId:   "_public_uid_",
				FirebaseUserId: "_firebase_uid",
			},
			selectByFirebaseUidCalls: 1,
			expectedError:            false,
		},
		{
			name: "Authenticate_異常_該当ユーザなし",
			mockUserRepo: &mock.MockUserRepository{
				SelectByFirebaseUidFunc: func(ctx context.Context, firebaseUid string) (*user.User, error) {
					return nil, repository.ErrNotFound
				},
			},
			inputUid:                 "_firebase_uid_",
			selectByFirebaseUidCalls: 1,
			expectedError:            true,
			expectedErrorOutput:      ErrUserNotFound,
		},
		{
			name: "Authenticate_異常_予期せぬDBエラー",
			mockUserRepo: &mock.MockUserRepository{
				SelectByFirebaseUidFunc: func(ctx context.Context, firebaseUid string) (*user.User, error) {
					return nil, repository.ErrTestUnexpected
				},
			},
			inputUid:                 "_firebase_uid_",
			selectByFirebaseUidCalls: 1,
			expectedError:            true,
			expectedErrorOutput:      repository.ErrTestUnexpected,
		},
		{
			name: "Authenticate_異常_退会ずみ",
			mockUserRepo: &mock.MockUserRepository{
				SelectByFirebaseUidFunc: func(ctx context.Context, firebaseUid string) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(2001),
						PublicUserId:   "_public_uid_",
						FirebaseUserId: "_firebase_uid",
						DeletedAt:      testutil.TimePtr(time.Now()),
					}, nil
				},
			},
			inputUid:                 "_firebase_uid_",
			selectByFirebaseUidCalls: 1,
			expectedError:            true,
			expectedErrorOutput:      ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			u := NewAuthUsecaseImpl(
				tt.mockUserRepo,
				nil,
			)

			actual, err := u.Authenticate(
				context.Background(),
				tt.inputUid,
			)

			// 呼び出し回数チェック
			assert.Equal(t, tt.selectByFirebaseUidCalls, tt.mockUserRepo.Called)

			// エラーケースチェック
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
				return
			}

			// 引数チェック
			assert.Equal(t, tt.inputUid, tt.mockUserRepo.LastSelectFirebaseUid)

			// outputチェック
			assert.Equal(t, tt.expectDto.InternalUserId, actual.InternalUserId)
			assert.Equal(t, tt.expectDto.PublicUserId, actual.PublicUserId)
			assert.Equal(t, tt.expectDto.FirebaseUserId, actual.FirebaseUserId)
		})
	}
}

func TestAuthUsecaseImpl_DeleteUser(t *testing.T) {
	tests := []struct {
		name                   string
		mockFirebaserepository *mock.MockFirebaseRepository
		inputUid               string
		deleteUserCalls        int
		expectError            bool
		expectedErrorOutput    error
	}{
		{
			name: "DeletedUser_正常",
			mockFirebaserepository: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputUid:        "_firebase_uid_",
			deleteUserCalls: 1,
			expectError:     false,
		},
		{
			name: "DeletedUser_異常_ユーザ削除失敗",
			mockFirebaserepository: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return repository.ErrTestUnexpected
				},
			},
			inputUid:            "_firebase_uid_",
			deleteUserCalls:     1,
			expectError:         true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			u := NewAuthUsecaseImpl(
				nil,
				tt.mockFirebaserepository,
			)

			err := u.DeleteUser(
				context.Background(),
				tt.inputUid,
			)

			// 呼び出し回数チェック
			assert.Equal(t, tt.deleteUserCalls, tt.mockFirebaserepository.Called)
			assert.Equal(t, tt.deleteUserCalls, tt.mockFirebaserepository.DeleteUserCalled)

			// エラーケースチェック
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
				return
			}

			// 引数チェック
			assert.Equal(t, tt.inputUid, tt.mockFirebaserepository.DeletedFirebaseUid)

			// outputチェック
			assert.NoError(t, err)
		})
	}
}
