package repository

import "errors"

// ErrTestUnexpected テスト用の予期せぬエラー
var ErrTestUnexpected = errors.New("test unexpected")

// ErrNotFound ユーザが存在しない
var ErrNotFound = errors.New("not found")

// ErrDuplicateEntry
var ErrDuplicateEntry = errors.New("duplicate entry")
