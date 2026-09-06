package middleware

import (
	"context"
	"fmt"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_FirebaseAuth(t *testing.T) {
	path := "/api/v0/test"
	authorizationKey := "Authorization"
	email := "sample@example.com"
	firebaseUid := "firebase_uid"

	tests := []struct {
		name                   string
		authorizationKey       string
		authorizationBody      string
		mockFirebaseAuthClient *MockFirebaseAuthClient
		callNext               bool
		expectedEmail          string
		expectedFirebaseUid    string
		expectedError          bool
		expectedErrorBody      error
	}{
		{
			name:              "正常_認証成功",
			authorizationKey:  authorizationKey,
			authorizationBody: "Bearer sampletoken",
			mockFirebaseAuthClient: &MockFirebaseAuthClient{
				VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return &auth.Token{
						UID: firebaseUid,
						Claims: map[string]interface{}{
							"email": email,
						},
					}, nil
				},
			},
			callNext:            true,
			expectedEmail:       email,
			expectedFirebaseUid: firebaseUid,
			expectedError:       false,
		},
		{
			name:                   "異常_Authorizationヘッダなし",
			authorizationKey:       "Invalid",
			authorizationBody:      "Bearer sampletoken",
			mockFirebaseAuthClient: &MockFirebaseAuthClient{},
			callNext:               false,
			expectedError:          true,
			expectedErrorBody:      ErrRequireAuthHeader,
		},
		{
			name:                   "異常_Bearerヘッダ形式不正",
			authorizationKey:       authorizationKey,
			authorizationBody:      "invalid sampletoken",
			mockFirebaseAuthClient: &MockFirebaseAuthClient{},
			callNext:               false,
			expectedError:          true,
			expectedErrorBody:      ErrRequireBearerToken,
		},
		{
			name:              "異常_FirebaseToken不正",
			authorizationKey:  authorizationKey,
			authorizationBody: "Bearer sampletoken",
			mockFirebaseAuthClient: &MockFirebaseAuthClient{
				VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return nil, ErrTestUnexpected
				},
			},
			callNext:          false,
			expectedError:     true,
			expectedErrorBody: ErrInvalidFirebaseToken,
		},
		{
			name:              "異常_Claims中のメールアドレス取得失敗",
			authorizationKey:  authorizationKey,
			authorizationBody: "Bearer sampletoken",
			mockFirebaseAuthClient: &MockFirebaseAuthClient{
				VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return &auth.Token{
						UID:    firebaseUid,
						Claims: nil,
					}, nil
				},
			},
			callNext:          false,
			expectedError:     true,
			expectedErrorBody: ErrRequireEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			// エラー実際値取得用middlewareを先に仕込む
			var actualError error
			r.Use(func(c *gin.Context) {
				c.Next()
				actualError = getLastGinError(c)
			})

			// テスト対象
			r.Use(FirebaseAuth(tt.mockFirebaseAuthClient))

			// 実際値取得用に仮Handlerを仕込む
			var actualEmail string
			var actualFirebaseUid string
			handlerCalled := false
			r.GET(path, func(c *gin.Context) {
				handlerCalled = true

				email, ok := c.Get(ctxutil.CtxKeyEmail)
				if ok {
					actualEmail, _ = email.(string)
				}

				firebaseUid, ok := c.Get(ctxutil.CtxKeyFirebaseUid)
				if ok {
					actualFirebaseUid, _ = firebaseUid.(string)
				}

				c.Status(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)
			req.Header.Set(tt.authorizationKey, tt.authorizationBody)

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.callNext, handlerCalled)

			// 異常系の検証
			if tt.expectedError {
				require.Error(t, actualError)
				assert.ErrorIs(t, actualError, tt.expectedErrorBody)
				return
			}
			// 正常系の検証
			assert.NoError(t, actualError)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.authorizationBody, "Bearer "+tt.mockFirebaseAuthClient.LastVerifiedIdToken)
			assert.Equal(t, tt.expectedEmail, actualEmail)
			assert.Equal(t, tt.expectedFirebaseUid, actualFirebaseUid)
		})
	}
}

func TestMiddleware_ChkUnregistered(t *testing.T) {
	path := "/api/v0/test"
	firebaseUid := "firebase_uid"

	tests := []struct {
		name              string
		ctxFirebaseUid    string
		mockAuthUsecase   *mock.MockAuthUsecase
		callNext          bool
		expectedError     bool
		expectedErrorBody error
	}{
		{
			name:           "正常_ユーザ未登録チェック",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return nil, usecase.ErrUserNotFound
				},
			},
			callNext:      true,
			expectedError: false,
		},
		{
			name:              "異常_Firebase uidなし",
			ctxFirebaseUid:    "",
			mockAuthUsecase:   &mock.MockAuthUsecase{},
			callNext:          false,
			expectedError:     true,
			expectedErrorBody: ErrRequireFirebaseUid,
		},
		{
			name:           "異常_ユーザ登録済み",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return &usecaseDto.AuthUserOutput{}, nil
				},
			},
			callNext:          false,
			expectedError:     true,
			expectedErrorBody: ErrUserAlreadyExists,
		},
		{
			name:           "異常_意図しないエラー",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return nil, ErrTestUnexpected
				},
			},
			callNext:          false,
			expectedError:     true,
			expectedErrorBody: ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			var actualError error
			// テスト用初期設定
			setErrorCatcher(r, &actualError)
			setCtxFirebaseUid(r, tt.ctxFirebaseUid)

			// テスト対象
			r.Use(ChkUnregistered(tt.mockAuthUsecase))

			// 後続で実行される仮Handler
			handlerCalled := false
			r.GET(path, func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})

			// テスト実行
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(rec, req)

			// 検証
			assert.Equal(t, tt.callNext, handlerCalled)

			// 異常系
			if tt.expectedError {
				require.Error(t, actualError)
				assert.ErrorIs(t, actualError, tt.expectedErrorBody)
				return
			}

			// 正常系
			assert.NoError(t, actualError)
			assert.Equal(t, http.StatusOK, rec.Code)

		})
	}
}

func TestMiddleware_FindCurrentUser(t *testing.T) {
	path := "/api/v0/test"
	firebaseUid := "firebase_uid"
	internalUid := int64(10010)
	publicUid := "public_uid"
	tests := []struct {
		name               string
		mockAuthUsecase    *mock.MockAuthUsecase
		ctxFirebaseUid     string
		callNext           bool
		authenticateCalled int
		DeleteUserCalled   int
		expectError        bool
		expectErrorBody    error
	}{
		{
			name:           "正常_ユーザ存在",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return &usecaseDto.AuthUserOutput{
						InternalUserId: internalUid,
						FirebaseUserId: firebaseUid,
						PublicUserId:   publicUid,
					}, nil
				},
				DeleteUserFunc: func(c context.Context, firebaseUid string) error {
					return nil
				},
			},
			authenticateCalled: 1,
			DeleteUserCalled:   0,
			callNext:           true,
			expectError:        false,
		},
		{
			name:               "異常_ctxにfirebaseUidなし",
			ctxFirebaseUid:     "",
			mockAuthUsecase:    &mock.MockAuthUsecase{},
			authenticateCalled: 0,
			DeleteUserCalled:   0,
			callNext:           false,
			expectError:        true,
			expectErrorBody:    ErrRequireFirebaseUid,
		},
		{
			name:           "異常_ユーザ退会済み",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return nil, usecase.ErrUserDeleted
				},
				DeleteUserFunc: func(c context.Context, firebaseUid string) error {
					return nil
				},
			},
			authenticateCalled: 1,
			DeleteUserCalled:   1,
			callNext:           false,
			expectError:        true,
			expectErrorBody:    ErrUserDeleted,
		},
		{
			name:           "異常_予期しないエラー",
			ctxFirebaseUid: firebaseUid,
			mockAuthUsecase: &mock.MockAuthUsecase{
				AuthenticateFunc: func(c context.Context, firebaseUid string) (*usecaseDto.AuthUserOutput, error) {
					return nil, ErrTestUnexpected
				},
				DeleteUserFunc: func(c context.Context, firebaseUid string) error {
					return nil
				},
			},
			authenticateCalled: 1,
			DeleteUserCalled:   0,
			callNext:           false,
			expectError:        true,
			expectErrorBody:    ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			var actualError error
			// テスト用初期設定
			setErrorCatcher(r, &actualError)
			setCtxFirebaseUid(r, tt.ctxFirebaseUid)

			// テスト対象
			r.Use(FindCurrentUser(tt.mockAuthUsecase))

			handlerCalled := false
			var actualInternalUid int64
			var actualPublicUid string
			// 後続で実行される仮Handler
			r.GET(path, func(c *gin.Context) {
				handlerCalled = true

				iu, ok := c.Get(ctxutil.CtxKeyUid)
				if ok {
					actualInternalUid, _ = iu.(int64)
				}

				pu, ok := c.Get(ctxutil.CtxKeyPublicUid)
				if ok {
					actualPublicUid = pu.(string)
				}

				c.Status(http.StatusOK)
			})

			// テスト実行
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(rec, req)

			// 検証
			assert.Equal(t, tt.callNext, handlerCalled)
			assert.Equal(t, tt.mockAuthUsecase.AuthenticateCalled, tt.authenticateCalled)
			assert.Equal(t, tt.mockAuthUsecase.DeleteUserCalled, tt.DeleteUserCalled)

			// 異常系
			if tt.expectError {
				require.Error(t, actualError)
				assert.ErrorIs(t, actualError, tt.expectErrorBody)
				return
			}

			// 正常系
			assert.NoError(t, actualError)
			assert.Equal(t, rec.Code, http.StatusOK)
			assert.Equal(t, firebaseUid, tt.mockAuthUsecase.LastAuthFirebaseUid)
			assert.Equal(t, actualInternalUid, internalUid)
			assert.Equal(t, actualPublicUid, publicUid)
		})
	}
}

type MockFirebaseAuthClient struct {
	VerifyIDTokenFunc   func(ctx context.Context, idToken string) (*auth.Token, error)
	Called              int
	VerifyIDTokenCalled int
	LastVerifiedIdToken string
}

func (m *MockFirebaseAuthClient) VerifyIDToken(
	ctx context.Context,
	idToken string,
) (*auth.Token, error) {
	m.Called++
	m.VerifyIDTokenCalled++
	m.LastVerifiedIdToken = idToken
	return m.VerifyIDTokenFunc(ctx, idToken)
}

func setErrorCatcher(r *gin.Engine, err *error) {
	r.Use(func(c *gin.Context) {
		c.Next()
		*err = getLastGinError(c)
	})
}

func getLastGinError(c *gin.Context) error {
	if len(c.Errors) == 0 {
		fmt.Println("errorなし")
		return nil
	}
	return c.Errors.Last().Err
}

func setCtxFirebaseUid(r *gin.Engine, firebaseUid string) {
	r.Use(func(c *gin.Context) {
		c.Set(ctxutil.CtxKeyFirebaseUid, firebaseUid)
	})
}
