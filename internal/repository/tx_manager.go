package repository

import "context"

type TxManager interface {

	// WithTransaction 引数fnに渡した更新処理をトランザクション付きで実行する
	WithTransaction(
		ctx context.Context,
		fn func(ctx context.Context, tx DBTX) error,
	) error
}
