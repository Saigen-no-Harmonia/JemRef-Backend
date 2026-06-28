package repository

import "context"

type FirebaseRepository interface {
	DeleteUser(ctx context.Context, uid string) error
}
