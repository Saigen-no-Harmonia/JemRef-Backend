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
		c.Title = strPtr(title)
	})
}

// withContributorName 貢献者名だけを指定した検索条件を返す
func withContributorName(name string) repository.RecordListCondition {
	return newListCondition(func(c *repository.RecordListCondition) {
		c.ContributorName = strPtr(name)
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
	assert.Equal(t, strPtr("その展開"), entry.SubTitle)
	assert.Equal(t, strPtr("1-300"), entry.PageRange)
	assert.Equal(t, strPtr("第1巻"), entry.Volume)
	assert.Equal(t, strPtr("第2号"), entry.Issue)
	assert.Equal(t, strPtr("音楽之友社"), entry.Publisher)
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
		c.Title = strPtr("学校")
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

// ---- 並び替え・ページング(records_list_002.sql) ----

func TestRecordRepository_SelectByCondition_正常_並び替え_出版年昇順でNULLは末尾(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	actual := selectListIds(t, repo, withSort(record.SortKeyPublishYear, record.SortOrderAsc))

	assert.Equal(t, []string{"SORT_B", "SORT_A", "SORT_D", "SORT_E1", "SORT_E2", "SORT_C", "SORT_F"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_並び替え_出版年降順でNULLは末尾(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	actual := selectListIds(t, repo, withSort(record.SortKeyPublishYear, record.SortOrderDesc))

	assert.Equal(t, []string{"SORT_E1", "SORT_E2", "SORT_D", "SORT_A", "SORT_B", "SORT_C", "SORT_F"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_並び替え_登録日昇順(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	actual := selectListIds(t, repo, withSort(record.SortKeyInsertDate, record.SortOrderAsc))

	assert.Equal(t, []string{"SORT_B", "SORT_D", "SORT_A", "SORT_C", "SORT_E1", "SORT_E2", "SORT_F"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_並び替え_登録日降順(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	actual := selectListIds(t, repo, withSort(record.SortKeyInsertDate, record.SortOrderDesc))

	assert.Equal(t, []string{"SORT_F", "SORT_E1", "SORT_E2", "SORT_C", "SORT_A", "SORT_D", "SORT_B"}, actual)
}

// TestRecordRepository_SelectByCondition_正常_並び替え_同順位はID昇順
// 降順でも、同順位の書誌はIDの昇順で並ぶ(offsetによるページングで順序がぶれないようにするため)。
func TestRecordRepository_SelectByCondition_正常_並び替え_同順位はID昇順(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	for _, key := range []record.SortKey{record.SortKeyPublishYear, record.SortKeyInsertDate} {
		actual := selectListIds(t, repo, withSort(key, record.SortOrderDesc))

		e1 := indexOf(actual, "SORT_E1")
		e2 := indexOf(actual, "SORT_E2")
		require.NotEqual(t, -1, e1)
		require.NotEqual(t, -1, e2)
		assert.Equal(t, e1+1, e2, "sort key=%s", key)
	}
}

// indexOf スライス内の位置を返す(見つからない場合は-1)
func indexOf(ids []string, id string) int {
	for i, v := range ids {
		if v == id {
			return i
		}
	}

	return -1
}

func TestRecordRepository_SelectByCondition_正常_ページング_limitで打ち切る(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	cond := newListCondition(func(c *repository.RecordListCondition) {
		c.SortKey = record.SortKeyInsertDate
		c.Limit = 2
	})
	actual := selectListIds(t, repo, cond)

	assert.Equal(t, []string{"SORT_B", "SORT_D"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_ページング_offset分読み飛ばす(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	cond := newListCondition(func(c *repository.RecordListCondition) {
		c.SortKey = record.SortKeyInsertDate
		c.Limit = 2
		c.Offset = 2
	})
	actual := selectListIds(t, repo, cond)

	assert.Equal(t, []string{"SORT_A", "SORT_C"}, actual)
}

func TestRecordRepository_SelectByCondition_正常_ページング_offsetが総件数以上なら空(t *testing.T) {
	repo := testRecordsSetup(t, recordListSortFixture)

	// フィクスチャは7件
	cond := newListCondition(func(c *repository.RecordListCondition) {
		c.Offset = 7
	})
	actual, err := repo.SelectByCondition(newRecordListTestCtx(), testDb, cond)
	require.NoError(t, err)

	assert.Empty(t, actual)
}

// -------------CountByCondition(context.Context, repository.DBTX, repository.RecordListCondition)のテスト-------------

func TestRecordRepository_CountByCondition_正常_条件なしの総件数(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual, err := repo.CountByCondition(newRecordListTestCtx(), testDb, newListCondition(nil))
	require.NoError(t, err)

	// 他ユーザ・論理削除済み・収録物を数えない
	assert.Equal(t, len(listAllIds), actual)
}

func TestRecordRepository_CountByCondition_正常_検索条件で絞り込んだ件数(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	// LST_ART_JOURNAL(書誌の主題)と LST_ART_BOOK(収録物の主題)
	actual, err := repo.CountByCondition(newRecordListTestCtx(), testDb, withTitle("学校"))
	require.NoError(t, err)

	assert.Equal(t, 2, actual)
}

func TestRecordRepository_CountByCondition_正常_limitとoffsetの影響を受けない(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	cond := withTitle("学校")
	cond.Limit = 1
	cond.Offset = 5

	actual, err := repo.CountByCondition(newRecordListTestCtx(), testDb, cond)
	require.NoError(t, err)

	assert.Equal(t, 2, actual)
}

func TestRecordRepository_CountByCondition_正常_該当なしは0(t *testing.T) {
	repo := testRecordsSetup(t, recordListFixture)

	actual, err := repo.CountByCondition(newRecordListTestCtx(), testDb, withTitle("存在しないタイトル"))
	require.NoError(t, err)

	assert.Equal(t, 0, actual)
}

// -------------SelectByPks(context.Context, repository.DBTX, []string, int64)のテスト-------------
// records_get_001.sql を使う

func TestRecordRepository_SelectByPks_正常_複数件を取得(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPks(
		newRecordListTestCtx(), testDb, []string{getFullId, getContainerId}, testUid,
	)
	require.NoError(t, err)

	// 結果の順序は保証しない
	assert.ElementsMatch(t, []string{getFullId, getContainerId}, listIds(actual))

	var full *record.UserRecordEntry
	for _, e := range actual {
		if e.RecordEntryId == getFullId {
			full = e
		}
	}
	require.NotNil(t, full)

	// 全カラムが SelectByPk と同じく復元されること
	assert.Equal(t, testUid, full.InternalUid)
	assert.Equal(t, record.RecordTypeJournalArticle, full.RecordType)
	assert.Equal(t, "第1章 「学校」制度の境界線", full.MainTitle)
	assert.Equal(t, strPtr("その形成と展開"), full.SubTitle)
	assert.Equal(t, strPtr("15-46"), full.PageRange)
	assert.Equal(t, strPtr("第1巻"), full.Volume)
	assert.Equal(t, strPtr("第11号"), full.Issue)
	assert.Equal(t, strPtr("東京大学出版会"), full.Publisher)
	require.NotNil(t, full.ReadStatus)
	assert.Equal(t, record.ReadStatusRead, *full.ReadStatus)
	require.NotNil(t, full.ContainerId)
	assert.Equal(t, getContainerId, *full.ContainerId)
	assert.Nil(t, full.MRecordId)
	assert.Equal(t, 0, full.DelFlg)

	require.NotNil(t, full.PublicationDate)
	expectedDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	assert.WithinDuration(t, expectedDate, *full.PublicationDate, time.Microsecond)
	require.NotNil(t, full.DatePrecision)
	assert.Equal(t, record.DatePrecisionDay, *full.DatePrecision)
}

func TestRecordRepository_SelectByPks_正常_他ユーザの書誌は含まれない(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPks(
		newRecordListTestCtx(), testDb, []string{getFullId, getOtherUserId}, testUid,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{getFullId}, listIds(actual))
}

func TestRecordRepository_SelectByPks_正常_論理削除済みは含まれない(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPks(
		newRecordListTestCtx(), testDb, []string{getFullId, getDeletedId}, testUid,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{getFullId}, listIds(actual))
}

func TestRecordRepository_SelectByPks_正常_存在しないIDは無視される(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPks(
		newRecordListTestCtx(), testDb, []string{getFullId, getNotExistsId}, testUid,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{getFullId}, listIds(actual))
}

// TestRecordRepository_SelectByPks_正常_空スライスなら空を返す
// IN ()は妥当なSQLではないため、クエリを発行せずに返す(buildPlaceholdersのpanicに到達しない)。
func TestRecordRepository_SelectByPks_正常_空スライスなら空を返す(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	actual, err := repo.SelectByPks(newRecordListTestCtx(), testDb, []string{}, testUid)
	require.NoError(t, err)

	assert.Nil(t, actual)
}

// -------------SelectContributorRelations(context.Context, repository.DBTX, []*record.UserRecordEntry)のテスト-------------
// records_get_001.sql を使う

// contributorNames 設定された貢献者の氏名を順に返す
func contributorNames(e *record.UserRecordEntry) []string {
	names := make([]string, 0, len(e.ContributorRelations))
	for _, r := range e.ContributorRelations {
		names = append(names, r.Contributor.RawName)
	}

	return names
}

func TestRecordRepository_SelectContributorRelations_正常_貢献者のみ設定される(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	// URL・メモ・項目値も持つ書誌を指定する
	target := newGetTargetEntry(getFullId)
	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	require.Len(t, target.ContributorRelations, 2)
	assert.Equal(t, "木村元", target.ContributorRelations[0].Contributor.RawName)
	assert.Equal(t, record.ContributorRoleAuthor, target.ContributorRelations[0].Contribution.Role)
	assert.Equal(t, "訳者B", target.ContributorRelations[1].Contributor.RawName)
	assert.Equal(t, record.ContributorRoleInterpreter, target.ContributorRelations[1].Contribution.Role)

	// 貢献者以外の構成要素は取得しない
	assert.Nil(t, target.Urls)
	assert.Nil(t, target.Memos)
	assert.Nil(t, target.OptionValues)
}

func TestRecordRepository_SelectContributorRelations_正常_書誌と収録物に振り分ける(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	article := newGetTargetEntry(getFullId)
	container := newGetTargetEntry(getContainerId)
	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{article, container},
	)
	require.NoError(t, err)

	assert.Equal(t, []string{"木村元", "訳者B"}, contributorNames(article))
	assert.Equal(t, []string{"編者A"}, contributorNames(container))
	assert.Equal(t, record.ContributorRoleCompiler, container.ContributorRelations[0].Contribution.Role)
}

func TestRecordRepository_SelectContributorRelations_正常_表示順の昇順(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	// フィクスチャでは表示順の逆順で登録している
	target := newGetTargetEntry(getOrderId)
	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{target},
	)
	require.NoError(t, err)

	assert.Equal(t, []string{"一番目の貢献者", "二番目の貢献者", "三番目の貢献者"}, contributorNames(target))
}

func TestRecordRepository_SelectContributorRelations_正常_論理削除済みの貢献と貢献者本体を除外(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	// 貢献側が削除済みの書誌と、貢献者本体が削除済みの書誌
	delComponent := newGetTargetEntry(getDelComponentId)
	delContrib := newGetTargetEntry(getDelContribId)
	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{delComponent, delContrib},
	)
	require.NoError(t, err)

	assert.Equal(t, []string{"生存の貢献者"}, contributorNames(delComponent))
	assert.Equal(t, []string{"生存の貢献者2"}, contributorNames(delContrib))
}

// TestRecordRepository_SelectContributorRelations_正常_nilを含むスライス
// 収録物を持たない書誌のために、nilの要素は読み飛ばす。
func TestRecordRepository_SelectContributorRelations_正常_nilを含むスライス(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	target := newGetTargetEntry(getFullId)
	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{target, nil},
	)
	require.NoError(t, err)

	assert.Equal(t, []string{"木村元", "訳者B"}, contributorNames(target))
}

func TestRecordRepository_SelectContributorRelations_正常_対象が空(t *testing.T) {
	repo := testRecordsSetup(t, recordGetFixture)

	err := repo.SelectContributorRelations(
		newRecordListTestCtx(), testDb, []*record.UserRecordEntry{},
	)
	require.NoError(t, err)
}
