package repository

import (
	"context"
	"jemref_go/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, uid int64) error
	SelectByInternalUid(ctx context.Context, uid int64) (*user.User, error)
	SelectByFirebaseUid(ctx context.Context, firebaseUid string) (*user.User, error)
}
