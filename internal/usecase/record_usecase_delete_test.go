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

// -------------Delete(context.Context, dto.DeleteRecordInput)のテスト-------------

// recordDeleteCase Deleteのテストケース
//
// 正常系・異常系は同じフィクスチャと実行部(runRecordDeleteCase)を共有する。
type recordDeleteCase struct {
	name  string
	input dto.DeleteRecordInput
	// selectResult / selectErr はSelectByPkの戻り値
	selectResult *record.UserRecordEntry
	selectErr    error
	// countResults / countErrs はCountActiveArticlesが呼び出し順に返す値
	countResults []int
	countErrs    []error
	// deleteErrs はLogicalDeleteが呼び出し順に返すエラー(nilは成功)
	deleteErrs []error
	// withTransactionFunc nilの場合は更新処理をそのまま実行する
	withTransactionFunc func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error
	// expectedCountedIds CountActiveArticlesに渡された書誌IDの呼び出し順
	expectedCountedIds []string
	// expectedDeletedIds LogicalDeleteに渡された書誌IDの呼び出し順
	expectedDeletedIds   []string
	expectedSelectCalled int
	expectedCountCalled  int
	expectedDeleteCalled int
	expectedTxCalled     int
	expectedErrorBody    error
	expectedErrorMsgs    []string
}

// recordDeleteCases Deleteの正常系・異常系のテストケースを返す
//
// フィクスチャを両方で共有するため、ケースの定義はここにまとめる。
func recordDeleteCases() (normal, abnormal []recordDeleteCase) {
	// SelectByPkが返す書誌(RecordEntryIdとContainerIdのみ設定される軽量SELECT)
	soloTarget := &record.UserRecordEntry{RecordEntryId: recordArticleId}
	articleTarget := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		ContainerId:   testutil.Ptr(recordContainerId),
	}

	input := dto.DeleteRecordInput{
		InternalUserId: recordTestUid,
		RecordId:       recordArticleId,
	}

	normal = []recordDeleteCase{
		{
			name:                 "正常_コンテナなしの書誌を削除",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			// コンテナがないため書誌件数の確認は1回(自身が収録物かの判定)のみ
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
		{
			name:         "正常_コンテナありで他に書誌がない場合は収録物も削除",
			input:        input,
			selectResult: articleTarget,
			// 1回目=自身の書誌件数, 2回目=収録物に残る書誌件数
			countResults: []int{0, 0},
			// 書誌 -> 収録物の順に削除される
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
		},
		{
			name:         "正常_コンテナありで他に書誌が残っている場合は収録物を残す",
			input:        input,
			selectResult: articleTarget,
			countResults: []int{0, 1},
			// 収録物は削除されない
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
		{
			name:                 "正常_入力がそのままrepositoryへ渡る",
			input:                input,
			selectResult:         articleTarget,
			countResults:         []int{0, 0},
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
		},
		{
			// 収録物ガードが書誌を持たない書誌を誤って弾かないことの確認
			name:                 "正常_書誌を持たない書誌は削除できる",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
	}

	abnormal = []recordDeleteCase{
		{
			name:      "異常_書誌が存在しない",
			input:     input,
			selectErr: repository.ErrNotFound,
			// 存在確認で中断するため以降は呼ばれない
			expectedSelectCalled: 1,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectedErrorBody:    ErrRecordNotFound,
			expectedErrorMsgs:    []string{"削除対象の書誌が存在しません"},
		},
		{
			name:         "異常_収録物を指定した場合は削除されない",
			input:        input,
			selectResult: soloTarget,
			// 自身が3件の書誌を持つ=収録物
			countResults:         []int{3},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectedErrorBody:    ErrRecordNotFound,
			// ログに残す必要がある情報がメッセージに含まれること
			expectedErrorMsgs: []string{
				"収録物の単体削除が要求されました",
				"record id=" + recordArticleId,
				"internal uid=1001",
				"article count=3",
			},
		},
		{
			name:                 "異常_SelectByPkが予期しないエラー",
			input:                input,
			selectErr:            repository.ErrTestUnexpected,
			expectedSelectCalled: 1,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"select record error"},
		},
		{
			name:                 "異常_対象の書誌件数取得に失敗",
			input:                input,
			selectResult:         soloTarget,
			countErrs:            []error{repository.ErrTestUnexpected},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"count articles of target error"},
		},
		{
			name:                 "異常_書誌の論理削除に失敗",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			deleteErrs:           []error{repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"delete article record error"},
		},
		{
			name:         "異常_収録物の書誌件数取得に失敗",
			input:        input,
			selectResult: articleTarget,
			// 1回目は成功、2回目(収録物の書誌件数)で失敗
			countResults:         []int{0, 0},
			countErrs:            []error{nil, repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"count container articles error"},
		},
		{
			name:         "異常_収録物の論理削除に失敗",
			input:        input,
			selectResult: articleTarget,
			countResults: []int{0, 0},
			// 2回目(収録物)の削除で失敗
			deleteErrs:           []error{nil, repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"delete container record error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: input,
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				// 更新処理を実行せずにエラーを返す
				return repository.ErrTestUnexpected
			},
			expectedSelectCalled: 0,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectedErrorBody:    repository.ErrTestUnexpected,
		},
	}

	return normal, abnormal
}

func TestRecordUsecase_Delete_正常(t *testing.T) {
	normal, _ := recordDeleteCases()
	for _, tt := range normal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordDeleteCase(t, tt)
		})
	}
}

func TestRecordUsecase_Delete_異常(t *testing.T) {
	_, abnormal := recordDeleteCases()
	for _, tt := range abnormal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordDeleteCase(t, tt)
		})
	}
}

// runRecordDeleteCase Deleteを実行し、repositoryの呼び出しと結果を検証する
func runRecordDeleteCase(t *testing.T, tt recordDeleteCase) {
	t.Helper()

	repo := &mock.MockRecordRepository{}

	repo.SelectByPkFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error) {
		if tt.selectErr != nil {
			return nil, tt.selectErr
		}
		return tt.selectResult, nil
	}

	repo.CountActiveArticlesFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		containerRecordId string,
	) (int, error) {
		// CountActiveArticlesの呼び出し回数に応じた戻り値を返す
		idx := repo.CountActiveArticlesCalled - 1
		if idx < len(tt.countErrs) && tt.countErrs[idx] != nil {
			return 0, tt.countErrs[idx]
		}
		require.Less(t, idx, len(tt.countResults), "想定外のCountActiveArticles呼び出し")
		return tt.countResults[idx], nil
	}

	repo.LogicalDeleteFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
	) error {
		// LogicalDeleteの呼び出し回数に応じたエラーを返す
		idx := repo.LogicalDeleteCalled - 1
		if idx < len(tt.deleteErrs) {
			return tt.deleteErrs[idx]
		}
		return nil
	}

	txManager := mock.NewMockTxManager()
	if tt.withTransactionFunc != nil {
		txManager.WithTransactionFunc = tt.withTransactionFunc
	}

	uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

	err := uc.Delete(context.Background(), tt.input)

	// 呼び出しチェック
	assert.Equal(t, tt.expectedSelectCalled, repo.SelectByPkCalled)
	assert.Equal(t, tt.expectedCountCalled, repo.CountActiveArticlesCalled)
	assert.Equal(t, tt.expectedDeleteCalled, repo.LogicalDeleteCalled)
	assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

	// 論理削除に渡された書誌IDと順序の検証
	assert.Equal(t, tt.expectedDeletedIds, repo.DeletedRecordIds)

	// 入力がそのままrepositoryへ渡ること
	if tt.expectedSelectCalled > 0 {
		assert.Equal(t, tt.input.RecordId, repo.LastSelectByPkRecordId)
		assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPkInternalUid)
	}
	// 書誌件数の確認は1回目が対象自身(収録物かの判定)、2回目が収録物に対して行われる
	assert.Equal(t, tt.expectedCountedIds, repo.CountedRecordIds)

	// 異常系の場合のassertion
	if tt.expectedErrorBody != nil {
		require.Error(t, err)
		assert.ErrorIs(t, err, tt.expectedErrorBody)
		for _, msg := range tt.expectedErrorMsgs {
			assert.Contains(t, err.Error(), msg)
		}
		return
	}

	// 正常系の場合のassertion
	require.NoError(t, err)
}
