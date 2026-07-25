package usecase

import "errors"

// ErrUserNotFound ユーザが存在しない
var ErrUserNotFound = errors.New("user not found")

// ErrPolicyNotFound 規約が存在しない
var ErrPolicyNotFound = errors.New("policy not found")

// ErrUserDeleted ユーザ退会済み
var ErrUserDeleted = errors.New("user has deleted")

// ErrInvalidUser ユーザデータ不整合
var ErrInvalidUser = errors.New("invalid user data")

// ErrUserAlreadyExists 既存のユーザあり
var ErrUserAlreadyExists = errors.New("duplicate user registration")

// ErrUserDeleteFailed ユーザ削除失敗（DBエラー）
var ErrUserDeleteFailed = errors.New("failed db update to delete user")

// ErrInvalidPolicyType ポリシー指定が不正
var ErrInvalidPolicyType = errors.New("invalid policy type")

// ErrInvalidPolicyVersion ポリシーバージョン指定が不正
var ErrInvalidPolicyVersion = errors.New("invalid policy version")

// ErrUnexpectedPolicy DBなどに規約マスタデータ異常がある場合
var ErrUnexpectedPolicy = errors.New("unexpected policy")
