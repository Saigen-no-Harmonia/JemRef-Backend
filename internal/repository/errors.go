package repository

import "errors"

// ErrNotFound ユーザが存在しない
var ErrNotFound = errors.New("not found")

// ErrDuplicateEntry
var ErrDuplicateEntry = errors.New("duplicate entry")
