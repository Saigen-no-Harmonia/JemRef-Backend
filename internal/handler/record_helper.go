package handler

import (
	"fmt"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/handler/dto"
	usecasedto "jemref_go/internal/usecase/dto"
)

func toCreateRecordInput(internalUid int64, req dto.CreateRecordRequest) (usecasedto.CreateRecordInput, error) {

	if err := chkCreateParam(req); err != nil {
		return usecasedto.CreateRecordInput{}, err
	}

	recordInput, err := toRecordInput(req.RecordDetailItem)
	if err != nil {
		return usecasedto.CreateRecordInput{}, err
	}

	containerInput, err := toContainerInput(req.ContainerDetailItem)
	if err != nil {
		return usecasedto.CreateRecordInput{}, err
	}

	return usecasedto.CreateRecordInput{
		InternalUserId: internalUid,
		RecordInput:    recordInput,
		ContainerInput: containerInput,
	}, nil
}

// createRecordRequiredErrors Createエンドポイントのbinding:"required"違反に対応するエラー
// 必須制約自体はDTOのbindingタグに委ねており、ここでは不足フィールドとエラーコードの対応表を定義。
var createRecordRequiredErrors = map[string]error{
	"Type":      ErrRecordTypeRequired,
	"MainTitle": ErrMainTitleRequired,
}

// chkCreateParam レコード登録APIのうち、bindingタグで表現できないバリデーションを行う
func chkCreateParam(req dto.CreateRecordRequest) error {
	// 子レコードへのチェック
	recordType := record.RecordType(req.Type)
	if !recordType.IsValid() {
		return ErrInvalidRecordType
	}

	// 親レコードへのチェック
	container := req.ContainerDetailItem
	if container != nil {
		if !recordType.HasParent() {
			return ErrContainerNotAllowed
		}
	}

	return nil
}

func toRecordInput(ri dto.RecordDetailItem) (usecasedto.RecordInput, error) {

	pubDatePtr := nilIfEmpty(ri.PublicationDate)
	publicationDate, datePrecision, err := record.GetDateAndPrecision(pubDatePtr)
	if err != nil {
		return usecasedto.RecordInput{}, fmt.Errorf("出版日が不正です。publicationDate=%v: %w", *pubDatePtr, ErrInvalidPublicationDate)
	}

	readStatus, err := toReadStatus(ri.ReadStatus)
	if err != nil {
		return usecasedto.RecordInput{}, err
	}

	contributors, err := toContributorsInput(ri.Contributors)
	if err != nil {
		return usecasedto.RecordInput{}, err
	}

	memos, err := toMemosInput(ri.Memos)
	if err != nil {
		return usecasedto.RecordInput{}, err
	}

	urls, err := toUrlsInput(ri.Urls)
	if err != nil {
		return usecasedto.RecordInput{}, err
	}

	return usecasedto.RecordInput{
		Type:            record.RecordType(ri.Type), // typeはハンドラでvalidateずみ
		MainTitle:       ri.MainTitle,
		SubTitle:        nilIfEmpty(ri.SubTitle),
		PageRange:       nilIfEmpty(ri.PageRange),
		Publisher:       nilIfEmpty(ri.Publisher),
		PublicationDate: publicationDate,
		DatePrecision:   datePrecision,
		ReadStatus:      readStatus,
		Contributors:    contributors,
		Memos:           memos,
		Urls:            urls,
		OptionValues:    toOptionValuesInput(ri.OptionValues),
	}, nil
}

func toContainerInput(ci *dto.ContainerDetailItem) (*usecasedto.ContainerInput, error) {
	if ci == nil {
		return nil, nil
	}

	readStatus, err := toReadStatus(ci.ReadStatus)
	if err != nil {
		return nil, err
	}

	contributorsInput, err := toContributorsInput(ci.Contributors)
	if err != nil {
		return nil, err
	}

	return &usecasedto.ContainerInput{
		MainTitle:    ci.MainTitle,
		SubTitle:     nilIfEmpty(ci.SubTitle),
		Volume:       nilIfEmpty(ci.Volume),
		Issue:        nilIfEmpty(ci.Issue),
		ReadStatus:   readStatus,
		Contributors: contributorsInput,
		OptionValues: toOptionValuesInput(ci.OptionValues),
	}, nil
}

func toContributorsInput(ci []dto.ContributorInfo) ([]usecasedto.ContributorInput, error) {
	var contributors []usecasedto.ContributorInput
	for _, c := range ci {
		if c.Role == "" {
			return nil, fmt.Errorf("貢献者情報がある場合、役割は必須です。%w", ErrInvalidRequestBody)
		} else if c.Name == "" {
			return nil, fmt.Errorf("貢献者情報がある場合、氏名は必須です。%w", ErrInvalidRequestBody)
		}

		role, err := toRole(c.Role)
		if err != nil {
			return nil, err
		}

		cont := usecasedto.ContributorInput{
			Name: c.Name,
			Role: role,
		}
		contributors = append(contributors, cont)
	}
	return contributors, nil
}

func toReadStatus(s *string) (*record.ReadStatus, error) {
	readStatusPtr := nilIfEmpty(s)
	if readStatusPtr == nil {
		return nil, nil
	}

	ret := record.ReadStatus(*readStatusPtr)
	if !ret.IsValid() {
		return nil, fmt.Errorf("read status=%s: %w", ret, ErrInvalidReadStatus)
	}

	return &ret, nil
}

func toRole(s string) (record.ContributorRole, error) {
	ret := record.ContributorRole(s)
	if !ret.IsValid() {
		return "", fmt.Errorf("contributor role=%s: %w", s, ErrInvalidContributorRole)
	}

	return ret, nil
}

func toMemosInput(mi []dto.MemoInfo) ([]usecasedto.MemoInput, error) {
	var memos []usecasedto.MemoInput
	for _, m := range mi {
		// メモの内容なしならエラー
		if m.Content == "" {
			return nil, fmt.Errorf("メモがある場合、内容は必須です。%w", ErrInvalidRequestBody)
		}

		memo := usecasedto.MemoInput{
			Label:   nilIfEmpty(m.Label),
			Content: m.Content,
		}
		memos = append(memos, memo)
	}
	return memos, nil
}

func toUrlsInput(ui []dto.UrlInfo) ([]usecasedto.UrlInput, error) {
	var urls []usecasedto.UrlInput
	for _, u := range ui {
		// URLなしならエラー
		if u.Value == "" {
			return nil, fmt.Errorf("URLがある場合、URLの値は必須です。%w", ErrInvalidRequestBody)
		}
		url := usecasedto.UrlInput{
			Label: nilIfEmpty(u.Label),
			Value: u.Value,
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func toOptionValuesInput(oi *dto.OptionValues) *usecasedto.OptionValuesInput {
	if oi == nil {
		return nil
	}

	isbn := nilIfEmpty(oi.Isbn)
	edition := nilIfEmpty(oi.Edition)
	originalTitle := nilIfEmpty(oi.OriginalTitle)
	if isbn == nil && edition == nil && originalTitle == nil {
		return nil
	}

	return &usecasedto.OptionValuesInput{
		Isbn:          isbn,
		Edition:       edition,
		OriginalTitle: originalTitle,
	}
}

// nilIfEmpty ポインタの指す値が空文字ならnilに変換する。DBのNULL登録用。
func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
