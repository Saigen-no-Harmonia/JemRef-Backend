package db

import (
	"context"
	"testing"
	"time"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------参照系のテスト-------------

const (
	// recordGetFixture 書誌詳細参照テスト用フィクスチャ
	recordGetFixture = "testdata/records_get_001.sql"

	// 以下はフィクスチャ(records_get_001.sql)で投入される書誌ID
	getContainerId     = "GET_CONTAINER_00000000001" // 収録物(雑誌)。構成要素あり
	getFullId          = "GET_FULL_00000000000001"   // 全カラムに値が入った書誌。収録物あり
	getMinimalId       = "GET_MINIMAL_0000000001"    // 必須項目のみ。構成要素なし
	getEmptyStrId      = "GET_EMPTY_STR_000000001"   // 任意項目が空文字
	getOrderId         = "GET_ORDER_000000000001"    // 構成要素を表示順の逆順で登録
	getDelComponentId  = "GET_DEL_COMPONENT_00001"   // 構成要素に生存/論理削除済みが混在
	getDelContribId    = "GET_DEL_CONTRIB_000001"    // 貢献者本体が論理削除済み
	getNullComponentId = "GET_NULL_COMPONENT_0001"   // 構成要素の任意項目がNULL
	getOtherUserId     = "GET_OTHER_USER_0000001"    // 他ユーザの書誌
	getDeletedId       = "GET_DELETED_0000000001"    // 論理削除済みの書誌
	getNotExistsId     = "GET_NOTEXISTS_000000001"   // 存在しない書誌ID
)

// newRecordGetTestCtx 書誌詳細参照APIのAPI IDを保持するcontextを返す
func newRecordGetTestCtx() context.Context {
	return newRecordTestCtx(ctxutil.REF_API_002)
}

// -------------SelectByPk(context.Context, repository.DBTX, string, int64)のテスト-------------

func TestRecordRepository_SelectByPk_正常_全カラムを取得(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getFullId, testUid,
	)
	require.NoError(t, err)

	assert.Equal(t, getFullId, actual.RecordEntryId)
	assert.Equal(t, testUid, actual.InternalUid)
	assert.Equal(t, "第1章 「学校」制度の境界線", actual.MainTitle)
	assert.Equal(t, strPtr("その形成と展開"), actual.SubTitle)
	assert.Equal(t, strPtr("15-46"), actual.PageRange)
	assert.Equal(t, strPtr("第1巻"), actual.Volume)
	assert.Equal(t, strPtr("第11号"), actual.Issue)
	assert.Equal(t, strPtr("東京大学出版会"), actual.Publisher)
	assert.Equal(t, 0, actual.DelFlg)

	// 書誌マスタ(m_records)はseedに存在しないためNULL固定
	assert.Nil(t, actual.MRecordId)

	// TYPE_ID("RT000012")から書誌タイプへ復元されること
	assert.Equal(t, record.RecordTypeJournalArticle, actual.RecordType)

	// READ_STATUS(1)から読了ステータスへ復元されること
	require.NotNil(t, actual.ReadStatus)
	assert.Equal(t, record.ReadStatusRead, *actual.ReadStatus)

	// 出版日と精度はそのまま取得する(文字列への整形はhandlerの責務)
	require.NotNil(t, actual.PublicationDate)
	expectedDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	assert.WithinDuration(t, expectedDate, *actual.PublicationDate, time.Microsecond)

	require.NotNil(t, actual.DatePrecision)
	assert.Equal(t, record.DatePrecisionDay, *actual.DatePrecision)
}

func TestRecordRepository_SelectByPk_正常_コンテナありの書誌を取得(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getFullId, testUid,
	)
	require.NoError(t, err)

	require.NotNil(t, actual.ContainerId)
	assert.Equal(t, getContainerId, *actual.ContainerId)
}

func TestRecordRepository_SelectByPk_正常_任意項目がNULL(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getMinimalId, testUid,
	)
	require.NoError(t, err)

	// 必須項目は取得できている
	assert.Equal(t, getMinimalId, actual.RecordEntryId)
	assert.Equal(t, "必須のみの単著書", actual.MainTitle)
	assert.Equal(t, record.RecordTypeAuthoredBook, actual.RecordType)

	// 任意項目はすべてnil(空文字やゼロ値ではない)
	assert.Nil(t, actual.SubTitle)
	assert.Nil(t, actual.PageRange)
	assert.Nil(t, actual.Volume)
	assert.Nil(t, actual.Issue)
	assert.Nil(t, actual.Publisher)
	assert.Nil(t, actual.PublicationDate)
	assert.Nil(t, actual.DatePrecision)
	assert.Nil(t, actual.ReadStatus)
	assert.Nil(t, actual.ContainerId)
	assert.Nil(t, actual.MRecordId)
}

// TestRecordRepository_SelectByPk_正常_任意項目が空文字
// DBの空文字はそのまま空文字のポインタで返す(NULL化はhandlerの責務)。
func TestRecordRepository_SelectByPk_正常_任意項目が空文字(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getEmptyStrId, testUid,
	)
	require.NoError(t, err)

	assert.Equal(t, strPtr(""), actual.SubTitle)
	assert.Equal(t, strPtr(""), actual.PageRange)
	assert.Equal(t, strPtr(""), actual.Volume)
	assert.Equal(t, strPtr(""), actual.Issue)
	assert.Equal(t, strPtr(""), actual.Publisher)
}

// TestRecordRepository_SelectByPk_正常_構成要素は取得しない
// 構成要素の取得は SelectComponents の責務であり、SelectByPk では設定しない。
func TestRecordRepository_SelectByPk_正常_構成要素は取得しない(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	// 構成要素を4種類とも持つ書誌を指定する
	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getFullId, testUid,
	)
	require.NoError(t, err)

	assert.Nil(t, actual.ContributorRelations)
	assert.Nil(t, actual.Urls)
	assert.Nil(t, actual.Memos)
	assert.Nil(t, actual.OptionValues)
}

func TestRecordRepository_SelectByPk_異常_存在しない書誌ID(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getNotExistsId, testUid,
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

func TestRecordRepository_SelectByPk_異常_他ユーザの書誌(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	// 他ユーザ(1002)の書誌を自ユーザ(1001)で指定する
	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getOtherUserId, testUid,
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

func TestRecordRepository_SelectByPk_異常_論理削除済みの書誌(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPk(
		newRecordGetTestCtx(), testDb, getDeletedId, testUid,
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

// -------------SelectComponents(context.Context, repository.DBTX, []*record.UserRecordEntry)のテスト-------------

// newGetTargetEntry SelectComponentsに渡す、構成要素が未設定の書誌を返す
func newGetTargetEntry(recordId string) *record.UserRecordEntry {
	return &record.UserRecordEntry{
		RecordEntryId: recordId,
		InternalUid:   testUid,
	}
}

func TestRecordRepository_SelectComponents_正常_全種類の構成要素を取得(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getFullId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	// 貢献者: 氏名はe_contributorsとのJOINで取得する
	require.Len(t, target.ContributorRelations, 2)

	first := target.ContributorRelations[0]
	assert.Equal(t, "木村元", first.Contributor.RawName)
	assert.Equal(t, testUid, first.Contributor.InternalUid)
	assert.Nil(t, first.Contributor.MContributorId)
	assert.NotZero(t, first.Contributor.ContributorEntryId)
	assert.Equal(t, getFullId, first.Contribution.RecordEntryId)
	// ROLE_ID("RL000010")から役割へ復元されること
	assert.Equal(t, record.ContributorRoleAuthor, first.Contribution.Role)
	assert.Equal(t, 0, first.Contribution.DisplayNumber)
	assert.NotZero(t, first.Contribution.ContributionEntryId)
	// 貢献は貢献者を指していること
	assert.Equal(t, first.Contributor.ContributorEntryId, first.Contribution.ContributorEntryId)

	second := target.ContributorRelations[1]
	assert.Equal(t, "訳者B", second.Contributor.RawName)
	assert.Equal(t, record.ContributorRoleInterpreter, second.Contribution.Role)
	assert.Equal(t, 1, second.Contribution.DisplayNumber)

	// URL
	require.Len(t, target.Urls, 2)
	assert.Equal(t, getFullId, target.Urls[0].RecordEntryId)
	assert.Equal(t, strPtr("オンラインPDF"), target.Urls[0].Label)
	assert.Equal(t, "https://example.com/pdf", target.Urls[0].Url)
	assert.Equal(t, 0, target.Urls[0].DisplayNumber)
	assert.Equal(t, testUid, target.Urls[0].InternalUid)
	assert.NotZero(t, target.Urls[0].UrlEntryId)
	assert.Equal(t, strPtr("書誌情報"), target.Urls[1].Label)

	// メモ
	require.Len(t, target.Memos, 2)
	assert.Equal(t, getFullId, target.Memos[0].RecordEntryId)
	assert.Equal(t, strPtr("要約"), target.Memos[0].Title)
	assert.Equal(t, "学校の戦後史。", target.Memos[0].Content)
	assert.Equal(t, 0, target.Memos[0].DisplayNumber)
	assert.Equal(t, 0, target.Memos[0].ShareLevel)
	assert.Equal(t, testUid, target.Memos[0].InternalUid)
	assert.NotZero(t, target.Memos[0].MemoEntryId)
	assert.Equal(t, strPtr("感想"), target.Memos[1].Title)

	// 項目値: RECORD_FIELD_IDから入力項目へ復元されること(3種類すべて)
	require.Len(t, target.OptionValues, 3)
	assert.Equal(t, getFullId, target.OptionValues[0].RecordEntryId)
	assert.Equal(t, record.RecordFieldIsbn, target.OptionValues[0].RecordField)
	assert.Equal(t, "9784130513555", target.OptionValues[0].RawValue)
	assert.Equal(t, 10, target.OptionValues[0].DisplayNumber)
	assert.Equal(t, testUid, target.OptionValues[0].InternalUid)
	assert.NotZero(t, target.OptionValues[0].OptionValueEntryId)
	assert.Equal(t, record.RecordFieldEdition, target.OptionValues[1].RecordField)
	assert.Equal(t, "第2版", target.OptionValues[1].RawValue)
	assert.Equal(t, record.RecordFieldOriginalTitle, target.OptionValues[2].RecordField)
	assert.Equal(t, "School History of Boundary", target.OptionValues[2].RawValue)
}

// TestRecordRepository_SelectComponents_正常_表示順で並ぶ
// フィクスチャは表示順の逆順でINSERTしており、登録順ではなくDISPLAY_NUMBERで並ぶことを検証する。
func TestRecordRepository_SelectComponents_正常_表示順で並ぶ(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getOrderId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	require.Len(t, target.ContributorRelations, 3)
	assert.Equal(t, "一番目の貢献者", target.ContributorRelations[0].Contributor.RawName)
	assert.Equal(t, "二番目の貢献者", target.ContributorRelations[1].Contributor.RawName)
	assert.Equal(t, "三番目の貢献者", target.ContributorRelations[2].Contributor.RawName)

	require.Len(t, target.Urls, 3)
	assert.Equal(t, strPtr("一番目のURL"), target.Urls[0].Label)
	assert.Equal(t, strPtr("二番目のURL"), target.Urls[1].Label)
	assert.Equal(t, strPtr("三番目のURL"), target.Urls[2].Label)

	require.Len(t, target.Memos, 3)
	assert.Equal(t, strPtr("一番目のメモ"), target.Memos[0].Title)
	assert.Equal(t, strPtr("二番目のメモ"), target.Memos[1].Title)
	assert.Equal(t, strPtr("三番目のメモ"), target.Memos[2].Title)

	require.Len(t, target.OptionValues, 3)
	assert.Equal(t, "一番目の項目値", target.OptionValues[0].RawValue)
	assert.Equal(t, "二番目の項目値", target.OptionValues[1].RawValue)
	assert.Equal(t, "三番目の項目値", target.OptionValues[2].RawValue)
}

func TestRecordRepository_SelectComponents_正常_論理削除済みの構成要素を除外(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getDelComponentId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	require.Len(t, target.ContributorRelations, 1)
	assert.Equal(t, "生存の貢献者", target.ContributorRelations[0].Contributor.RawName)

	require.Len(t, target.Urls, 1)
	assert.Equal(t, strPtr("生存のURL"), target.Urls[0].Label)

	require.Len(t, target.Memos, 1)
	assert.Equal(t, strPtr("生存のメモ"), target.Memos[0].Title)

	require.Len(t, target.OptionValues, 1)
	assert.Equal(t, "生存の項目値", target.OptionValues[0].RawValue)
}

// TestRecordRepository_SelectComponents_正常_貢献者本体が論理削除済み
// e_record_contributions は生存していても、e_contributors が論理削除済みなら除外する
// (論理削除済みの構成要素を除外するケースとは別の分岐)。
func TestRecordRepository_SelectComponents_正常_貢献者本体が論理削除済み(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getDelContribId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	require.Len(t, target.ContributorRelations, 1)
	assert.Equal(t, "生存の貢献者2", target.ContributorRelations[0].Contributor.RawName)
}

func TestRecordRepository_SelectComponents_正常_構成要素がない書誌(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getMinimalId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	// 空スライスではなくnilのままとする
	assert.Nil(t, target.ContributorRelations)
	assert.Nil(t, target.Urls)
	assert.Nil(t, target.Memos)
	assert.Nil(t, target.OptionValues)
}

// TestRecordRepository_SelectComponents_正常_2件を同時に取得
// 書誌と収録物を1回のクエリでまとめて取得するため、書誌IDでの振り分けを検証する。
func TestRecordRepository_SelectComponents_正常_2件を同時に取得(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	article := newGetTargetEntry(getFullId)
	container := newGetTargetEntry(getContainerId)

	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{article, container},
	)
	require.NoError(t, err)

	// 書誌側
	require.Len(t, article.ContributorRelations, 2)
	assert.Equal(t, "木村元", article.ContributorRelations[0].Contributor.RawName)
	require.Len(t, article.OptionValues, 3)
	assert.Equal(t, "9784130513555", article.OptionValues[0].RawValue)
	assert.Len(t, article.Urls, 2)
	assert.Len(t, article.Memos, 2)

	// 収録物側(書誌の構成要素が混入しないこと)
	require.Len(t, container.ContributorRelations, 1)
	assert.Equal(t, "編者A", container.ContributorRelations[0].Contributor.RawName)
	assert.Equal(t, record.ContributorRoleCompiler, container.ContributorRelations[0].Contribution.Role)
	require.Len(t, container.OptionValues, 1)
	assert.Equal(t, "9784130519999", container.OptionValues[0].RawValue)
	assert.Nil(t, container.Urls)
	assert.Nil(t, container.Memos)
}

// TestRecordRepository_SelectComponents_正常_nilを含むスライス
// 収録物を持たない書誌のため、usecaseは [対象書誌, nil] の形で呼び出す。
func TestRecordRepository_SelectComponents_正常_nilを含むスライス(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	article := newGetTargetEntry(getFullId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{article, nil},
	)
	require.NoError(t, err)

	assert.Len(t, article.ContributorRelations, 2)
	assert.Len(t, article.Urls, 2)
	assert.Len(t, article.Memos, 2)
	assert.Len(t, article.OptionValues, 3)
}

// TestRecordRepository_SelectComponents_正常_対象が空
// 対象が0件のときはクエリを発行しない(IN ()は構文エラーであり、buildPlaceholdersもpanicする)。
func TestRecordRepository_SelectComponents_正常_対象が空(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)
	ctx := newRecordGetTestCtx()

	assert.NoError(t, repo.SelectComponents(ctx, testDb, []*record.UserRecordEntry{}))
	assert.NoError(t, repo.SelectComponents(ctx, testDb, nil))
	assert.NoError(t, repo.SelectComponents(ctx, testDb, []*record.UserRecordEntry{nil}))
}

func TestRecordRepository_SelectComponents_正常_構成要素の任意項目がNULL(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getNullComponentId)
	err := repo.SelectComponents(
		newRecordGetTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	require.Len(t, target.ContributorRelations, 1)
	assert.Nil(t, target.ContributorRelations[0].Contributor.MContributorId)

	require.Len(t, target.Urls, 1)
	assert.Nil(t, target.Urls[0].Label)
	assert.Equal(t, "https://example.com/nolabel", target.Urls[0].Url)

	require.Len(t, target.Memos, 1)
	assert.Nil(t, target.Memos[0].Title)
	assert.Equal(t, "見出しなしのメモ", target.Memos[0].Content)
}
