package usecase

// 一般Usecase実装

import (
	"context"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
)

type GeneralUsecaseImpl struct {
	generalRepo repository.GeneralRepository
}

// コンストラクタ
func NewGeneralUsecaseImpl(r repository.GeneralRepository) *GeneralUsecaseImpl {
	return &GeneralUsecaseImpl{generalRepo: r}
}

// 規約情報取得Usecase
func (g *GeneralUsecaseImpl) GetPolicies(
	ctx context.Context,
	gp dto.GetPoliciesInput,
) (*dto.GetPoliciesOutput, error) {
	// 規約情報を取得
	p, err := g.generalRepo.SelectLatestById(ctx, gp.PolicyId)
	if err != nil {
		return nil, err
	}

	// 規約情報を返却
	return &dto.GetPoliciesOutput{
		PolicyId:      p.Id,
		Label:         p.Name,
		LatestVersion: p.Version,
		EffectiveDate: p.EffectiveDate,
		Content:       p.Content,
	}, nil
}
