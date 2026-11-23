package db

import (
	"testing"
	"time"

	"jemref_go/internal/domain/record"
	"jemref_go/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Equal(t, testutil.Ptr("その形成と展開"), full.SubTitle)
	assert.Equal(t, testutil.Ptr("15-46"), full.PageRange)
	assert.Equal(t, testutil.Ptr("第1巻"), full.Volume)
	assert.Equal(t, testutil.Ptr("第11号"), full.Issue)
	assert.Equal(t, testutil.Ptr("東京大学出版会"), full.Publisher)
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
