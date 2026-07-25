package mock

import (
	"context"
	usecaseDto "jemref_go/internal/usecase/dto"
)

type MockUserUsecase struct {
	CreateUserFunc func(
		context.Context,
		usecaseDto.CreateUserInput,
	) (*usecaseDto.CreateUserOutput, error)

	DeleteUserFunc func(
		context.Context,
		usecaseDto.DeleteUserInput,
	) error

	GetUserAgreementsFunc func(
		context.Context,
		int64,
	) (*usecaseDto.GetUserAgreementsOutput, error)

	UpdateUserAgreementsFunc func(
		context.Context,
		usecaseDto.UpdateUserAgreementsInput,
	) error

	LoginFunc func(
		context.Context,
		usecaseDto.UserLoginInput,
	) (*usecaseDto.UserLoginOutput, error)

	CreateUserCalled           bool
	DeleteUserCalled           bool
	GetUserAgreementsCalled    bool
	UpdateUserAgreementsCalled bool
	LoginCalled                bool
	LastCreateUserInput        usecaseDto.CreateUserInput
	LastDeleteUserInput        usecaseDto.DeleteUserInput
	LastUpdateUserInput        usecaseDto.UpdateUserAgreementsInput
}

func (m *MockUserUsecase) Create(
	ctx context.Context,
	input usecaseDto.CreateUserInput,
) (*usecaseDto.CreateUserOutput, error) {
	m.CreateUserCalled = true
	m.LastCreateUserInput = input
	return m.CreateUserFunc(ctx, input)
}

func (m *MockUserUsecase) Delete(
	ctx context.Context,
	input usecaseDto.DeleteUserInput,
) error {
	m.DeleteUserCalled = true
	m.LastDeleteUserInput = input
	return m.DeleteUserFunc(ctx, input)
}

func (m *MockUserUsecase) GetUserAgreements(
	ctx context.Context,
	uid int64,
) (*usecaseDto.GetUserAgreementsOutput, error) {
	m.GetUserAgreementsCalled = true
	return m.GetUserAgreementsFunc(ctx, uid)
}

func (m *MockUserUsecase) UpdateUserAgreements(
	ctx context.Context,
	ua usecaseDto.UpdateUserAgreementsInput,
) error {
	m.UpdateUserAgreementsCalled = true
	m.LastUpdateUserInput = ua
	return m.UpdateUserAgreementsFunc(ctx, ua)
}

func (m *MockUserUsecase) Login(
	ctx context.Context,
	input usecaseDto.UserLoginInput,
) (*usecaseDto.UserLoginOutput, error) {
	m.LoginCalled = true
	return m.LoginFunc(ctx, input)
}
