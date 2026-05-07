package repository

import (
	"context"
	"jemref_go/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
}
