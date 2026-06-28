package repository

import "errors"

// ErrUserNotFound ユーザが存在しない
var ErrUserNotFound = errors.New("user not found")

// ErrPolicyNotFound 規約が存在しない
var ErrPolicyNotFound = errors.New("policy not found")
