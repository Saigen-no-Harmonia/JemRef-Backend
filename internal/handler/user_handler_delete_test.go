package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------DeleteUser(*gin.Context)のテスト-------------

// newDeleteUserUsecaseMock 指定したエラーを返すMockUserUsecaseを返す(nilなら成功)
func newDeleteUserUsecaseMock(err error) *mock.MockUserUsecase {
	return &mock.MockUserUsecase{
		DeleteUserFunc: func(ctx context.Context, input usecaseDto.DeleteUserInput) error {
			return err
		},
	}
}

// serveDeleteUser DeleteUserを実行する
func serveDeleteUser(uc *mock.MockUserUsecase, uctx userCtx) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodDelete, "/api/v0/users", nil)
	return serveUser(uc, (*UserHandler).DeleteUser, req, uctx)
}

// assertDeleteUserInput contextの内部ユーザIDとFirebase UIDがそのままusecaseへ渡ったこと
func assertDeleteUserInput(t *testing.T, uc *mock.MockUserUsecase) {
	t.Helper()

	assert.Equal(t, userTestInternalUid, uc.LastDeleteUserInput.InternalUserid)
	assert.Equal(t, userTestFirebaseUid, uc.LastDeleteUserInput.FirebaseUserId)
}

func TestUserHandler_DeleteUser_正常(t *testing.T) {
	t.Run("正常_ユーザ退会", func(t *testing.T) {
		uc := newDeleteUserUsecaseMock(nil)

		_, c := serveDeleteUser(uc, userCtx{})

		assertUserUsecaseCalled(t, uc, "DeleteUser", true)
		assertDeleteUserInput(t, uc)

		require.Len(t, c.Errors, 0)
		assert.Equal(t, ctxutil.MEM_API_004, ctxutil.GetApiId(uc.LastCtx))
	})
}

// TestUserHandler_DeleteUser_異常 contextの値が取得できなければ、usecaseを呼ばずにエラーを返すこと
func TestUserHandler_DeleteUser_異常(t *testing.T) {

	tests := []struct {
		name              string
		ctx               userCtx
		expectedErrorBody error
	}{
		{
			name:              "異常_InternalUIDの取得に失敗",
			ctx:               userCtx{omitInternalUid: true},
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_FirebaseUIDの取得に失敗",
			ctx:               userCtx{omitFirebaseUid: true},
			expectedErrorBody: ErrFirebaseUidNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newDeleteUserUsecaseMock(nil)

			_, c := serveDeleteUser(uc, tt.ctx)

			assertUserUsecaseCalled(t, uc, "DeleteUser", false)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestUserHandler_DeleteUser_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestUserHandler_DeleteUser_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		usecaseErr error
	}{
		{
			name:       "異常_未入会ユーザ",
			usecaseErr: usecase.ErrUserNotFound,
		},
		{
			name:       "異常_ユーザ情報に不整合あり",
			usecaseErr: usecase.ErrUserDataInconsistent,
		},
		{
			name:       "異常_予期しないエラー",
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newDeleteUserUsecaseMock(tt.usecaseErr)

			_, c := serveDeleteUser(uc, userCtx{})

			assertUserUsecaseCalled(t, uc, "DeleteUser", true)
			assertDeleteUserInput(t, uc)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
