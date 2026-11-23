package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------Login(*gin.Context)のテスト-------------
//
// Loginはusecaseを呼ばず、contextの公開用ユーザIDをそのまま返す。

// serveLogin Loginを実行する
func serveLogin(uc *mock.MockUserUsecase, uctx userCtx) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodPut, "/api/v0/users/login", nil)
	return serveUser(uc, (*UserHandler).Login, req, uctx)
}

func TestUserHandler_Login_正常(t *testing.T) {
	t.Run("正常_ユーザログイン", func(t *testing.T) {
		uc := &mock.MockUserUsecase{}

		rec, c := serveLogin(uc, userCtx{})

		assertUserUsecaseCalled(t, uc, "Login", false)
		require.Len(t, c.Errors, 0)
		// usecaseを呼ばないためcontextの値を直接検証
		assert.Equal(t, ctxutil.MEM_API_007, c.Keys[ctxutil.CtxKeyApiId])
		testutil.AssertResponse(t, rec, handlerDto.UserLoginResponse{
			PublicUserId: userTestPublicUid,
		})
	})
}

func TestUserHandler_Login_異常(t *testing.T) {
	t.Run("異常_公開用ユーザID取得失敗", func(t *testing.T) {
		uc := &mock.MockUserUsecase{}

		_, c := serveLogin(uc, userCtx{omitPublicUid: true})

		assertHandlerError(t, c, ErrPublicUidNotFound)
	})
}
