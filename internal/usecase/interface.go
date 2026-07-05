package usecase

import (
	"context"

	usecaseDto "jemref_go/internal/usecase/dto"
)

type UserUsecase interface {
	CreateUser(
		ctx context.Context,
		input usecaseDto.CreateUserInput,
	) (*usecaseDto.CreateUserOutput, error)

	DeleteUser(
		ctx context.Context,
		input usecaseDto.DeleteUserInput,
	) error

	Login(
		ctx context.Context,
		input usecaseDto.UserLoginInput,
	) (*usecaseDto.UserLoginOutput, error)
}
