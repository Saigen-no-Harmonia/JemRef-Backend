package repository

import (
	"context"
	"jemref_go/internal/domain/policy"
)

type GeneralRepository interface {
	// SelectLatestPolicyById 規約IDをもとに、規約１件を取得する
	SelectLatestPolicyById(
		context.Context,
		string,
	) (*policy.Policy, error)

	// SelectPolicyByPrimaryKey 規約IDとバージョンをもとに、規約1件を取得する
	SelectPolicyByPrimaryKey(
		ctx context.Context,
		id string,
		version string,
	) (*policy.Policy, error)
}
