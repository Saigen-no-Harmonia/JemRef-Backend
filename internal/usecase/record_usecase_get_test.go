package usecase

import (
	"context"
	"strconv"
	"testing"

	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------Get(context.Context, dto.GetRecordInput)のテスト-------------

// recordGetCase Getのテストケース
//
// 正常系・異常系は同じフィクスチャと実行部(runRecordGetCase)を共有する。
type recordGetCase struct {
	name  string
	input dto.GetRecordInput
	// selectResults / selectErrs はSelectByPkが呼び出し順に返す値(最大2回)
	selectResults []*record.UserRecordEntry
	selectErrs    []error
	// countResult / countErr はCountActiveArticlesの戻り値(必ず1回のみ呼ばれる)
	countResult int
	countErr    error
	// selectComponentsFunc nilの場合は構成要素を設定せず成功させる
	selectComponentsFunc func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error
	// withTransactionFunc nilの場合は処理をそのまま実行する
	withTransactionFunc func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error
	// expectedSelectedIds SelectByPkに渡された書誌IDの呼び出し順
	expectedSelectedIds []string
	// expectedCountedIds CountActiveArticlesに渡された書誌IDの呼び出し順
	expectedCountedIds []string
	// expectedComponentIds SelectComponentsに渡された書誌IDの組(nilは空文字で記録される)
	expectedComponentIds     [][]string
	expectedSelectCalled     int
	expectedCountCalled      int
	expectedComponentsCalled int
	expectedTxCalled         int
	// expectedContainerId 収録物の書誌ID。空文字の場合はContainerがnilであることを検証する
	expectedContainerId string
	// expectComponentsSet 構成要素が設定されていることを検証するか
	expectComponentsSet bool
	expectedErrorBody   error
	expectedErrorMsgs   []string
}

// recordGetCases Getの正常系・異常系のテストケースを返す
//
// フィクスチャを両方で共有するため、ケースの定義はここにまとめる。
func recordGetCases() (normal, abnormal []recordGetCase) {
	// SelectByPkが返す書誌(構成要素はSelectComponentsが設定するためここでは持たない)
	soloArticle := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeAuthoredBook,
		MainTitle:     "単著書のタイトル",
	}
	containedArticle := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeJournalArticle,
		MainTitle:     "論文のタイトル",
		ContainerId:   testutil.Ptr(recordContainerId),
	}
	container := &record.UserRecordEntry{
		RecordEntryId: recordContainerId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "雑誌のタイトル",
	}

	// setComponents 渡されたentryに構成要素を設定する(repositoryの振る舞いの模倣)
	// 書誌と収録物を区別できるよう、書誌IDを貢献者名に含める。
	setComponents := func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		for _, e := range entries {
			if e == nil {
				continue
			}
			e.ContributorRelations = []record.ContributorRelation{
				{
					Contributor:  record.ContributorEntry{RawName: "貢献者_" + e.RecordEntryId},
					Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor},
				},
			}
			e.Urls = []record.UrlEntry{{Url: "https://example.com/" + e.RecordEntryId}}
			e.Memos = []record.MemoEntry{{Content: "メモ_" + e.RecordEntryId}}
			e.OptionValues = []record.OptionValue{
				{RecordField: record.RecordFieldIsbn, RawValue: "9784130513555"},
			}
		}
		return nil
	}

	normal = []recordGetCase{
		{
			name:  "正常_収録物なしの書誌を取得",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      "",
		},
		{
			name:  "正常_収録物ありの書誌を取得",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{containedArticle, container},
			countResult:   0,

			// 1回目が書誌、2回目が収録物の順で引かれる
			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, recordContainerId}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      recordContainerId,
		},
		{
			name:  "正常_書誌と収録物に構成要素が設定される",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults:        []*record.UserRecordEntry{containedArticle, container},
			countResult:          0,
			selectComponentsFunc: setComponents,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, recordContainerId}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      recordContainerId,
			expectComponentsSet:      true,
		},
		{
			name:  "正常_収録物が論理削除済み",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			// 2回目(収録物)のSelectByPkがErrNotFoundを返す
			selectResults: []*record.UserRecordEntry{containedArticle, nil},
			selectErrs:    []error{nil, repository.ErrNotFound},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			// 収録物を持たない書誌として返す(エラーにしない)
			expectedContainerId: "",
		},
		{
			name:  "正常_入力がそのままrepositoryへ渡る",
			input: dto.GetRecordInput{InternalUserId: 2002, RecordId: "OTHERRECORDID0000000000001"},

			selectResults: []*record.UserRecordEntry{
				{RecordEntryId: "OTHERRECORDID0000000000001", InternalUid: 2002},
			},
			countResult: 0,

			expectedSelectedIds:      []string{"OTHERRECORDID0000000000001"},
			expectedCountedIds:       []string{"OTHERRECORDID0000000000001"},
			expectedComponentIds:     [][]string{{"OTHERRECORDID0000000000001", ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      "",
		},
	}

	abnormal = []recordGetCase{
		{
			name:  "異常_書誌が存在しない",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{nil},
			selectErrs:    []error{repository.ErrNotFound},

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       nil,
			expectedComponentIds:     nil,
			expectedSelectCalled:     1,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrRecordNotFound,
			expectedErrorMsgs:        []string{"参照対象の書誌が存在しません", recordArticleId},
		},
		{
			name:  "異常_収録物を指定した場合は参照できない",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordContainerId},

			selectResults: []*record.UserRecordEntry{container},
			// 生きている書誌が1件いる = 収録物である
			countResult: 1,

			expectedSelectedIds:      []string{recordContainerId},
			expectedCountedIds:       []string{recordContainerId},
			expectedComponentIds:     nil,
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrRecordNotFound,
			// リクエストログに残すため、書誌ID・ユーザID・書誌件数をメッセージに含める
			expectedErrorMsgs: []string{
				"収録物の単体参照が要求されました",
				recordContainerId,
				strconv.FormatInt(recordTestUid, 10),
				"article count=1",
			},
		},
		{
			name:  "異常_SelectByPkが予期しないエラー",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{nil},
			selectErrs:    []error{ErrTestUnexpected},

			expectedSelectedIds:      []string{recordArticleId},
			expectedSelectCalled:     1,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select record error"},
		},
		{
			name:  "異常_書誌件数の取得に失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countErr:      ErrTestUnexpected,

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"count articles of target error"},
		},
		{
			name:  "異常_収録物のSelectByPkが予期しないエラー",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			// ErrNotFound以外は収録物なしとして扱わず、エラーとして返す
			selectResults: []*record.UserRecordEntry{containedArticle, nil},
			selectErrs:    []error{nil, ErrTestUnexpected},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select container record error"},
		},
		{
			name:  "異常_SelectComponentsが失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countResult:   0,
			selectComponentsFunc: func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				return ErrTestUnexpected
			},

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select components error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				return ErrTestUnexpected
			},

			// クロージャが実行されないため、repositoryは一度も呼ばれない
			expectedSelectCalled:     0,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectedErrorBody:        ErrTestUnexpected,
		},
	}

	return normal, abnormal
}

func TestRecordUsecase_Get_正常(t *testing.T) {
	normal, _ := recordGetCases()
	for _, tt := range normal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordGetCase(t, tt)
		})
	}
}

func TestRecordUsecase_Get_異常(t *testing.T) {
	_, abnormal := recordGetCases()
	for _, tt := range abnormal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordGetCase(t, tt)
		})
	}
}

// runRecordGetCase Getを実行し、repositoryの呼び出しと結果を検証する
func runRecordGetCase(t *testing.T, tt recordGetCase) {
	t.Helper()

	repo := &mock.MockRecordRepository{}

	repo.SelectByPkFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error) {
		// SelectByPkの呼び出し回数に応じた戻り値を返す
		idx := repo.SelectByPkCalled - 1
		if idx < len(tt.selectErrs) && tt.selectErrs[idx] != nil {
			return nil, tt.selectErrs[idx]
		}
		require.Less(t, idx, len(tt.selectResults), "想定外のSelectByPk呼び出し")

		// 書誌はテストケース間で共有するため、複製を返して副作用を防ぐ
		src := tt.selectResults[idx]
		if src == nil {
			return nil, nil
		}
		copied := *src
		return &copied, nil
	}

	repo.CountActiveArticlesFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		containerRecordId string,
	) (int, error) {
		if tt.countErr != nil {
			return 0, tt.countErr
		}
		return tt.countResult, nil
	}

	repo.SelectComponentsFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		if tt.selectComponentsFunc != nil {
			return tt.selectComponentsFunc(ctx, tx, entries)
		}
		return nil
	}

	txManager := mock.NewMockTxManager()
	if tt.withTransactionFunc != nil {
		txManager.WithTransactionFunc = tt.withTransactionFunc
	}

	uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

	actual, err := uc.Get(context.Background(), tt.input)

	// 呼び出しチェック
	assert.Equal(t, tt.expectedSelectCalled, repo.SelectByPkCalled)
	assert.Equal(t, tt.expectedCountCalled, repo.CountActiveArticlesCalled)
	assert.Equal(t, tt.expectedComponentsCalled, repo.SelectComponentsCalled)
	assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

	// repositoryへ渡された書誌IDと順序の検証
	assert.Equal(t, tt.expectedSelectedIds, repo.SelectedRecordIds)
	assert.Equal(t, tt.expectedCountedIds, repo.CountedRecordIds)

	// SelectComponentsには常に[対象書誌, 収録物]の2要素を渡す(収録物なしはnil)
	assert.Equal(t, tt.expectedComponentIds, repo.ComponentTargetIds)

	// 入力がそのままrepositoryへ渡ること
	if tt.expectedSelectCalled > 0 {
		assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPkInternalUid)
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

	assert.Equal(t, tt.input.RecordId, actual.Record.RecordEntryId)

	if tt.expectedContainerId == "" {
		assert.Nil(t, actual.Container)
	} else {
		require.NotNil(t, actual.Container)
		assert.Equal(t, tt.expectedContainerId, actual.Container.RecordEntryId)
	}

	if !tt.expectComponentsSet {
		return
	}

	// SelectComponentsが設定した構成要素がOutputに反映されていること
	// (Outputへのコピーが SelectComponents の後に行われることの検証)
	assertComponentsSet(t, actual.Record)
	require.NotNil(t, actual.Container)
	assertComponentsSet(t, *actual.Container)
}

// assertComponentsSet 構成要素が設定され、書誌IDに対応した値になっていることを検証する
func assertComponentsSet(t *testing.T, e record.UserRecordEntry) {
	t.Helper()

	require.Len(t, e.ContributorRelations, 1)
	assert.Equal(t, "貢献者_"+e.RecordEntryId, e.ContributorRelations[0].Contributor.RawName)

	require.Len(t, e.Urls, 1)
	assert.Equal(t, "https://example.com/"+e.RecordEntryId, e.Urls[0].Url)

	require.Len(t, e.Memos, 1)
	assert.Equal(t, "メモ_"+e.RecordEntryId, e.Memos[0].Content)

	require.Len(t, e.OptionValues, 1)
	assert.Equal(t, record.RecordFieldIsbn, e.OptionValues[0].RecordField)
}
