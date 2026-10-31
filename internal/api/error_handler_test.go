package api

import (
	"errors"
	"fmt"
	"jemref_go/internal/handler"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/middleware"
	"jemref_go/internal/testutil"
	"jemref_go/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errApiTestUnregistered toApiErrorに登録されていないエラー
var errApiTestUnregistered = errors.New("api test unregistered error")

func TestErrorHandler(t *testing.T) {

	path := "/api/v0/test"

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
			name:          "異常_middleware層のエラーを変換する",
			handlerErrs:   []error{middleware.ErrInvalidFirebaseToken},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrUnAuthorized.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrUnAuthorized.Code,
				Message: ErrUnAuthorized.Message,
			},
		},
		{
			name:          "異常_handler層のエラーを変換する",
			handlerErrs:   []error{handler.ErrMainTitleRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrMainTitleRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrMainTitleRequired.Code,
				Message: ErrMainTitleRequired.Message,
			},
		},
		{
			name:          "異常_usecase層のエラーを変換する",
			handlerErrs:   []error{usecase.ErrReachRegistrationLimit},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrReachRegistrationLimit.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrReachRegistrationLimit.Code,
				Message: ErrReachRegistrationLimit.Message,
			},
		},
		{
			name:          "異常_ラップされたエラーも変換する",
			handlerErrs:   []error{fmt.Errorf("書誌タイプ=foo: %w", handler.ErrInvalidRecordType)},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrInvalidRecordType.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrInvalidRecordType.Code,
				Message: ErrInvalidRecordType.Message,
			},
		},
		{
			name: "異常_複数エラーがある場合は最後のエラーを変換する",
			handlerErrs: []error{
				handler.ErrMainTitleRequired,
				usecase.ErrUserNotFound,
			},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrUserNotFound.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrUserNotFound.Code,
				Message: ErrUserNotFound.Message,
			},
		},
		{
			name:          "異常_規約情報必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrPolicyRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrPolicyRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrPolicyRequired.Code,
				Message: ErrPolicyRequired.Message,
			},
		},
		{
			name:          "異常_規約同意バージョン必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrAgreedVersionRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrAgreedVersionRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrAgreedVersionRequired.Code,
				Message: ErrAgreedVersionRequired.Message,
			},
		},
		{
			name:          "異常_規約タイプ不正エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrPolicyTypeInvalid},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrInvalidPolicyType.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrInvalidPolicyType.Code,
				Message: ErrInvalidPolicyType.Message,
			},
		},
		{
			name:          "異常_貢献者氏名必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrContributorNameRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrContributorNameRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrContributorNameRequired.Code,
				Message: ErrContributorNameRequired.Message,
			},
		},
		{
			name:          "異常_貢献者役割必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrContributorRoleRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrContributorRoleRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrContributorRoleRequired.Code,
				Message: ErrContributorRoleRequired.Message,
			},
		},
		{
			name:          "異常_メモ内容必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrMemoContentRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrMemoContentRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrMemoContentRequired.Code,
				Message: ErrMemoContentRequired.Message,
			},
		},
		{
			name:          "異常_URL値必須エラーを専用コードに変換する",
			handlerErrs:   []error{handler.ErrUrlValueRequired},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrUrlValueRequired.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrUrlValueRequired.Code,
				Message: ErrUrlValueRequired.Message,
			},
		},
		{
			name:          "異常_未登録のエラーは500に変換する",
			handlerErrs:   []error{errApiTestUnregistered},
			handlerStatus: http.StatusOK,
			expectedCode:  ErrInternal.StatusCode,
			expectedBody: handlerDto.ErrorResponse{
				Code:    ErrInternal.Code,
				Message: ErrInternal.Message,
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
		})
	}
}

// -------------toApiError(error)のテスト-------------

// 各レイヤーの代表的なエラーと、ラップ・未登録の扱いのみを確認する。
// 個々のsentinel errorとApiErrorの対応はErrorHandler経由のテストで担保する。
func TestToApiError(t *testing.T) {

	tests := []struct {
		name     string
		input    error
		expected ApiError
	}{
		{
			name:     "正常_middleware層のエラー",
			input:    middleware.ErrRequireAuthHeader,
			expected: ErrUnAuthorized,
		},
		{
			name:     "正常_handler層のエラー",
			input:    handler.ErrInvalidRequestBody,
			expected: ErrBadRequest,
		},
		{
			name:     "正常_usecase層のエラー",
			input:    usecase.ErrPolicyNotFound,
			expected: ErrPolicyNotFound,
		},
		{
			name:     "正常_ラップされたエラー",
			input:    fmt.Errorf("階層1: %w", fmt.Errorf("階層2: %w", handler.ErrInvalidReadStatus)),
			expected: ErrInvalidReadStatus,
		},
		{
			name:     "異常_未登録のエラーは予期せぬエラーになる",
			input:    errApiTestUnregistered,
			expected: ErrInternal,
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
