package mock

import (
	"context"
	"jemref_go/internal/domain/policy"
)

type MockGeneralRepository struct {
	SelectLatestByIdFunc func(
		ctx context.Context,
		id string,
	) (*policy.Policy, error)

	SelectLatestByIdFuncCalled int
}

func (m *MockGeneralRepository) SelectLatestById(
	ctx context.Context,
	id string,
) (*policy.Policy, error) {
	m.SelectLatestByIdFuncCalled++
	return m.SelectLatestByIdFunc(ctx, id)
}
