package db

import (
	"context"
	"fmt"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
)

// SelectComponents 渡された書誌に、構成要素(貢献者/URL/メモ/項目値)を設定する
//
// 書誌と収録物を同時に渡せるよう、書誌IDのIN句で一括取得してから振り分ける。
// 4テーブルを個別に引くと最大8クエリになるため、4クエリに抑えている。
func (r *RecordRepositoryImpl) SelectComponents(
	ctx context.Context,
	tx repository.DBTX,
	entries []*record.UserRecordEntry,
) error {

	// 収録物を持たない書誌のためにnilを許容する
	targets := make(map[string]*record.UserRecordEntry, len(entries))
	ids := make([]any, 0, len(entries))

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		targets[entry.RecordEntryId] = entry
		ids = append(ids, entry.RecordEntryId)
	}

	if len(ids) == 0 {
		return nil
	}

	selectors := []struct {
		name string
		fn   func(context.Context, repository.DBTX, []any, map[string]*record.UserRecordEntry) error
	}{
		{"contributor relations", r.selectContributorRelations},
		{"urls", r.selectUrls},
		{"memos", r.selectMemos},
		{"option values", r.selectOptionValues},
	}

	for _, selector := range selectors {
		if err := selector.fn(ctx, tx, ids, targets); err != nil {
			return fmt.Errorf("select %s: %w", selector.name, err)
		}
	}

	return nil
}

// selectContributorRelations 貢献と貢献者のペアを取得する
//
// 氏名は e_contributors 側にあるため、両テーブルをJOINして1件のペアに復元する。
func (r *RecordRepositoryImpl) selectContributorRelations(
	ctx context.Context,
	tx repository.DBTX,
	ids []any,
	targets map[string]*record.UserRecordEntry,
) error {

	query := selectContributionsByRecordIdsQueryBase +
		buildPlaceholders(len(ids)) +
		selectContributionsByRecordIdsQueryOrder

	rows, err := tx.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			recordEntryId string
			roleId        string
			contribution  record.ContributionEntry
			contributor   record.ContributorEntry
		)

		if err := rows.Scan(
			&recordEntryId,
			&contribution.ContributionEntryId,
			&roleId,
			&contribution.DisplayNumber,
			&contributor.ContributorEntryId,
			&contributor.RawName,
			&contributor.InternalUid,
			&contributor.MContributorId,
		); err != nil {
			return err
		}

		target, ok := targets[recordEntryId]
		if !ok {
			continue
		}

		contribution.RecordEntryId = recordEntryId
		contribution.ContributorEntryId = contributor.ContributorEntryId
		// ROLE_IDはマスタへのFKがあるため未知の値は入らない
		contribution.Role = record.RoleFromId(roleId)

		target.ContributorRelations = append(
			target.ContributorRelations,
			record.ContributorRelation{
				Contributor:  contributor,
				Contribution: contribution,
			},
		)
	}

	return rows.Err()
}

func (r *RecordRepositoryImpl) selectUrls(
	ctx context.Context,
	tx repository.DBTX,
	ids []any,
	targets map[string]*record.UserRecordEntry,
) error {

	query := selectUrlsByRecordIdsQueryBase +
		buildPlaceholders(len(ids)) +
		selectUrlsByRecordIdsQueryOrder

	rows, err := tx.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			recordEntryId string
			url           record.UrlEntry
		)

		if err := rows.Scan(
			&recordEntryId,
			&url.UrlEntryId,
			&url.DisplayNumber,
			&url.Label,
			&url.Url,
			&url.InternalUid,
		); err != nil {
			return err
		}

		target, ok := targets[recordEntryId]
		if !ok {
			continue
		}

		url.RecordEntryId = recordEntryId
		target.Urls = append(target.Urls, url)
	}

	return rows.Err()
}

func (r *RecordRepositoryImpl) selectMemos(
	ctx context.Context,
	tx repository.DBTX,
	ids []any,
	targets map[string]*record.UserRecordEntry,
) error {

	query := selectMemosByRecordIdsQueryBase +
		buildPlaceholders(len(ids)) +
		selectMemosByRecordIdsQueryOrder

	rows, err := tx.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			recordEntryId string
			memo          record.MemoEntry
		)

		if err := rows.Scan(
			&recordEntryId,
			&memo.MemoEntryId,
			&memo.DisplayNumber,
			&memo.Title,
			&memo.Content,
			&memo.InternalUid,
			&memo.ShareLevel,
		); err != nil {
			return err
		}

		target, ok := targets[recordEntryId]
		if !ok {
			continue
		}

		memo.RecordEntryId = recordEntryId
		target.Memos = append(target.Memos, memo)
	}

	return rows.Err()
}

func (r *RecordRepositoryImpl) selectOptionValues(
	ctx context.Context,
	tx repository.DBTX,
	ids []any,
	targets map[string]*record.UserRecordEntry,
) error {

	query := selectFieldValuesByRecordIdsQueryBase +
		buildPlaceholders(len(ids)) +
		selectFieldValuesByRecordIdsQueryOrder

	rows, err := tx.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			recordEntryId string
			fieldId       string
			optionValue   record.OptionValue
		)

		if err := rows.Scan(
			&recordEntryId,
			&optionValue.OptionValueEntryId,
			&fieldId,
			&optionValue.DisplayNumber,
			&optionValue.RawValue,
			&optionValue.InternalUid,
		); err != nil {
			return err
		}

		target, ok := targets[recordEntryId]
		if !ok {
			continue
		}

		optionValue.RecordEntryId = recordEntryId
		// RECORD_FIELD_IDはマスタへのFKがあるため未知の値は入らない
		optionValue.RecordField = record.RecordFieldFromId(fieldId)

		target.OptionValues = append(target.OptionValues, optionValue)
	}

	return rows.Err()
}
