package mock

import (
	"context"
	usecaseDto "jemref_go/internal/usecase/dto"
)

// MockRecordUsecase usecase.RecordUsecaseのモック
type MockRecordUsecase struct {
	CreateFunc func(
		context.Context,
		usecaseDto.CreateRecordInput,
	) (*usecaseDto.CreateRecordOutput, error)

	DeleteFunc func(
		context.Context,
		usecaseDto.DeleteRecordInput,
	) error

	CreateCalled    bool
	DeleteCalled    bool
	LastCtx         context.Context
	LastCreateInput usecaseDto.CreateRecordInput
	LastDeleteInput usecaseDto.DeleteRecordInput
}

func (m *MockRecordUsecase) Delete(
	ctx context.Context,
	di usecaseDto.DeleteRecordInput,
) error {
	m.LastCtx = ctx
	m.DeleteCalled = true
	m.LastDeleteInput = di
	return m.DeleteFunc(ctx, di)
}

func (m *MockRecordUsecase) Create(
	ctx context.Context,
	ci usecaseDto.CreateRecordInput,
) (*usecaseDto.CreateRecordOutput, error) {
	m.LastCtx = ctx
	m.CreateCalled = true
	m.LastCreateInput = ci
	return m.CreateFunc(ctx, ci)
}
