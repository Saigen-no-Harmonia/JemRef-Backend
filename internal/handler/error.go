package handler

import (
	"errors"
)

var (
	ErrTestUnexpected      = errors.New("handler test unexpected")
	ErrInternalUidNotFound = errors.New("contextからinternal uidが取得できませんでした。")
	ErrFirebaseUidNotFound = errors.New("contextからfirebase uidが取得できませんでした。")
	ErrPublicUidNotFound   = errors.New("contextから公開用UIDが取得できませんでした。")
	ErrEmailNotFound       = errors.New("contextからメールアドレスが取得できませんでした。")
	ErrPolicyRequired      = errors.New("リクエストに規約情報がありませんでした。")
	ErrInvalidRequestBody  = errors.New("リクエストボディの形式が不正です。")
	ErrPolicyTypeInvalid   = errors.New("規約タイプ指定が不正です。")
)
