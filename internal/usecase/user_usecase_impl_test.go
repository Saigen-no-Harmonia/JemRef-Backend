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
	usecaseDto "jemref_go/internal/usecase/dto"
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
				SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, nil
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
				FirebaseUserId:             "firebase_uid",
				Email:                      "sample@example.com",
				TermsAgreedVersion:         "0.01",
				PrivacyPolicyAgreedVersion: "0.02",
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
				SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, ErrPolicyNotFound
					case policy.PolicyIdPrivacyPolicy:
						return nil, nil
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
				FirebaseUserId:             "firebase_uid",
				Email:                      "sample@example.com",
				TermsAgreedVersion:         "0.01",
				PrivacyPolicyAgreedVersion: "0.02",
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
				SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, nil
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
				FirebaseUserId:             "firebase_uid",
				Email:                      "sample@example.com",
				TermsAgreedVersion:         "0.01",
				PrivacyPolicyAgreedVersion: "0.02",
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
				SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, nil
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
				FirebaseUserId:             "firebase_uid",
				Email:                      "sample@example.com",
				TermsAgreedVersion:         "0.01",
				PrivacyPolicyAgreedVersion: "0.02",
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
					return repository.ErrTestUnexpected
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, nil
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
				FirebaseUserId:             "firebase_uid",
				Email:                      "sample@example.com",
				TermsAgreedVersion:         "0.01",
				PrivacyPolicyAgreedVersion: "0.02",
			},
			expected: dto.CreateUserOutput{
				PublicUserId: "_sample_id_",
			},
			userRepoCalls:       1,
			generalRepoCalls:    2,
			idGenCalls:          1,
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
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
			assert.Equal(t, tt.generalRepoCalls, tt.mockGeneralRepo.SelectByPrimaryKeyCalled)
			assert.Equal(t, tt.userRepoCalls, tt.mockUserRepo.CreateCalled)
			assert.Equal(t, tt.idGenCalls, tt.mockIdGen.GenerateCalls)
			assert.Equal(t, 0, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByFirebaseUidCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByInternalUidCalled)

			// エラーケースチェック-------------------
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
				assert.Nil(t, actual)
				return
			}

			// repo引数チェック-------------------
			assert.Equal(t, tt.expected.PublicUserId, tt.mockUserRepo.LastInsertUser.PublicUserId)
			assert.Equal(t, tt.inputDto.FirebaseUserId, tt.mockUserRepo.LastInsertUser.FirebaseUserId)
			assert.Equal(t, tt.inputDto.TermsAgreedVersion, tt.mockUserRepo.LastInsertUser.TermsVersion)
			assert.Equal(t, tt.inputDto.PrivacyPolicyAgreedVersion, tt.mockUserRepo.LastInsertUser.PrivacyPolicyVersion)
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
					return repository.ErrTestUnexpected
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
			expectedErrorOutput: nil,
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
					return nil, repository.ErrTestUnexpected
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
			expectedErrorOutput: repository.ErrTestUnexpected,
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
			expectedErrorOutput: ErrUserDataInconsistent,
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
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
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

func TestUserUsecaseImpl_GetUserAgreements(t *testing.T) {

	tests := []struct {
		name                string
		mockUserRepo        *mock.MockUserRepository
		mockGeneralRepo     *mock.MockGeneralRepository
		inputUid            int64
		userRepoCalls       int
		userRepoArgs        []int64
		generalRepoCalls    int
		generalRepoArgs     []string
		expected            *usecaseDto.GetUserAgreementsOutput
		expectedError       bool
		expectedErrorOutput error
	}{
		{
			name: "GetUserAgreements_正常",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:       uid,
						TermsVersion:         "1.0",
						PrivacyPolicyVersion: "2.0",
					}, nil
				},
			},
			userRepoArgs: []int64{2014},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "2.0",
							Name:    "_terms_label_",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
							Name:    "_privacy_policy_label_",
						}, nil
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoArgs: []string{
				policy.PolicyIdTermsOfService,
				policy.PolicyIdPrivacyPolicy,
			},
			inputUid:         int64(2014),
			userRepoCalls:    1,
			generalRepoCalls: 2,
			expected: &usecaseDto.GetUserAgreementsOutput{
				Agreements: []usecaseDto.GetUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						Label:         "_terms_label_",
						LatestVersion: "2.0",
						AgreedVersion: "1.0",
						Status:        user.PolicyAgreementStatusUpdateRequired,
					},
					{
						PolicyType:    policy.PolicyTypePrivacyPolicy,
						Label:         "_privacy_policy_label_",
						LatestVersion: "2.0",
						AgreedVersion: "2.0",
						Status:        user.PolicyAgreementStatusAgreed,
					},
				},
			},
		},
		{
			name: "GetUserAgreements_異常_ユーザマスタ取得失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, repository.ErrNotFound
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					return nil, nil
				},
			},
			userRepoCalls:       1,
			generalRepoCalls:    0,
			expectedError:       true,
			expectedErrorOutput: ErrUserNotFound,
		},
		{
			name: "GetUserAgreements_異常_ユーザ情報取得時に予期せぬエラー",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, repository.ErrTestUnexpected
				},
			},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					return nil, nil
				},
			},
			userRepoCalls:       1,
			generalRepoCalls:    0,
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
		{
			name: "GetUserAgreements_異常_ユーザ規約取得失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:       uid,
						TermsVersion:         "1.0",
						PrivacyPolicyVersion: "2.0",
					}, nil
				},
			},
			userRepoArgs: []int64{2014},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, repository.ErrNotFound
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
							Name:    "_privacy_policy_label_",
						}, nil
					}
					return nil, errors.New("failed")
				},
			},
			generalRepoArgs: []string{
				policy.PolicyIdTermsOfService,
				policy.PolicyIdPrivacyPolicy,
			},
			inputUid:            int64(2014),
			userRepoCalls:       1,
			generalRepoCalls:    1,
			expectedError:       true,
			expectedErrorOutput: ErrPolicyNotFound,
		},
		{
			name: "GetUserAgreements_異常_ユーザ規約取得時に予期せぬエラー",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:       uid,
						TermsVersion:         "1.0",
						PrivacyPolicyVersion: "2.0",
					}, nil
				},
			},
			userRepoArgs: []int64{2014},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, repository.ErrTestUnexpected
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
							Name:    "_privacy_policy_label_",
						}, nil
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoArgs: []string{
				policy.PolicyIdTermsOfService,
				policy.PolicyIdPrivacyPolicy,
			},
			inputUid:            int64(2014),
			userRepoCalls:       1,
			generalRepoCalls:    1,
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
		{
			name: "GetUserAgreements_異常_プラポリ取得失敗",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:       uid,
						TermsVersion:         "1.0",
						PrivacyPolicyVersion: "2.0",
					}, nil
				},
			},
			userRepoArgs: []int64{2014},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "2.0",
							Name:    "_terms_label_",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, repository.ErrNotFound
					}
					return nil, repository.ErrNotFound
				},
			},
			generalRepoArgs: []string{
				policy.PolicyIdTermsOfService,
				policy.PolicyIdPrivacyPolicy,
			},
			inputUid:            int64(2014),
			userRepoCalls:       1,
			generalRepoCalls:    2,
			expectedError:       true,
			expectedErrorOutput: ErrPolicyNotFound,
		},
		{
			name: "GetUserAgreements_異常_プラポリ取得時に予期せぬエラー",
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:       uid,
						TermsVersion:         "1.0",
						PrivacyPolicyVersion: "2.0",
					}, nil
				},
			},
			userRepoArgs: []int64{2014},
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "2.0",
							Name:    "_terms_label_",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return nil, repository.ErrTestUnexpected
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoArgs: []string{
				policy.PolicyIdTermsOfService,
				policy.PolicyIdPrivacyPolicy,
			},
			inputUid:            int64(2014),
			userRepoCalls:       1,
			generalRepoCalls:    2,
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(
				tt.mockUserRepo,
				tt.mockGeneralRepo,
				nil,
				nil,
			)

			actual, err := u.GetUserAgreements(
				context.Background(),
				tt.inputUid,
			)

			// 呼び出し回数チェック
			assert.Equal(t, tt.userRepoCalls, tt.mockUserRepo.SelectByInternalUidCalled)
			assert.Equal(t, tt.generalRepoCalls, tt.mockGeneralRepo.SelectLatestByIdFuncCalled)
			assert.Equal(t, 0, tt.mockUserRepo.CreateCalled)
			assert.Equal(t, 0, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByFirebaseUidCalled)
			assert.Equal(t, 0, tt.mockGeneralRepo.SelectByPrimaryKeyCalled)

			// エラーケースチェック
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
				assert.Nil(t, actual)
				return
			}

			// repo引数チェック
			assert.Equal(t, tt.inputUid, tt.mockUserRepo.LastSelectInternalUid)
			assert.Equal(t, tt.userRepoArgs[0], tt.mockUserRepo.LastSelectInternalUid)
			assert.Equal(t, tt.generalRepoArgs, tt.mockGeneralRepo.SelectedIds)

			// outputチェック
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)

		})
	}
}

func TestUserUsecaseImpl_UpdateUserAgreements(t *testing.T) {

	testdDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	before := time.Now()
	tests := []struct {
		name                      string
		mockUserRepo              *mock.MockUserRepository
		mockGeneralRepo           *mock.MockGeneralRepository
		inputDto                  *usecaseDto.UpdateUserAgreementsInput
		selectByInternalUidCalls  int
		selectByInternalUidArgs   int64
		updateUserAgreementsCalls int
		updateUserAgreementsArgs  *user.User
		generalRepoCalls          int
		generalRepoArgs           []mock.PolicyPrimaryKeyCall
		expectedError             bool
		expectedErrorOutput       error
	}{
		{
			name: "UpdateUserAgreements_正常",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "1.0",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoCalls: 2,
			generalRepoArgs: []mock.PolicyPrimaryKeyCall{
				{
					Id:      policy.PolicyIdTermsOfService,
					Version: "1.0",
				},
				{
					Id:      policy.PolicyIdPrivacyPolicy,
					Version: "2.0",
				},
			},
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
				UpdateUserAgreementFunc: func(ctx context.Context, u *user.User) (int64, error) {
					return 1, nil
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 1,
			selectByInternalUidArgs:   int64(10001),
			updateUserAgreementsArgs: &user.User{
				InternalUserId:       int64(10001),
				TermsVersion:         "1.0",
				PrivacyPolicyVersion: "2.0",
			},
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
					{
						PolicyType:    policy.PolicyTypePrivacyPolicy,
						AgreedVersion: "2.0",
					},
				},
			},
			expectedError: false,
		},
		{
			name:             "UpdateUserAgreements_異常_ユーザが存在しない",
			mockGeneralRepo:  &mock.MockGeneralRepository{},
			generalRepoCalls: 0,
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, repository.ErrNotFound
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
					{
						PolicyType:    policy.PolicyTypePrivacyPolicy,
						AgreedVersion: "2.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: ErrUserNotFound,
		},
		{
			name:             "UpdateUserAgreements_異常_ユーザ取得処理で予期せぬエラー",
			mockGeneralRepo:  &mock.MockGeneralRepository{},
			generalRepoCalls: 0,
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return nil, repository.ErrTestUnexpected
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
					{
						PolicyType:    policy.PolicyTypePrivacyPolicy,
						AgreedVersion: "2.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
		{
			name: "UpdateUserAgreements_異常_リクエストされた規約タイプが不正",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "1.0",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, errors.New("test failed")
				},
			},
			generalRepoCalls: 1,
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
					{
						PolicyType:    "不正な規約",
						AgreedVersion: "2.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: ErrInvalidPolicyType,
		},
		{
			name: "UpdateUserAgreements_異常_規約が存在しない（バージョン指定誤り）",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, repository.ErrNotFound
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoCalls: 1,
			generalRepoArgs: []mock.PolicyPrimaryKeyCall{
				{
					Id:      policy.PolicyIdTermsOfService,
					Version: "1.0",
				},
			},
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			updateUserAgreementsArgs: &user.User{
				InternalUserId: int64(10001),
				TermsVersion:   "1.0",
			},
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: ErrInvalidPolicyVersion,
		},
		{
			name: "UpdateUserAgreements_異常_規約取得時に予期せぬエラー",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return nil, repository.ErrTestUnexpected
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, repository.ErrTestUnexpected
				},
			},
			generalRepoCalls: 1,
			generalRepoArgs: []mock.PolicyPrimaryKeyCall{
				{
					Id:      policy.PolicyIdTermsOfService,
					Version: "1.0",
				},
			},
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			updateUserAgreementsArgs: &user.User{
				InternalUserId: int64(10001),
				TermsVersion:   "1.0",
			},
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
		{
			name: "UpdateUserAgreements_異常_DBが不正な規約情報を返却",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      "P999",
							Version: "1.0",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, errors.New("test failed")
				},
			},
			generalRepoCalls: 1,
			generalRepoArgs: []mock.PolicyPrimaryKeyCall{
				{
					Id:      policy.PolicyIdTermsOfService,
					Version: "1.0",
				},
			},
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 0,
			selectByInternalUidArgs:   int64(10001),
			updateUserAgreementsArgs: &user.User{
				InternalUserId:       int64(10001),
				TermsVersion:         "1.0",
				PrivacyPolicyVersion: "2.0",
			},
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: ErrUnexpectedPolicy,
		},
		{
			name: "UpdateUserAgreements_異常_ユーザマスタ更新に失敗",
			mockGeneralRepo: &mock.MockGeneralRepository{
				SelectByPrimaryKeyFunc: func(ctx context.Context, id, version string) (*policy.Policy, error) {
					switch id {
					case policy.PolicyIdTermsOfService:
						return &policy.Policy{
							Id:      policy.PolicyIdTermsOfService,
							Version: "1.0",
						}, nil
					case policy.PolicyIdPrivacyPolicy:
						return &policy.Policy{
							Id:      policy.PolicyIdPrivacyPolicy,
							Version: "2.0",
						}, nil
					}
					return nil, errors.New("test failed")
				},
			},
			generalRepoCalls: 2,
			generalRepoArgs: []mock.PolicyPrimaryKeyCall{
				{
					Id:      policy.PolicyIdTermsOfService,
					Version: "1.0",
				},
				{
					Id:      policy.PolicyIdPrivacyPolicy,
					Version: "2.0",
				},
			},
			mockUserRepo: &mock.MockUserRepository{
				SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
					return &user.User{
						InternalUserId:        int64(10001),
						TermsAgreedDt:         &testdDate,
						TermsVersion:          "_tmp1_",
						PrivacyPolicyAgreedDt: &testdDate,
						PrivacyPolicyVersion:  "_tmp2_",
					}, nil
				},
				UpdateUserAgreementFunc: func(ctx context.Context, u *user.User) (int64, error) {
					return 0, repository.ErrTestUnexpected
				},
			},
			selectByInternalUidCalls:  1,
			updateUserAgreementsCalls: 1,
			selectByInternalUidArgs:   int64(10001),
			updateUserAgreementsArgs: &user.User{
				InternalUserId:       int64(10001),
				TermsVersion:         "1.0",
				PrivacyPolicyVersion: "2.0",
			},
			inputDto: &usecaseDto.UpdateUserAgreementsInput{
				InternalUid: int64(10001),
				Agreements: []dto.UpdateUserAgreement{
					{
						PolicyType:    policy.PolicyTypeTermsOfService,
						AgreedVersion: "1.0",
					},
					{
						PolicyType:    policy.PolicyTypePrivacyPolicy,
						AgreedVersion: "2.0",
					},
				},
			},
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUserUsecase(
				tt.mockUserRepo,
				tt.mockGeneralRepo,
				nil,
				nil,
			)

			err := u.UpdateUserAgreements(
				context.Background(),
				*tt.inputDto,
			)

			after := time.Now()

			// 呼び出し回数チェック
			assert.Equal(t, tt.generalRepoCalls, tt.mockGeneralRepo.SelectByPrimaryKeyCalled)
			assert.Equal(t, tt.updateUserAgreementsCalls, tt.mockUserRepo.UpdateUserAgreementCalled)
			assert.Equal(t, tt.selectByInternalUidCalls, tt.mockUserRepo.SelectByInternalUidCalled)
			assert.Equal(t, 0, tt.mockGeneralRepo.SelectLatestByIdFuncCalled)
			assert.Equal(t, 0, tt.mockUserRepo.CreateCalled)
			assert.Equal(t, 0, tt.mockUserRepo.DeleteCalled)
			assert.Equal(t, 0, tt.mockUserRepo.SelectByFirebaseUidCalled)

			// エラーケースチェック
			if tt.expectedError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
				return
			}

			// repo引数チェック
			assert.Equal(t, tt.generalRepoArgs, tt.mockGeneralRepo.SelectedPrimaryKeys)
			assert.Equal(t, tt.inputDto.InternalUid, tt.mockUserRepo.LastSelectInternalUid)
			assert.Equal(t, tt.updateUserAgreementsArgs.InternalUserId, tt.mockUserRepo.LastUpdatedUser.InternalUserId)
			assert.True(t, before.Before(*tt.mockUserRepo.LastUpdatedUser.TermsAgreedDt))
			assert.True(t, after.After(*tt.mockUserRepo.LastUpdatedUser.TermsAgreedDt))
			assert.True(t, before.Before(*tt.mockUserRepo.LastUpdatedUser.PrivacyPolicyAgreedDt))
			assert.True(t, after.After(*tt.mockUserRepo.LastUpdatedUser.PrivacyPolicyAgreedDt))
			assert.Equal(t, tt.updateUserAgreementsArgs.TermsVersion, tt.mockUserRepo.LastUpdatedUser.TermsVersion)
			assert.Equal(t, tt.updateUserAgreementsArgs.PrivacyPolicyVersion, tt.mockUserRepo.LastUpdatedUser.PrivacyPolicyVersion)

			// outputチェック
			assert.Nil(t, err)
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
					return nil, repository.ErrTestUnexpected
				},
			},
			inputDto: dto.UserLoginInput{
				InternalUserId: int64(1002),
			},
			userRepoCalls:       1,
			expectedError:       true,
			expectedErrorOutput: repository.ErrTestUnexpected,
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
				assert.ErrorIs(t, err, tt.expectedErrorOutput)
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

func uids(uids ...int64) []int64 {
	return uids
}

func policyPk(id, version string) mock.PolicyPrimaryKeyCall {
	return mock.PolicyPrimaryKeyCall{
		Id:      id,
		Version: version,
	}
}
