package usecase

import (
	"context"
	"errors"
	"fmt"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
)

type GeneralUsecaseImpl struct {
	generalRepo repository.GeneralRepository
}

func NewGeneralUsecaseImpl(r repository.GeneralRepository) *GeneralUsecaseImpl {
	return &GeneralUsecaseImpl{generalRepo: r}
}

// GetPolicies 最新の規約情報１件を返却する
func (g *GeneralUsecaseImpl) GetPolicies(
	ctx context.Context,
	gp dto.GetPoliciesInput,
) (*dto.GetPoliciesOutput, error) {
	p, err := g.generalRepo.SelectLatestPolicyById(ctx, gp.PolicyId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("指定した規約が存在しません。:%w", ErrPolicyNotFound)
		}

		return nil, err
	}

	return &dto.GetPoliciesOutput{
		PolicyId:      p.Id,
		Label:         p.Name,
		LatestVersion: p.Version,
		EffectiveDate: p.EffectiveDate,
		Content:       p.Content,
	}, nil
}
