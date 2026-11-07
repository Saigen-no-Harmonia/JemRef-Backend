package usecase

import (
	"context"

	"jemref_go/internal/usecase/dto"
	usecaseDto "jemref_go/internal/usecase/dto"
)

type GeneralUsecase interface {
	GetPolicies(
		ctx context.Context,
		input usecaseDto.GetPoliciesInput,
	) (*usecaseDto.GetPoliciesOutput, error)
}

type UserUsecase interface {
	Create(
		ctx context.Context,
		input usecaseDto.CreateUserInput,
	) (*usecaseDto.CreateUserOutput, error)

	Delete(
		ctx context.Context,
		input usecaseDto.DeleteUserInput,
	) error

	GetUserAgreements(
		ctx context.Context,
		uid int64,
	) (*usecaseDto.GetUserAgreementsOutput, error)

	UpdateUserAgreements(
		ctx context.Context,
		ua usecaseDto.UpdateUserAgreementsInput,
	) error

	Login(
		ctx context.Context,
		input usecaseDto.UserLoginInput,
	) (*usecaseDto.UserLoginOutput, error)
}

type AuthUsecase interface {
	Authenticate(
		ctx context.Context,
		firebaseUid string,
	) (*dto.AuthUserOutput, error)

	DeleteUser(
		ctx context.Context,
		firebaseUid string,
	) error
}

type RecordUsecase interface {
	Create(
		ctx context.Context,
		ci dto.CreateRecordInput,
	) (*dto.CreateRecordOutput, error)

	Get(
		ctx context.Context,
		gi dto.GetRecordInput,
	) (*dto.GetRecordOutput, error)

	Delete(
		ctx context.Context,
		di dto.DeleteRecordInput,
	) error
}
