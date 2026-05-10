package usecase

import "time"

// 規約取得Usecase Input
type GetPoliciesInput struct {
	PolicyType string
}

// 規約取得Usecase Output
type GetPoliciesOutput struct {
	PolicyType    string
	Label         string
	LatestVersion string
	EffectiveDate time.Time
	Content       string
}
