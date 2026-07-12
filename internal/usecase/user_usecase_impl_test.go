package usecase

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestUserUsecaseImpl_CreateUser(t *testing.T) {

	tests := []struct {
		name                string
		mockUserRepo        *mock.MockUserRepository
		mockGeneralRepo     *mock.MockGeneralRepository
		mockFirebaseRepo    *mock.MockFirebaseRepository
		mockIdGen           *mock.MockIdGen
		inputDto            dto.CreateUserInput
		expected            dto.CreateUserOutput
		userRepoCalls       int
		generalRepoCalls    int
		idGenCalls          int
		expectedError       bool
		expectedErrorOutput error
	}{
		{
			name: "CreateUser_正常",
			mockUserRepo: &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, u *user.User) error {
					return nil
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Version: "0.01",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Version: "0.02",
						}, nil
					}
					return nil, ErrPolicyNotFound
				},
			},
			mockIdGen: &mock.MockIdGen{
				GenerateFunc: func() string {
					return "_sample_id_"
				},
			},
			inputDto: dto.CreateUserInput{
				FirebaseUserId: "firebase_uid",
				Email:          "sample@example.com",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:    1,
			generalRepoCalls: 2,
			idGenCalls:       1,
			expectedError:    false,
		},
		{
			name: "CreateUser_異常_ユーザ規約チェック失敗",
			mockUserRepo: &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, u *user.User) error {
					return nil
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, ErrPolicyNotFound
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Version: "0.02",
						}, nil
					}
					return nil, ErrPolicyNotFound
				},
			},
			mockIdGen: &mock.MockIdGen{
				GenerateFunc: func() string {
					return "_sample_id_"
				},
			},
			inputDto: dto.CreateUserInput{
				FirebaseUserId: "firebase_uid",
				Email:          "sample@example.com",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:       0,
			generalRepoCalls:    1,
			idGenCalls:          1,
			expectedError:       true,
			expectedErrorOutput: ErrPolicyNotFound,
		},
		{
			name: "CreateUser_異常_プラポリ規約チェック失敗",
			mockUserRepo: &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, u *user.User) error {
					return nil
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Version: "0.01",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, ErrPolicyNotFound
					}
					return nil, ErrPolicyNotFound
				},
			},
			mockIdGen: &mock.MockIdGen{
				GenerateFunc: func() string {
					return "_sample_id_"
				},
			},
			inputDto: dto.CreateUserInput{
				FirebaseUserId: "firebase_uid",
				Email:          "sample@example.com",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:       0,
			generalRepoCalls:    2,
			idGenCalls:          1,
			expectedError:       true,
			expectedErrorOutput: ErrPolicyNotFound,
		},
		{
			name: "CreateUser_異常_ユーザ重複エラー",
			mockUserRepo: &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, u *user.User) error {
					return &mysql.MySQLError{
						Number: 1062,
					}
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Version: "0.01",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Version: "0.02",
						}, nil
					}
					return nil, ErrPolicyNotFound
				},
			},
			mockIdGen: &mock.MockIdGen{
				GenerateFunc: func() string {
					return "_sample_id_"
				},
			},
			inputDto: dto.CreateUserInput{
				FirebaseUserId: "firebase_uid",
				Email:          "sample@example.com",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:       1,
			generalRepoCalls:    2,
			idGenCalls:          1,
			expectedError:       true,
			expectedErrorOutput: ErrUserAlreadyExists,
		},
		{
			name: "CreateUser_異常_その他のユーザRepoエラー",
			mockUserRepo: &mock.MockUserRepository{
				CreateFunc: func(ctx context.Context, u *user.User) error {
					return errors.New("unexpeted error")
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Version: "0.01",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Version: "0.02",
						}, nil
					}
					return nil, ErrPolicyNotFound
				},
			},
			mockIdGen: &mock.MockIdGen{
				GenerateFunc: func() string {
					return "_sample_id_"
				},
			},
			inputDto: dto.CreateUserInput{
				FirebaseUserId: "firebase_uid",
				Email:          "sample@example.com",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:       1,
			generalRepoCalls:    2,
			idGenCalls:          1,
			expectedError:       true,
			expectedErrorOutput: errors.New("unexpeted error"),
		},
	}

	for _, tt := range tests {
		timestampBefore := time.Now()
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(
				tt.mockUserRepo,
				tt.mockGeneralRepo,
				tt.mockFirebaseRepo,
				tt.mockIdGen,
			)

			actual, err := u.Create(
				context.Background(),
				tt.inputDto,
			)
			timestampAfter := time.Now()

			// 呼び出し回数チェック -------------------
			assert.Equal(t, tt.generalRepoCalls, tt.mockGeneralRepo.SelectLatestByIdFuncCalled)
			assert.Equal(t, tt.userRepoCalls, tt.mockUserRepo.CreateCalled)
			assert.Equal(t, tt.idGenCalls, tt.mockIdGen.GenerateCalls)
			assert.Equal(t, 0, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByFirebaseUidCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByInternalUidCalled)

			// エラーケースチェック-------------------
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorOutput, err)
				assert.Nil(t, actual)
				return
			}

			// repo引数チェック-------------------
			assert.Equal(t, tt.expected.PublicUserId, tt.mockUserRepo.LastInsertUser.PublicUserId)
			assert.Equal(t, tt.inputDto.FirebaseUserId, tt.mockUserRepo.LastInsertUser.FirebaseUserId)
			assert.Equal(t, tt.inputDto.Email, tt.mockUserRepo.LastInsertUser.Email)
			assert.Equal(t, "0.01", tt.mockUserRepo.LastInsertUser.TermsVersion)
			assert.Equal(t, "0.02", tt.mockUserRepo.LastInsertUser.PrivacyPolicyVersion)
			assert.True(t, timestampBefore.Before(*tt.mockUserRepo.LastInsertUser.TermsAgreedDt))
			assert.True(t, timestampAfter.After(*tt.mockUserRepo.LastInsertUser.TermsAgreedDt))
			assert.True(t, timestampBefore.Before(*tt.mockUserRepo.LastInsertUser.PrivacyPolicyAgreedDt))
			assert.True(t, timestampAfter.After(*tt.mockUserRepo.LastInsertUser.PrivacyPolicyAgreedDt))

			// outputチェック-------------------
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.PublicUserId, actual.PublicUserId)
		})
	}
}

func TestUserUsecaseImpl_DeleteUser(t *testing.T) {

	tests := []struct {
		name                string
		mockUserRepo        *mock.MockUserRepository
		mockFirebaseRepo    *mock.MockFirebaseRepository
		inputDto            dto.DeleteUserInput
		userDeleted         bool
		userSelectCalls     int
		userDeleteCalls     int
		firebaseDeleteCalls int
		expectedError       bool
		expectedErrorOutput error
	}{
		{
			name: "DeleteUser_正常",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(1001),
						FirebaseUserId: "_firebase_uid_",
					}, nil
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     1,
			firebaseDeleteCalls: 1,
			expectedError:       false,
		},
		{
			name: "DeleteUser_準正常_ユーザ論理削除済み",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(1001),
						FirebaseUserId: "_firebase_uid_",
						DeletedAt:      testutil.TimePtr(time.Now()),
					}, nil
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userDeleted:         true,
			userSelectCalls:     1,
			userDeleteCalls:     0,
			firebaseDeleteCalls: 1,
			expectedError:       false,
		},
		{
			name: "DeleteUser_準正常_firebaseユーザ削除失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(1001),
						FirebaseUserId: "_firebase_uid_",
					}, nil
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return errors.New("error: failed to delete firebase user")
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     1,
			firebaseDeleteCalls: 1,
			expectedError:       false,
		},
		{
			name: "DeleteUser_異常_該当ユーザなし",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, sql.ErrNoRows
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     0,
			firebaseDeleteCalls: 0,
			expectedError:       true,
			expectedErrorOutput: ErrUserNotFound,
		},
		{
			name: "DeleteUser_異常_ユーザSELECT失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, errors.New("error")
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     0,
			firebaseDeleteCalls: 0,
			expectedError:       true,
			expectedErrorOutput: errors.New("error"),
		},
		{
			name: "DeleteUser_異常_パラメータ不正",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(1001),
						FirebaseUserId: "_firebase_uid_",
					}, nil
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return nil
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_invalid_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     0,
			firebaseDeleteCalls: 0,
			expectedError:       true,
			expectedErrorOutput: ErrInvalidUser,
		},
		{
			name: "DeleteUser_異常_DBデータ削除失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId: int64(1001),
						FirebaseUserId: "_firebase_uid_",
					}, nil
				},
				DeleteFunc: func(ctx context.Context, uid int64) error {
					return errors.New("error delete failed")
				},
			},
			mockFirebaseRepo: &mock.MockFirebaseRepository{
				DeleteUserFunc: func(ctx context.Context, uid string) error {
					return nil
				},
			},
			inputDto: dto.DeleteUserInput{
				InternalUserid: int64(1001),
				FirebaseUserId: "_firebase_uid_",
			},
			userSelectCalls:     1,
			userDeleteCalls:     1,
			firebaseDeleteCalls: 0,
			expectedError:       true,
			expectedErrorOutput: ErrUserDeleteFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(
				tt.mockUserRepo,
				nil,
				tt.mockFirebaseRepo,
				nil,
			)

			err := u.Delete(
				context.Background(),
				tt.inputDto,
			)

			// 呼び出し回数チェック -------------------
			assert.Equal(t, tt.userSelectCalls, tt.mockUserRepo.SelectByInternalUidCalled)
			assert.Equal(t, tt.userDeleteCalls, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, tt.firebaseDeleteCalls, tt.mockFirebaseRepo.DeleteUserCalled)
			assert.Equal(t, tt.userSelectCalls+tt.userDeleteCalls, tt.mockUserRepo.Called)
			assert.Equal(t, tt.firebaseDeleteCalls, tt.mockFirebaseRepo.DeleteUserCalled)

			// エラーケースチェック-------------------
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorOutput, err)
				return
			}

			// repo引数チェック
			if !tt.userDeleted {
				assert.Equal(t, tt.inputDto.InternalUserid, tt.mockUserRepo.LastDeleteUserId)
			}
			assert.Equal(t, tt.inputDto.FirebaseUserId, tt.mockFirebaseRepo.DeletedFirebaseUid)

			// outputチェック
			assert.NoError(t, err)
		})
	}
}

func TestUserUseraseImpl_Login(t *testing.T) {

	tests := []struct {
		name                string
		mockUserRepo        *mock.MockUserRepository
		inputDto            dto.UserLoginInput
		expected            dto.UserLoginOutput
		userRepoCalls       int
		expectedError       bool
		expectedErrorOutput error
	}{
		{
			name: "Login_正常",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						PublicUserId: "_public_uid_",
					}, nil
				},
			},
			inputDto: dto.UserLoginInput{
				InternalUserId: int64(1002),
			},
			expected: dto.UserLoginOutput{
				PublicUserId: "_public_uid_",
			},
			userRepoCalls: 1,
			expectedError: false,
		},
		{
			name: "Login_異常_該当ユーザなし",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, repository.ErrNotFound
				},
			},
			inputDto: dto.UserLoginInput{
				InternalUserId: int64(1002),
			},
			userRepoCalls:       1,
			expectedError:       true,
			expectedErrorOutput: ErrUserNotFound,
		},
		{
			name: "Login_異常_予期せぬDBエラー",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, errors.New("error")
				},
			},
			inputDto: dto.UserLoginInput{
				InternalUserId: int64(1002),
			},
			userRepoCalls:       1,
			expectedError:       true,
			expectedErrorOutput: errors.New("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(
				tt.mockUserRepo,
				nil,
				nil,
				nil,
			)

			actual, err := u.Login(
				context.Background(),
				tt.inputDto,
			)

			// 呼び出し回数チェック -------------------
			assert.Equal(t, tt.userRepoCalls, tt.mockUserRepo.SelectByInternalUidCalled)
			assert.Equal(t, 0, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, 0, tt.mockUserRepo.CreateCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByFirebaseUidCalled)

			// エラーケースチェック-------------------
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorOutput, err)
				assert.Nil(t, actual)
				return
			}

			// repo引数チェック-------------------
			assert.Equal(t, tt.inputDto.InternalUserId, tt.mockUserRepo.LastSelectInternalUid)

			// outputチェック
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.PublicUserId, actual.PublicUserId)
		})
	}

}
