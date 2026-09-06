package api

import (
	"errors"
	"jemref_go/internal/handler"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/middleware"
	"jemref_go/internal/usecase"

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
func toApiError(err error) ApiError {
	switch {
	// Middlewareが投げるエラーの変換
	case errors.Is(err, middleware.ErrRequireAuthHeader):
		return ErrUnAuthorized
	case errors.Is(err, middleware.ErrRequireBearerToken):
		return ErrUnAuthorized
	case errors.Is(err, middleware.ErrRequireFirebaseUid):
		return ErrUnAuthorized
	case errors.Is(err, middleware.ErrInvalidFirebaseToken):
		return ErrUnAuthorized
	case errors.Is(err, middleware.ErrRequireEmail):
		return ErrUnAuthorized
	case errors.Is(err, middleware.ErrUserAlreadyExists):
		return ErrUserAlreadyExists

	// Handler層が投げるエラーの変換
	case errors.Is(err, handler.ErrInternalUidNotFound):
		return ErrInternal
	case errors.Is(err, handler.ErrFirebaseUidNotFound):
		return ErrInternal
	case errors.Is(err, handler.ErrPublicUidNotFound):
		return ErrInternal
	case errors.Is(err, handler.ErrEmailNotFound):
		return ErrInternal
	case errors.Is(err, handler.ErrPolicyRequired):
		return ErrBadRequest
	case errors.Is(err, handler.ErrInvalidRequestBody):
		return ErrBadRequest
	case errors.Is(err, handler.ErrPolicyTypeInvalid):
		return ErrBadRequest
	case errors.Is(err, middleware.ErrUserDeleted):
		return ErrUserDeleted

		// Usecaseが投げるエラーの変換
	case errors.Is(err, usecase.ErrPolicyNotFound):
		return ErrPolicyNotFound
	case errors.Is(err, usecase.ErrUserAlreadyExists):
		return ErrUserAlreadyExists
	case errors.Is(err, usecase.ErrUserNotFound):
		return ErrUserNotFound
	case errors.Is(err, usecase.ErrUserDeleted):
		return ErrUserDeleted
	case errors.Is(err, usecase.ErrInvalidPolicyType):
		return ErrInvalidPolicyType
	case errors.Is(err, usecase.ErrInvalidPolicyVersion):
		return ErrInvalidPolicyVersion

		// 予期せぬエラー
	default:
		return ErrInternal
	}
}
