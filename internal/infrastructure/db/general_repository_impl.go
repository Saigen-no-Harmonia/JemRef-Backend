package db

import (
	"database/sql"
	"jemref_go/internal/domain/policy"
)

// 一般リポジトリ実装

type GeneralRepositoryImpl struct {
	db *sql.DB
}

func NewGeneralRepositoryImpl(db *sql.DB) *GeneralRepositoryImpl {
	return &GeneralRepositoryImpl{
		db: db,
	}
}

func (r *GeneralRepositoryImpl) GetPolicies(p *policy.Policy) error {
	// TODO スケルトン
	return nil
}
