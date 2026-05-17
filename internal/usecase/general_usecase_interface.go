package usecase

import (
	"context"
	"jemref_go/internal/usecase/dto"
)

type GeneralUsecase interface {
	GetPolicies(
		ctx context.Context,
		input dto.GetPoliciesInput,
	) (*dto.GetPoliciesOutput, error)
}
