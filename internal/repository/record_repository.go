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
}
