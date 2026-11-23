package usecase

import (
	"jemref_go/internal/domain/record"
)

// RecordUsecaseImplのテストは、メソッドごとに record_usecase_<メソッド>_test.go に置く。
// このファイルには、それらが共有する定数とヘルパを置く。

const (
	// recordTestUid テスト用の内部ユーザID
	recordTestUid = int64(1001)
	// recordArticleId idGenが2回目(コンテナなしの場合は1回目)に生成するID
	recordArticleId = "ARTICLEID00000000000000001"
	// recordContainerId idGenが1回目に生成するID
	recordContainerId = "CONTAINERID000000000000001"
)

// entryIds Create/Updateに渡された書誌のIDを呼び出し順に返す
func entryIds(entries []record.UserRecordEntry) []string {
	if entries == nil {
		return nil
	}

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.RecordEntryId)
	}

	return ids
}
