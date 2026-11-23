package usecase

import (
	"context"
	"testing"

	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------List(context.Context, dto.GetRecordsInput)のテスト-------------

const (
	// listSoloId1/2 収録物を持たない書誌のID
	listSoloId1 = "LISTSOLO000000000000000001"
	listSoloId2 = "LISTSOLO000000000000000002"
	// listArticleAId/BId 収録物1を持つ書誌のID(同じ収録物を共有する)
	listArticleAId = "LISTARTICLEA00000000000001"
	listArticleBId = "LISTARTICLEB00000000000001"
	// listArticleCId 収録物2を持つ書誌のID
	listArticleCId = "LISTARTICLEC00000000000001"
	// listContainerId1/2 収録物のID
	listContainerId1 = "LISTCONTAINER0000000000001"
	listContainerId2 = "LISTCONTAINER0000000000002"
)

// copyEntries 書誌の複製を返す
// mockの戻り値はテストケース間で共有するため、usecaseによる書き換え(貢献者の設定)が漏れないようにする。
func copyEntries(src []*record.UserRecordEntry) []*record.UserRecordEntry {
	if src == nil {
		return nil
	}

	copied := make([]*record.UserRecordEntry, 0, len(src))
	for _, e := range src {
		c := *e
		copied = append(copied, &c)
	}

	return copied
}

// recordListCase Listのテストケース
//
// 正常系・異常系は同じフィクスチャと実行部(runRecordListCase)を共有する。
type recordListCase struct {
	name  string
	input dto.GetRecordsInput

	// countResult / countErr CountByConditionの戻り値
	countResult int
	countErr    error
	// selectResults / selectErr SelectByConditionの戻り値
	selectResults []*record.UserRecordEntry
	selectErr     error
	// pkResults / pkErr SelectByPksの戻り値
	pkResults []*record.UserRecordEntry
	pkErr     error
	// contributorsFunc nilの場合は貢献者を設定せず成功させる
	contributorsFunc func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error
	// withTransactionFunc nilの場合は処理をそのまま実行する
	withTransactionFunc func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error

	expectedCountCalled        int
	expectedSelectCalled       int
	expectedPkCalled           int
	expectedContributorsCalled int
	expectedTxCalled           int
	// expectedCondition repositoryへ渡る検索条件(nilの場合は検証しない)
	expectedCondition *repository.RecordListCondition
	// expectedPkIds SelectByPksに渡された書誌IDの組
	expectedPkIds [][]string
	// expectedContributorIds SelectContributorRelationsに渡された書誌IDの組
	expectedContributorIds [][]string

	expectedTotal int
	// expectedRecordIds 出力の書誌IDの順。nilの場合はRecordsが空であることを検証する
	expectedRecordIds []string
	// expectedContainerIds 出力の各書誌に付いた収録物のID(収録物なしは "")
	expectedContainerIds []string
	// expectContributorsSet 書誌と収録物に、自身のIDの貢献者が1人ずつ付いていることを検証するか
	expectContributorsSet bool

	expectedErrorBody error
	expectedErrorMsgs []string
}

// recordListCases Listの正常系・異常系のテストケースを返す
//
// フィクスチャを両方で共有するため、ケースの定義はここにまとめる。
func recordListCases() (normal, abnormal []recordListCase) {
	newSolo := func(recordId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeAuthoredBook,
			MainTitle:     "単著書のタイトル",
		}
	}
	newContained := func(recordId string, containerId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeJournalArticle,
			MainTitle:     "論文のタイトル",
			ContainerId:   testutil.Ptr(containerId),
		}
	}
	newContainer := func(recordId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeJournal,
			MainTitle:     "雑誌のタイトル",
		}
	}

	soloArticle1 := newSolo(listSoloId1)
	soloArticle2 := newSolo(listSoloId2)
	containedArticleA := newContained(listArticleAId, listContainerId1)
	containedArticleB := newContained(listArticleBId, listContainerId1)
	containedArticleC := newContained(listArticleCId, listContainerId2)
	container1 := newContainer(listContainerId1)
	container2 := newContainer(listContainerId2)

	// baseInput デフォルト値相当の入力(handlerで補完済みの想定)
	baseInput := dto.GetRecordsInput{
		InternalUserId: recordTestUid,
		SortKey:        record.SortKeyPublishYear,
		SortOrder:      record.SortOrderAsc,
		Limit:          20,
		Offset:         0,
	}
	withOffset := func(offset int) dto.GetRecordsInput {
		in := baseInput
		in.Offset = offset
		return in
	}

	// fullInput 全項目を指定した入力(詰め替えの検証用)
	publishYear := 2022
	fullInput := dto.GetRecordsInput{
		InternalUserId:  2002,
		ContributorName: testutil.Ptr("木村"),
		Title:           testutil.Ptr("学校"),
		PublishYear:     &publishYear,
		SortKey:         record.SortKeyInsertDate,
		SortOrder:       record.SortOrderDesc,
		Limit:           50,
		Offset:          20,
	}

	// setContributors 渡されたentryに貢献者を1人追加する(repositoryの振る舞いの模倣)
	// どのentryに付いたかを判別できるよう、書誌IDを貢献者名に含める。
	// appendで追加するため、同じentryが2回渡されると貢献者が2人になり検出できる。
	setContributors := func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		for _, e := range entries {
			if e == nil {
				continue
			}
			e.ContributorRelations = append(e.ContributorRelations, record.ContributorRelation{
				Contributor:  record.ContributorEntry{RawName: "貢献者_" + e.RecordEntryId},
				Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor},
			})
		}
		return nil
	}

	normal = []recordListCase{
		{
			name:          "正常_収録物なしの書誌のみ",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{soloArticle1, soloArticle2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              nil,
			expectedContributorIds:     [][]string{{listSoloId1, listSoloId2}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listSoloId1, listSoloId2},
			expectedContainerIds:       []string{"", ""},
		},
		{
			name:          "正常_収録物ありの書誌",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			// 書誌の後に収録物を渡す
			expectedContributorIds: [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:          1,
			expectedRecordIds:      []string{listArticleAId},
			expectedContainerIds:   []string{listContainerId1},
		},
		{
			name:          "正常_収録物あり・なしの混在",
			input:         baseInput,
			countResult:   3,
			selectResults: []*record.UserRecordEntry{soloArticle1, containedArticleA, containedArticleC},
			pkResults:     []*record.UserRecordEntry{container1, container2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			// 収録物IDは書誌の出現順に集める
			expectedPkIds: [][]string{{listContainerId1, listContainerId2}},
			expectedContributorIds: [][]string{{
				listSoloId1, listArticleAId, listArticleCId, listContainerId1, listContainerId2,
			}},
			expectedTotal:        3,
			expectedRecordIds:    []string{listSoloId1, listArticleAId, listArticleCId},
			expectedContainerIds: []string{"", listContainerId1, listContainerId2},
		},
		{
			name:          "正常_同じ収録物を持つ書誌はIDを1回だけ渡す",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{containedArticleA, containedArticleB},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			// 収録物は1回だけ渡す(同じIDの別entryを渡すと貢献者が片方にしか付かないため)
			expectedContributorIds: [][]string{{listArticleAId, listArticleBId, listContainerId1}},
			expectedTotal:          2,
			expectedRecordIds:      []string{listArticleAId, listArticleBId},
			expectedContainerIds:   []string{listContainerId1, listContainerId1},
		},
		{
			name:          "正常_収録物が見つからなければ収録物なし",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{containedArticleA, containedArticleC},
			// 収録物1は論理削除済みで返らない想定
			pkResults: []*record.UserRecordEntry{container2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1, listContainerId2}},
			expectedContributorIds:     [][]string{{listArticleAId, listArticleCId, listContainerId2}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listArticleAId, listArticleCId},
			// エラーにせず、収録物を持たない書誌として返す(詳細参照と同じ扱い)
			expectedContainerIds: []string{"", listContainerId2},
		},
		{
			name:             "正常_書誌と収録物に貢献者が設定される",
			input:            baseInput,
			countResult:      1,
			selectResults:    []*record.UserRecordEntry{containedArticleA},
			pkResults:        []*record.UserRecordEntry{container1},
			contributorsFunc: setContributors,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:              1,
			expectedRecordIds:          []string{listArticleAId},
			expectedContainerIds:       []string{listContainerId1},
			// 出力へのコピーが貢献者の取得より後に行われていること
			expectContributorsSet: true,
		},
		{
			name:             "正常_同じ収録物を共有する書誌に同じ貢献者が付く",
			input:            baseInput,
			countResult:      2,
			selectResults:    []*record.UserRecordEntry{containedArticleA, containedArticleB},
			pkResults:        []*record.UserRecordEntry{container1},
			contributorsFunc: setContributors,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listArticleBId, listContainerId1}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listArticleAId, listArticleBId},
			expectedContainerIds:       []string{listContainerId1, listContainerId1},
			// 両方の書誌の収録物に、貢献者がちょうど1人ずつ付いていること
			expectContributorsSet: true,
		},
		{
			name:        "正常_総件数が0なら一覧を取得しない",
			input:       baseInput,
			countResult: 0,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedTotal:              0,
			expectedRecordIds:          nil,
		},
		{
			name:        "正常_offsetが総件数以上なら一覧を取得しない",
			input:       withOffset(40),
			countResult: 25,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			// 総件数はそのまま返す
			expectedTotal:     25,
			expectedRecordIds: nil,
		},
		{
			name:        "正常_offsetが総件数と同じなら一覧を取得しない",
			input:       withOffset(20),
			countResult: 20,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedTotal:              20,
			expectedRecordIds:          nil,
		},
		{
			name:          "正常_offsetが総件数未満なら一覧を取得する",
			input:         withOffset(20),
			countResult:   21,
			selectResults: []*record.UserRecordEntry{soloArticle1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedContributorIds:     [][]string{{listSoloId1}},
			expectedTotal:              21,
			expectedRecordIds:          []string{listSoloId1},
			expectedContainerIds:       []string{""},
		},
		{
			name:          "正常_総件数はlimitで切り詰めない",
			input:         baseInput,
			countResult:   25,
			selectResults: []*record.UserRecordEntry{soloArticle1, soloArticle2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedContributorIds:     [][]string{{listSoloId1, listSoloId2}},
			expectedTotal:              25,
			expectedRecordIds:          []string{listSoloId1, listSoloId2},
			expectedContainerIds:       []string{"", ""},
		},
		{
			name:          "正常_入力が変換されてrepositoryへ渡る",
			input:         fullInput,
			countResult:   21,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedCondition: &repository.RecordListCondition{
				InternalUid:     2002,
				ContributorName: testutil.Ptr("木村"),
				Title:           testutil.Ptr("学校"),
				PublishYear:     &publishYear,
				SortKey:         record.SortKeyInsertDate,
				SortOrder:       record.SortOrderDesc,
				Limit:           50,
				Offset:          20,
			},
			expectedPkIds:          [][]string{{listContainerId1}},
			expectedContributorIds: [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:          21,
			expectedRecordIds:      []string{listArticleAId},
			expectedContainerIds:   []string{listContainerId1},
		},
	}

	abnormal = []recordListCase{
		{
			name:     "異常_件数取得に失敗",
			input:    baseInput,
			countErr: ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"count records error"},
		},
		{
			name:        "異常_一覧取得に失敗",
			input:       baseInput,
			countResult: 1,
			selectErr:   ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select records error"},
		},
		{
			name:          "異常_収録物の取得に失敗",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkErr:         ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select container records error"},
		},
		{
			name:          "異常_貢献者の取得に失敗",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},
			contributorsFunc: func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				return ErrTestUnexpected
			},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listContainerId1}},
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select contributor relations error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: baseInput,
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				return ErrTestUnexpected
			},

			// クロージャが実行されないため、repositoryは一度も呼ばれない
			expectedCountCalled:        0,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedErrorBody:          ErrTestUnexpected,
		},
	}

	return normal, abnormal
}

func TestRecordUsecase_List_正常(t *testing.T) {
	normal, _ := recordListCases()
	for _, tt := range normal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordListCase(t, tt)
		})
	}
}

func TestRecordUsecase_List_異常(t *testing.T) {
	_, abnormal := recordListCases()
	for _, tt := range abnormal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordListCase(t, tt)
		})
	}
}

// runRecordListCase Listを実行し、repositoryの呼び出しと結果を検証する
func runRecordListCase(t *testing.T, tt recordListCase) {
	t.Helper()

	repo := &mock.MockRecordRepository{}

	repo.CountByConditionFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		cond repository.RecordListCondition,
	) (int, error) {
		if tt.countErr != nil {
			return 0, tt.countErr
		}
		return tt.countResult, nil
	}

	repo.SelectByConditionFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		cond repository.RecordListCondition,
	) ([]*record.UserRecordEntry, error) {
		if tt.selectErr != nil {
			return nil, tt.selectErr
		}
		return copyEntries(tt.selectResults), nil
	}

	repo.SelectByPksFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordIds []string,
		internalUid int64,
	) ([]*record.UserRecordEntry, error) {
		if tt.pkErr != nil {
			return nil, tt.pkErr
		}
		return copyEntries(tt.pkResults), nil
	}

	repo.SelectContributorRelationsFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		if tt.contributorsFunc != nil {
			return tt.contributorsFunc(ctx, tx, entries)
		}
		return nil
	}

	txManager := mock.NewMockTxManager()
	if tt.withTransactionFunc != nil {
		txManager.WithTransactionFunc = tt.withTransactionFunc
	}

	uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

	actual, err := uc.List(context.Background(), tt.input)

	// 呼び出しチェック
	assert.Equal(t, tt.expectedCountCalled, repo.CountByConditionCalled)
	assert.Equal(t, tt.expectedSelectCalled, repo.SelectByConditionCalled)
	assert.Equal(t, tt.expectedPkCalled, repo.SelectByPksCalled)
	assert.Equal(t, tt.expectedContributorsCalled, repo.SelectContributorRelationsCalled)
	assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

	// repositoryへ渡された値の検証
	assert.Equal(t, tt.expectedPkIds, repo.SelectedPkIds)
	assert.Equal(t, tt.expectedContributorIds, repo.ContributorTargetIds)

	if tt.expectedPkCalled > 0 {
		assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPksInternalUid)
	}

	if tt.expectedCondition != nil {
		assert.Equal(t, []repository.RecordListCondition{*tt.expectedCondition}, repo.CountedConditions)
		assert.Equal(t, []repository.RecordListCondition{*tt.expectedCondition}, repo.SelectedConditions)
	}

	// 件数と一覧は同じ条件で引く
	if tt.expectedSelectCalled > 0 {
		assert.Equal(t, repo.CountedConditions, repo.SelectedConditions)
	}

	// 異常系の場合のassertion
	if tt.expectedErrorBody != nil {
		require.Error(t, err)
		assert.ErrorIs(t, err, tt.expectedErrorBody)
		for _, msg := range tt.expectedErrorMsgs {
			assert.Contains(t, err.Error(), msg)
		}
		assert.Nil(t, actual)
		return
	}

	// 正常系の場合のassertion
	require.NoError(t, err)
	require.NotNil(t, actual)

	assert.Equal(t, tt.expectedTotal, actual.TotalCount)

	if tt.expectedRecordIds == nil {
		assert.Empty(t, actual.Records)
		return
	}

	recordIds := make([]string, 0, len(actual.Records))
	containerIds := make([]string, 0, len(actual.Records))
	for _, item := range actual.Records {
		recordIds = append(recordIds, item.Record.RecordEntryId)
		if item.Container == nil {
			containerIds = append(containerIds, "")
			continue
		}
		containerIds = append(containerIds, item.Container.RecordEntryId)
	}
	assert.Equal(t, tt.expectedRecordIds, recordIds)
	assert.Equal(t, tt.expectedContainerIds, containerIds)

	if !tt.expectContributorsSet {
		return
	}

	// 書誌と収録物に、自身のIDを含む貢献者がちょうど1人ずつ付いていること
	for _, item := range actual.Records {
		assertListContributorSet(t, item.Record)
		if item.Container != nil {
			assertListContributorSet(t, *item.Container)
		}
	}
}

// assertListContributorSet 書誌IDに対応した貢献者がちょうど1人付いていることを検証する
func assertListContributorSet(t *testing.T, e record.UserRecordEntry) {
	t.Helper()

	require.Len(t, e.ContributorRelations, 1, "record id=%s", e.RecordEntryId)
	assert.Equal(t, "貢献者_"+e.RecordEntryId, e.ContributorRelations[0].Contributor.RawName)
}
