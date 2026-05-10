package usecase

import "jemref_go/internal/repository"

type GeneralUsecase struct {
	repo repository.GeneralRepository
}

func NewGeneralUsecase(r repository.GeneralRepository) *GeneralUsecase {
	return &GeneralUsecase{repo: r}
}

func (g *GeneralUsecase) GetPolicies(gi GetPoliciesInput) (*GetPoliciesOutput, error) {

	// スケルトン
	return &GetPoliciesOutput{}, nil
}
