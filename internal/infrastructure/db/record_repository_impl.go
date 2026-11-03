package db

import (
	"context"
	"database/sql"
	"fmt"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"strings"
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
	const query = `
		INSERT INTO e_records (
			ID,
			USER_ID,
			TYPE_ID,
            MAIN_TITLE,
            SUB_TITLE,
            PAGE_RANGE,
            VOLUME,
            ISSUE,
            PUBLISHER,
            PUBLICATION_DATE,
            DATE_PRECISION,
            READ_STATUS,
            PARENT_ID,
            M_RECORD_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?,
			?, 
			?, 
			?,
			?,
			?,
			?, 
			?, 
			?,
			?,
			?,
			?, 
			?, 
			?,
			'system',
			?,
			'system'
		)
	`
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
		re.ParentId,
		re.MRecordId,
		re.DelFlg,
		apiId,
		apiId,
	)

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert record entry: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertContributor(ctx context.Context, tx repository.DBTX, c record.ContributorEntry) (int64, error) {
	const query = `
		INSERT INTO e_contributors (
			RAW_NAME,
			USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?, 
			?,
			'system',
			?,
			'system'
		)
	`
	args := make([]any, 0, 5)
	apiId := ctxutil.GetApiId(ctx)

	args = append(args,
		c.RawName,
		c.InternalUid,
		c.DelFlg,
		apiId,
		apiId,
	)

	result, err := tx.ExecContext(ctx, query, args...)
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
	const query = `
		INSERT INTO e_record_contributions (
			E_RECORD_ID,
			E_CONTRIBUTOR_ID,
            ROLE_ID,
			DISPLAY_NUMBER,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?,
			?, 
			?, 
			?,
			'system',
			?,
			'system'
		)
	`
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

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert contribution: %w", err)
	}

	return nil
}

func (r *RecordRepositoryImpl) insertURLs(ctx context.Context, tx repository.DBTX, urls []record.UrlEntry) error {
	if len(urls) == 0 {
		return nil
	}

	const queryBase = `
        INSERT INTO e_record_urls (
            E_RECORD_ID,
            DISPLAY_NUMBER,
            RAW_LABEL,
            RAW_URL,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

	values := make([]string, 0, len(urls))
	args := make([]any, 0, len(urls)*8)
	apiId := ctxutil.GetApiId(ctx)

	for _, u := range urls {
		values = append(values, "(?, ?, ?, ?, ?, ?,?,'system',?,'system')")

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

	query := queryBase + strings.Join(values, ", ")

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

	const queryBase = `
        INSERT INTO e_record_memos (
            E_RECORD_ID,
            DISPLAY_NUMBER,
            TITLE,
            CONTENT,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

	values := make([]string, 0, len(memos))
	args := make([]any, 0, len(memos)*9)
	apiId := ctxutil.GetApiId(ctx)

	for _, m := range memos {
		values = append(values, "(?, ?, ?, ?, ?, ?,?,'system',?,'system')")

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

	query := queryBase + strings.Join(values, ", ")

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

	const queryBase = `
        INSERT INTO e_record_field_values (
            E_RECORD_ID,
            RECORD_FIELD_ID,
            DISPLAY_NUMBER,
            RAW_VALUE,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

	values := make([]string, 0, len(options))
	args := make([]any, 0, len(options)*8)
	apiId := ctxutil.GetApiId(ctx)

	for _, o := range options {
		values = append(values, "(?, ?, ?, ?, ?, ?,?,'system',?,'system')")

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

	query := queryBase + strings.Join(values, ", ")

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert option value: %w", err)
	}

	return nil
}
