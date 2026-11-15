package record

// SortKey 書誌一覧の並び替えキー
//
// DBに保存しない値なので、id⇔値のマップは持たない。
// どのカラムで並べるかの対応は infrastructure 層(ORDER BYのホワイトリスト)の責務とする。
type SortKey string

const (
	// SortKeyPublishYear 出版年月日で並べる
	SortKeyPublishYear SortKey = "publish_year"
	// SortKeyInsertDate 登録日時で並べる
	SortKeyInsertDate SortKey = "insert_date"
)

// IsValid 並び替えキーが正しいものであればtrueを返却する
func (k SortKey) IsValid() bool {
	switch k {
	case SortKeyPublishYear,
		SortKeyInsertDate:
		return true
	default:
		return false
	}
}

// SortOrder 書誌一覧の並び順
//
// リクエストのパラメータ名は sort_type だが、書誌タイプ(RecordType)との混同を避けるため
// 内部では SortOrder と呼ぶ。
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// IsValid 並び順が正しいものであればtrueを返却する
func (o SortOrder) IsValid() bool {
	switch o {
	case SortOrderAsc,
		SortOrderDesc:
		return true
	default:
		return false
	}
}
