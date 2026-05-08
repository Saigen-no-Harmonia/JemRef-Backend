package repository

import (
	"context"
	"jemref_go/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
}
