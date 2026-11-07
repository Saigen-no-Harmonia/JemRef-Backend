package db

import "strings"

// buildBulkInsertQuery 一括INSERTのクエリを行数分に展開する
//
// base は「INSERT INTO ... (カラム列挙) VALUES」まで、
// values は1行分のプレースホルダ("(?, ?, ...)")を渡す。
//
// 呼び出し側はvalues内のプレースホルダと同じ順序・同じ個数の引数を
// 行ごとに積む必要がある(引数の順序ズレはテーブル定義とのズレとして
// internal/infrastructure/db の統合テストで検出される)。
//
// rowsが0以下の場合はpanicする。VALUES句のない INSERT は妥当なSQLではないため
// 返せる値がなく、空文字を返すと実行時にSQL構文エラーとして現れて原因が追いにくくなる。
// 呼び出し側で件数0を除外しておくこと(呼び出し側のバグとして落とす)。
func buildBulkInsertQuery(base string, values string, rows int) string {
	if rows <= 0 {
		panic("buildBulkInsertQuery: rows must be greater than 0")
	}

	placeholders := make([]string, 0, rows)
	for range rows {
		placeholders = append(placeholders, values)
	}

	return base + strings.Join(placeholders, ", ")
}
