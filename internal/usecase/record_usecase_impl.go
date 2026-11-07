package usecase

import (
	"context"
	"errors"
	"fmt"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
)

type RecordUsecaseImpl struct {
	recordRepo repository.RecordRepository
	txManager  repository.TxManager
	idGen      id.Generator
}

func NewRecordUsecaseImpl(
	r repository.RecordRepository,
	tx repository.TxManager,
	ig id.Generator,
) *RecordUsecaseImpl {
	return &RecordUsecaseImpl{
		recordRepo: r,
		txManager:  tx,
		idGen:      ig,
	}
}

var _ RecordUsecase = (*RecordUsecaseImpl)(nil)

func (u *RecordUsecaseImpl) Create(
	ctx context.Context,
	ci dto.CreateRecordInput,
) (*dto.CreateRecordOutput, error) {
	if err := chkInsertLimit(ci); err != nil {
		return nil, err
	}

	containerId := ""
	var container *record.UserRecordEntry
	if ci.ContainerInput != nil {
		container = extractContainer(ci, u.idGen.Generate())
	}

	// container idはコンテナがある場合のみ値を入れる
	if container != nil {
		containerId = container.RecordEntryId
	}

	recordId := u.idGen.Generate()
	article := extractArticle(ci, recordId, containerId)

	err := u.txManager.WithTransaction(
		ctx,
		func(ctx context.Context, tx repository.DBTX) error {
			if container != nil {
				if err := u.recordRepo.Create(
					ctx, tx, *container,
				); err != nil {
					return fmt.Errorf("create container record error: %w", err)
				}
			}

			if err := u.recordRepo.Create(
				ctx, tx, *article,
			); err != nil {
				return fmt.Errorf("create article record error: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &dto.CreateRecordOutput{
		RecordId: recordId,
	}, nil
}

// recordInsertLimit 子要素1種あたりの登録件数上限
// 子要素のDISPLAY_NUMBERはTINYINT(0起点)なので128件までしか登録できない。
const recordInsertLimit = 128

func chkInsertLimit(ci dto.CreateRecordInput) error {
	if len(ci.RecordInput.Contributors) > recordInsertLimit {
		return fmt.Errorf("貢献者情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if ci.ContainerInput != nil && len(ci.ContainerInput.Contributors) > recordInsertLimit {
		return fmt.Errorf("収録物の貢献者情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if len(ci.RecordInput.Memos) > recordInsertLimit {
		return fmt.Errorf("メモ情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if len(ci.RecordInput.Urls) > recordInsertLimit {
		return fmt.Errorf("URL情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	return nil
}

func extractContainer(input dto.CreateRecordInput, containerId string) *record.UserRecordEntry {
	ci := input.ContainerInput

	uid := input.InternalUserId
	optionValues := extractOptionValues(ci.OptionValues, uid, containerId)
	relations := extractContributorRelations(ci.Contributors)

	return &record.UserRecordEntry{
		RecordEntryId:   containerId,
		InternalUid:     uid,
		RecordType:      input.RecordInput.Type.GetParent(),
		MainTitle:       ci.MainTitle,
		SubTitle:        ci.SubTitle,
		Volume:          ci.Volume,
		Issue:           ci.Issue,
		Publisher:       input.RecordInput.Publisher,
		PublicationDate: input.RecordInput.PublicationDate,
		DatePrecision:   input.RecordInput.DatePrecision,
		ReadStatus:      ci.ReadStatus,
		DelFlg:          0,

		ContributorRelations: relations,
		OptionValues:         optionValues,
	}
}

func extractArticle(ci dto.CreateRecordInput, recordId string, containerId string) *record.UserRecordEntry {
	uid := ci.InternalUserId
	ri := ci.RecordInput

	optionValues := extractOptionValues(ri.OptionValues, uid, recordId)
	relations := extractContributorRelations(ri.Contributors)
	urls := extractUrls(ri.Urls, uid, recordId)
	memos := extractMemos(ri.Memos, uid, recordId)

	var containerIdPtr *string
	if containerId != "" {
		containerIdPtr = &containerId
	}

	return &record.UserRecordEntry{
		RecordEntryId:   recordId,
		InternalUid:     uid,
		RecordType:      ri.Type,
		MainTitle:       ri.MainTitle,
		SubTitle:        ri.SubTitle,
		PageRange:       ri.PageRange,
		Publisher:       ri.Publisher,
		PublicationDate: ri.PublicationDate,
		DatePrecision:   ri.DatePrecision,
		ReadStatus:      ri.ReadStatus,
		ParentId:        containerIdPtr,
		DelFlg:          0,

		ContributorRelations: relations,
		Urls:                 urls,
		Memos:                memos,
		OptionValues:         optionValues,
	}
}

// extractContributorRelations 貢献者と貢献情報のペアを取得する
// 書誌ID・内部用ユーザID・表示順はrepositoryが払い出すor取得可能なので移譲。
func extractContributorRelations(
	contributorInput []dto.ContributorInput,
) []record.ContributorRelation {
	var relations []record.ContributorRelation

	for _, c := range contributorInput {
		relations = append(relations, record.ContributorRelation{
			Contributor: record.ContributorEntry{
				RawName: c.Name,
				DelFlg:  0,
			},
			Contribution: record.ContributionEntry{
				Role:   c.Role,
				DelFlg: 0,
			},
		})
	}

	return relations
}

func extractMemos(mi []dto.MemoInput, uid int64, recordId string) []record.MemoEntry {
	var memos []record.MemoEntry
	for i, m := range mi {
		memos = append(memos, record.MemoEntry{
			RecordEntryId: recordId,
			DisplayNumber: i,
			Title:         m.Label,
			Content:       m.Content,
			InternalUid:   uid,
			DelFlg:        0,
		})
	}
	return memos
}

func extractUrls(ui []dto.UrlInput, uid int64, recordId string) []record.UrlEntry {
	var urls []record.UrlEntry
	for i, u := range ui {
		urls = append(urls, record.UrlEntry{
			RecordEntryId: recordId,
			DisplayNumber: i,
			Label:         u.Label,
			Url:           u.Value,
			InternalUid:   uid,
			DelFlg:        0,
		})
	}

	return urls
}

func extractOptionValues(oi *dto.OptionValuesInput, internalUid int64, recordId string) []record.OptionValue {
	if oi == nil {
		return nil
	}

	var optionValues []record.OptionValue
	if oi.Isbn != nil {
		optionValues = append(optionValues, createOptionValue(recordId, internalUid, record.RecordFieldIsbn, 10, *oi.Isbn))
	}

	if oi.Edition != nil {
		optionValues = append(optionValues, createOptionValue(recordId, internalUid, record.RecordFieldEdition, 20, *oi.Edition))
	}

	if oi.OriginalTitle != nil {
		optionValues = append(optionValues, createOptionValue(recordId, internalUid, record.RecordFieldOriginalTitle, 30, *oi.OriginalTitle))
	}

	return optionValues
}

func createOptionValue(recordId string, internalUid int64, field record.RecordField, displayNum int, value string) record.OptionValue {
	return record.OptionValue{
		RecordEntryId: recordId,
		RecordField:   field,
		DisplayNumber: displayNum,
		RawValue:      value,
		InternalUid:   internalUid,
		DelFlg:        0,
	}
}

// Delete 書誌情報を論理削除する
//
// コンテナ(親レコード)は、他に論理削除されていない子が残っていない場合のみ削除する。
// e_contributors はユーザ単位の入力辞書であり書誌に属さないため削除しない。
func (u *RecordUsecaseImpl) Delete(
	ctx context.Context,
	di dto.DeleteRecordInput,
) error {

	return u.txManager.WithTransaction(
		ctx,
		func(ctx context.Context, tx repository.DBTX) error {
			// 存在確認と所有権確認を兼ねる
			target, err := u.recordRepo.SelectByPk(ctx, tx, di.RecordId, di.InternalUserId)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return fmt.Errorf(
						"削除対象の書誌が存在しません。record id=%s :%w",
						di.RecordId,
						ErrRecordNotFound,
					)
				}
				return fmt.Errorf("select record error: %w", err)
			}

			// 収録物(他の書誌の親になっている書誌)は単体の削除対象にしない
			//
			// 収録物のIDはAPIレスポンスに露出しないため、指定された場合は
			// 不正なリクエストとして記録を残す(このエラーはリクエストログのerrorsに出力される)。
			// レスポンスは存在を漏らさないため404に寄せる。
			children, err := u.recordRepo.CountActiveChildren(ctx, tx, target.RecordEntryId)
			if err != nil {
				return fmt.Errorf("count children of target error: %w", err)
			}

			if children > 0 {
				return fmt.Errorf(
					"収録物の単体削除が要求されました。収録物は削除できません。record id=%s, internal uid=%d, child count=%d :%w",
					di.RecordId,
					di.InternalUserId,
					children,
					ErrRecordNotFound,
				)
			}

			// 子要素の論理削除処理
			if err := u.recordRepo.LogicalDelete(ctx, tx, target.RecordEntryId); err != nil {
				return fmt.Errorf("delete article record error: %w", err)
			}

			if target.ParentId == nil {
				return nil
			}

			// 親要素がある場合、親の論理削除処理に進む
			count, err := u.recordRepo.CountActiveChildren(ctx, tx, *target.ParentId)
			if err != nil {
				return fmt.Errorf("count container children error: %w", err)
			}

			// 親に紐づく生きている子が残っていれば削除しない
			if count > 0 {
				return nil
			}

			if err := u.recordRepo.LogicalDelete(ctx, tx, *target.ParentId); err != nil {
				return fmt.Errorf("delete container record error: %w", err)
			}

			return nil
		},
	)
}

// Get 書誌情報を1件取得する
//
// 本体・収録物・子要素を複数回のSELECTで読むため、一貫したスナップショットを得る目的で
// トランザクション内で実行する(SelectByPkは削除・更新と共用のためtxを必須としている)。
func (u *RecordUsecaseImpl) Get(
	ctx context.Context,
	gi dto.GetRecordInput,
) (*dto.GetRecordOutput, error) {

	var output dto.GetRecordOutput

	err := u.txManager.WithTransaction(
		ctx,
		func(ctx context.Context, tx repository.DBTX) error {
			// 存在確認と所有権確認を兼ねる
			article, err := u.recordRepo.SelectByPk(ctx, tx, gi.RecordId, gi.InternalUserId)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return fmt.Errorf(
						"参照対象の書誌が存在しません。record id=%s :%w",
						gi.RecordId,
						ErrRecordNotFound,
					)
				}
				return fmt.Errorf("select record error: %w", err)
			}

			// 収録物(他の書誌の親になっている書誌)は単体の参照対象にしない
			//
			// 収録物のIDはAPIレスポンスに露出しないため、指定された場合は
			// 不正なリクエストとして記録を残す(このエラーはリクエストログのerrorsに出力される)。
			// レスポンスは存在を漏らさないため404に寄せる。
			children, err := u.recordRepo.CountActiveChildren(ctx, tx, article.RecordEntryId)
			if err != nil {
				return fmt.Errorf("count children of target error: %w", err)
			}

			if children > 0 {
				return fmt.Errorf(
					"収録物の単体参照が要求されました。収録物は参照できません。record id=%s, internal uid=%d, child count=%d :%w",
					gi.RecordId,
					gi.InternalUserId,
					children,
					ErrRecordNotFound,
				)
			}

			var container *record.UserRecordEntry
			if article.ParentId != nil {
				container, err = u.recordRepo.SelectByPk(ctx, tx, *article.ParentId, gi.InternalUserId)
				if err != nil {
					if !errors.Is(err, repository.ErrNotFound) {
						return fmt.Errorf("select container record error: %w", err)
					}

					// 収録物が論理削除済みの場合は、収録物を持たない書誌として返す
					container = nil
				}
			}

			// containerがnilの場合も含めて渡す(repository側でnilを読み飛ばす)
			if err := u.recordRepo.SelectChildren(
				ctx, tx, []*record.UserRecordEntry{article, container},
			); err != nil {
				return fmt.Errorf("select children error: %w", err)
			}

			output.Record = *article
			output.Container = container

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
