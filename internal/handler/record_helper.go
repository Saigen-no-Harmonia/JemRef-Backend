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
	"Name":      ErrContributorNameRequired,
	"Role":      ErrContributorRoleRequired,
	"Content":   ErrMemoContentRequired,
	"Value":     ErrUrlValueRequired,
}

// chkCreateParam レコード登録APIのうち、bindingタグで表現できないバリデーションを行う
func chkCreateParam(req dto.CreateRecordRequest) error {
	// 書誌へのチェック
	recordType := record.RecordType(req.Type)
	if !recordType.IsValid() {
		return ErrInvalidRecordType
	}

	// 収録物へのチェック
	container := req.ContainerDetailItem
	if container != nil {
		if !recordType.IsArticle() {
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
		Memos:           toMemosInput(ri.Memos),
		Urls:            toUrlsInput(ri.Urls),
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

// toContributorsInput 貢献者情報を変換する
// 必須チェックはDTOのbindingタグ(とtoBindingError)に委ねる。
func toContributorsInput(ci []dto.ContributorInfo) ([]usecasedto.ContributorInput, error) {
	var contributors []usecasedto.ContributorInput
	for _, c := range ci {
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

// toMemosInput メモ情報を変換する
//
// 内容の必須チェックはbindingタグ(とtoBindingError)に委ねる。
func toMemosInput(mi []dto.MemoInfo) []usecasedto.MemoInput {
	var memos []usecasedto.MemoInput
	for _, m := range mi {
		memo := usecasedto.MemoInput{
			Label:   nilIfEmpty(m.Label),
			Content: m.Content,
		}
		memos = append(memos, memo)
	}
	return memos
}

// toUrlsInput URL情報を変換する
//
// URLの値の必須チェックはbindingタグ(とtoBindingError)に委ねる。
func toUrlsInput(ui []dto.UrlInfo) []usecasedto.UrlInput {
	var urls []usecasedto.UrlInput
	for _, u := range ui {
		url := usecasedto.UrlInput{
			Label: nilIfEmpty(u.Label),
			Value: u.Value,
		}
		urls = append(urls, url)
	}
	return urls
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

// toGetRecordResponse 書誌詳細参照のレスポンスを組み立てる
//
// domainモデルには露出してはいけない項目(内部用ユーザID、削除フラグ等)が
// 含まれるため、ここで必要な項目だけを明示的に詰め替える。
func toGetRecordResponse(output *usecasedto.GetRecordOutput) dto.GetRecordResponse {
	return dto.GetRecordResponse{
		RecordId:            output.Record.RecordEntryId,
		RecordDetailItem:    toRecordDetailItem(output.Record),
		ContainerDetailItem: toContainerDetailItem(output.Container),
	}
}

func toRecordDetailItem(e record.UserRecordEntry) dto.RecordDetailItem {
	return dto.RecordDetailItem{
		Type:            string(e.RecordType),
		MainTitle:       e.MainTitle,
		SubTitle:        nilIfEmpty(e.SubTitle),
		PageRange:       nilIfEmpty(e.PageRange),
		Publisher:       nilIfEmpty(e.Publisher),
		PublicationDate: record.FormatPublicationDate(e.PublicationDate, e.DatePrecision),
		ReadStatus:      toReadStatusString(e.ReadStatus),
		Contributors:    toContributorInfos(e.ContributorRelations),
		Memos:           toMemoInfos(e.Memos),
		Urls:            toUrlInfos(e.Urls),
		OptionValues:    toOptionValuesResponse(e.OptionValues),
	}
}

// toContainerDetailItem 収録物のレスポンスを組み立てる
//
// 収録物のIDは露出しない。出版者・出版日は登録時に書誌側からコピーしているが、
// 書誌側が正であるためレスポンスには含めない(ContainerDetailItemにもフィールドがない)。
func toContainerDetailItem(e *record.UserRecordEntry) *dto.ContainerDetailItem {
	if e == nil {
		return nil
	}

	return &dto.ContainerDetailItem{
		MainTitle:    e.MainTitle,
		SubTitle:     nilIfEmpty(e.SubTitle),
		Volume:       nilIfEmpty(e.Volume),
		Issue:        nilIfEmpty(e.Issue),
		ReadStatus:   toReadStatusString(e.ReadStatus),
		Contributors: toContributorInfos(e.ContributorRelations),
		OptionValues: toOptionValuesResponse(e.OptionValues),
	}
}

func toContributorInfos(relations []record.ContributorRelation) []dto.ContributorInfo {
	var contributors []dto.ContributorInfo
	for _, r := range relations {
		contributors = append(contributors, dto.ContributorInfo{
			Name: r.Contributor.RawName,
			Role: string(r.Contribution.Role),
		})
	}

	return contributors
}

func toMemoInfos(entries []record.MemoEntry) []dto.MemoInfo {
	var memos []dto.MemoInfo
	for _, m := range entries {
		memos = append(memos, dto.MemoInfo{
			Label:   nilIfEmpty(m.Title),
			Content: m.Content,
		})
	}

	return memos
}

func toUrlInfos(entries []record.UrlEntry) []dto.UrlInfo {
	var urls []dto.UrlInfo
	for _, u := range entries {
		urls = append(urls, dto.UrlInfo{
			Label: nilIfEmpty(u.Label),
			Value: u.Url,
		})
	}

	return urls
}

// toOptionValuesResponse 項目値を入力項目ごとのフィールドへ振り分ける
//
// 1件も値がない場合はnilを返す(登録時のtoOptionValuesInputと揃える)。
func toOptionValuesResponse(values []record.OptionValue) *dto.OptionValues {
	if len(values) == 0 {
		return nil
	}

	var optionValues dto.OptionValues
	for _, v := range values {
		value := v.RawValue
		if value == "" {
			continue
		}

		switch v.RecordField {
		case record.RecordFieldIsbn:
			optionValues.Isbn = &value

		case record.RecordFieldEdition:
			optionValues.Edition = &value

		case record.RecordFieldOriginalTitle:
			optionValues.OriginalTitle = &value
		}
	}

	if optionValues.Isbn == nil &&
		optionValues.Edition == nil &&
		optionValues.OriginalTitle == nil {
		return nil
	}

	return &optionValues
}

// toReadStatusString 読了ステータスをレスポンス用の文字列に変換する
func toReadStatusString(s *record.ReadStatus) *string {
	if s == nil {
		return nil
	}

	status := string(*s)

	return &status
}
