package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------更新系のテスト-------------

const (
	// recordUpdateFixture 書誌情報更新テスト用フィクスチャ
	recordUpdateFixture = "testdata/records_update_001.sql"

	// 以下はフィクスチャ(records_update_001.sql)で投入される書誌ID
	updFullId          = "UPD_FULL_00000000000001" // 全カラムに値+構成要素フル(論理削除済みを含む)
	updContainerId     = "UPD_CONTAINER_00000001"  // updFullIdの収録物。構成要素あり
	updFreeContainerId = "UPD_FREE_CONTAINER_001"  // 参照されていない収録物(付け替え先)
	updMinimalId       = "UPD_MINIMAL_000000001"   // 必須項目のみ。構成要素なし
	updSharedAId       = "UPD_SHARED_A_0000001"    // 貢献者をupdSharedBIdと共有する書誌
	updSharedBId       = "UPD_SHARED_B_0000001"    // 同上
	updOtherUserId     = "UPD_OTHER_USER_00001"    // 他ユーザの書誌。構成要素あり
	updDeletedId       = "UPD_DELETED_00000001"    // 論理削除済みの書誌
	updUntouchedId     = "UPD_UNTOUCHED_000001"    // 更新対象外。構成要素フル
	updNotExistsId     = "UPD_NOTEXISTS_0000001"   // 存在しない書誌ID

	// untouchedPg 更新されていないことの判別用のUPD_PG
	untouchedPg = "untouched"
)

// updSharedContributorId updSharedAId と updSharedBId が共有する貢献者ID
const updSharedContributorId = int64(9211)

// newRecordUpdateTestCtx 書誌情報更新APIのAPI IDを保持するcontextを返す
func newRecordUpdateTestCtx() context.Context {
	return newRecordTestCtx(ctxutil.REF_API_004)
}

// updOptionFields 項目値に使う入力項目
//
// UNIQUE制約が (書誌ID, 入力項目ID, 表示順) のため、件数分だけ別の入力項目を割り当てる。
var updOptionFields = []record.RecordField{
	record.RecordFieldIsbn,
	record.RecordFieldEdition,
	record.RecordFieldOriginalTitle,
}

// newUpdateEntry 必須項目のみを設定した更新用のユーザ入力書誌を返す
func newUpdateEntry(recordId string, recordType record.RecordType, mainTitle string) record.UserRecordEntry {
	return record.UserRecordEntry{
		RecordEntryId: recordId,
		InternalUid:   testUid,
		RecordType:    recordType,
		MainTitle:     mainTitle,
		DelFlg:        0,
	}
}

// withComponents 構成要素を各count件(最大3件)持たせたユーザ入力書誌を返す
//
// prefixは洗い替えの前後を値で区別するための接頭辞。
// 表示順は0起点の連番とし、フィクスチャと同じ表示順で再登録されることを検証できるようにする
// (子テーブルのUNIQUE制約に DEL_FLG が含まれないため、論理削除では Error 1062 になる)。
func withComponents(e record.UserRecordEntry, count int, prefix string) record.UserRecordEntry {
	for i := range count {
		e.ContributorRelations = append(
			e.ContributorRelations,
			newMinimalRelation(fmt.Sprintf("%s貢献者%d", prefix, i), record.ContributorRoleAuthor),
		)
		e.Urls = append(e.Urls, record.UrlEntry{
			RecordEntryId: e.RecordEntryId,
			DisplayNumber: i,
			Label:         strPtr(fmt.Sprintf("%sラベル%d", prefix, i)),
			Url:           fmt.Sprintf("https://example.com/%s/%d", prefix, i),
			InternalUid:   testUid,
		})
		e.Memos = append(e.Memos, record.MemoEntry{
			RecordEntryId: e.RecordEntryId,
			DisplayNumber: i,
			Title:         strPtr(fmt.Sprintf("%s見出し%d", prefix, i)),
			Content:       fmt.Sprintf("%s内容%d", prefix, i),
			InternalUid:   testUid,
		})
		e.OptionValues = append(e.OptionValues, record.OptionValue{
			RecordEntryId: e.RecordEntryId,
			RecordField:   updOptionFields[i],
			DisplayNumber: i,
			RawValue:      fmt.Sprintf("%s項目値%d", prefix, i),
			InternalUid:   testUid,
		})
	}

	return e
}

// componentIds 構成要素の行ID
//
// 洗い替えが UPDATE ではなく「物理削除 + 再INSERT」で行われたことは、
// 行IDが入れ替わっていること(AUTO_INCREMENTは再利用されない)で判別する。
type componentIds struct {
	contributions []int64
	urls          []int64
	memos         []int64
	fieldValues   []int64
}

// selectComponentIds 指定書誌の構成要素の行IDを取得する(論理削除済みも含む)
func selectComponentIds(t *testing.T, recordId string) componentIds {
	t.Helper()

	var ids componentIds
	for _, c := range testutil.SelectContributionsByRecordId(t, testDb, recordId) {
		ids.contributions = append(ids.contributions, c.Id)
	}
	for _, u := range testutil.SelectUrlsByRecordId(t, testDb, recordId) {
		ids.urls = append(ids.urls, u.Id)
	}
	for _, m := range testutil.SelectMemosByRecordId(t, testDb, recordId) {
		ids.memos = append(ids.memos, m.Id)
	}
	for _, f := range testutil.SelectFieldValuesByRecordId(t, testDb, recordId) {
		ids.fieldValues = append(ids.fieldValues, f.Id)
	}

	return ids
}

// assertComponentIdsReplaced 構成要素の行IDが洗い替えで総入れ替えされたことを検証する
func assertComponentIdsReplaced(t *testing.T, before, after componentIds) {
	t.Helper()

	assertNoSharedId(t, "contributions", before.contributions, after.contributions)
	assertNoSharedId(t, "urls", before.urls, after.urls)
	assertNoSharedId(t, "memos", before.memos, after.memos)
	assertNoSharedId(t, "field values", before.fieldValues, after.fieldValues)
}

func assertNoSharedId(t *testing.T, name string, before, after []int64) {
	t.Helper()

	require.NotEmpty(t, before, "%s: 洗い替え前の行がない", name)
	require.NotEmpty(t, after, "%s: 洗い替え後の行がない", name)

	for _, b := range before {
		assert.NotContains(t, after, b, "%s: 行ID %d が洗い替え後も残っている", name, b)
	}
}

// findContributor 指定IDの貢献者を返す(存在しない場合はnil)
func findContributor(rows []testutil.ContributorRow, id int64) *testutil.ContributorRow {
	for i := range rows {
		if rows[i].Id == id {
			return &rows[i]
		}
	}

	return nil
}

// assertComponentCounts 指定書誌の構成要素の件数を検証する
func assertComponentCounts(t *testing.T, recordId string, want int) {
	t.Helper()

	assert.Len(t, testutil.SelectContributionsByRecordId(t, testDb, recordId), want, "貢献")
	assert.Len(t, testutil.SelectUrlsByRecordId(t, testDb, recordId), want, "URL")
	assert.Len(t, testutil.SelectMemosByRecordId(t, testDb, recordId), want, "メモ")
	assert.Len(t, testutil.SelectFieldValuesByRecordId(t, testDb, recordId), want, "項目値")
}

// -------------Update(context.Context, repository.DBTX, record.UserRecordEntry)のテスト-------------

func TestRecordRepository_Update_正常_本体の全カラムが上書きされる(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	now := time.Now().Truncate(time.Microsecond)

	before, err := testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)

	publicationDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionYear
	readStatus := record.ReadStatusReading

	input := newUpdateEntry(updFullId, record.RecordTypeContributedBookArticle, "更新後の主題")
	input.SubTitle = strPtr("更新後の副題")
	input.PageRange = strPtr("100-200")
	input.Volume = strPtr("更新後の巻")
	input.Issue = strPtr("更新後の号")
	input.Publisher = strPtr("更新後の出版者")
	input.PublicationDate = &publicationDate
	input.DatePrecision = &datePrecision
	input.ReadStatus = &readStatus
	input.ContainerId = strPtr(updFreeContainerId)

	err = repo.Update(newRecordUpdateTestCtx(), testDb, input)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)

	// SET対象の11カラムすべてが上書きされる。
	// 書誌タイプの変更を禁止するのはusecaseの責務であり、repositoryは止めない。
	assert.Equal(t, record.RecordTypeIdContributedBookArticle, actual.TypeId)
	assert.Equal(t, "更新後の主題", actual.MainTitle)
	assert.Equal(t, "更新後の副題", actual.SubTitle.String)
	assert.Equal(t, "100-200", actual.PageRange.String)
	assert.Equal(t, "更新後の巻", actual.Volume.String)
	assert.Equal(t, "更新後の号", actual.Issue.String)
	assert.Equal(t, "更新後の出版者", actual.Publisher.String)
	assert.WithinDuration(t, publicationDate, actual.PublicationDate.Time, time.Microsecond)
	assert.Equal(t, int64(record.DatePrecisionYear), actual.DatePrecision.Int64)
	assert.Equal(t, int64(record.ReadStatusIdReading), actual.ReadStatus.Int64)
	assert.Equal(t, updFreeContainerId, actual.ContainerId.String)

	// 監査列(更新)はAPI IDと'system'で上書きされる
	assert.Equal(t, ctxutil.REF_API_004, actual.UpdPg)
	assert.Equal(t, "system", actual.UpdId)
	assert.WithinDuration(t, now, actual.UpdDt, 3*time.Second)

	// SET対象外のカラムは変化しない
	assert.Equal(t, updFullId, actual.Id)
	assert.Equal(t, testUid, actual.UserId)
	assert.False(t, actual.MRecordId.Valid, "M_RECORD_IDはPh0未使用のため更新対象外")
	assert.Equal(t, 0, actual.DelFlg, "DEL_FLGは削除APIの責務のため更新対象外")
	assert.Equal(t, before.InsPg, actual.InsPg)
	assert.Equal(t, before.InsId, actual.InsId)
	assert.WithinDuration(t, before.InsDt, actual.InsDt, time.Microsecond)
}

func TestRecordRepository_Update_正常_任意カラムのNULLと値が双方向に上書きされる(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	ctx := newRecordUpdateTestCtx()

	// 値あり -> NULL。「リクエストどおりそのまま上書きする」ため、
	// 空文字ではなくNULLになる(NOT NULLカラムのみ更新するようなロジックは持たない)。
	err := repo.Update(ctx, testDb, newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "必須のみに戻した主題"))
	require.NoError(t, err)

	cleared, err := testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)
	assert.False(t, cleared.SubTitle.Valid)
	assert.False(t, cleared.PageRange.Valid)
	assert.False(t, cleared.Volume.Valid)
	assert.False(t, cleared.Issue.Valid)
	assert.False(t, cleared.Publisher.Valid)
	assert.False(t, cleared.PublicationDate.Valid)
	assert.False(t, cleared.DatePrecision.Valid)
	assert.False(t, cleared.ReadStatus.Valid, "ReadStatus.GetNullableId()がnilを返す")
	assert.False(t, cleared.ContainerId.Valid)

	// NULL -> 値あり(逆方向)
	publicationDate := time.Date(2023, 5, 20, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionMonth
	readStatus := record.ReadStatusUnread

	filled := newUpdateEntry(updMinimalId, record.RecordTypeAuthoredBook, "値を入れた主題")
	filled.SubTitle = strPtr("入れた副題")
	filled.PageRange = strPtr("1-2")
	filled.Volume = strPtr("入れた巻")
	filled.Issue = strPtr("入れた号")
	filled.Publisher = strPtr("入れた出版者")
	filled.PublicationDate = &publicationDate
	filled.DatePrecision = &datePrecision
	filled.ReadStatus = &readStatus
	filled.ContainerId = strPtr(updFreeContainerId)

	err = repo.Update(ctx, testDb, filled)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, updMinimalId)
	require.NoError(t, err)
	assert.Equal(t, "入れた副題", actual.SubTitle.String)
	assert.Equal(t, "1-2", actual.PageRange.String)
	assert.Equal(t, "入れた巻", actual.Volume.String)
	assert.Equal(t, "入れた号", actual.Issue.String)
	assert.Equal(t, "入れた出版者", actual.Publisher.String)
	assert.WithinDuration(t, publicationDate, actual.PublicationDate.Time, time.Microsecond)
	assert.Equal(t, int64(record.DatePrecisionMonth), actual.DatePrecision.Int64)
	assert.Equal(t, int64(record.ReadStatusIdUnread), actual.ReadStatus.Int64)
	assert.Equal(t, updFreeContainerId, actual.ContainerId.String)
}

func TestRecordRepository_Update_正常_同じ内容で更新しても成功する(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)

	before, err := testutil.SelectRecordByPk(t, testDb, updMinimalId)
	require.NoError(t, err)

	// フィクスチャとまったく同じ内容で更新する。
	// UPD_DT が必ず変わるため RowsAffected() は0にならず、ErrNotFound と誤判定されない
	// (= RowsAffected()==0 を「該当行なし」と解釈してよい根拠)。
	err = repo.Update(
		newRecordUpdateTestCtx(),
		testDb,
		newUpdateEntry(updMinimalId, record.RecordTypeAuthoredBook, before.MainTitle),
	)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, updMinimalId)
	require.NoError(t, err)
	assert.Equal(t, ctxutil.REF_API_004, actual.UpdPg)
	assert.True(t, actual.UpdDt.After(before.UpdDt), "UPD_DTが進んでいること")
}

func TestRecordRepository_Update_正常_構成要素が洗い替えられる(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	now := time.Now().Truncate(time.Microsecond)

	before := selectComponentIds(t, updFullId)
	// フィクスチャは各テーブルに論理削除済みの行を1件含む
	require.Len(t, before.contributions, 3)
	require.Len(t, before.urls, 3)
	require.Len(t, before.memos, 3)
	require.Len(t, before.fieldValues, 4)

	// フィクスチャと同じ表示順(0,1)で別の内容を登録する。
	// 論理削除では子テーブルのUNIQUE制約により Error 1062 になるため、
	// ここがエラーにならないこと自体が「物理削除している」ことの検証になる。
	input := withComponents(
		newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "洗い替え後の主題"),
		2,
		"after",
	)

	err := repo.Update(newRecordUpdateTestCtx(), testDb, input)
	require.NoError(t, err)

	// 論理削除済みだった行も物理削除されるため、残るのは新しい2件だけ
	assertComponentCounts(t, updFullId, 2)

	after := selectComponentIds(t, updFullId)
	assertComponentIdsReplaced(t, before, after)

	// 表示順は0起点で振り直され、値も新しいものに入れ替わる
	contributions := testutil.SelectContributionsByRecordId(t, testDb, updFullId)
	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	for i, c := range contributions {
		assert.Equal(t, i, c.DisplayNumber)
		assert.Equal(t, 0, c.DelFlg)
		assert.Equal(t, ctxutil.REF_API_004, c.InsPg)
		assert.Equal(t, ctxutil.REF_API_004, c.UpdPg)
		assert.WithinDuration(t, now, c.UpdDt, 3*time.Second)

		contributor := findContributor(contributors, c.ContributorId)
		require.NotNil(t, contributor)
		assert.Equal(t, fmt.Sprintf("after貢献者%d", i), contributor.RawName)
	}

	urls := testutil.SelectUrlsByRecordId(t, testDb, updFullId)
	for i, u := range urls {
		assert.Equal(t, i, u.DisplayNumber)
		assert.Equal(t, fmt.Sprintf("afterラベル%d", i), u.RawLabel.String)
		assert.Equal(t, 0, u.DelFlg)
		assert.Equal(t, ctxutil.REF_API_004, u.InsPg)
	}

	memos := testutil.SelectMemosByRecordId(t, testDb, updFullId)
	for i, m := range memos {
		assert.Equal(t, i, m.DisplayNumber)
		assert.Equal(t, fmt.Sprintf("after内容%d", i), m.Content)
		assert.Equal(t, 0, m.DelFlg)
		assert.Equal(t, ctxutil.REF_API_004, m.InsPg)
	}

	fieldValues := testutil.SelectFieldValuesByRecordId(t, testDb, updFullId)
	for i, f := range fieldValues {
		assert.Equal(t, i, f.DisplayNumber)
		assert.Equal(t, updOptionFields[i].GetId(), f.RecordFieldId)
		assert.Equal(t, fmt.Sprintf("after項目値%d", i), f.RawValue)
		assert.Equal(t, 0, f.DelFlg)
		assert.Equal(t, ctxutil.REF_API_004, f.InsPg)
	}
}

func TestRecordRepository_Update_正常_構成要素の件数が増減する(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	ctx := newRecordUpdateTestCtx()
	base := newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "件数検証の主題")

	// 減る(3件 -> 1件)
	err := repo.Update(ctx, testDb, withComponents(base, 1, "one"))
	require.NoError(t, err)
	assertComponentCounts(t, updFullId, 1)

	// 空になる(1件 -> 0件)
	err = repo.Update(ctx, testDb, base)
	require.NoError(t, err)
	assertComponentCounts(t, updFullId, 0)

	// 増える(0件 -> 3件)。
	// 貢献が0件の状態からの洗い替えでは退避した貢献者IDが空になるため、
	// buildPlaceholders(0) のpanicを避ける早期returnが効いていることの検証も兼ねる。
	err = repo.Update(ctx, testDb, withComponents(base, 3, "three"))
	require.NoError(t, err)
	assertComponentCounts(t, updFullId, 3)
}

func TestRecordRepository_Update_正常_参照されなくなった貢献者のみ物理削除される(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	ctx := newRecordUpdateTestCtx()

	sharedBefore := findContributor(
		testutil.SelectContributorsByUserId(t, testDb, testUid), updSharedContributorId,
	)
	require.NotNil(t, sharedBefore)

	// Ph0では貢献者と書誌は1対1なので、洗い替えで旧貢献者はすべて参照されなくなり物理削除される
	err := repo.Update(
		ctx, testDb,
		withComponents(newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "貢献者検証の主題"), 1, "new"),
	)
	require.NoError(t, err)

	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	for _, id := range []int64{9201, 9202, 9203} {
		assert.Nil(t, findContributor(contributors, id), "貢献者 %d が物理削除されていない", id)
	}

	// 他の書誌(updSharedBId)から参照されている貢献者は、洗い替えでも削除しない。
	// Ph0では発生しない状態だが、Ph1の名寄せ導入時に他の書誌の貢献者を消さないための条件
	// (deleteUnreferencedContributorsQueryBase の not exists)を固定する。
	err = repo.Update(
		ctx, testDb,
		withComponents(newUpdateEntry(updSharedAId, record.RecordTypeAuthoredBook, "共有貢献者の主題"), 1, "shared"),
	)
	require.NoError(t, err)

	sharedAfter := findContributor(
		testutil.SelectContributorsByUserId(t, testDb, testUid), updSharedContributorId,
	)
	require.NotNil(t, sharedAfter, "他の書誌から参照される貢献者が削除された")
	assert.Equal(t, untouchedPg, sharedAfter.UpdPg)
	assert.WithinDuration(t, sharedBefore.UpdDt, sharedAfter.UpdDt, time.Microsecond)

	// 参照元(updSharedBId)の貢献も残っている
	assert.Len(t, testutil.SelectContributionsByRecordId(t, testDb, updSharedBId), 1)
}

func TestRecordRepository_Update_正常_コンテナIDの付け外しができる(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	ctx := newRecordUpdateTestCtx()

	// なし -> あり
	attach := newUpdateEntry(updMinimalId, record.RecordTypeAuthoredBook, "コンテナを付けた主題")
	attach.ContainerId = strPtr(updFreeContainerId)
	require.NoError(t, repo.Update(ctx, testDb, attach))

	actual, err := testutil.SelectRecordByPk(t, testDb, updMinimalId)
	require.NoError(t, err)
	assert.Equal(t, updFreeContainerId, actual.ContainerId.String)

	// あり -> 別のコンテナ
	swap := newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "コンテナを付け替えた主題")
	swap.ContainerId = strPtr(updFreeContainerId)
	require.NoError(t, repo.Update(ctx, testDb, swap))

	actual, err = testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)
	assert.Equal(t, updFreeContainerId, actual.ContainerId.String)

	// あり -> なし(収録物自体の論理削除はusecaseの責務で、repositoryは参照を外すだけ)
	require.NoError(t, repo.Update(
		ctx, testDb,
		newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "コンテナを外した主題"),
	))

	actual, err = testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)
	assert.False(t, actual.ContainerId.Valid)

	// 外された収録物は残ったまま(削除しない)
	container, err := testutil.SelectRecordByPk(t, testDb, updContainerId)
	require.NoError(t, err)
	assert.Equal(t, 0, container.DelFlg)
}

func TestRecordRepository_Update_正常_収録物自体を更新できる(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)

	// 収録物も書誌と同じメソッドで更新する(usecaseのapplyContainerから呼ばれる経路)。
	// 収録物を単体の更新対象にしない判断はusecaseの責務であり、repositoryは区別しない。
	input := withComponents(
		newUpdateEntry(updContainerId, record.RecordTypeJournal, "更新後の雑誌"),
		1,
		"container",
	)

	err := repo.Update(newRecordUpdateTestCtx(), testDb, input)
	require.NoError(t, err)

	actual, err := testutil.SelectRecordByPk(t, testDb, updContainerId)
	require.NoError(t, err)
	assert.Equal(t, "更新後の雑誌", actual.MainTitle)
	assert.Equal(t, ctxutil.REF_API_004, actual.UpdPg)
	assertComponentCounts(t, updContainerId, 1)

	// 収録物を参照する書誌のCONTAINER_IDは変化しない
	article, err := testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)
	assert.Equal(t, updContainerId, article.ContainerId.String)
	assert.Equal(t, "upd_pg", article.UpdPg)
}

func TestRecordRepository_Update_正常_他の書誌と収録物に影響しない(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)

	untouchedBefore, err := testutil.SelectRecordByPk(t, testDb, updUntouchedId)
	require.NoError(t, err)
	untouchedIdsBefore := selectComponentIds(t, updUntouchedId)
	containerIdsBefore := selectComponentIds(t, updContainerId)

	err = repo.Update(
		newRecordUpdateTestCtx(), testDb,
		withComponents(newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "影響範囲検証の主題"), 2, "scoped"),
	)
	require.NoError(t, err)

	// 無関係な書誌の本体
	untouchedAfter, err := testutil.SelectRecordByPk(t, testDb, updUntouchedId)
	require.NoError(t, err)
	assert.Equal(t, untouchedPg, untouchedAfter.UpdPg)
	assert.WithinDuration(t, untouchedBefore.UpdDt, untouchedAfter.UpdDt, time.Microsecond)
	assert.Equal(t, untouchedIdsBefore, selectComponentIds(t, updUntouchedId))

	// 収録物の構成要素(削除条件はE_RECORD_IDであり、CONTAINER_IDを辿らない)
	assert.Equal(t, containerIdsBefore, selectComponentIds(t, updContainerId))

	// 他の書誌の貢献者も残っている
	contributors := testutil.SelectContributorsByUserId(t, testDb, testUid)
	for _, id := range []int64{9204, updSharedContributorId, 9241, 9242} {
		assert.NotNil(t, findContributor(contributors, id), "貢献者 %d が削除された", id)
	}
}

func TestRecordRepository_Update_異常_更新対象が見つからない場合はErrNotFound(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)
	ctx := newRecordUpdateTestCtx()

	// 存在しない / 他ユーザ / 論理削除済み はいずれもWHERE条件に一致せず、
	// RowsAffected()==0 から ErrNotFound に変換される。
	cases := []struct {
		name     string
		recordId string
	}{
		{"存在しない書誌ID", updNotExistsId},
		{"他ユーザの書誌", updOtherUserId},
		{"論理削除済みの書誌", updDeletedId},
	}

	for _, c := range cases {
		err := repo.Update(
			ctx, testDb,
			withComponents(newUpdateEntry(c.recordId, record.RecordTypeAuthoredBook, "更新されない主題"), 1, "ng"),
		)
		assert.ErrorIs(t, err, repository.ErrNotFound, c.name)
	}

	// updateRecordEntryで即returnするため、構成要素の物理削除まで進まない(処理順の固定)。
	for _, recordId := range []string{updOtherUserId, updDeletedId} {
		actual, err := testutil.SelectRecordByPk(t, testDb, recordId)
		require.NoError(t, err)
		assert.NotEqual(t, ctxutil.REF_API_004, actual.UpdPg, "%s の本体が更新された", recordId)

		assertComponentCounts(t, recordId, 1)
	}

	// 貢献者も消えていない
	assert.NotNil(t, findContributor(testutil.SelectContributorsByUserId(t, testDb, 1002), 9221))
	assert.NotNil(t, findContributor(testutil.SelectContributorsByUserId(t, testDb, testUid), 9231))
}

func TestRecordRepository_Update_異常_存在しないコンテナIDはFK違反(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)

	input := newUpdateEntry(updMinimalId, record.RecordTypeAuthoredBook, "FK違反の主題")
	input.ContainerId = strPtr("UPD_NOTEXISTS_CONTAINER01")

	// コンテナの存在はusecaseが保証するが、fk_er_containerが最後の防波堤として効くこと。
	// 該当行はあるため ErrNotFound ではなく、生のDBエラーが返る。
	err := repo.Update(newRecordUpdateTestCtx(), testDb, input)
	require.Error(t, err)
	assert.NotErrorIs(t, err, repository.ErrNotFound)

	actual, err := testutil.SelectRecordByPk(t, testDb, updMinimalId)
	require.NoError(t, err)
	assert.Equal(t, "更新前の単著書", actual.MainTitle)
	assert.False(t, actual.ContainerId.Valid)
	assert.Equal(t, "upd_pg", actual.UpdPg)
}

func TestRecordRepository_Update_異常_APIIDが未設定(t *testing.T) {
	repo := testRecordsSetup(t, recordUpdateFixture)

	before := selectComponentIds(t, updFullId)

	// API IDはctxutil.GetApiIdがpanicする。SQL発行前に落ちるため1行も変化しない。
	assert.Panics(t, func() {
		_ = repo.Update(
			context.Background(), testDb,
			withComponents(newUpdateEntry(updFullId, record.RecordTypeJournalArticle, "panicする主題"), 1, "panic"),
		)
	})

	actual, err := testutil.SelectRecordByPk(t, testDb, updFullId)
	require.NoError(t, err)
	assert.Equal(t, "更新前の主題", actual.MainTitle)
	assert.Equal(t, "upd_pg", actual.UpdPg)
	assert.Equal(t, before, selectComponentIds(t, updFullId))
}
