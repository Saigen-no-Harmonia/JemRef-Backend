package api

import (
	"net/http"

	"jemref_go/internal/apperr"
)

// ApiError エラーレスポンスの構成要素
//
// コードと公開メッセージの定義は internal/apperr の台帳が持つ。
// この層の責務は Kind をHTTPステータスへ対応づけることだけ。
type ApiError struct {
	StatusCode int
	Code       string
	Message    string
}

// fromDef apperrの定義をレスポンス用の構造体へ変換する
func fromDef(d apperr.Def) ApiError {
	return ApiError{
		StatusCode: kindToStatus(d.Kind),
		Code:       d.Code,
		Message:    d.Message,
	}
}

// kindToStatus エラーの意味的な分類をHTTPステータスへ対応づける
//
// HTTPを知るのはこの層だけ。未知のKindは安全側に倒して500とする。
func kindToStatus(k apperr.Kind) int {
	switch k {
	case apperr.KindBadRequest:
		return http.StatusBadRequest

	case apperr.KindUnauthorized:
		return http.StatusUnauthorized

	case apperr.KindNotFound:
		return http.StatusNotFound

	case apperr.KindConflict:
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
