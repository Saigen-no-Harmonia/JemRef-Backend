package handler

import (
	"context"
	"errors"
	"fmt"
	ctxutil "jemref_go/internal/context"
	handlerDto "jemref_go/internal/handler/dto"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_CreateUser(t *testing.T) {

	path := "/api/v0/join"
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		body              handlerDto.CreateUserRequest
		setUpContext      gin.HandlerFunc
		mockUsecase       *MockUserUsecase
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
			mockUsecase: &MockUserUsecase{
				createUserFunc: func(
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
			mockUsecase:    &MockUserUsecase{},
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
			mockUsecase:    &MockUserUsecase{},
			usecaseCalled:  false,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedErrorBody: handlerDto.ErrorResponse{
				Code:    "A0001",
				Message: "fatal:Eメールアドレスの取得処理に異常があります。",
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
			mockUsecase: &MockUserUsecase{
				createUserFunc: func(
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
			mockUsecase: &MockUserUsecase{
				createUserFunc: func(
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
				assert.True(t, tt.mockUsecase.createUserCalled)
			} else {
				assert.False(t, tt.mockUsecase.createUserCalled)
			}
			assert.False(t, tt.mockUsecase.deleteUserCalled)
			assert.False(t, tt.mockUsecase.loginCalled)

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

func TestUserHandler_DeleteUser(t *testing.T) {
	path := "/api/v0/users"
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		setUpContext      gin.HandlerFunc
		mockUsecase       *MockUserUsecase
		usecaseCalled     bool
		expectedStatus    int
		expectError       bool
		expectedErrorBody handlerDto.ErrorResponse
	}{
		{
			name: "正常_ユーザ退会",
			setUpContext: func(c *gin.Context) {
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1234))
				c.Next()
			},
			mockUsecase: &MockUserUsecase{
				deleteUserFunc: func(
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
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Next()
			},
			mockUsecase:    &MockUserUsecase{},
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
				c.Set(ctxutil.CtxKeyUid, int64(1234))
				c.Next()
			},
			mockUsecase:    &MockUserUsecase{},
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
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1234))
				c.Next()
			},
			mockUsecase: &MockUserUsecase{
				deleteUserFunc: func(
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
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1234))
				c.Next()
			},
			mockUsecase: &MockUserUsecase{
				deleteUserFunc: func(
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
				c.Set(ctxutil.CtxKeyFirebaseUid, "firebase_uid")
				c.Set(ctxutil.CtxKeyUid, int64(1234))
				c.Next()
			},
			mockUsecase: &MockUserUsecase{
				deleteUserFunc: func(
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
				assert.True(t, tt.mockUsecase.deleteUserCalled)
			} else {
				assert.False(t, tt.mockUsecase.deleteUserCalled)
			}
			assert.False(t, tt.mockUsecase.createUserCalled)
			assert.False(t, tt.mockUsecase.loginCalled)

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
		mockUsecase     *MockUserUsecase
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

type MockUserUsecase struct {
	createUserFunc func(
		context.Context,
		usecaseDto.CreateUserInput,
	) (*usecaseDto.CreateUserOutput, error)

	deleteUserFunc func(
		context.Context,
		usecaseDto.DeleteUserInput,
	) error

	loginFunc func(
		context.Context,
		usecaseDto.UserLoginInput,
	) (*usecaseDto.UserLoginOutput, error)

	createUserCalled bool
	deleteUserCalled bool
	loginCalled      bool
}

func (m *MockUserUsecase) Create(
	ctx context.Context,
	input usecaseDto.CreateUserInput,
) (*usecaseDto.CreateUserOutput, error) {
	m.createUserCalled = true
	return m.createUserFunc(ctx, input)
}

func (m *MockUserUsecase) Delete(
	ctx context.Context,
	input usecaseDto.DeleteUserInput,
) error {
	m.deleteUserCalled = true
	return m.deleteUserFunc(ctx, input)
}

func (m *MockUserUsecase) Login(
	ctx context.Context,
	input usecaseDto.UserLoginInput,
) (*usecaseDto.UserLoginOutput, error) {
	m.loginCalled = true
	return m.loginFunc(ctx, input)
}
