package repository

import (
	"context"
	"jemref_go/internal/domain/policy"
)

type GeneralRepository interface {
	// 規約IDをもとに、規約１件を取得する
	SelectLatestById(context.Context, string) (*policy.Policy, error)
}
