package db

import (
	"context"
	"fmt"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
)

// Update 書誌を更新する
//
// リクエストの内容でそのまま上書きし、構成要素は洗い替える(物理削除 -> 再登録)。
// 該当する書誌がない(他ユーザ・論理削除済み・存在しない)場合は ErrNotFound を返す。
func (r *RecordRepositoryImpl) Update(
	ctx context.Context,
	tx repository.DBTX,
	input record.UserRecordEntry,
) error {

	if err := r.updateRecordEntry(ctx, tx, input); err != nil {
		return err
	}

	// 構成要素は差分を取らず、すべて消してから入れ直す
	if err := r.deleteComponents(ctx, tx, input.RecordEntryId); err != nil {
		return err
	}

	return r.insertComponents(ctx, tx, input)
}

func (r *RecordRepositoryImpl) updateRecordEntry(
	ctx context.Context,
	tx repository.DBTX,
	re record.UserRecordEntry,
) error {

	args := make([]any, 0, 14)
	apiId := ctxutil.GetApiId(ctx)

	args = append(args,
		re.RecordType.GetId(),
		re.MainTitle,
		re.SubTitle,
		re.PageRange,
		re.Volume,
		re.Issue,
		re.Publisher,
		re.PublicationDate,
		re.DatePrecision,
		re.ReadStatus.GetNullableId(),
		re.ContainerId,
		apiId,
		re.RecordEntryId,
		re.InternalUid,
	)

	result, err := tx.ExecContext(ctx, updateRecordEntryQuery, args...)
	if err != nil {
		return fmt.Errorf("update record entry record id=%s: %w", re.RecordEntryId, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update record entry record id=%s: %w", re.RecordEntryId, err)
	}

	// 内容が同じ場合も更新対象になる(UPD_DTが必ず変わるため0件にはならない)
	if affected == 0 {
		return fmt.Errorf(
			"更新対象の書誌情報が存在しません。record id=%s, internal uid=%d :%w",
			re.RecordEntryId,
			re.InternalUid,
			repository.ErrNotFound,
		)
	}

	return nil
}

// deleteComponentTargets 構成要素の物理削除クエリ
//
// 貢献(e_record_contributions)を最後に置くと貢献者の削除で参照が残るため、
// 貢献者IDを先に退避してから貢献 -> 貢献者の順で消す(deleteContributors側で扱う)。
var deleteComponentTargets = []struct {
	name  string
	query string
}{
	{"field values", deleteRecordFieldValuesQuery},
	{"memos", deleteRecordMemosQuery},
	{"urls", deleteRecordUrlsQuery},
}

// deleteComponents 書誌の構成要素を物理削除する
func (r *RecordRepositoryImpl) deleteComponents(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
) error {

	for _, target := range deleteComponentTargets {
		if _, err := tx.ExecContext(ctx, target.query, recordId); err != nil {
			return fmt.Errorf("delete %s record id=%s: %w", target.name, recordId, err)
		}
	}

	return r.deleteContributors(ctx, tx, recordId)
}

// deleteContributors 貢献と、他の書誌から参照されなくなった貢献者を物理削除する
//
// 貢献を消すと貢献者への参照の手がかりが失われるため、先に貢献者IDを退避する。
func (r *RecordRepositoryImpl) deleteContributors(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
) error {

	contributorIds, err := r.selectContributorIds(ctx, tx, recordId)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, deleteRecordContributionsQuery, recordId); err != nil {
		return fmt.Errorf("delete contributions record id=%s: %w", recordId, err)
	}

	if len(contributorIds) == 0 {
		return nil
	}

	query := deleteUnreferencedContributorsQueryBase +
		buildPlaceholders(len(contributorIds)) +
		deleteUnreferencedContributorsQueryEnd

	if _, err := tx.ExecContext(ctx, query, contributorIds...); err != nil {
		return fmt.Errorf("delete contributors record id=%s: %w", recordId, err)
	}

	return nil
}

// selectContributorIds 書誌が参照している貢献者IDを取得する
func (r *RecordRepositoryImpl) selectContributorIds(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
) ([]any, error) {

	rows, err := tx.QueryContext(ctx, selectContributorIdsByRecordIdQuery, recordId)
	if err != nil {
		return nil, fmt.Errorf("select contributor ids record id=%s: %w", recordId, err)
	}
	defer rows.Close()

	var ids []any
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("select contributor ids record id=%s: %w", recordId, err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select contributor ids record id=%s: %w", recordId, err)
	}

	return ids, nil
}
