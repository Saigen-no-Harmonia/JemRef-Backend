package repository

import "jemref_go/internal/domain"

type UserRepository interface {
	Create(u *domain.User) error
}
