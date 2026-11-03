package db

import (
	"context"
	"database/sql"
	"errors"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errTxTest WithTransactionのロールバック確認用エラー
var errTxTest = errors.New("tx test error")

// -------------WithTransaction(context.Context, func(context.Context, repository.DBTX) error)のテスト-------------

func TestTxManager_WithTransaction_正常_コミットされる(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)
	txManager := NewTxManager(testDb)

	input := newMinimalRecordEntry(testRecordId)
	beforeCnt := testutil.CountRecords(t, testDb)

	called := false
	err := txManager.WithTransaction(
		newRecordTestCtx(),
		func(ctx context.Context, tx repository.DBTX) error {
			called = true
			return repo.Create(ctx, tx, input)
		},
	)
	require.NoError(t, err)
	assert.True(t, called)

	assert.Equal(t, 1, testutil.CountRecords(t, testDb)-beforeCnt)

	actual, err := testutil.SelectRecordByPk(t, testDb, testRecordId)
	require.NoError(t, err)
	assert.Equal(t, testRecordId, actual.Id)
	assert.Equal(t, ctxutil.REF_API_003, actual.InsPg)
	assert.Equal(t, "system", actual.InsId)
}

func TestTxManager_WithTransaction_異常_関数がエラーならロールバックされる(t *testing.T) {
	repo := testRecordsSetup(t, recordFixture)
	txManager := NewTxManager(testDb)

	input := newMinimalRecordEntry(testRecordId)
	input.ContributorRelations = []record.ContributorRelation{
		newMinimalRelation("貢献者A", record.ContributorRoleAuthor),
	}

	beforeCnt := testutil.CountRecords(t, testDb)

	err := txManager.WithTransaction(
		newRecordTestCtx(),
		func(ctx context.Context, tx repository.DBTX) error {
			// 登録に成功した後でエラーを返す
			if err := repo.Create(ctx, tx, input); err != nil {
				return err
			}
			return errTxTest
		},
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, errTxTest)

	// 登録ずみのデータが残っていないこと
	assert.Equal(t, beforeCnt, testutil.CountRecords(t, testDb))
	assert.Equal(t, 0, testutil.CountContributors(t, testDb))
	assert.Equal(t, 0, testutil.CountContributions(t, testDb))

	_, err = testutil.SelectRecordByPk(t, testDb, testRecordId)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestTxManager_WithTransaction_異常_トランザクション開始失敗(t *testing.T) {
	// 接続できないDSNを指定してBeginTxを失敗させる
	db, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:1)/testdb")
	require.NoError(t, err)
	defer db.Close()

	txManager := NewTxManager(db)

	called := false
	err = txManager.WithTransaction(
		context.Background(),
		func(ctx context.Context, tx repository.DBTX) error {
			called = true
			return nil
		},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction error")
	assert.False(t, called)
}
