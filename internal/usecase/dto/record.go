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
