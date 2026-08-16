package db

// 一般リポジトリ実装

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/repository"
)

type GeneralRepositoryImpl struct {
	db *sql.DB
}

func NewGeneralRepositoryImpl(db *sql.DB) *GeneralRepositoryImpl {
	return &GeneralRepositoryImpl{
		db: db,
	}
}

// SelectPolicyByPrimaryKey PKで指定した規約情報を取得する
func (r *GeneralRepositoryImpl) SelectPolicyByPrimaryKey(ctx context.Context, id string, version string) (*policy.Policy, error) {
	var p policy.Policy

	err := r.db.QueryRowContext(
		ctx,
		selectPolicyByPrimaryKeyQuery,
		id,
		version,
	).Scan(
		&p.Id,
		&p.Version,
		&p.Name,
		&p.Content,
		&p.EffectiveDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("規約が存在しません。:%w", repository.ErrNotFound)
		}
		return nil, fmt.Errorf("select policy by id=%s, version=%s :%w", id, version, err)
	}
	return &p, nil
}

// SelectById 規約IDをキーとして、最新の規約情報１件を返却する
func (r *GeneralRepositoryImpl) SelectLatestPolicyById(ctx context.Context, id string) (*policy.Policy, error) {

	var p policy.Policy

	err := r.db.QueryRowContext(
		ctx,
		selectPolicyByIdQuery,
		id,
	).Scan(
		&p.Id,
		&p.Version,
		&p.Name,
		&p.Content,
		&p.EffectiveDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("規約が存在しません。:%w", repository.ErrNotFound)
		}
		return nil, fmt.Errorf("select latest policy by id=%s :%w", id, err)
	}

	return &p, nil
}
