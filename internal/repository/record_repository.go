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
	// 返却するのは RecordEntryId と ParentId のみ(全カラムは取得しない)。
	SelectByPk(
		ctx context.Context,
		tx DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error)

	// CountActiveChildren 指定した書誌を親とする、論理削除されていない書誌の件数を返す
	CountActiveChildren(
		ctx context.Context,
		tx DBTX,
		parentRecordId string,
	) (int, error)

	// LogicalDelete 書誌とその子要素(貢献/URL/メモ/項目値)を論理削除する
	//
	// e_contributors はユーザ単位の入力辞書であり書誌に属さないため削除しない。
	LogicalDelete(
		ctx context.Context,
		tx DBTX,
		recordId string,
	) error
}
