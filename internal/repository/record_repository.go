package repository

import (
	"context"
	"jemref_go/internal/domain/record"
)

type RecordRepository interface {
	Create(
		ctx context.Context,
		tx DBTX,
		ur record.UserRecordEntry,
	) error

	// SelectByPk 書誌を1件取得する
	//
	// 所有権の確認を兼ねるため、内部用ユーザIDも条件に含める。
	// 論理削除済みの書誌は取得対象外。該当なしの場合は ErrNotFound を返す。
	// 詳細参照・削除・更新で共用するため全カラムを取得する(監査列を除く)。
	// 構成要素は含まないので、必要な場合は SelectComponents を続けて呼ぶ。
	SelectByPk(
		ctx context.Context,
		tx DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error)

	// SelectComponents 渡された書誌に、構成要素(貢献者/URL/メモ/項目値)を設定する
	//
	// 書誌と収録物をまとめて渡せるよう、書誌IDのIN句で一括取得して振り分ける。
	// 論理削除済みの構成要素と、貢献者本体(e_contributors)が論理削除済みの貢献は除外する。
	// 各スライスは表示順(DISPLAY_NUMBER)の昇順で設定し、該当なしの場合はnilのままとする。
	// entriesの要素にnilが含まれていても良い(収録物を持たない書誌のために許容する)。
	//
	// 結果は引数のentryにappendするため、同じentryに対して2回呼ぶと構成要素が重複する。
	// 1つのentryに対して1回だけ呼ぶこと。
	SelectComponents(
		ctx context.Context,
		tx DBTX,
		entries []*record.UserRecordEntry,
	) error

	// Update 書誌を更新する
	//
	// リクエストの内容でそのまま上書きし、構成要素(貢献者/URL/メモ/項目値)は洗い替える。
	// 子テーブルのUNIQUE制約に DEL_FLG が含まれないため、構成要素は物理削除してから再登録する。
	// 貢献者(e_contributors)は他の書誌から参照されていないものだけを削除する。
	// 該当なし(他ユーザ・論理削除済み・存在しない)の場合は ErrNotFound を返す。
	//
	// 検証するのは更新対象の書誌の所有権だけで、ur.ContainerId が同じユーザの書誌かは見ない。
	// コンテナIDは所有権を確認してから渡すこと(呼び出し側の責務)。
	Update(
		ctx context.Context,
		tx DBTX,
		ur record.UserRecordEntry,
	) error

	// CountActiveArticles 指定した書誌をコンテナとする、論理削除されていない書誌の件数を返す
	CountActiveArticles(
		ctx context.Context,
		tx DBTX,
		containerRecordId string,
	) (int, error)

	// LogicalDelete 書誌とその構成要素(貢献/URL/メモ/項目値)を論理削除する
	//
	// e_contributors はユーザ単位の入力辞書であり書誌に属さないため削除しない。
	LogicalDelete(
		ctx context.Context,
		tx DBTX,
		recordId string,
	) error

	// CountByCondition 一覧の検索条件に一致する書誌の件数を返す
	//
	// SelectByCondition と同じ条件で数える。並び順とlimit/offsetは件数に影響しない。
	CountByCondition(
		ctx context.Context,
		tx DBTX,
		cond RecordListCondition,
	) (int, error)

	// SelectByCondition 一覧の検索条件に一致する書誌を、並び替えとlimit/offsetを適用して返す
	//
	// 自ユーザの論理削除されていない書誌が対象。収録物(論理削除されていない書誌から
	// CONTAINER_IDで参照されている書誌)は含まない(CountActiveArticles > 0 と同じ定義)。
	// 構成要素は含まないので、必要な場合は SelectContributorRelations 等を続けて呼ぶ。
	// 該当なしの場合は空(nil)を返す。
	SelectByCondition(
		ctx context.Context,
		tx DBTX,
		cond RecordListCondition,
	) ([]*record.UserRecordEntry, error)

	// SelectByPks 書誌を複数件取得する
	//
	// 一覧で表示する収録物をまとめて取得するために使う。
	// 所有権の確認を兼ねるため内部用ユーザIDも条件に含め、論理削除済みは取得対象外とする。
	// 該当しないIDは結果に含めない(エラーにしない)。結果の順序は保証しない。
	// recordIdsが空の場合はクエリを発行せずに空(nil)を返す。
	SelectByPks(
		ctx context.Context,
		tx DBTX,
		recordIds []string,
		internalUid int64,
	) ([]*record.UserRecordEntry, error)

	// SelectContributorRelations 渡された書誌に、貢献者情報だけを設定する
	//
	// 一覧の表示に必要な構成要素は貢献者だけなので、URL・メモ・項目値は取得しない。
	// 取得条件・並び順・nilの扱いは SelectComponents と同じ。
	// 結果は引数のentryにappendするため、1つのentryに対して1回だけ呼ぶこと。
	// また同じ書誌IDの別のentryを渡すと、貢献者はどちらか一方にしか設定されない。
	SelectContributorRelations(
		ctx context.Context,
		tx DBTX,
		entries []*record.UserRecordEntry,
	) error
}

// RecordListCondition 書誌一覧の検索条件
//
// 検索条件(ContributorName / Title / PublishYear)はnilなら条件に含めず、指定されたものはANDで組み合わせる。
// 値の妥当性(並び替えキー・並び順・limitの範囲など)は呼び出し側で検証済みであることを前提とする
// (未知の並び替えキー・並び順はpanicする)。
type RecordListCondition struct {
	// 内部用ユーザID。自ユーザの書誌だけを対象にする
	InternalUid int64
	// 貢献者名。書誌または収録物の貢献者名の部分一致
	ContributorName *string
	// タイトル。書誌または収録物の主題・副題の部分一致
	Title *string
	// 出版年。書誌の出版日の年が一致するもの(出版日が未登録の書誌は含まない)
	PublishYear *int
	// 並び替えキー
	SortKey record.SortKey
	// 並び順
	SortOrder record.SortOrder
	// 取得件数
	Limit int
	// 取得開始位置
	Offset int
}
