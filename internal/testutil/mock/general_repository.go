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
	SelectByPrimaryKeyFunc func(
		ctx context.Context,
		id string,
		version string,
	) (*policy.Policy, error)

	SelectLatestByIdFuncCalled int
	SelectByPrimaryKeyCalled   int
	SelectedIds                []string
	SelectedPrimaryKeys        []PolicyPrimaryKeyCall
}

func (m *MockGeneralRepository) SelectLatestPolicyById(
	ctx context.Context,
	id string,
) (*policy.Policy, error) {
	m.SelectLatestByIdFuncCalled++
	m.SelectedIds = append(m.SelectedIds, id)
	return m.SelectLatestByIdFunc(ctx, id)
}

func (m *MockGeneralRepository) SelectPolicyByPrimaryKey(
	ctx context.Context,
	id string,
	version string,
) (*policy.Policy, error) {
	m.SelectByPrimaryKeyCalled++
	m.SelectedPrimaryKeys = append(m.SelectedPrimaryKeys, PolicyPrimaryKeyCall{
		Id:      id,
		Version: version,
	})
	return m.SelectByPrimaryKeyFunc(ctx, id, version)
}

type PolicyPrimaryKeyCall struct {
	Id      string
	Version string
}
