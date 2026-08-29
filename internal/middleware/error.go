package middleware

import (
	"errors"
)

var (
	ErrRequireAuthHeader    = errors.New("Authorizationヘッダがありません。")
	ErrRequireBearerToken   = errors.New("Bearerトークンがありません。")
	ErrRequireFirebaseUid   = errors.New("Firebase Uidがありません。")
	ErrInvalidFirebaseToken = errors.New("Firebase tokenが不正です。")
	ErrRequireEmail         = errors.New("Firebaseからのメールアドレス取得に失敗しました。")
	ErrUserAlreadyExists    = errors.New("会員登録済みのユーザです。")
	ErrUserDeleted          = errors.New("削除済みのユーザです。")
)
