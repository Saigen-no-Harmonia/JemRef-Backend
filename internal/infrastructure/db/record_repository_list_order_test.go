package db

import (
	"testing"

	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------SelectByCondition(context.Context, repository.DBTX, repository.RecordListCondition)のテスト(並び替え・ページング)-------------
// records_list_002.sql を使う

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
