package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------CreateUser(*gin.Context)のテスト-------------

// newCreateUserUsecaseMock 指定したエラーを返すMockUserUsecaseを返す(nilなら公開用ユーザIDを返して成功)
func newCreateUserUsecaseMock(err error) *mock.MockUserUsecase {
	return &mock.MockUserUsecase{
		CreateUserFunc: func(ctx context.Context, input usecaseDto.CreateUserInput) (*usecaseDto.CreateUserOutput, error) {
			if err != nil {
				return nil, err
			}
			return &usecaseDto.CreateUserOutput{PublicUserId: "_test_public_userid_"}, nil
		},
	}
}

// serveCreateUser CreateUserを実行する
func serveCreateUser(
	t *testing.T,
	uc *mock.MockUserUsecase,
	body handlerDto.CreateUserRequest,
	uctx userCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/join", testutil.NewJsonReader(t, body))
	return serveUser(uc, (*UserHandler).CreateUser, req, uctx)
}

// assertCreateUserInput リクエストの同意バージョンがそのままusecaseへ渡ったこと
func assertCreateUserInput(t *testing.T, uc *mock.MockUserUsecase, body handlerDto.CreateUserRequest) {
	t.Helper()

	assert.Equal(t, body.TermsAgreedVersion, uc.LastCreateUserInput.TermsAgreedVersion)
	assert.Equal(t, body.PrivacyPolicyAgreedVersion, uc.LastCreateUserInput.PrivacyPolicyAgreedVersion)
}

func TestUserHandler_CreateUser_正常(t *testing.T) {
	t.Run("正常_ユーザ登録", func(t *testing.T) {
		uc := newCreateUserUsecaseMock(nil)
		body := handlerDto.CreateUserRequest{
			TermsAgreedVersion:         "0.1",
			PrivacyPolicyAgreedVersion: "0.1",
		}

		rec, c := serveCreateUser(t, uc, body, userCtx{})

		assertUserUsecaseCalled(t, uc, "CreateUser", true)
		assertCreateUserInput(t, uc, body)

		require.Len(t, c.Errors, 0)
		assert.Equal(t, ctxutil.MEM_API_002, ctxutil.GetApiId(uc.LastCtx))
		testutil.AssertResponse(t, rec, handlerDto.CreateUserResponse{
			PublicUserId: "_test_public_userid_",
		})
	})
}

// TestUserHandler_CreateUser_異常 contextの値やリクエストボディが不正なら、usecaseを呼ばずにエラーを返すこと
func TestUserHandler_CreateUser_異常(t *testing.T) {

	tests := []struct {
		name              string
		ctx               userCtx
		body              handlerDto.CreateUserRequest
		expectedErrorBody error
	}{
		{
			name: "異常_FirebaseUID取得失敗",
			ctx:  userCtx{omitFirebaseUid: true},
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			expectedErrorBody: ErrFirebaseUidNotFound,
		},
		{
			name: "異常_Email取得失敗",
			ctx:  userCtx{omitEmail: true},
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			expectedErrorBody: ErrEmailNotFound,
		},
		{
			name:              "異常_規約同意バージョン未指定",
			body:              handlerDto.CreateUserRequest{},
			expectedErrorBody: ErrAgreedVersionRequired,
		},
		{
			name: "異常_プライバシーポリシー同意バージョン未指定",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion: "0.1",
			},
			expectedErrorBody: ErrAgreedVersionRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newCreateUserUsecaseMock(nil)

			_, c := serveCreateUser(t, uc, tt.body, tt.ctx)

			assertUserUsecaseCalled(t, uc, "CreateUser", false)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestUserHandler_CreateUser_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestUserHandler_CreateUser_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		body       handlerDto.CreateUserRequest
		usecaseErr error
	}{
		{
			name: "異常_既存ユーザあり",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			usecaseErr: usecase.ErrUserAlreadyExists,
		},
		{
			name: "異常_規約指定不正",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.3",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			usecaseErr: usecase.ErrPolicyNotFound,
		},
		{
			name: "異常_予期しないエラー",
			body: handlerDto.CreateUserRequest{
				TermsAgreedVersion:         "0.1",
				PrivacyPolicyAgreedVersion: "0.1",
			},
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newCreateUserUsecaseMock(tt.usecaseErr)

			_, c := serveCreateUser(t, uc, tt.body, userCtx{})

			assertUserUsecaseCalled(t, uc, "CreateUser", true)
			assertCreateUserInput(t, uc, tt.body)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
