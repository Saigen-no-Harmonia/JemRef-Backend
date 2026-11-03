package db

import (
	"context"
	"database/sql"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
func newRecordTestCtx() context.Context {
	return context.WithValue(
		context.Background(),
		ctxutil.CtxKeyApiId,
		ctxutil.REF_API_003,
	)
}

// strPtr string型のポインタを返却する
func strPtr(s string) *string {
	return &s
}

// newMinimalRelation 貢献者と貢献のペアを生成する
// 書誌ID・内部用ユーザID・表示順はrepositoryが親レコードと登録順から設定する。Ph0では氏名と役割のみを指定。
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

// -------------Create(context.Context, repository.DBTX, record.UserRecordEntry)のテスト-------------

func TestRecordRepository_Create_正常_必須項目のみ(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)
	now := time.Now().Truncate(time.Microsecond)
	input := newMinimalRecordEntry(testRecordId)

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)
	assert.Equal(t, 1, testutil.CountRecords(t, testDb)-beforeCnt)

	actual, err := testutil.SelectRecordByPk(t, testDb, testRecordId)
	require.NoError(t, err)

	assert.Equal(t, testRecordId, actual.Id)
	assert.Equal(t, testUid, actual.UserId)
	assert.Equal(t, record.RecordTypeIdAuthoredBook, actual.TypeId)
	assert.Equal(t, "単著書のタイトル", actual.MainTitle)

	// 任意項目は全てNULLで登録されること
	assert.False(t, actual.SubTitle.Valid)
	assert.False(t, actual.PageRange.Valid)
	assert.False(t, actual.Volume.Valid)
	assert.False(t, actual.Issue.Valid)
	assert.False(t, actual.Publisher.Valid)
	assert.False(t, actual.PublicationDate.Valid)
	assert.False(t, actual.DatePrecision.Valid)
	assert.False(t, actual.ReadStatus.Valid)
	assert.False(t, actual.ParentId.Valid)
	assert.False(t, actual.MRecordId.Valid)

	assert.Equal(t, 0, actual.DelFlg)
	assert.Equal(t, ctxutil.REF_API_003, actual.InsPg)
	assert.Equal(t, "system", actual.InsId)
	assert.Equal(t, ctxutil.REF_API_003, actual.UpdPg)
	assert.Equal(t, "system", actual.UpdId)
	assert.WithinDuration(t, now, actual.InsDt, 3*time.Second)
	assert.WithinDuration(t, now, actual.UpdDt, 3*time.Second)

	// 子要素は1件も登録されないこと
	assert.Equal(t, 0, testutil.CountContributors(t, testDb))
	assert.Equal(t, 0, testutil.CountContributions(t, testDb))
	assert.Equal(t, 0, testutil.CountUrls(t, testDb))
	assert.Equal(t, 0, testutil.CountMemos(t, testDb))
	assert.Equal(t, 0, testutil.CountFieldValues(t, testDb))
}

func TestRecordRepository_Create_正常_全項目設定(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionDay
	readStatus := record.ReadStatusRead

	input := record.UserRecordEntry{
		RecordEntryId:   testRecordId,
		InternalUid:     testUid,
		RecordType:      record.RecordTypeJournalArticle,
		MainTitle:       "第1章 「学校」制度の境界線",
		SubTitle:        strPtr("その形成と展開"),
		PageRange:       strPtr("15-46"),
		Volume:          strPtr("日本教育史講座第1巻"),
		Issue:           strPtr("第11号"),
		Publisher:       strPtr("東京大学出版会"),
		PublicationDate: &publicationDate,
		DatePrecision:   &datePrecision,
		ReadStatus:      &readStatus,
		ParentId:        strPtr(testExistingContainerId),
		DelFlg:          0,
		ContributorRelations: []record.ContributorRelation{
			newMinimalRelation("木村元", record.ContributorRoleAuthor),
			newMinimalRelation("山田太郎", record.ContributorRoleCompiler),
		},
		Urls: []record.UrlEntry{
			{RecordEntryId: testRecordId, DisplayNumber: 0, Label: strPtr("PDF"), Url: "https://example.com/1", InternalUid: testUid, DelFlg: 0},
			{RecordEntryId: testRecordId, DisplayNumber: 1, Label: strPtr("書影"), Url: "https://example.com/2", InternalUid: testUid, DelFlg: 0},
		},
		Memos: []record.MemoEntry{
			{RecordEntryId: testRecordId, DisplayNumber: 0, Title: strPtr("要約"), Content: "学校の戦後史。", InternalUid: testUid, DelFlg: 0},
			{RecordEntryId: testRecordId, DisplayNumber: 1, Title: strPtr("所感"), Content: "少し難しい。", InternalUid: testUid, DelFlg: 0},
		},
		OptionValues: []record.OptionValue{
			{RecordEntryId: testRecordId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "9784130513555", InternalUid: testUid, DelFlg: 0},
			{RecordEntryId: testRecordId, RecordField: record.RecordFieldEdition, DisplayNumber: 20, RawValue: "第2版", InternalUid: testUid, DelFlg: 0},
			{RecordEntryId: testRecordId, RecordField: record.RecordFieldOriginalTitle, DisplayNumber: 30, RawValue: "School History of Boundary", InternalUid: testUid, DelFlg: 0},
		},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, testRecordId)
	require.NoError(t, err)

	assert.Equal(t, record.RecordTypeIdJournalArticle, actual.TypeId)
	assert.Equal(t, input.MainTitle, actual.MainTitle)
	assert.Equal(t, *input.SubTitle, actual.SubTitle.String)
	assert.Equal(t, *input.PageRange, actual.PageRange.String)
	assert.Equal(t, *input.Volume, actual.Volume.String)
	assert.Equal(t, *input.Issue, actual.Issue.String)
	assert.Equal(t, *input.Publisher, actual.Publisher.String)
	assert.WithinDuration(t, publicationDate, actual.PublicationDate.Time, time.Microsecond)
	assert.Equal(t, int64(record.DatePrecisionDay), actual.DatePrecision.Int64)
	assert.Equal(t, int64(record.ReadStatusIdRead), actual.ReadStatus.Int64)
	assert.Equal(t, testExistingContainerId, actual.ParentId.String)
	assert.False(t, actual.MRecordId.Valid)
	assert.Equal(t, 0, actual.DelFlg)

	// 子要素の件数
	assert.Equal(t, 2, testutil.CountContributors(t, testDb))
	assert.Equal(t, 2, testutil.CountContributions(t, testDb))
	assert.Equal(t, 2, testutil.CountUrls(t, testDb))
	assert.Equal(t, 2, testutil.CountMemos(t, testDb))
	assert.Equal(t, 3, testutil.CountFieldValues(t, testDb))

	// オプション書誌項目値は入力項目IDに変換されて登録されること
	fieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, testRecordId)
	require.Len(t, fieldValues, 3)
	assert.Equal(t, record.RecordFieldIdIsbn, fieldValues[0].RecordFieldId)
	assert.Equal(t, "9784130513555", fieldValues[0].RawValue)
	assert.Equal(t, 10, fieldValues[0].DisplayNumber)
	assert.Equal(t, record.RecordFieldIdEdition, fieldValues[1].RecordFieldId)
	assert.Equal(t, 20, fieldValues[1].DisplayNumber)
	assert.Equal(t, record.RecordFieldIdOriginalTitle, fieldValues[2].RecordFieldId)
	assert.Equal(t, 30, fieldValues[2].DisplayNumber)
	assert.Equal(t, testUid, fieldValues[0].UserId)
	assert.Equal(t, ctxutil.REF_API_003, fieldValues[0].InsPg)
	assert.Equal(t, "system", fieldValues[0].InsId)
}

func TestRecordRepository_Create_正常_子要素が空スライス(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.ContributorRelations = []record.ContributorRelation{}
	input.Urls = []record.UrlEntry{}
	input.Memos = []record.MemoEntry{}
	input.OptionValues = []record.OptionValue{}

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)
	assert.Equal(t, 1, testutil.CountRecords(t, testDb)-beforeCnt)

	assert.Equal(t, 0, testutil.CountContributors(t, testDb))
	assert.Equal(t, 0, testutil.CountContributions(t, testDb))
	assert.Equal(t, 0, testutil.CountUrls(t, testDb))
	assert.Equal(t, 0, testutil.CountMemos(t, testDb))
	assert.Equal(t, 0, testutil.CountFieldValues(t, testDb))
}

func TestRecordRepository_Create_正常_出版日精度_年のみ(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	publicationDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionYear

	input := newMinimalRecordEntry(testRecordId)
	input.PublicationDate = &publicationDate
	input.DatePrecision = &datePrecision

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, testRecordId)
	require.NoError(t, err)
	assert.WithinDuration(t, publicationDate, actual.PublicationDate.Time, time.Microsecond)
	assert.Equal(t, int64(record.DatePrecisionYear), actual.DatePrecision.Int64)
}

func TestRecordRepository_Create_正常_貢献者複数(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.ContributorRelations = []record.ContributorRelation{
		newMinimalRelation("貢献者A", record.ContributorRoleAuthor),
		newMinimalRelation("貢献者B", record.ContributorRoleInterpreter),
		newMinimalRelation("貢献者C", record.ContributorRoleCompiler),
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	require.Len(t, contributors, 3)
	for i, c := range contributors {
		assert.Equal(t, input.ContributorRelations[i].Contributor.RawName, c.RawName)
		assert.Equal(t, testUid, c.UserId)
		assert.False(t, c.MContributorId.Valid)
		assert.Equal(t, 0, c.DelFlg)
		assert.Equal(t, ctxutil.REF_API_003, c.InsPg)
		assert.Equal(t, "system", c.InsId)
	}

	contributions := testutil.SelectContributionsByRecordId(t, testDb, testRecordId)
	require.Len(t, contributions, 3)

	expectedRoleIds := []string{
		record.ContributorRoleIdAuthor,
		record.ContributorRoleIdInterpreter,
		record.ContributorRoleIdCompiler,
	}
	for i, c := range contributions {
		assert.Equal(t, testRecordId, c.RecordId)
		assert.Equal(t, i, c.DisplayNumber)
		assert.Equal(t, expectedRoleIds[i], c.RoleId)
		// 名前と役割の組み合わせが入れ替わらないこと(採番されたIDとの対応)
		assert.Equal(t, contributors[i].Id, c.ContributorId)
		assert.Equal(t, 0, c.DelFlg)
	}
}

func TestRecordRepository_Create_正常_貢献者の派生項目が親レコードから設定される(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// 氏名と役割のみを指定する(書誌ID/ユーザID/表示順は未設定)
	input := newMinimalRecordEntry(testRecordId)
	input.ContributorRelations = []record.ContributorRelation{
		newMinimalRelation("貢献者A", record.ContributorRoleAuthor),
		newMinimalRelation("貢献者B", record.ContributorRoleCompiler),
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	// 内部用ユーザIDが親レコードから設定されること
	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	require.Len(t, contributors, 2)
	for _, c := range contributors {
		assert.Equal(t, testUid, c.UserId)
	}

	// 書誌IDが親レコードから、表示順が登録順から設定されること
	contributions := testutil.SelectContributionsByRecordId(t, testDb, testRecordId)
	require.Len(t, contributions, 2)
	for i, c := range contributions {
		assert.Equal(t, testRecordId, c.RecordId)
		assert.Equal(t, i, c.DisplayNumber)
	}
}

func TestRecordRepository_Create_正常_同名貢献者は別レコードで登録される(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.ContributorRelations = []record.ContributorRelation{
		newMinimalRelation("同姓同名", record.ContributorRoleAuthor),
		newMinimalRelation("同姓同名", record.ContributorRoleCompiler),
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	// Ph0では名寄せを行わないため、同名でも別レコードとして登録される
	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	require.Len(t, contributors, 2)
	assert.Equal(t, "同姓同名", contributors[0].RawName)
	assert.Equal(t, "同姓同名", contributors[1].RawName)
	assert.NotEqual(t, contributors[0].Id, contributors[1].Id)

	contributions := testutil.SelectContributionsByRecordId(t, testDb, testRecordId)
	require.Len(t, contributions, 2)
	assert.NotEqual(t, contributions[0].ContributorId, contributions[1].ContributorId)
}

func TestRecordRepository_Create_正常_URLのラベルあり無し混在(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.Urls = []record.UrlEntry{
		{RecordEntryId: testRecordId, DisplayNumber: 0, Label: strPtr("オンラインPDF"), Url: "https://example.com/1", InternalUid: testUid, DelFlg: 0},
		{RecordEntryId: testRecordId, DisplayNumber: 1, Label: nil, Url: "https://example.com/2", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	urls := testutil.SelectUrlsByRecordId(t, testDb, testRecordId)
	require.Len(t, urls, 2)

	assert.Equal(t, 0, urls[0].DisplayNumber)
	assert.True(t, urls[0].RawLabel.Valid)
	assert.Equal(t, "オンラインPDF", urls[0].RawLabel.String)
	assert.Equal(t, "https://example.com/1", urls[0].RawUrl)
	assert.Equal(t, testUid, urls[0].UserId)
	assert.Equal(t, 0, urls[0].DelFlg)
	assert.Equal(t, ctxutil.REF_API_003, urls[0].InsPg)
	assert.Equal(t, "system", urls[0].InsId)

	assert.Equal(t, 1, urls[1].DisplayNumber)
	// ラベル未設定はNULLで登録されること
	assert.False(t, urls[1].RawLabel.Valid)
	assert.Equal(t, "https://example.com/2", urls[1].RawUrl)
}

func TestRecordRepository_Create_正常_メモのラベルあり無し混在(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.Memos = []record.MemoEntry{
		{RecordEntryId: testRecordId, DisplayNumber: 0, Title: strPtr("要約"), Content: "学校の戦後史。", InternalUid: testUid, DelFlg: 0},
		{RecordEntryId: testRecordId, DisplayNumber: 1, Title: nil, Content: "ラベルなしのメモ。", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	memos := testutil.SelectMemosByRecordId(t, testDb, testRecordId)
	require.Len(t, memos, 2)

	assert.Equal(t, 0, memos[0].DisplayNumber)
	assert.True(t, memos[0].Title.Valid)
	assert.Equal(t, "要約", memos[0].Title.String)
	assert.Equal(t, "学校の戦後史。", memos[0].Content)
	assert.Equal(t, testUid, memos[0].UserId)
	// 公開許諾レベルはPh0では0固定(DDLのデフォルト値)
	assert.Equal(t, 0, memos[0].ShareLevel)
	assert.Equal(t, 0, memos[0].DelFlg)

	assert.Equal(t, 1, memos[1].DisplayNumber)
	// 見出し未設定はNULLで登録されること
	assert.False(t, memos[1].Title.Valid)
	assert.Equal(t, "ラベルなしのメモ。", memos[1].Content)
}

func TestRecordRepository_Create_正常_オプション書誌項目値が一部のみ(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.OptionValues = []record.OptionValue{
		{RecordEntryId: testRecordId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "9784130513555", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	fieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, testRecordId)
	require.Len(t, fieldValues, 1)
	assert.Equal(t, record.RecordFieldIdIsbn, fieldValues[0].RecordFieldId)
	assert.Equal(t, 10, fieldValues[0].DisplayNumber)
	assert.Equal(t, "9784130513555", fieldValues[0].RawValue)
	assert.Equal(t, testUid, fieldValues[0].UserId)
	assert.Equal(t, 0, fieldValues[0].DelFlg)
}

func TestRecordRepository_Create_正常_コンテナ付き親子(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)
	ctx := newRecordTestCtx()

	container := record.UserRecordEntry{
		RecordEntryId: testContainerId,
		InternalUid:   testUid,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "境界線の学校史",
		DelFlg:        0,
	}

	article := record.UserRecordEntry{
		RecordEntryId: testRecordId,
		InternalUid:   testUid,
		RecordType:    record.RecordTypeJournalArticle,
		MainTitle:     "第1章 「学校」制度の境界線",
		ParentId:      strPtr(testContainerId),
		DelFlg:        0,
	}

	beforeCnt := testutil.CountRecords(t, testDb)

	// コンテナ -> 記事の順に登録する(記事のPARENT_IDがFK制約を満たすため)
	err := repo.Create(ctx, testDb, container)
	require.NoError(t, err)
	err = repo.Create(ctx, testDb, article)
	require.NoError(t, err)

	assert.Equal(t, 2, testutil.CountRecords(t, testDb)-beforeCnt)

	actualContainer, err := testutil.SelectRecordByPk(t, testDb, testContainerId)
	require.NoError(t, err)
	assert.Equal(t, record.RecordTypeIdJournal, actualContainer.TypeId)
	assert.False(t, actualContainer.ParentId.Valid)

	actualArticle, err := testutil.SelectRecordByPk(t, testDb, testRecordId)
	require.NoError(t, err)
	assert.Equal(t, record.RecordTypeIdJournalArticle, actualArticle.TypeId)
	assert.True(t, actualArticle.ParentId.Valid)
	assert.Equal(t, actualContainer.Id, actualArticle.ParentId.String)
}

func TestRecordRepository_Create_正常_貢献者が上限件数(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// DISPLAY_NUMBERはTINYINTなので0起点で128件(0..127)まで登録できる
	const limit = 128

	input := newMinimalRecordEntry(testRecordId)
	for i := 0; i < limit; i++ {
		input.ContributorRelations = append(
			input.ContributorRelations,
			newMinimalRelation("貢献者"+strconv.Itoa(i), record.ContributorRoleAuthor),
		)
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.NoError(t, err)

	assert.Equal(t, limit, testutil.CountContributors(t, testDb))

	contributions := testutil.SelectContributionsByRecordId(t, testDb, testRecordId)
	require.Len(t, contributions, limit)
	assert.Equal(t, 0, contributions[0].DisplayNumber)
	assert.Equal(t, limit-1, contributions[limit-1].DisplayNumber)
}

func TestRecordRepository_Create_異常_書誌IDが重複(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// フィクスチャで登録ずみのIDを指定する
	input := newMinimalRecordEntry(testExistingContainerId)

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert record entry")
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
}

func TestRecordRepository_Create_異常_存在しないユーザID(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.InternalUid = int64(9999)

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert record entry")
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
}

func TestRecordRepository_Create_異常_存在しない親ID(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	input := newMinimalRecordEntry(testRecordId)
	input.RecordType = record.RecordTypeJournalArticle
	input.ParentId = strPtr("NOTEXISTSCONTAINER00000001")

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert record entry")
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
}

func TestRecordRepository_Create_異常_主題が長さ超過(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// MAIN_TITLEはVARCHAR(1024)
	input := newMinimalRecordEntry(testRecordId)
	input.MainTitle = strings.Repeat("あ", 1025)

	beforeCnt := testutil.CountRecords(t, testDb)
	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert record entry")
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
}

func TestRecordRepository_Create_異常_URLの表示順が重複(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// uk_url_order (E_RECORD_ID, DISPLAY_NUMBER) 違反
	input := newMinimalRecordEntry(testRecordId)
	input.Urls = []record.UrlEntry{
		{RecordEntryId: testRecordId, DisplayNumber: 0, Url: "https://example.com/1", InternalUid: testUid, DelFlg: 0},
		{RecordEntryId: testRecordId, DisplayNumber: 0, Url: "https://example.com/2", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert urls")
	assert.Equal(t, 0, testutil.CountUrls(t, testDb))
}

func TestRecordRepository_Create_異常_メモの表示順が重複(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// uk_memo_order (E_RECORD_ID, DISPLAY_NUMBER) 違反
	input := newMinimalRecordEntry(testRecordId)
	input.Memos = []record.MemoEntry{
		{RecordEntryId: testRecordId, DisplayNumber: 0, Content: "メモ1", InternalUid: testUid, DelFlg: 0},
		{RecordEntryId: testRecordId, DisplayNumber: 0, Content: "メモ2", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert memo")
	assert.Equal(t, 0, testutil.CountMemos(t, testDb))
}

func TestRecordRepository_Create_異常_オプション書誌項目値の表示順が重複(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)

	// uk_field_value (E_RECORD_ID, RECORD_FIELD_ID, DISPLAY_NUMBER) 違反
	input := newMinimalRecordEntry(testRecordId)
	input.OptionValues = []record.OptionValue{
		{RecordEntryId: testRecordId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "1111111111111", InternalUid: testUid, DelFlg: 0},
		{RecordEntryId: testRecordId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "2222222222222", InternalUid: testUid, DelFlg: 0},
	}

	err := repo.Create(newRecordTestCtx(), testDb, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert option value")
	assert.Equal(t, 0, testutil.CountFieldValues(t, testDb))
}

func TestRecordRepository_Create_異常_トランザクション内でロールバックされる(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)
	txManager := NewTxManager(testDb)
	ctx := newRecordTestCtx()

	container := record.UserRecordEntry{
		RecordEntryId: testContainerId,
		InternalUid:   testUid,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "境界線の学校史",
		DelFlg:        0,
		ContributorRelations: []record.ContributorRelation{
			newMinimalRelation("貢献者A", record.ContributorRoleCompiler),
		},
	}

	// 記事側はIDが重複しており必ず失敗する
	article := newMinimalRecordEntry(testExistingContainerId)

	beforeCnt := testutil.CountRecords(t, testDb)

	err := txManager.WithTransaction(
		ctx,
		func(ctx context.Context, tx repository.DBTX) error {
			if err := repo.Create(ctx, tx, container); err != nil {
				return err
			}
			return repo.Create(ctx, tx, article)
		},
	)
	require.Error(t, err)

	// コンテナ側の登録もロールバックされていること
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
	assert.Equal(t, 0, testutil.CountContributors(t, testDb))
	assert.Equal(t, 0, testutil.CountContributions(t, testDb))

	_, err = testutil.SelectRecordByPk(t, testDb, testContainerId)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
