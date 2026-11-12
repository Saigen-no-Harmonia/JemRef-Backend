package api

import (
	"errors"

	"jemref_go/internal/apperr"
	handlerDto "jemref_go/internal/handler/dto"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		apiErr := toApiError(err)

		c.JSON(
			apiErr.StatusCode,
			handlerDto.ErrorResponse{
				Code:    apiErr.Code,
				Message: apiErr.Message,
			},
		)

	}
}

// toApiError エラー情報を元にレスポンス用構造体を返却する
//
// 各層のsentinel errorは apperr.Error として定義されており、コードと公開メッセージを
// 自身が持つ。したがってここに変換表は不要で、連鎖から取り出すだけでよい。
// apperr.Error でないエラー(想定外の失敗)は、内容を漏らさず500に寄せる。
func toApiError(err error) ApiError {
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		return fromDef(apperr.Internal)
	}

	return fromDef(appErr.Def())
}
