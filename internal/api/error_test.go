package api

import (
	"net/http"
	"testing"

	"jemref_go/internal/apperr"

	"github.com/stretchr/testify/assert"
)

// -------------kindToStatus(apperr.Kind)のテスト-------------

// TestKindToStatus
//
// 台帳の各定義がどのステータスになるかは「Def.Kind」と「この対応表」の組み合わせで決まる。
// Kindごとの対応はここで固定し、個々のDefのKindが妥当かは台帳側の不変条件テスト
// (apperr.TestAll_台帳の不変条件)とレビューで担保する。
func TestKindToStatus(t *testing.T) {

	tests := []struct {
		name     string
		kind     apperr.Kind
		expected int
	}{
		{
			name:     "正常_BadRequest",
			kind:     apperr.KindBadRequest,
			expected: http.StatusBadRequest,
		},
		{
			name:     "正常_Unauthorized",
			kind:     apperr.KindUnauthorized,
			expected: http.StatusUnauthorized,
		},
		{
			name:     "正常_NotFound",
			kind:     apperr.KindNotFound,
			expected: http.StatusNotFound,
		},
		{
			name:     "正常_Conflict",
			kind:     apperr.KindConflict,
			expected: http.StatusConflict,
		},
		{
			name:     "正常_Internal",
			kind:     apperr.KindInternal,
			expected: http.StatusInternalServerError,
		},
		{
			// Kindの指定漏れ。台帳側のテストで検出するが、実行時は安全側に倒す
			name:     "異常_未設定のKind",
			kind:     apperr.KindUnspecified,
			expected: http.StatusInternalServerError,
		},
		{
			name:     "異常_未知のKind",
			kind:     apperr.Kind(99),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, kindToStatus(tt.kind))
		})
	}
}

// -------------fromDef(apperr.Def)のテスト-------------

// TestFromDef 台帳の定義がそのままレスポンス用構造体へ載ることを確認する
//
// 個々のコードごとのケースは作らない(台帳を期待値の出所にすると検証にならないため)。
func TestFromDef(t *testing.T) {

	tests := []struct {
		name     string
		def      apperr.Def
		expected ApiError
	}{
		{
			name:     "正常_Kindに応じたステータスとコード・メッセージが載る",
			def:      apperr.Def{Kind: apperr.KindNotFound, Code: "T0001", Message: "テスト用のメッセージ"},
			expected: ApiError{http.StatusNotFound, "T0001", "テスト用のメッセージ"},
		},
		{
			name:     "異常_Kindが未設定の定義は500になる",
			def:      apperr.Def{Code: "T0002", Message: "テスト用のメッセージ"},
			expected: ApiError{http.StatusInternalServerError, "T0002", "テスト用のメッセージ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, fromDef(tt.def))
		})
	}
}
