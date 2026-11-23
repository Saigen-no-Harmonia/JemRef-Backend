package db

import (
	"context"
	"testing"
	"time"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------一覧検索系のテスト-------------

const (
	// recordListFixture 検索条件・除外条件の検証用フィクスチャ
	recordListFixture = "testdata/records_list_001.sql"
	// recordListSortFixture 並び替え・ページングの検証用フィクスチャ
	recordListSortFixture = "testdata/records_list_002.sql"
)

// listAllIds records_list_001.sql で、条件なしの一覧に出る書誌(16件)
// 他ユーザ・論理削除済み・収録物は含まない。
var listAllIds = []string{
	"LST_SOLO_FULL",
	"LST_ART_JOURNAL",
	"LST_ART_BOOK",
	"LST_STANDALONE",
	"LST_PCT",
	"LST_PCT_PLAIN",
	"LST_USC",
	"LST_USC_PLAIN",
	"LST_ESC",
	"LST_ESC_PLAIN",
	"LST_MULTI",
	"LST_DEL_CB",
	"LST_DEL_CR",
	"LST_ART_DELCON",
	"LST_YEAR_START",
	"LST_YEAR_NEXT",
}

// newRecordListTestCtx 書誌一覧参照APIのAPI IDを保持するcontextを返す
func newRecordListTestCtx() context.Context {
	return newRecordTestCtx(ctxutil.REF_API_001)
}

// newListCondition 既定の検索条件(自ユーザ・条件なし・出版年昇順・全件)を書き換えて返す
// limitはフィクスチャの全件が収まる値にしている。
func newListCondition(modify func(c *repository.RecordListCondition)) repository.RecordListCondition {
	cond := repository.RecordListCondition{
		InternalUid: testUid,
		SortKey:     record.SortKeyPublishYear,
		SortOrder:   record.SortOrderAsc,
		Limit:       100,
		Offset:      0,
	}
	if modify != nil {
		modify(&cond)
	}

	return cond
}

// withTitle タイトルだけを指定した検索条件を返す
func withTitle(title string) repository.RecordListCondition {
	return newListCondition(func(c *repository.RecordListCondition) {
		c.Title = testutil.Ptr(title)
	})
}

// withContributorName 貢献者名だけを指定した検索条件を返す
func withContributorName(name string) repository.RecordListCondition {
	return newListCondition(func(c *repository.RecordListCondition) {
		c.ContributorName = testutil.Ptr(name)
	})
}

// withSort 並び替えだけを指定した検索条件を返す
func withSort(key record.SortKey, order record.SortOrder) repository.RecordListCondition {
	return newListCondition(func(c *repository.RecordListCondition) {
		c.SortKey = key
		c.SortOrder = order
	})
}

// listIds 取得した書誌のIDを順に返す
func listIds(entries []*record.UserRecordEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.RecordEntryId)
	}

	return ids
}

// selectListIds 検索条件で一覧を取得し、書誌のIDを順に返す
func selectListIds(t *testing.T, repo *RecordRepositoryImpl, cond repository.RecordListCondition) []string {
	t.Helper()

	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, cond)
	require.NoError(t, err)

	return listIds(actual)
}

// -------------SelectByCondition(context.Context, repository.DBTX, repository.RecordListCondition)のテスト-------------

func TestRecordRepository_SelectByCondition_正常_条件なしは自ユーザの生存書誌のみ(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, newListCondition(nil))

	// 他ユーザ・論理削除済み・収録物(雑誌・編著書・論理削除済みの雑誌)を含まない
	assert.ElementsMatch(t, listAllIds, actual)
}

func TestRecordRepository_SelectByCondition_正常_全カラムを取得(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, withTitle("音楽教育の歴史"))
	require.NoError(t, err)
	require.Len(t, actual, 1)

	entry := actual[0]
	assert.Equal(t, "LST_SOLO_FULL", entry.RecordEntryId)
	assert.Equal(t, testUid, entry.InternalUid)
	assert.Equal(t, "音楽教育の歴史", entry.MainTitle)
	assert.Equal(t, testutil.Ptr("その展開"), entry.SubTitle)
	assert.Equal(t, testutil.Ptr("1-300"), entry.PageRange)
	assert.Equal(t, testutil.Ptr("第1巻"), entry.Volume)
	assert.Equal(t, testutil.Ptr("第2号"), entry.Issue)
	assert.Equal(t, testutil.Ptr("音楽之友社"), entry.Publisher)
	assert.Nil(t, entry.ContainerId)
	assert.Nil(t, entry.MRecordId)
	assert.Equal(t, 0, entry.DelFlg)

	// TYPE_ID("RT000001")から書誌タイプへ復元されること
	assert.Equal(t, record.RecordTypeAuthoredBook, entry.RecordType)

	// READ_STATUS(1)から読了ステータスへ復元されること
	require.NotNil(t, entry.ReadStatus)
	assert.Equal(t, record.ReadStatusRead, *entry.ReadStatus)

	require.NotNil(t, entry.PublicationDate)
	expectedDate := time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC)
	assert.WithinDuration(t, expectedDate, *entry.PublicationDate, time.Microsecond)

	require.NotNil(t, entry.DatePrecision)
	assert.Equal(t, record.DatePrecisionDay, *entry.DatePrecision)
}

func TestRecordRepository_SelectByCondition_正常_収録物のIDを持つ書誌はContainerIdが入る(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, withTitle("戦後の学校"))
	require.NoError(t, err)
	require.Len(t, actual, 1)

	assert.Equal(t, record.RecordTypeJournalArticle, actual[0].RecordType)
	require.NotNil(t, actual[0].ContainerId)
	assert.Equal(t, "LST_CON_JOURNAL", *actual[0].ContainerId)
}

// TestRecordRepository_SelectByCondition_正常_収録物は含まれない
// 収録物の主題で検索しても、収録物そのものではなく、それを収録する書誌が返る。
func TestRecordRepository_SelectByCondition_正常_収録物は含まれない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("教育学研究"))

	assert.Equal(t, []string{"LST_ART_JOURNAL"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_単独で登録した雑誌は含まれる
// 雑誌型でも、生きている書誌から参照されていなければ収録物ではない(CountActiveArticlesと同じ定義)。
func TestRecordRepository_SelectByCondition_正常_単独で登録した雑誌は含まれる(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("単独で登録"))

	assert.Equal(t, []string{"LST_STANDALONE"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_タイトル_書誌の主題で部分一致(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// 他ユーザ・論理削除済みの書誌も「音楽教育」を含むが返らない
	actual := selectListIds(t, repo, withTitle("音楽教育"))

	assert.Equal(t, []string{"LST_SOLO_FULL"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_タイトル_書誌の副題で部分一致(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("制度の形成"))

	assert.Equal(t, []string{"LST_ART_BOOK"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_タイトル_収録物の主題でヒット(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("境界線の学校史"))

	assert.Equal(t, []string{"LST_ART_BOOK"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_タイトル_収録物の副題でヒット(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("特集 学級"))

	assert.Equal(t, []string{"LST_ART_JOURNAL"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_タイトル_論理削除済みの収録物ではヒットしない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// LST_ART_DELCON は「削除済の雑誌」を収録物として指しているが、収録物が論理削除済み
	actual := selectListIds(t, repo, withTitle("削除済の雑誌"))

	assert.Empty(t, actual)
}

// TestRecordRepository_SelectByCondition_正常_タイトル_ワイルドカード文字は文字どおり扱う
// エスケープしないと「100%」のパターンが「100点の理解」にも一致する。
func TestRecordRepository_SelectByCondition_正常_タイトル_ワイルドカード文字は文字どおり扱う(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("100%"))

	assert.Equal(t, []string{"LST_PCT"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_タイトル_アンダースコアは文字どおり扱う
// エスケープしないと「_」が任意の1文字になり、「axbの研究」にも一致する。
func TestRecordRepository_SelectByCondition_正常_タイトル_アンダースコアは文字どおり扱う(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("a_b"))

	assert.Equal(t, []string{"LST_USC"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_タイトル_エスケープ文字自体を含む検索語
// エスケープしないと「!重」が「重」と解釈され、「注意重要」にだけ一致する。
func TestRecordRepository_SelectByCondition_正常_タイトル_エスケープ文字自体を含む検索語(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("意!重"))

	assert.Equal(t, []string{"LST_ESC"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_SQLとして解釈される文字列も文字どおり検索する
// 値はプレースホルダで渡すため、SQLとして解釈されず、他ユーザ・論理削除済みの書誌も返らない。
func TestRecordRepository_SelectByCondition_正常_SQLとして解釈される文字列も文字どおり検索する(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withTitle("' OR 1=1 --"))

	assert.Empty(t, actual)
}

func TestRecordRepository_SelectByCondition_正常_貢献者名_書誌の貢献者で部分一致(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// 他ユーザにも同名の貢献者がいるが返らない
	actual := selectListIds(t, repo, withContributorName("山田"))

	assert.Equal(t, []string{"LST_SOLO_FULL"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_貢献者名_収録物の貢献者でヒット(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// 木村元は編著書(LST_CON_BOOK)の編者
	actual := selectListIds(t, repo, withContributorName("木村"))

	assert.Equal(t, []string{"LST_ART_BOOK"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_貢献者名_論理削除済みの収録物の貢献者ではヒットしない
// 貢献・貢献者本体は生存しているため、収録物のDEL_FLGで除外されていることを検証できる。
func TestRecordRepository_SelectByCondition_正常_貢献者名_論理削除済みの収録物の貢献者ではヒットしない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withContributorName("削除雑誌の編者"))

	assert.Empty(t, actual)
}

// TestRecordRepository_SelectByCondition_正常_貢献者名_論理削除済みの貢献ではヒットしない
// 「消された貢献」は「消された貢献者」にも部分一致するが、そちらは貢献者本体が論理削除済み。
func TestRecordRepository_SelectByCondition_正常_貢献者名_論理削除済みの貢献ではヒットしない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withContributorName("消された貢献"))

	assert.Empty(t, actual)
}

func TestRecordRepository_SelectByCondition_正常_貢献者名_貢献者本体が論理削除済みならヒットしない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withContributorName("消された貢献者"))

	assert.Empty(t, actual)
}

// TestRecordRepository_SelectByCondition_正常_貢献者名_複数の貢献者が一致しても1件
// EXISTSで判定するため、貢献者の数だけ書誌が重複することはない。
func TestRecordRepository_SelectByCondition_正常_貢献者名_複数の貢献者が一致しても1件(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual := selectListIds(t, repo, withContributorName("田中"))

	assert.Equal(t, []string{"LST_MULTI"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_出版年_年の境界
// 1月1日と12月31日は含み、翌年の1月1日は含まない。
func TestRecordRepository_SelectByCondition_正常_出版年_年の境界(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	cond := newListCondition(func(c *repository.RecordListCondition) {
		year := 2020
		c.PublishYear = &year
	})
	actual := selectListIds(t, repo, cond)

	// 出版年の昇順
	assert.Equal(t, []string{"LST_YEAR_START", "LST_SOLO_FULL", "LST_ART_BOOK"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_出版年_出版日がNULLの書誌は含まれない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	cond := newListCondition(func(c *repository.RecordListCondition) {
		year := 2020
		c.PublishYear = &year
	})
	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, cond)
	require.NoError(t, err)
	require.NotEmpty(t, actual)

	for _, e := range actual {
		assert.NotNil(t, e.PublicationDate, "record id=%s", e.RecordEntryId)
	}
}

// TestRecordRepository_SelectByCondition_正常_複合条件はANDで絞り込む
// タイトルだけなら JOURNAL と BOOK、出版年だけなら JOURNAL と YEAR_NEXT がヒットする。
func TestRecordRepository_SelectByCondition_正常_複合条件はANDで絞り込む(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	cond := newListCondition(func(c *repository.RecordListCondition) {
		year := 2021
		c.Title = testutil.Ptr("学校")
		c.PublishYear = &year
	})
	actual := selectListIds(t, repo, cond)

	assert.Equal(t, []string{"LST_ART_JOURNAL"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_該当なしは空(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, withTitle("存在しないタイトル"))
	require.NoError(t, err)

	assert.Empty(t, actual)
}

// TestRecordRepository_SelectByCondition_正常_構成要素は取得しない
// 構成要素の取得は SelectContributorRelations 等の責務であり、SelectByCondition では設定しない。
func TestRecordRepository_SelectByCondition_正常_構成要素は取得しない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// 貢献者を持つ書誌を指定する
	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, withTitle("音楽教育の歴史"))
	require.NoError(t, err)
	require.Len(t, actual, 1)

	assert.Nil(t, actual[0].ContributorRelations)
	assert.Nil(t, actual[0].Urls)
	assert.Nil(t, actual[0].Memos)
	assert.Nil(t, actual[0].OptionValues)
}
