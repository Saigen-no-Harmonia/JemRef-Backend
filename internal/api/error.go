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

	ErrMainTitleRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0009",
		Message:    "主題は必須項目です。",
	}

	ErrRecordTypeRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0010",
		Message:    "書誌タイプ指定は必須です。",
	}

	ErrInvalidRecordType = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0011",
		Message:    "書誌タイプ指定が不正です。",
	}

	ErrContainerNotAllowed = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0012",
		Message:    "この書誌タイプは親レコードを持つことができません。",
	}

	ErrInvalidPublicationDate = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0013",
		Message:    "出版日付の指定が不正です。",
	}

	ErrInvalidReadStatus = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0014",
		Message:    "読了ステータスが不正です。",
	}

	ErrInvalidContributorRole = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0015",
		Message:    "貢献者役割が不正です。",
	}

	ErrReachRegistrationLimit = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0016",
		Message:    "DB登録件数上限を超えています。",
	}

	ErrPolicyRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0017",
		Message:    "規約情報は必須です。",
	}

	ErrAgreedVersionRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0018",
		Message:    "規約同意バージョンは必須です。",
	}

	ErrContributorNameRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0019",
		Message:    "貢献者の氏名は必須です。",
	}

	ErrContributorRoleRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0020",
		Message:    "貢献者の役割は必須です。",
	}

	ErrMemoContentRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0021",
		Message:    "メモの内容は必須です。",
	}

	ErrUrlValueRequired = ApiError{
		StatusCode: http.StatusBadRequest,
		Code:       "E0022",
		Message:    "URLの値は必須です。",
	}
)
