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

// RecordHandlerのテストは、メソッドごとに record_handler_<メソッド>_test.go に置く。
// このファイルには、それらが共有する定数とヘルパを置く。

const (
	// recordTestInternalUid テスト用の内部ユーザID
	recordTestInternalUid = int64(1001)
	// recordTestPublicUid テスト用の公開用ユーザID
	recordTestPublicUid = "PUBLICUID00000000000000001"
	// recordTestRecordId usecaseが返す書誌ID
	recordTestRecordId = "RECORDID000000000000000001"
)

// recordCtx 書誌APIのhandlerに渡すcontextの状態
//
// ゼロ値は、認証ミドルウェアが値をすべて設定した正常な状態を表す。
// 取得失敗のケースだけ omitXxx を立てる。
type recordCtx struct {
	omitInternalUid bool
	omitPublicUid   bool
}

// serveRecord recordCtxに従ってgin.Contextを組み立て、handlerを実行する
//
// recordId はパスパラメータ id に設定する。空文字の場合はパラメータを設定しない。
func serveRecord(
	uc *mock.MockRecordUsecase,
	serve func(h *RecordHandler, c *gin.Context),
	req *http.Request,
	recordId string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	values := map[string]any{}
	if !rc.omitInternalUid {
		values[ctxutil.CtxKeyUid] = recordTestInternalUid
	}
	if !rc.omitPublicUid {
		values[ctxutil.CtxKeyPublicUid] = recordTestPublicUid
	}

	var params gin.Params
	if recordId != "" {
		params = gin.Params{{Key: ParamRecordId, Value: recordId}}
	}

	rec, c := newTestContext(req, values, params)
	serve(NewRecordHandler(uc), c)
	return rec, c
}

func TestRecordHandler_RegisterRoutes(t *testing.T) {
	r := gin.New()
	group := r.Group("/api/v0")

	h := NewRecordHandler(nil)
	h.RegisterRoutes(group)

	routes := r.Routes()

	expected := map[string]bool{
		"GET /api/v0/records":        false,
		"GET /api/v0/records/:id":    false,
		"POST /api/v0/records":       false,
		"PUT /api/v0/records/:id":    false,
		"DELETE /api/v0/records/:id": false,
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
