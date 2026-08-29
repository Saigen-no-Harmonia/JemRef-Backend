package api

import (
	"net/http"
)

type ApiError struct {
	StatusCode int
	Code       string
	Message    string
}

var (
	ErrInternal = ApiError{
		StatusCode: http.StatusInternalServerError,
		Code:       "F0001",
		Message:    "予期せぬエラーが発生しました。",
	}
	ErrUnAuthorized = ApiError{
		StatusCode: http.StatusUnauthorized,
		Code:       "E0001",
		Message:    "認証情報が不正です。",
	}
	ErrBadRequest = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0002",
		Message:    "リクエストが不正です。",
	}
	ErrPolicyNotFound = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0003",
		Message:    "規約が存在しません。",
	}
	ErrInvalidPolicyType = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0004",
		Message:    "規約タイプが不正です。",
	}
	ErrInvalidPolicyVersion = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0005",
		Message:    "規約バージョンが不正です。",
	}
	ErrUserAlreadyExists = ApiError{
		StatusCode: http.StatusConflict,
		Code:       "E0006",
		Message:    "すでに存在するFirebase UIDまたはEmailです。",
	}
	ErrUserDeleted = ApiError{
		StatusCode: http.StatusUnauthorized,
		Code:       "E0007",
		Message:    "退会済みのユーザです。",
	}
	ErrUserNotFound = ApiError{
		StatusCode: http.StatusUnauthorized,
		Code:       "E0008",
		Message:    "未登録ユーザです。",
	}
)
