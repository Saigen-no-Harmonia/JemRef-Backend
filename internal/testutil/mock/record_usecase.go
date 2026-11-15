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

	UpdateFunc func(
		context.Context,
		usecaseDto.UpdateRecordInput,
	) error

	GetFunc func(
		context.Context,
		usecaseDto.GetRecordInput,
	) (*usecaseDto.GetRecordOutput, error)

	ListFunc func(
		context.Context,
		usecaseDto.GetRecordsInput,
	) (*usecaseDto.GetRecordsOutput, error)

	CreateCalled    bool
	UpdateCalled    bool
	DeleteCalled    bool
	GetCalled       bool
	ListCalled      bool
	LastCtx         context.Context
	LastCreateInput usecaseDto.CreateRecordInput
	LastUpdateInput usecaseDto.UpdateRecordInput
	LastDeleteInput usecaseDto.DeleteRecordInput
	LastGetInput    usecaseDto.GetRecordInput
	LastListInput   usecaseDto.GetRecordsInput
}

func (m *MockRecordUsecase) List(
	ctx context.Context,
	li usecaseDto.GetRecordsInput,
) (*usecaseDto.GetRecordsOutput, error) {
	m.LastCtx = ctx
	m.ListCalled = true
	m.LastListInput = li
	return m.ListFunc(ctx, li)
}

func (m *MockRecordUsecase) Update(
	ctx context.Context,
	ui usecaseDto.UpdateRecordInput,
) error {
	m.LastCtx = ctx
	m.UpdateCalled = true
	m.LastUpdateInput = ui
	return m.UpdateFunc(ctx, ui)
}

func (m *MockRecordUsecase) Get(
	ctx context.Context,
	gi usecaseDto.GetRecordInput,
) (*usecaseDto.GetRecordOutput, error) {
	m.LastCtx = ctx
	m.GetCalled = true
	m.LastGetInput = gi
	return m.GetFunc(ctx, gi)
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
