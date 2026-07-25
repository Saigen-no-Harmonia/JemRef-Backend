package db

// 一般リポジトリ実装

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/repository"
	"log"
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
			log.Print("SelectByPrimaryKey:規約が存在しませんでした。")
			return nil, repository.ErrNotFound
		}
		log.Print("SelectByPrimaryKey:DB処理に失敗しました。")
		return nil, err
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
			log.Print("SelectLatestById:規約が取得できませんでした。")
			return nil, repository.ErrNotFound
		}
		log.Print("SelectLatestById:DB処理に異常があります。")
		return nil, err
	}

	return &p, nil
}
