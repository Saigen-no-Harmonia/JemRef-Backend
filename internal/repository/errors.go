package repository

import "errors"

var (
	ErrTestUnexpected = errors.New("test unexpected")
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEntry = errors.New("duplicate entry")
)
