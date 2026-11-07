package db

// 書誌情報repositoryテストの共通セットアップとヘルパ
// 各メソッド群のテストは record_repository_<メソッド群>_test.go に置く

import (
	"context"
	"testing"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/testutil"

	"github.com/stretchr/testify/require"
)

// seedPath 書誌タイプ/役割/入力項目のマスタデータ。本番と同じseedを使う
const seedPath = "../../../sql/seed.sql"

const (
	// recordFixture 書誌情報登録テスト用フィクスチャ
	recordFixture = "testdata/records_create_001.sql"
	// testUid フィクスチャの登録対象ユーザ
	testUid = int64(1001)
	// testExistingContainerId フィクスチャで登録ずみの収録物(雑誌)のID
	testExistingContainerId = "EXISTINGCONTAINER000000001"
	// testRecordId テストで登録する書誌のID
	testRecordId = "TESTRECORD0000000000000001"
	// testContainerId テストで登録する収録物のID
	testContainerId = "TESTCONTAINER00000000000001"
)

// testRecordsSetup テーブルを初期化し、マスタとフィクスチャを投入したrepositoryを返す
func testRecordsSetup(t *testing.T, fixturePath string) *RecordRepositoryImpl {
	t.Helper()

	testutil.CleanUpTables(t, testDb)

	err := testutil.LoadFixture(testDb, seedPath)
	require.NoError(t, err)

	if fixturePath != "" {
		err = testutil.LoadFixture(testDb, fixturePath)
		require.NoError(t, err)
	}

	return NewRecordRepositoryImpl(testDb)
}

// newRecordTestCtx 書誌情報登録APIのAPI IDを保持するcontextを返す
// 本番では*gin.Context経由で渡るが、Value()の挙動は同じ。
func newRecordTestCtx(apiId string) context.Context {
	return context.WithValue(
		context.Background(),
		ctxutil.CtxKeyApiId,
		apiId,
	)
}

// strPtr string型のポインタを返却する
func strPtr(s string) *string {
	return &s
}

// newMinimalRelation 貢献者と貢献のペアを生成する
// 書誌ID・内部用ユーザID・表示順はrepositoryが収録物と登録順から設定する。Ph0では氏名と役割のみを指定。
func newMinimalRelation(rawName string, role record.ContributorRole) record.ContributorRelation {
	return record.ContributorRelation{
		Contributor:  record.ContributorEntry{RawName: rawName, DelFlg: 0},
		Contribution: record.ContributionEntry{Role: role, DelFlg: 0},
	}
}

// newMinimalRecordEntry 必須項目のみを設定したユーザ入力書誌を返す
func newMinimalRecordEntry(recordId string) record.UserRecordEntry {
	return record.UserRecordEntry{
		RecordEntryId: recordId,
		InternalUid:   testUid,
		RecordType:    record.RecordTypeAuthoredBook,
		MainTitle:     "単著書のタイトル",
		DelFlg:        0,
	}
}
