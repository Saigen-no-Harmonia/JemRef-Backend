package mock

import (
	"context"
	"database/sql"
	"jemref_go/internal/repository"
)

type MockTxManager struct {
	WithTransactionFunc func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error

	Called                int
	WithTransactionCalled int
}

func (m *MockTxManager) WithTransaction(
	ctx context.Context,
	fn func(ctx context.Context, tx repository.DBTX) error,
) error {
	m.Called++
	m.WithTransactionCalled++
	return m.WithTransactionFunc(ctx, fn)
}

var _ repository.TxManager = (*MockTxManager)(nil)

// NewMockTxManager 渡された更新処理をそのまま実行するMockTxManagerを返す
// トランザクション開始失敗やロールバックを再現したい場合はWithTransactionFuncを差し替える。
func NewMockTxManager() *MockTxManager {
	m := &MockTxManager{}
	m.WithTransactionFunc = func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error {
		return fn(ctx, &MockDBTX{})
	}

	return m
}

// MockDBTX repository.DBTXのスタブ
// usecaseはtxをrepositoryへ受け渡すだけで自身では使わないため何もしない
type MockDBTX struct{}

func (m *MockDBTX) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	return nil, nil
}

func (m *MockDBTX) QueryContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	return nil, nil
}

func (m *MockDBTX) QueryRowContext(
	ctx context.Context,
	query string,
	args ...any,
) *sql.Row {
	return nil
}

var _ repository.DBTX = (*MockDBTX)(nil)
