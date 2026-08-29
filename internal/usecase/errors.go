package usecase

import "errors"

var (
	ErrTestUnexpected       = errors.New("usecase test unexpected")
	ErrUserNotFound         = errors.New("user not found")
	ErrPolicyNotFound       = errors.New("policy not found")
	ErrUserDeleted          = errors.New("user has deleted")
	ErrUserDataInconsistent = errors.New("invalid user data")
	ErrUserAlreadyExists    = errors.New("duplicate user registration")
	ErrUserDeleteFailed     = errors.New("failed db update to delete user")
	ErrInvalidPolicyType    = errors.New("invalid policy type")
	ErrInvalidPolicyVersion = errors.New("invalid policy version")
	ErrUnexpectedPolicy     = errors.New("unexpected policy")
)
