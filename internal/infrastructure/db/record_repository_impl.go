package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
)

type RecordRepositoryImpl struct {
	db *sql.DB
}

func NewRecordRepositoryImpl(db *sql.DB) *RecordRepositoryImpl {
	return &RecordRepositoryImpl{
		db: db,
	}
}

func (r *RecordRepositoryImpl) Create(
	ctx context.Context,
	tx repository.DBTX,
	input record.UserRecordEntry,
) error {
	err := r.insertRecordEntry(ctx, tx, input)
	if err != nil {
		return err
	}

	// 貢献者と貢献情報はPh0では純粋な1:1となるため、必要なIDと登録順はここで設定する
	for i, relation := range input.ContributorRelations {
		contributor := relation.Contributor
		contributor.InternalUid = input.InternalUid

		contributorId, err := r.insertContributor(ctx, tx, contributor)
		if err != nil {
			return err
		}

		contribution := relation.Contribution
		contribution.RecordEntryId = input.RecordEntryId
		contribution.ContributorEntryId = contributorId
		contribution.DisplayNumber = i

		if err := r.insertContribution(ctx, tx, contribution); err != nil {
			return err
		}
	}

	if err = r.insertURLs(ctx, tx, input.Urls); err != nil {
		return err
	}

	if err = r.insertMemos(ctx, tx, input.Memos); err != nil {
		return err
	}

	if err = r.insertOptionValues(ctx, tx, input.OptionValues); err != nil {
		return err
	}

	return nil
}

func (r *RecordRepositoryImpl) insertRecordEntry(ctx context.Context, tx repository.DBTX, re record.UserRecordEntry) error {
	args := make([]any, 0, 17)
	apiId := ctxutil.GetApiId(ctx)

	args = append(args,
		re.RecordEntryId,
		re.InternalUid,
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
		re.MRecordId,
		re.DelFlg,
		apiId,
		apiId,
	)

	_, err := tx.ExecContext(ctx, insertRecordEntryQuery, args...)
	if err != nil {
		return fmt.Errorf("insert record entry: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertContributor(ctx context.Context, tx repository.DBTX, c record.ContributorEntry) (int64, error) {
	args := make([]any, 0, 5)
	apiId := ctxutil.GetApiId(ctx)

	args = append(args,
		c.RawName,
		c.InternalUid,
		c.DelFlg,
		apiId,
		apiId,
	)

	result, err := tx.ExecContext(ctx, insertContributorQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("insert contributor: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert contributor: %w", err)
	}

	return id, nil
}

func (r *RecordRepositoryImpl) insertContribution(ctx context.Context, tx repository.DBTX, c record.ContributionEntry) error {
	args := make([]any, 0, 7)
	apiId := ctxutil.GetApiId(ctx)

	args = append(args,
		c.RecordEntryId,
		c.ContributorEntryId,
		c.Role.GetId(),
		c.DisplayNumber,
		c.DelFlg,
		apiId,
		apiId,
	)

	_, err := tx.ExecContext(ctx, insertContributionQuery, args...)
	if err != nil {
		return fmt.Errorf("insert contribution: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertURLs(ctx context.Context, tx repository.DBTX, urls []record.UrlEntry) error {
	if len(urls) == 0 {
		return nil
	}

	args := make([]any, 0, len(urls)*8)
	apiId := ctxutil.GetApiId(ctx)

	for _, u := range urls {
		args = append(args,
			u.RecordEntryId,
			u.DisplayNumber,
			u.Label,
			u.Url,
			u.InternalUid,
			u.DelFlg,
			apiId,
			apiId,
		)
	}

	query := buildBulkInsertQuery(
		insertRecordUrlsQueryBase,
		insertRecordUrlsQueryValues,
		len(urls),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert urls: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertMemos(ctx context.Context, tx repository.DBTX, memos []record.MemoEntry) error {
	if len(memos) == 0 {
		return nil
	}

	args := make([]any, 0, len(memos)*8)
	apiId := ctxutil.GetApiId(ctx)

	for _, m := range memos {
		args = append(args,
			m.RecordEntryId,
			m.DisplayNumber,
			m.Title,
			m.Content,
			m.InternalUid,
			m.DelFlg,
			apiId,
			apiId,
		)
	}

	query := buildBulkInsertQuery(
		insertRecordMemosQueryBase,
		insertRecordMemosQueryValues,
		len(memos),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert memo: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertOptionValues(ctx context.Context, tx repository.DBTX, options []record.OptionValue) error {
	if len(options) == 0 {
		return nil
	}

	args := make([]any, 0, len(options)*8)
	apiId := ctxutil.GetApiId(ctx)

	for _, o := range options {
		args = append(args,
			o.RecordEntryId,
			o.RecordField.GetId(),
			o.DisplayNumber,
			o.RawValue,
			o.InternalUid,
			o.DelFlg,
			apiId,
			apiId,
		)
	}

	query := buildBulkInsertQuery(
		insertRecordFieldValuesQueryBase,
		insertRecordFieldValuesQueryValues,
		len(options),
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert option value: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) SelectByPk(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
	internalUid int64,
) (*record.UserRecordEntry, error) {

	var (
		entry      record.UserRecordEntry
		typeId     string
		readStatus *int
	)

	err := tx.QueryRowContext(
		ctx,
		selectRecordByPkQuery,
		recordId,
		internalUid,
	).Scan(
		&entry.RecordEntryId,
		&entry.InternalUid,
		&typeId,
		&entry.MainTitle,
		&entry.SubTitle,
		&entry.PageRange,
		&entry.Volume,
		&entry.Issue,
		&entry.Publisher,
		&entry.PublicationDate,
		&entry.DatePrecision,
		&readStatus,
		&entry.ContainerId,
		&entry.MRecordId,
		&entry.DelFlg,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"書誌情報が存在しません。record id=%s, internal uid=%d :%w",
				recordId,
				internalUid,
				repository.ErrNotFound,
			)
		}
		return nil, fmt.Errorf("select record by pk record id=%s: %w", recordId, err)
	}

	// IDから値オブジェクトへ復元する(登録時のGetId()の逆変換)。
	// TYPE_IDはマスタへのFKがあるため未知の値は入らない。
	entry.RecordType = record.RecordTypeFromId(typeId)

	// READ_STATUSはFKを持たないTINYINTだが、不正値はバグとして落とす方針のため検証せずに変換する。
	if readStatus != nil {
		status := record.ReadStatusFromId(*readStatus)
		entry.ReadStatus = &status
	}

	return &entry, nil
}

func (r *RecordRepositoryImpl) CountActiveArticles(
	ctx context.Context,
	tx repository.DBTX,
	containerRecordId string,
) (int, error) {

	var count int
	err := tx.QueryRowContext(
		ctx,
		countActiveArticlesQuery,
		containerRecordId,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("count active articles container id=%s: %w", containerRecordId, err)
	}

	return count, nil
}

// logicalDeleteTargets 論理削除の対象クエリ
//
// 構成要素から消す順に並べている(構成要素 -> 書誌本体の順)。
var logicalDeleteTargets = []struct {
	name  string
	query string
}{
	{"field values", logicalDeleteRecordFieldValuesQuery},
	{"memos", logicalDeleteRecordMemosQuery},
	{"urls", logicalDeleteRecordUrlsQuery},
	{"contributions", logicalDeleteRecordContributionsQuery},
	{"record entry", logicalDeleteRecordEntryQuery},
}

func (r *RecordRepositoryImpl) LogicalDelete(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
) error {

	apiId := ctxutil.GetApiId(ctx)

	for _, target := range logicalDeleteTargets {
		_, err := tx.ExecContext(ctx, target.query, apiId, recordId)
		if err != nil {
			return fmt.Errorf("logical delete %s record id=%s: %w", target.name, recordId, err)
		}
	}

	return nil
}
