package dto

import "time"

// 規約取得Usecase Input
type GetPoliciesInput struct {
	PolicyId string
}

// 規約取得Usecase Output
type GetPoliciesOutput struct {
	PolicyId      string
	Label         string
	LatestVersion string
	EffectiveDate time.Time
	Content       string
}
