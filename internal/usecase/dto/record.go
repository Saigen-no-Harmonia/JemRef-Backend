package dto

import (
	"jemref_go/internal/domain/record"
	"time"
)

type CreateRecordInput struct {
	InternalUserId int64
	RecordInput    RecordInput
	ContainerInput *ContainerInput
}

type RecordInput struct {
	Type            record.RecordType
	MainTitle       string
	SubTitle        *string
	PageRange       *string
	Publisher       *string
	PublicationDate *time.Time
	DatePrecision   *record.DatePrecision
	ReadStatus      *record.ReadStatus
	Contributors    []ContributorInput
	Memos           []MemoInput
	Urls            []UrlInput
	OptionValues    *OptionValuesInput
}

type ContainerInput struct {
	MainTitle    string
	SubTitle     *string
	Volume       *string
	Issue        *string
	ReadStatus   *record.ReadStatus
	Contributors []ContributorInput
	OptionValues *OptionValuesInput
}

type ContributorInput struct {
	Name string
	Role record.ContributorRole
}

type MemoInput struct {
	Label   *string
	Content string
}

type UrlInput struct {
	Label *string
	Value string
}

type OptionValuesInput struct {
	Isbn          *string
	Edition       *string
	OriginalTitle *string
}

type CreateRecordOutput struct {
	RecordId string
}

// [REF-API-004] 書誌情報更新
//
// 登録と同じ入力を受け取り、リクエストの内容でそのまま上書きする。
type UpdateRecordInput struct {
	InternalUserId int64
	RecordId       string
	RecordInput    RecordInput
	ContainerInput *ContainerInput
}

type DeleteRecordInput struct {
	InternalUserId int64
	RecordId       string
}

// [REF-API-002] 書誌詳細参照
type GetRecordInput struct {
	InternalUserId int64
	RecordId       string
}

// GetRecordOutput 書誌詳細参照の結果
//
// domainモデルをそのまま保持するが、DBに存在しない派生値(出典フォーマット等)が必要になった場合は、
// domainのメソッドで算出するか、このstructにフィールドを追加して対応する。
type GetRecordOutput struct {
	// 書誌本体
	Record record.UserRecordEntry
	// 収録物。持たない場合と、収録物が論理削除済みの場合はnil
	Container *record.UserRecordEntry
}

// [REF-API-001] 書誌一覧参照
//
// 検索条件(ContributorName / Title / PublishYear)はnilなら条件に含めない。
// 値の妥当性(並び替えキー・並び順・limitの範囲など)はhandlerで検証済み。
type GetRecordsInput struct {
	InternalUserId int64
	// 貢献者名。書誌または収録物の貢献者名の部分一致
	ContributorName *string
	// タイトル。書誌または収録物の主題・副題の部分一致
	Title *string
	// 出版年
	PublishYear *int
	SortKey     record.SortKey
	SortOrder   record.SortOrder
	Limit       int
	Offset      int
}

// GetRecordsOutput 書誌一覧参照の結果
type GetRecordsOutput struct {
	// 検索条件に一致する総件数(limit/offset適用前)
	TotalCount int
	// 並び替えとlimit/offsetを適用した書誌。該当なしの場合は空(nil)
	Records []RecordListItem
}

// RecordListItem 書誌一覧の1件
//
// 構成要素は貢献者だけを持つ(URL・メモ・項目値は取得しない)。
type RecordListItem struct {
	// 書誌本体
	Record record.UserRecordEntry
	// 収録物。持たない場合と、収録物が論理削除済みの場合はnil
	Container *record.UserRecordEntry
}
