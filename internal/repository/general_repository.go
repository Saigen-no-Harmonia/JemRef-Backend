package repository

import "jemref_go/internal/domain/policy"

type GeneralRepository interface {
	GetPolicies(p *policy.Policy) error
}
