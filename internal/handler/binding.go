package handler

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// toBindingError ShouldBindJSON等のbindエラーを、対応するsentinel errorへ変換する
//
// 必須制約の宣言はリクエストDTOの binding:"required" タグが唯一の正であり、
// requiredErrorsは必須エラーとエラーコードの対応表にすぎない。
//
// requiredErrorsのキーは構造体のフィールド名(validator.FieldError.StructField())。
// 埋め込み構造体をまたいで同名フィールドを区別する必要が出た場合は、
// Namespace()で判別する実装に切り替える。
func toBindingError(err error, requiredErrors map[string]error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fieldErr := range validationErrs {
			if fieldErr.Tag() != "required" {
				continue
			}

			if sentinel, ok := requiredErrors[fieldErr.StructField()]; ok {
				// Namespaceはログ用。レスポンスのメッセージはApiError側が持つ
				return fmt.Errorf("%s: %w", fieldErr.Namespace(), sentinel)
			}
		}
	}

	return fmt.Errorf("%w: %v", ErrInvalidRequestBody, err)
}
