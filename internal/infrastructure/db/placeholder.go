package db

import "strings"

// buildPlaceholders IN句に渡すプレースホルダを件数分生成する
//
// 例: 3 -> "?, ?, ?"
// クエリは「IN (」まで(Base)と「)」以降(Order)に分けて定義し、その間に挟んで使う。
//
// countが0以下の場合はpanicする。IN ()は妥当なSQLではないため返せる値がなく、
// 空文字を返すと実行時にSQL構文エラーとして現れて原因が追いにくくなる。
// 呼び出し側で件数0を除外しておくこと(呼び出し側のバグとして落とす)。
func buildPlaceholders(count int) string {
	if count <= 0 {
		panic("buildPlaceholders: count must be greater than 0")
	}

	placeholders := make([]string, 0, count)
	for range count {
		placeholders = append(placeholders, "?")
	}

	return strings.Join(placeholders, ", ")
}
