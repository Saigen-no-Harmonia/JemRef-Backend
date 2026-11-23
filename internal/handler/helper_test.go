package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------handlerテスト共通のヘルパ-------------

// newTestContext リクエストとcontextの値からgin.Contextを組み立てる
//
// values に含めないキーは「ミドルウェアが値を設定しなかった」状態になる。
// params が nil の場合はパスパラメータを設定しない。
func newTestContext(
	req *http.Request,
	values map[string]any,
	params gin.Params,
) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	for k, v := range values {
		c.Set(k, v)
	}
	if params != nil {
		c.Params = params
	}
	c.Request = req

	return rec, c
}

// newJsonRequest JSONボディを持つリクエストを返す
//
// body は生のJSON文字列で渡す(json tagとbindingの挙動をそのまま検証するため)。
func newJsonRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// assertHandlerError handlerがエラーを1件だけ c.Errors に積んだことを検証する
func assertHandlerError(t *testing.T, c *gin.Context, expected error) {
	t.Helper()

	require.Len(t, c.Errors, 1)
	assert.ErrorIs(t, c.Errors.Last().Err, expected)
}
