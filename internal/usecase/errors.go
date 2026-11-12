package usecase

import (
	"errors"

	"jemref_go/internal/apperr"
)

// ErrTestUnexpected 分岐でカバーしていない予期しないエラーのテスト用
// apperr.Errorにしないことで、変換されず500に落ちることを検証できる。
var ErrTestUnexpected = errors.New("usecase test unexpected")

// 業務エラー。レスポンスとログ出力は台帳の公開メッセージを使う。
var (
	ErrUserNotFound         = apperr.UserNotFound.Err()
	ErrUserDeleted          = apperr.UserDeleted.Err()
	ErrUserAlreadyExists    = apperr.UserAlreadyExists.Err()
	ErrPolicyNotFound       = apperr.PolicyNotFound.Err()
	ErrInvalidPolicyType    = apperr.InvalidPolicyType.Err()
	ErrInvalidPolicyVersion = apperr.InvalidPolicyVersion.Err()

	ErrReachRegistrationLimit = apperr.ReachRegistrationLimit.Err()
	ErrRecordNotFound         = apperr.RecordNotFound.Err()
)

// データ不整合・更新失敗はユーザが対処できないため、意図して500を返す。
// 原因はログのみに残すため、Err()ではなくNew()で生成する。
var (
	ErrUserDataInconsistent = apperr.Internal.New("invalid user data")
	ErrUserDeleteFailed     = apperr.Internal.New("failed db update to delete user")
	ErrUnexpectedPolicy     = apperr.Internal.New("unexpected policy")
)
