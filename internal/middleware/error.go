package middleware

import (
	"errors"

	"jemref_go/internal/apperr"
)

// ErrTestUnexpected 分岐でカバーしていない予期しないエラーのテスト用
// apperr.Errorにしないことで、変換されず500に落ちることを検証できる。
var ErrTestUnexpected = errors.New("middleware test unexpected")

// 認証失敗の各原因はレスポンスを共通化する(どの段階で失敗したかを漏らさない)。
// 原因はログにのみ残すため、Err()ではなくNew()で定義
var (
	ErrRequireAuthHeader    = apperr.Unauthorized.New("Authorizationヘッダがありません。")
	ErrRequireBearerToken   = apperr.Unauthorized.New("Bearerトークンがありません。")
	ErrRequireFirebaseUid   = apperr.Unauthorized.New("Firebase Uidがありません。")
	ErrInvalidFirebaseToken = apperr.Unauthorized.New("Firebase tokenが不正です。")
	ErrRequireEmail         = apperr.Unauthorized.New("Firebaseからのメールアドレス取得に失敗しました。")
)

var (
	ErrUserAlreadyExists = apperr.UserAlreadyExists.New("会員登録済みのユーザです。")
	ErrUserDeleted       = apperr.UserDeleted.New("削除済みのユーザです。")
)
