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

	CreateCalled    bool
	LastCreateInput usecaseDto.CreateRecordInput
}

func (m *MockRecordUsecase) Create(
	ctx context.Context,
	ci usecaseDto.CreateRecordInput,
) (*usecaseDto.CreateRecordOutput, error) {
	m.CreateCalled = true
	m.LastCreateInput = ci
	return m.CreateFunc(ctx, ci)
}
