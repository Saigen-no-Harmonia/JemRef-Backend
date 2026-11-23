package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/testutil/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// UserHandlerのテストは、メソッドごとに user_handler_<メソッド>_test.go に置く。
// このファイルには、それらが共有する定数とヘルパを置く。

const (
	// userTestInternalUid テスト用の内部ユーザID
	userTestInternalUid = int64(2001)
	// userTestPublicUid テスト用の公開用ユーザID
	userTestPublicUid = "public_uid"
	// userTestFirebaseUid テスト用のFirebase UID
	userTestFirebaseUid = "firebase_uid"
	// userTestEmail テスト用のメールアドレス
	userTestEmail = "test@example.com"
)

// userCtx ユーザAPIのhandlerに渡すcontextの状態
//
// ゼロ値は、認証ミドルウェアが値をすべて設定した正常な状態を表す。
// 取得失敗のケースだけ omitXxx を立てる。
type userCtx struct {
	omitInternalUid bool
	omitPublicUid   bool
	omitFirebaseUid bool
	omitEmail       bool
}

// serveUser userCtxに従ってgin.Contextを組み立て、handlerを実行する
func serveUser(
	uc *mock.MockUserUsecase,
	serve func(h *UserHandler, c *gin.Context),
	req *http.Request,
	uctx userCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	values := map[string]any{}
	if !uctx.omitInternalUid {
		values[ctxutil.CtxKeyUid] = userTestInternalUid
	}
	if !uctx.omitPublicUid {
		values[ctxutil.CtxKeyPublicUid] = userTestPublicUid
	}
	if !uctx.omitFirebaseUid {
		values[ctxutil.CtxKeyFirebaseUid] = userTestFirebaseUid
	}
	if !uctx.omitEmail {
		values[ctxutil.CtxKeyEmail] = userTestEmail
	}

	rec, c := newTestContext(req, values, nil)
	serve(NewUserHandler(uc), c)
	return rec, c
}

// assertUserUsecaseCalled 指定したメソッドだけが呼ばれ(called=falseなら何も呼ばれず)、他のメソッドが呼ばれていないこと
func assertUserUsecaseCalled(t *testing.T, uc *mock.MockUserUsecase, method string, called bool) {
	t.Helper()

	actual := map[string]bool{
		"CreateUser":           uc.CreateUserCalled,
		"DeleteUser":           uc.DeleteUserCalled,
		"GetUserAgreements":    uc.GetUserAgreementsCalled,
		"UpdateUserAgreements": uc.UpdateUserAgreementsCalled,
		"Login":                uc.LoginCalled,
	}
	for name, got := range actual {
		assert.Equal(t, name == method && called, got, "%s の呼び出し有無", name)
	}
}

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
