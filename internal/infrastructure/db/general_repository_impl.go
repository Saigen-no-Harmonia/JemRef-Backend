package db

// 一般リポジトリ実装

import (
	"context"
	"database/sql"
	"errors"
	"jemref_go/internal/domain/policy"
)

// 取得件数0件の場合のエラー
var ErrPolicyNotFound = errors.New("policy not found")

type GeneralRepositoryImpl struct {
	db *sql.DB
}

func NewGeneralRepositoryImpl(db *sql.DB) *GeneralRepositoryImpl {
	return &GeneralRepositoryImpl{
		db: db,
	}
}

// SelectById 規約IDをキーとして、最新の規約情報１件を返却する
func (r *GeneralRepositoryImpl) SelectLatestById(ctx context.Context, id string) (*policy.Policy, error) {

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

		// 取得件数が0件の場合
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPolicyNotFound
		}

		return nil, err
	}

	return &p, nil
}
