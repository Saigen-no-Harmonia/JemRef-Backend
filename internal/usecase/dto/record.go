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
