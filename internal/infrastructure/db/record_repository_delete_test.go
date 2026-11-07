package db

import (
	"context"
	"testing"
	"time"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------削除系のテスト-------------

const (
	// recordDeleteFixture 書誌情報削除テスト用フィクスチャ
	recordDeleteFixture = "testdata/records_delete_001.sql"

	// 以下はフィクスチャ(records_delete_001.sql)で投入される書誌ID
	delSoloId        = "DEL_SOLO_0000000000000001" // コンテナなし。構成要素フル
	delContainer1Id  = "DEL_CONTAINER_000000001"   // 書誌1件の収録物
	delContainer2Id  = "DEL_CONTAINER_000000002"   // 書誌2件(生1/削除済1)の収録物
	delArticleB1Id   = "DEL_ARTICLE_B_000000001"   // 生存している書誌
	delDeletedId     = "DEL_DELETED_00000000001"   // すでに論理削除済み
	delPartialId     = "DEL_PARTIAL_00000000001"   // 構成要素の一部が論理削除済み
	delNotExistsId   = "DEL_NOTEXISTS_0000000001"  // 存在しない書誌ID
	alreadyDeletedPg = "already_deleted"           // 削除済み行のUPD_PG(更新されていないことの判別用)
)

// -------------CountActiveArticles(context.Context, repository.DBTX, string)のテスト-------------

func TestRecordRepository_CountActiveArticles_正常_書誌が0件(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	count, err := repo.CountActiveArticles(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delSoloId,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRecordRepository_CountActiveArticles_正常_論理削除済みの書誌は数えない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)
	ctx := newRecordTestCtx(ctxutil.REF_API_005)

	// 書誌1件(生存)のみの収録物
	count, err := repo.CountActiveArticles(ctx, testDb, delContainer1Id)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// 書誌2件のうち1件が論理削除済みの収録物
	count, err = repo.CountActiveArticles(ctx, testDb, delContainer2Id)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRecordRepository_CountActiveArticles_正常_存在しないコンテナIDは0件(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	count, err := repo.CountActiveArticles(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delNotExistsId,
	)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// -------------LogicalDelete(context.Context, repository.DBTX, string)のテスト-------------

func TestRecordRepository_LogicalDelete_正常_書誌と構成要素が論理削除される(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)
	now := time.Now().Truncate(time.Microsecond)

	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delSoloId,
	)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, delSoloId)
	require.NoError(t, err)
	assert.Equal(t, 1, actual.DelFlg)
	assert.Equal(t, ctxutil.REF_API_005, actual.UpdPg)
	assert.Equal(t, "system", actual.UpdId)
	assert.WithinDuration(t, now, actual.UpdDt, 3*time.Second)
	// 登録時の監査列は書き換わらない
	assert.Equal(t, "ins_pg", actual.InsPg)

	contributions := testutil.SelectContributionsByRecordId(t, testDb, delSoloId)
	require.Len(t, contributions, 2)
	for _, c := range contributions {
		assert.Equal(t, 1, c.DelFlg)
		assert.Equal(t, ctxutil.REF_API_005, c.UpdPg)
		assert.Equal(t, "system", c.UpdId)
		assert.WithinDuration(t, now, c.UpdDt, 3*time.Second)
	}

	urls := testutil.SelectUrlsByRecordId(t, testDb, delSoloId)
	require.Len(t, urls, 2)
	for _, u := range urls {
		assert.Equal(t, 1, u.DelFlg)
		assert.Equal(t, ctxutil.REF_API_005, u.UpdPg)
		assert.Equal(t, "system", u.UpdId)
	}

	memos := testutil.SelectMemosByRecordId(t, testDb, delSoloId)
	require.Len(t, memos, 2)
	for _, m := range memos {
		assert.Equal(t, 1, m.DelFlg)
		assert.Equal(t, ctxutil.REF_API_005, m.UpdPg)
		assert.Equal(t, "system", m.UpdId)
	}

	fieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, delSoloId)
	require.Len(t, fieldValues, 2)
	for _, f := range fieldValues {
		assert.Equal(t, 1, f.DelFlg)
		assert.Equal(t, ctxutil.REF_API_005, f.UpdPg)
		assert.Equal(t, "system", f.UpdId)
	}
}

func TestRecordRepository_LogicalDelete_正常_貢献者マスタは削除されない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	before := testutil.SelectContributorsByUserId(t, testDb, testUid)
	require.NotEmpty(t, before)

	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delSoloId,
	)
	require.NoError(t, err)

	// e_contributorsはユーザ単位の入力辞書であり書誌に属さないため削除しない
	after := testutil.SelectContributorsByUserId(t, testDb, testUid)
	require.Len(t, after, len(before))
	for i, c := range after {
		assert.Equal(t, 0, c.DelFlg)
		assert.Equal(t, before[i].UpdPg, c.UpdPg)
		assert.WithinDuration(t, before[i].UpdDt, c.UpdDt, time.Microsecond)
	}
}

func TestRecordRepository_LogicalDelete_正常_構成要素がない書誌(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	// 構成要素行を持たない収録物
	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delContainer2Id,
	)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, delContainer2Id)
	require.NoError(t, err)
	assert.Equal(t, 1, actual.DelFlg)
	assert.Equal(t, ctxutil.REF_API_005, actual.UpdPg)
}

func TestRecordRepository_LogicalDelete_正常_他の書誌の構成要素に影響しない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	beforeContributions := testutil.SelectContributionsByRecordId(t, testDb, delContainer1Id)
	beforeFieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, delContainer1Id)
	require.NotEmpty(t, beforeContributions)
	require.NotEmpty(t, beforeFieldValues)

	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delSoloId,
	)
	require.NoError(t, err)

	// 別の書誌に紐づく構成要素は生存したまま
	afterContributions := testutil.SelectContributionsByRecordId(t, testDb, delContainer1Id)
	require.Len(t, afterContributions, len(beforeContributions))
	for i, c := range afterContributions {
		assert.Equal(t, 0, c.DelFlg)
		assert.WithinDuration(t, beforeContributions[i].UpdDt, c.UpdDt, time.Microsecond)
	}

	afterFieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, delContainer1Id)
	require.Len(t, afterFieldValues, len(beforeFieldValues))
	for i, f := range afterFieldValues {
		assert.Equal(t, 0, f.DelFlg)
		assert.WithinDuration(t, beforeFieldValues[i].UpdDt, f.UpdDt, time.Microsecond)
	}

	// 書誌本体も影響を受けない
	container, err := testutil.SelectRecordByPk(t, testDb, delContainer1Id)
	require.NoError(t, err)
	assert.Equal(t, 0, container.DelFlg)
}

func TestRecordRepository_LogicalDelete_正常_削除済みの構成要素は更新しない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)
	now := time.Now().Truncate(time.Microsecond)

	beforeContributions := testutil.SelectContributionsByRecordId(t, testDb, delPartialId)
	require.Len(t, beforeContributions, 2)

	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delPartialId,
	)
	require.NoError(t, err)

	// 表示順0=生存していた行、表示順1=すでに論理削除済みの行
	contributions := testutil.SelectContributionsByRecordId(t, testDb, delPartialId)
	require.Len(t, contributions, 2)
	assert.Equal(t, 1, contributions[0].DelFlg)
	assert.Equal(t, ctxutil.REF_API_005, contributions[0].UpdPg)
	assert.WithinDuration(t, now, contributions[0].UpdDt, 3*time.Second)
	// すでに削除済みの行はUPD_PG/UPD_DTが更新されない(DEL_FLG=0条件の効果)
	assert.Equal(t, alreadyDeletedPg, contributions[1].UpdPg)
	assert.WithinDuration(t, beforeContributions[1].UpdDt, contributions[1].UpdDt, time.Microsecond)

	urls := testutil.SelectUrlsByRecordId(t, testDb, delPartialId)
	require.Len(t, urls, 2)
	assert.Equal(t, ctxutil.REF_API_005, urls[0].UpdPg)
	assert.Equal(t, alreadyDeletedPg, urls[1].UpdPg)

	memos := testutil.SelectMemosByRecordId(t, testDb, delPartialId)
	require.Len(t, memos, 2)
	assert.Equal(t, ctxutil.REF_API_005, memos[0].UpdPg)
	assert.Equal(t, alreadyDeletedPg, memos[1].UpdPg)

	fieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, delPartialId)
	require.Len(t, fieldValues, 2)
	assert.Equal(t, ctxutil.REF_API_005, fieldValues[0].UpdPg)
	assert.Equal(t, alreadyDeletedPg, fieldValues[1].UpdPg)
}

func TestRecordRepository_LogicalDelete_正常_削除済みの書誌を再実行してもエラーにならない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	before, err := testutil.SelectRecordByPk(t, testDb, delDeletedId)
	require.NoError(t, err)
	require.Equal(t, 1, before.DelFlg)

	err = repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delDeletedId,
	)
	require.NoError(t, err)

	// 冪等であり、監査列も更新されない
	after, err := testutil.SelectRecordByPk(t, testDb, delDeletedId)
	require.NoError(t, err)
	assert.Equal(t, 1, after.DelFlg)
	assert.Equal(t, alreadyDeletedPg, after.UpdPg)
	assert.WithinDuration(t, before.UpdDt, after.UpdDt, time.Microsecond)
}

func TestRecordRepository_LogicalDelete_正常_存在しない書誌IDでもエラーにならない(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	beforeCnt := testutil.CountRecords(t, testDb)

	// 存在確認はSelectByPkの責務であり、LogicalDeleteは更新0件を許容する
	err := repo.LogicalDelete(
		newRecordTestCtx(ctxutil.REF_API_005), testDb, delNotExistsId,
	)
	require.NoError(t, err)

	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))

	// 他の書誌が巻き込まれていないこと
	solo, err := testutil.SelectRecordByPk(t, testDb, delSoloId)
	require.NoError(t, err)
	assert.Equal(t, 0, solo.DelFlg)
}

func TestRecordRepository_LogicalDelete_異常_APIIDが未設定(t *testing.T) {
	repo := testRecordsSetup(t, recordDeleteFixture)

	assert.Panics(t, func() {
		_ = repo.LogicalDelete(context.Background(), testDb, delSoloId)
	})

	// 1行も更新されていないこと
	actual, err := testutil.SelectRecordByPk(t, testDb, delSoloId)
	require.NoError(t, err)
	assert.Equal(t, 0, actual.DelFlg)

	contributions := testutil.SelectContributionsByRecordId(t, testDb, delSoloId)
	require.Len(t, contributions, 2)
	for _, c := range contributions {
		assert.Equal(t, 0, c.DelFlg)
	}
}
