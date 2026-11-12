package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"jemref_go/internal/apperr"
	"jemref_go/internal/handler"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/middleware"
	"jemref_go/internal/testutil"
	"jemref_go/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errApiTestPlain apperr.Errorではないエラー(想定外の失敗を表す)
var errApiTestPlain = errors.New("api test plain error")

// -------------ErrorHandler()のテスト-------------

// TestErrorHandler
//
// 検証対象はミドルウェアの振る舞い(エラーの有無の判定、連鎖の辿り方、
// 想定外エラーの隠蔽、ボディの書き込み)であり、個々のエラーコードの対応ではない。
// toApiErrorに変換表はなく、どのエラーでも同一の経路を通るため、
// エラーを追加してもこのテストは修正不要。
func TestErrorHandler(t *testing.T) {

	path := "/api/v0/test"

	// 代表として1件だけ使う。期待値は台帳から組み立てる
	wantRecordNotFound := fromDef(apperr.RecordNotFound)
	wantInternal := fromDef(apperr.Internal)

	tests := []struct {
		name string
		// handlerErrs ハンドラがc.Errorで積むエラー
		handlerErrs []error
		// handlerStatus ハンドラが設定するステータスコード
		handlerStatus int
		expectedCode  int
		expectedBody  handlerDto.ErrorResponse
		// expectNoBody レスポンスボディが書き込まれないことを期待する
		expectNoBody bool
	}{
		{
			name:          "正常_エラーなしの場合はレスポンスを上書きしない",
			handlerErrs:   nil,
			handlerStatus: http.StatusOK,
			expectedCode:  http.StatusOK,
			expectNoBody:  true,
		},
		{
			name:          "異常_apperrのエラーを台帳どおりのレスポンスへ変換する",
			handlerErrs:   []error{usecase.ErrRecordNotFound},
			handlerStatus: http.StatusOK,
			expectedCode:  wantRecordNotFound.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    wantRecordNotFound.Code,
				Message: wantRecordNotFound.Message,
			},
		},
		{
			// 呼び出し側が文脈を付けてラップしても、連鎖を辿って変換できること
			name: "異常_多段にラップされたエラーも変換する",
			handlerErrs: []error{
				fmt.Errorf("階層1: %w", fmt.Errorf("階層2: %w", usecase.ErrRecordNotFound)),
			},
			handlerStatus: http.StatusOK,
			expectedCode:  wantRecordNotFound.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    wantRecordNotFound.Code,
				Message: wantRecordNotFound.Message,
			},
		},
		{
			name: "異常_複数エラーがある場合は最後のエラーを変換する",
			handlerErrs: []error{
				middleware.ErrInvalidFirebaseToken,
				handler.ErrMainTitleRequired,
				usecase.ErrRecordNotFound,
			},
			handlerStatus: http.StatusOK,
			expectedCode:  wantRecordNotFound.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    wantRecordNotFound.Code,
				Message: wantRecordNotFound.Message,
			},
		},
		{
			// 想定外の失敗は内容を漏らさず500に寄せる
			name:          "異常_apperr以外のエラーは内容を漏らさず500にする",
			handlerErrs:   []error{errApiTestPlain},
			handlerStatus: http.StatusOK,
			expectedCode:  wantInternal.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    wantInternal.Code,
				Message: wantInternal.Message,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(ErrorHandler())

			router.GET(path, func(c *gin.Context) {
				for _, err := range tt.handlerErrs {
					c.Error(err)
				}
				c.Status(tt.handlerStatus)
			})

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectNoBody {
				assert.Empty(t, rec.Body.String())
				return
			}

			testutil.AssertResponse(t, rec, tt.expectedBody)

			// 想定外エラーのログ用文言がレスポンスへ漏れていないこと
			assert.NotContains(t, rec.Body.String(), errApiTestPlain.Error())
		})
	}
}

// -------------toApiError(error)のテスト-------------

// TestToApiError 各層のsentinelが台帳に繋がっていることを層ごとに1件だけ確認する
//
// 個々のエラーごとのケースは作らない(変換表がなく、どのエラーでも同じ経路を通るため)。
func TestToApiError(t *testing.T) {

	tests := []struct {
		name     string
		input    error
		expected ApiError
	}{
		{
			name:     "正常_middleware層のエラー",
			input:    middleware.ErrRequireAuthHeader,
			expected: fromDef(apperr.Unauthorized),
		},
		{
			name:     "正常_handler層のエラー",
			input:    handler.ErrInvalidRequestBody,
			expected: fromDef(apperr.BadRequest),
		},
		{
			name:     "正常_usecase層のエラー",
			input:    usecase.ErrPolicyNotFound,
			expected: fromDef(apperr.PolicyNotFound),
		},
		{
			name:     "異常_apperr以外のエラーは予期せぬエラーになる",
			input:    errApiTestPlain,
			expected: fromDef(apperr.Internal),
		},
		{
			// %v でラップすると連鎖に入らないため変換できない(500へ寄せる)
			name:     "異常_文字列化されたエラーは変換できない",
			input:    fmt.Errorf("文脈: %v", usecase.ErrRecordNotFound),
			expected: fromDef(apperr.Internal),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toApiError(tt.input)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestErrorHandler_ハンドラのレスポンスボディは上書きされない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := "/api/v0/test"

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET(path, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"result": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"result":"ok"}`, rec.Body.String())
}
