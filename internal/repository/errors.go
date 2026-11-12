package repository

import "errors"

// DBエラーを定義する
// これらのエラーはusecase側で業務エラーに包み直してから返却する
var (
	ErrTestUnexpected = errors.New("test unexpected")
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEntry = errors.New("duplicate entry")
)
