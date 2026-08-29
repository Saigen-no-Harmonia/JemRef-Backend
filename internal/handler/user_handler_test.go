package handler

import (
	"context"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	handlerDto "jemref_go/internal/handler/dto"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_RegisterRoutes(t *testing.T) {
	r := gin.New()
	group := r.Group("/api/v0")

	h := NewUserHandler(nil)
	h.RegisterRoutes(group)

	routes := r.Routes()

	expected := map[string]bool{
		"DELETE /api/v0/users":         false,
		"GET /api/v0/users/agreements": false,
		"PUT /api/v0/users/agreements": false,
		"PUT /api/v0/users/login":      false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}

	for key, found := range expected {
		assert.True(t, found, "%s is not registered", key)
	}
}

func TestUserHandler_CreateUser(t *testing.T) {

	path := "/api/v0/join"
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		body              handlerDto.CreateUserRequest
		CtxFirebaseUid    string
		CtxEmail          string
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectError       bool
		expectedBody      handlerDto.CreateUserResponse
		expectedErrorBody error
	}{
		{
			name:           "正常_ユーザ登録",
			CtxFirebaseUid: "firebase_uid",
			CtxEmail:       "test@example.com",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return &usecaseDto.CreateUserOutput{
						PublicUserId: "_test_public_userid_",
					}, nil
				},
			},
			usecaseCalled: true,
			expectError:   false,
			expectedBody: handlerDto.CreateUserResponse{
				PublicUserId: "_test_public_userid_",
			},
		},
		{
			name:     "異常_ユーザ登録_FirebaseUID取得失敗",
			CtxEmail: "test@example.com",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrFirebaseUidNotFound,
		},
		{
			name:           "異常_ユーザ登録_Email取得失敗",
			CtxFirebaseUid: "firebase_uid",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrEmailNotFound,
		},
		{
			name:           "異常_ユーザ登録_リクエスト不正",
			CtxFirebaseUid: "firebase_uid",
			CtxEmail:       "test@example.com",
			body:           handlerDto.CreateUserRequest{},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return &usecaseDto.CreateUserOutput{
						PublicUserId: "_test_public_userid_",
					}, nil
				},
			},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:           "異常_ユーザ登録_既存ユーザあり",
			CtxFirebaseUid: "firebase_uid",
			CtxEmail:       "test@example.com",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, usecase.ErrUserAlreadyExists
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrUserAlreadyExists,
		},
		{
			name:           "異常_ユーザ登録_規約指定不正",
			CtxFirebaseUid: "firebase_uid",
			CtxEmail:       "test@example.com",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.3",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, usecase.ErrPolicyNotFound
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrPolicyNotFound,
		},
		{
			name:           "異常_ユーザ登録_予期しないエラー",
			CtxFirebaseUid: "firebase_uid",
			CtxEmail:       "test@example.com",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, usecase.ErrTestUnexpected
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyFirebaseUid, tt.CtxFirebaseUid)
			c.Set(ctxutil.CtxKeyEmail, tt.CtxEmail)

			req := httptest.NewRequest(
				http.MethodPost,
				path,
				testutil.NewJsonReader(t, tt.body),
			)

			c.Request = req

			h := NewUserHandler(tt.mockUsecase)
			h.CreateUser(c)

			// 呼び出しチェック
			if tt.usecaseCalled {
				assert.True(t, tt.mockUsecase.CreateUserCalled)
				assert.Equal(t, tt.body.TermsAgreedVersion, tt.mockUsecase.LastCreateUserInput.TermsAgreedVersion)
				assert.Equal(t, tt.body.PrivacyPolicyAgreedVersion, tt.mockUsecase.LastCreateUserInput.PrivacyPolicyAgreedVersion)
			} else {
				assert.False(t, tt.mockUsecase.CreateUserCalled)
			}
			assert.False(t, tt.mockUsecase.DeleteUserCalled)
			assert.False(t, tt.mockUsecase.LoginCalled)

			// 異常系の場合のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}

			// 正常系の場合のassertion
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	path := "/api/v0/users"
	gin.SetMode(gin.TestMode)
	firebaseUid := "firebase_uid"
	internalUid := int64(1234)

	tests := []struct {
		name              string
		ctxInternalUid    int64
		ctxFirebaseUid    string
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_ユーザ退会",
			ctxInternalUid: internalUid,
			ctxFirebaseUid: firebaseUid,
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return nil
				},
			},
			usecaseCalled: true,
			expectError:   false,
		},
		{
			name:              "異常_ユーザ退会_InternalUIDの取得に失敗",
			ctxInternalUid:    int64(0),
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_ユーザ退会_FirebaseUIDの取得に失敗",
			ctxInternalUid:    internalUid,
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrFirebaseUidNotFound,
		},
		{
			name:           "異常_ユーザ退会_未入会ユーザ",
			ctxInternalUid: internalUid,
			ctxFirebaseUid: firebaseUid,
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return usecase.ErrUserNotFound
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrUserNotFound,
		},
		{
			name:           "異常_ユーザ退会_ユーザ情報に不整合あり",
			ctxInternalUid: internalUid,
			ctxFirebaseUid: firebaseUid,
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return usecase.ErrUserDataInconsistent
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrUserDataInconsistent,
		},
		{
			name:           "異常_ユーザ退会_予期しないエラー",
			ctxInternalUid: internalUid,
			ctxFirebaseUid: firebaseUid,
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return usecase.ErrTestUnexpected
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)
			c.Set(ctxutil.CtxKeyFirebaseUid, tt.ctxFirebaseUid)

			req := httptest.NewRequest(
				http.MethodDelete,
				path,
				nil,
			)

			c.Request = req

			h := NewUserHandler(tt.mockUsecase)
			h.DeleteUser(c)

			// 呼び出しチェック
			if tt.usecaseCalled {
				assert.True(t, tt.mockUsecase.DeleteUserCalled)
				assert.Equal(t, internalUid, tt.mockUsecase.LastDeleteUserInput.InternalUserid)
				assert.Equal(t, firebaseUid, tt.mockUsecase.LastDeleteUserInput.FirebaseUserId)

			} else {
				assert.False(t, tt.mockUsecase.DeleteUserCalled)
			}
			assert.False(t, tt.mockUsecase.CreateUserCalled)
			assert.False(t, tt.mockUsecase.LoginCalled)

			// 異常系の場合
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}
		})
	}
}

func TestUserHandler_GetUserAgreements(t *testing.T) {
	path := "/api/v0/users/agreements"
	internalUid := int64(101)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		ctxInternalUid    int64
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectError       bool
		expectedBody      handlerDto.GetUserAgreementsResponse
		expectedErrorBody error
	}{
		{
			name:           "正常_ユーザ規約同意状況参照",
			ctxInternalUid: internalUid,
			mockUsecase: &mock.MockUserUsecase{
				GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
					agreements := []usecaseDto.GetUserAgreement{
						{
							PolicyType:    policy.PolicyTypeTermsOfService,
							Label:         "_terms_",
							LatestVersion: "_1.0_",
							AgreedVersion: "_1.0_",
							Status:        user.PolicyAgreementStatusAgreed,
						},
						{
							PolicyType:    policy.PolicyTypePrivacyPolicy,
							Label:         "_プラポリ_",
							LatestVersion: "_2.0_",
							AgreedVersion: "_1.0_",
							Status:        user.PolicyAgreementStatusUpdateRequired,
						},
					}
					return &usecaseDto.GetUserAgreementsOutput{
						Agreements: agreements,
					}, nil
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.GetUserAgreementsResponse{
				Agreements: []handlerDto.UserAgreementResponse{
					{
						PolicyType:    string(policy.PolicyTypeTermsOfService),
						Label:         "_terms_",
						LatestVersion: "_1.0_",
						AgreedVersion: "_1.0_",
						Status:        user.PolicyAgreementStatusAgreed,
					},
					{
						PolicyType:    string(policy.PolicyTypePrivacyPolicy),
						Label:         "_プラポリ_",
						LatestVersion: "_2.0_",
						AgreedVersion: "_1.0_",
						Status:        user.PolicyAgreementStatusUpdateRequired,
					},
				},
			},
			expectError: false,
		},
		{
			name:              "異常_ユーザ規約同意状況参照_internal uid取得失敗",
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:           "異常_ユーザ規約同意状況参照_ユーザ非存在",
			ctxInternalUid: internalUid,
			mockUsecase: &mock.MockUserUsecase{
				GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
					return nil, usecase.ErrUserNotFound
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrUserNotFound,
		},
		{
			name:           "異常_ユーザ規約同意状況参照_ユーザ情報取得時に予期せぬエラー",
			ctxInternalUid: internalUid,
			mockUsecase: &mock.MockUserUsecase{
				GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
					return nil, usecase.ErrTestUnexpected
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)

			req := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)

			c.Request = req

			h := NewUserHandler(tt.mockUsecase)
			h.GetUserAgreements(c)

			// 呼び出しチェック
			if tt.usecaseCalled {
				assert.True(t, tt.mockUsecase.GetUserAgreementsCalled)
			} else {
				assert.False(t, tt.mockUsecase.GetUserAgreementsCalled)
			}
			assert.False(t, tt.mockUsecase.CreateUserCalled)
			assert.False(t, tt.mockUsecase.DeleteUserCalled)

			// 異常系のチェック
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}

			// 正常系のチェック
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}

func TestUserHandler_UpdateUserAgreements(t *testing.T) {
	path := "/api/v0/user/agreements"
	gin.SetMode(gin.TestMode)
	uid := int64(2001)

	tests := []struct {
		name              string
		ctxInternalUid    int64
		body              handlerDto.UpdateUserAgreementsRequest
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_ユーザ規約同意状況更新",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{
						PolicyType: string(policy.PolicyTypeTermsOfService),
						Version:    "0.1",
					},
					{
						PolicyType: string(policy.PolicyTypePrivacyPolicy),
						Version:    "0.2",
					},
				},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled: true,
			expectError:   false,
		},
		{
			name:              "異常_ユーザ規約同意状況更新_ユーザID取得失敗",
			body:              handlerDto.UpdateUserAgreementsRequest{},
			mockUsecase:       &mock.MockUserUsecase{},
			usecaseCalled:     false,
			expectError:       false,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_リクエスト形式不正",
			ctxInternalUid: uid,
			body:           handlerDto.UpdateUserAgreementsRequest{},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_リクエストが空の配列",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrPolicyRequired,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_ユーザ取得処理異常",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{
						PolicyType: string(policy.PolicyTypeTermsOfService),
						Version:    "0.1",
					},
					{
						PolicyType: string(policy.PolicyTypePrivacyPolicy),
						Version:    "0.2",
					},
				},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrUserNotFound
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrUserNotFound,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_規約タイプ指定不正",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{
						PolicyType: string(policy.PolicyTypeTermsOfService),
						Version:    "0.1",
					},
					{
						PolicyType: string(policy.PolicyTypePrivacyPolicy),
						Version:    "0.2",
					},
				},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrInvalidPolicyType
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrInvalidPolicyType,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_規約バージョン指定不正",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{
						PolicyType: string(policy.PolicyTypeTermsOfService),
						Version:    "0.1",
					},
					{
						PolicyType: string(policy.PolicyTypePrivacyPolicy),
						Version:    "0.2",
					},
				},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrInvalidPolicyVersion
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrInvalidPolicyVersion,
		},
		{
			name:           "異常_ユーザ規約同意状況更新_予期せぬエラー",
			ctxInternalUid: uid,
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{
						PolicyType: string(policy.PolicyTypeTermsOfService),
						Version:    "0.1",
					},
					{
						PolicyType: string(policy.PolicyTypePrivacyPolicy),
						Version:    "0.2",
					},
				},
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrTestUnexpected
				},
			},
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)

			req := httptest.NewRequest(
				http.MethodPut,
				path,
				testutil.NewJsonReader(t, tt.body),
			)
			c.Request = req

			h := NewUserHandler(tt.mockUsecase)
			h.UpdateUserAgreements(c)

			// 呼び出しチェック
			if tt.usecaseCalled {
				assert.True(t, tt.mockUsecase.UpdateUserAgreementsCalled)
				assert.Equal(t, uid, tt.mockUsecase.LastUpdateUserInput.InternalUid)
				for i, p := range tt.body.Policies {
					arg := tt.mockUsecase.LastUpdateUserInput.Agreements[i]
					assert.Equal(t, p.PolicyType, string(arg.PolicyType))
					assert.Equal(t, p.Version, string(arg.AgreedVersion))
				}
			} else {
				assert.False(t, tt.mockUsecase.UpdateUserAgreementsCalled)
			}
			assert.False(t, tt.mockUsecase.CreateUserCalled)
			assert.False(t, tt.mockUsecase.DeleteUserCalled)
			assert.False(t, tt.mockUsecase.GetUserAgreementsCalled)
			assert.False(t, tt.mockUsecase.LoginCalled)

			// 異常系の婆のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	path := "/api/v0/users/login"
	gin.SetMode(gin.TestMode)
	publicUid := "public_uid"

	tests := []struct {
		name            string
		ctxPublicUid    string
		mockUsecase     *mock.MockUserUsecase
		usecaseCalled   bool
		expectedBody    handlerDto.UserLoginResponse
		expectError     bool
		expectErrorBody error
	}{
		{
			name:         "正常_ユーザログイン",
			ctxPublicUid: publicUid,
			expectedBody: handlerdto.UserLoginResponse{
				PublicUserId: "public_uid",
			},
			expectError: false,
		},
		{
			name:            "異常_ユーザログイン_公開用ユーザID取得失敗",
			expectError:     true,
			expectErrorBody: ErrPublicUidNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyPublicUid, tt.ctxPublicUid)

			req := httptest.NewRequest(
				http.MethodPut,
				path,
				nil,
			)
			c.Request = req

			h := NewUserHandler(tt.mockUsecase)
			h.Login(c)

			// 異常系の場合
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectErrorBody)
				return
			}

			// 正常系の場合
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}
