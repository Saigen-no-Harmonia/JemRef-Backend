package handler

import (
	"context"
	"errors"
	"fmt"
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
		setUpContext      gin.HandlerFunc
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectedStatus    int
		expectError       bool
		expectedBody      handlerDto.CreateUserResponse
		expectedErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザ登録",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
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
			usecaseCalled:  true,
			expectedStatus: http.StatusCreated,
			expectError:    false,
			expectedBody: handlerDto.CreateUserResponse{
				PublicUserId: "_test_public_userid_",
			},
		},
		{
			name: "異常_ユーザ登録_FirebaseUID取得失敗",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:FirebaseUIDの取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ登録_Email取得失敗",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "")
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:Eメールアドレスの取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ登録_リクエスト不正",
			body: handlerDto.CreateUserRequest{},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
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
			usecaseCalled:  false,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code:    "E0002",
				Message: "リクエストが不正です。",
			},
		},
		{
			name: "異常_ユーザ登録_既存ユーザあり",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, usecase.ErrUserAlreadyExists
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusConflict,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0005",
				Message: fmt.Sprintf("すでに存在するFirebase UIDまたはEmailです。FirebaseUID: %s, Email: %s", "firebase_uid", "test@example.com"),
			},
		},
		{
			name: "異常_ユーザ登録_規約指定不正",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.3",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, usecase.ErrPolicyNotFound
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0004",
				Message: "不正な規約情報です。",
			},
		},
		{
			name: "異常_ユーザ登録_予期しないエラー",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyEmail, "test@example.com")
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				CreateUserFunc: func(
					ctx context.Context,
					input usecaseDto.CreateUserInput,
				) (*usecaseDto.CreateUserOutput, error) {
					return nil, errors.New("意図しないエラー")
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:ユーザ作成処理に失敗しました。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := gin.New()
			if tt.setUpContext != nil {
				r.Use(tt.setUpContext)
			}
			h := NewUserHandler(tt.mockUsecase)

			r.POST(
				path,
				h.CreateUser,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				path,
				testutil.NewJsonReader(t, tt.body),
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

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
				testutil.AssertResponse(t, rec, tt.expectedErrorBody)
				return
			}

			// 正常系の場合のassertion
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
		body              handlerDto.UpdateUserAgreementsRequest
		setUpContext      gin.HandlerFunc
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectedStatus    int
		expectError       bool
		expectedErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザ規約同意状況更新",
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
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "異常_ユーザ規約同意状況更新_ユーザID取得失敗",
			body: handlerDto.UpdateUserAgreementsRequest{},
			setUpContext: func(c *gin.Context) {
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    false,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal: 内部用UID取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_リクエスト形式不正",
			body: handlerDto.UpdateUserAgreementsRequest{},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled:  false,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0002",
				Message: "リクエストが不正です。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_リクエストが空の配列",
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{},
			},
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return nil
				},
			},
			usecaseCalled:  false,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0002",
				Message: "リクエストが不正です。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_ユーザ取得処理異常",
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
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrUserNotFound
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal:ユーザ取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_規約タイプ指定不正",
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
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrInvalidPolicyType
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0011",
				Message: "規約タイプが不正です。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_規約バージョン指定不正",
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
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return usecase.ErrInvalidPolicyVersion
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0012",
				Message: "規約バージョンが不正です。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況更新_予期せぬエラー",
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
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, uid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
					return errors.New("予期せぬエラー")
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal:予期せぬエラー",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := gin.New()
			if tt.setUpContext != nil {
				r.Use(tt.setUpContext)
			}
			h := NewUserHandler(tt.mockUsecase)

			r.PUT(
				path,
				h.UpdateUserAgreements,
			)

			req := httptest.NewRequest(
				http.MethodPut,
				path,
				testutil.NewJsonReader(t, tt.body),
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

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
				testutil.AssertResponse(t, rec, tt.expectedErrorBody)
			}
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
		setUpContext      gin.HandlerFunc
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectedStatus    int
		expectError       bool
		expectedErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザ退会",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
				c.Set(ctxutil.CtxKeyUid, internalUid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return nil
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "異常_ユーザ退会_InternalUIDの取得に失敗",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:ユーザID取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ退会_FirebaseUIDの取得に失敗",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, internalUid)
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "firebaseユーザID取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ退会_未入会ユーザ",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
				c.Set(ctxutil.CtxKeyUid, internalUid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return usecase.ErrUserNotFound
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "E0010",
				Message: "未登録ユーザです。ユーザ登録を行ってください。",
			},
		},
		{
			name: "異常_ユーザ退会_ユーザ情報に不整合あり",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
				c.Set(ctxutil.CtxKeyUid, internalUid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return usecase.ErrInvalidUser
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal: ユーザデータに不整合があります。",
			},
		},
		{
			name: "異常_ユーザ退会_予期しないエラー",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
				c.Set(ctxutil.CtxKeyUid, internalUid)
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				DeleteUserFunc: func(
					ctx context.Context,
					input usecaseDto.DeleteUserInput,
				) error {
					return errors.New("予期しないエラー")
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal: internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			if tt.setUpContext != nil {
				r.Use(tt.setUpContext)

			}
			h := NewUserHandler(tt.mockUsecase)

			r.DELETE(
				path,
				h.DeleteUser,
			)

			req := httptest.NewRequest(
				http.MethodDelete,
				path,
				nil,
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

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
				testutil.AssertResponse(t, rec, tt.expectedErrorBody)
				return
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	path := "/api/v0/users/login"
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		setUpContext    gin.HandlerFunc
		mockUsecase     *mock.MockUserUsecase
		usecaseCalled   bool
		expectedStatus  int
		expectedBody    handlerDto.UserLoginResponse
		expectError     bool
		expectErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザログイン",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1432))
				c.Set(ctxutil.CtxKeyPublicUid, "public_uid")
				c.Next()
			},
			expectedStatus: http.StatusOK,
			expectedBody: handlerdto.UserLoginResponse{
				PublicUserId: "public_uid",
			},
			expectError: false,
		},
		{
			name: "異常_ユーザログイン_公開用ユーザID取得失敗",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1432))
				c.Next()
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectErrorBody: handlerdto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal: 公開用ユーザID取得処理に異常があります。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			if tt.setUpContext != nil {
				r.Use(tt.setUpContext)
			}
			h := NewUserHandler(tt.mockUsecase)

			r.PUT(
				path,
				h.Login,
			)

			req := httptest.NewRequest(
				http.MethodPut,
				path,
				nil,
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			// 異常系の場合
			if tt.expectError {
				testutil.AssertResponse(t, rec, tt.expectErrorBody)
				return
			}

			// 正常系の場合
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}

func TestUserHandler_GetUserAgreements(t *testing.T) {
	path := "/api/v0/users/agreements"
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		setUpContext      gin.HandlerFunc
		mockUsecase       *mock.MockUserUsecase
		usecaseCalled     bool
		expectedStatus    int
		expectError       bool
		expectedBody      handlerDto.GetUserAgreementsResponse
		expectedErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザ規約同意状況参照",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, int64(1001))
				c.Next()
			},
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
			usecaseCalled:  true,
			expectedStatus: http.StatusOK,
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
			name: "異常_ユーザ規約同意状況参照_internal uid取得失敗",
			setUpContext: func(c *gin.Context) {
				c.Next()
			},
			mockUsecase:    &mock.MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:ユーザID取得処理に異常があります。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況参照_ユーザ非存在",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, int64(1001))
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
					return nil, usecase.ErrUserNotFound
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:ユーザマスタ情報に異常があります。",
			},
		},
		{
			name: "異常_ユーザ規約同意状況参照_ユーザ非存在",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyUid, int64(1001))
				c.Next()
			},
			mockUsecase: &mock.MockUserUsecase{
				GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
					return nil, errors.New("error")
				},
			},
			usecaseCalled:  true,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:予期せぬエラー",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := gin.New()
			if tt.setUpContext != nil {
				r.Use(tt.setUpContext)
			}
			h := NewUserHandler(tt.mockUsecase)

			r.GET(
				path,
				h.GetUserAgreements,
			)

			req := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

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
				testutil.AssertResponse(t, rec, tt.expectedErrorBody)
				return
			}

			// 正常系のチェック
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}
