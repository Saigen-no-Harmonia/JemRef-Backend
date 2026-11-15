package db

import (
	"context"
	"fmt"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"strings"
	"time"
)

// 書誌一覧の検索(REF-API-001)
//
// SQLは record_query.go の断片を組み合わせて作る。ユーザの入力は必ずargsに積み、SQLには連結しない。
// 断片のプレースホルダの順とargsの順がずれてもコンパイルエラーにならないため、
// 検索条件ごとの統合テストで担保する。

// CountByCondition 一覧の検索条件に一致する書誌の件数を返す
func (r *RecordRepositoryImpl) CountByCondition(
	ctx context.Context,
	tx repository.DBTX,
	cond repository.RecordListCondition,
) (int, error) {

	where, args := buildRecordListWhere(cond)

	var count int
	err := tx.QueryRowContext(
		ctx,
		countRecordListQueryBase+where,
		args...,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("count records by condition internal uid=%d: %w", cond.InternalUid, err)
	}

	return count, nil
}

// SelectByCondition 一覧の検索条件に一致する書誌を、並び替えとlimit/offsetを適用して返す
func (r *RecordRepositoryImpl) SelectByCondition(
	ctx context.Context,
	tx repository.DBTX,
	cond repository.RecordListCondition,
) ([]*record.UserRecordEntry, error) {

	where, args := buildRecordListWhere(cond)
	query := selectRecordListQueryBase +
		where +
		buildRecordListOrderBy(cond.SortKey, cond.SortOrder) +
		recordListQueryLimit
	args = append(args, cond.Limit, cond.Offset)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select records by condition internal uid=%d: %w", cond.InternalUid, err)
	}
	defer rows.Close()

	entries, err := scanRecordEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scan records by condition internal uid=%d: %w", cond.InternalUid, err)
	}

	return entries, nil
}

// SelectByPks 書誌を複数件取得する
func (r *RecordRepositoryImpl) SelectByPks(
	ctx context.Context,
	tx repository.DBTX,
	recordIds []string,
	internalUid int64,
) ([]*record.UserRecordEntry, error) {

	// IN ()は妥当なSQLではない(buildPlaceholdersもpanicする)ため、クエリを発行しない
	if len(recordIds) == 0 {
		return nil, nil
	}

	args := make([]any, 0, len(recordIds)+1)
	args = append(args, internalUid)
	for _, id := range recordIds {
		args = append(args, id)
	}

	query := selectRecordsByPksQueryBase +
		buildPlaceholders(len(recordIds)) +
		selectRecordsByPksQueryEnd

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select records by pks internal uid=%d: %w", internalUid, err)
	}
	defer rows.Close()

	entries, err := scanRecordEntries(rows)
	if err != nil {
		return nil, fmt.Errorf("scan records by pks internal uid=%d: %w", internalUid, err)
	}

	return entries, nil
}

// buildRecordListWhere 一覧の検索条件からWHERE句とその引数を組み立てる
//
// 件数クエリと一覧クエリで共用する。指定された条件の断片だけを連結し(条件はANDになる)、
// 断片のプレースホルダの順に引数を積む。
func buildRecordListWhere(cond repository.RecordListCondition) (string, []any) {
	var where strings.Builder
	args := make([]any, 0, 9)

	where.WriteString(recordListWhereBase)
	args = append(args, cond.InternalUid)

	if cond.ContributorName != nil {
		pattern := toContainsPattern(*cond.ContributorName)

		where.WriteString(recordListWhereContributorName)
		// 書誌の貢献者、収録物の貢献者の順
		args = append(args, pattern, pattern)
	}

	if cond.Title != nil {
		pattern := toContainsPattern(*cond.Title)

		where.WriteString(recordListWhereTitle)
		// 書誌の主題・副題、収録物の主題・副題の順
		args = append(args, pattern, pattern, pattern, pattern)
	}

	if cond.PublishYear != nil {
		from := time.Date(*cond.PublishYear, time.January, 1, 0, 0, 0, 0, time.UTC)

		where.WriteString(recordListWherePublishYear)
		// [その年の1月1日, 翌年の1月1日)
		args = append(args, from, from.AddDate(1, 0, 0))
	}

	return where.String(), args
}

// recordListOrderBy 一覧の並び順のホワイトリスト
//
// ORDER BYはプレースホルダを使えないため、入力値をSQLに連結せず、
// 並び替えキー×並び順から固定の句を引く。
var recordListOrderBy = map[record.SortKey]map[record.SortOrder]string{
	record.SortKeyPublishYear: {
		record.SortOrderAsc:  recordListOrderByPublishYearAsc,
		record.SortOrderDesc: recordListOrderByPublishYearDesc,
	},
	record.SortKeyInsertDate: {
		record.SortOrderAsc:  recordListOrderByInsertDateAsc,
		record.SortOrderDesc: recordListOrderByInsertDateDesc,
	},
}

// buildRecordListOrderBy 並び替えキーと並び順に対応するORDER BY句を返す
//
// 未知の並び替えキー・並び順の場合はpanicする。値はhandlerで IsValid() により検証済みの前提であり、
// ここに到達した場合は呼び出し側のバグとして落とす(XxxFromId と同じ方針)。
func buildRecordListOrderBy(key record.SortKey, order record.SortOrder) string {
	orders, ok := recordListOrderBy[key]
	if !ok {
		panic("buildRecordListOrderBy: invalid sort key")
	}

	clause, ok := orders[order]
	if !ok {
		panic("buildRecordListOrderBy: invalid sort order")
	}

	return clause
}

// likeEscaper LIKEパターンの特殊文字(エスケープ文字 '!' と、ワイルドカード '%' '_')をエスケープする
//
// エスケープ文字はクエリの escape '!' と一致させる。
// strings.Replacer は入力を1回だけ走査して置換するため、置換で付けた '!' が再び置換されることはない
// (エスケープ文字を先に処理する、という順序の問題が起きない)。
var likeEscaper = strings.NewReplacer(
	"!", "!!",
	"%", "!%",
	"_", "!_",
)

// escapeLike 値をLIKEパターンの中で文字どおりに扱えるようエスケープする
func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}

// toContainsPattern 部分一致のLIKEパターンを返す
//
// 前後のワイルドカードは、値をエスケープした後に付ける。
func toContainsPattern(s string) string {
	return "%" + escapeLike(s) + "%"
}
