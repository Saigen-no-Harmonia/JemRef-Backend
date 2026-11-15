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
	if err := chkInsertLimit(ci.RecordInput, ci.ContainerInput); err != nil {
		return nil, err
	}

	// 収録物がある場合のみ生成し、書誌にそのIDを持たせる
	var (
		container   *record.UserRecordEntry
		containerId *string
	)
	if ci.ContainerInput != nil {
		container = extractContainer(
			ci.RecordInput, *ci.ContainerInput, ci.InternalUserId, u.idGen.Generate(),
		)
		containerId = &container.RecordEntryId
	}

	recordId := u.idGen.Generate()
	article := extractArticle(ci.RecordInput, ci.InternalUserId, recordId, containerId)

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

// recordInsertLimit 構成要素1種あたりの登録件数上限
// 構成要素のDISPLAY_NUMBERはTINYINT(0起点)なので128件までしか登録できない。
const recordInsertLimit = 128

// chkInsertLimit 子要素の件数が登録上限を超えていないか検証する
//
// 登録・更新の両方から呼ぶため、入力DTOそのものではなく中身を受け取る。
func chkInsertLimit(ri dto.RecordInput, ci *dto.ContainerInput) error {
	if len(ri.Contributors) > recordInsertLimit {
		return fmt.Errorf("貢献者情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if ci != nil && len(ci.Contributors) > recordInsertLimit {
		return fmt.Errorf("収録物の貢献者情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if len(ri.Memos) > recordInsertLimit {
		return fmt.Errorf("メモ情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	if len(ri.Urls) > recordInsertLimit {
		return fmt.Errorf("URL情報の件数が登録上限を超えています。%w", ErrReachRegistrationLimit)
	}

	return nil
}

// extractContainer 収録物のエントリを組み立てる
//
// 出版情報(出版者・出版日・精度)は書誌側が正だが、収録物を単独で参照する場合に備えて
// 収録物にもコピーする。登録・更新の両方から呼ぶ。
func extractContainer(
	ri dto.RecordInput,
	ci dto.ContainerInput,
	uid int64,
	containerId string,
) *record.UserRecordEntry {

	optionValues := extractOptionValues(ci.OptionValues, uid, containerId)
	relations := extractContributorRelations(ci.Contributors)

	return &record.UserRecordEntry{
		RecordEntryId:   containerId,
		InternalUid:     uid,
		RecordType:      ri.Type.GetContainerType(),
		MainTitle:       ci.MainTitle,
		SubTitle:        ci.SubTitle,
		Volume:          ci.Volume,
		Issue:           ci.Issue,
		Publisher:       ri.Publisher,
		PublicationDate: ri.PublicationDate,
		DatePrecision:   ri.DatePrecision,
		ReadStatus:      ci.ReadStatus,
		DelFlg:          0,

		ContributorRelations: relations,
		OptionValues:         optionValues,
	}
}

// extractArticle 書誌本体のエントリを組み立てる
//
// containerIdがnilの場合は収録物を持たない書誌になる。登録・更新の両方から呼ぶ。
func extractArticle(
	ri dto.RecordInput,
	uid int64,
	recordId string,
	containerId *string,
) *record.UserRecordEntry {

	optionValues := extractOptionValues(ri.OptionValues, uid, recordId)
	relations := extractContributorRelations(ri.Contributors)
	urls := extractUrls(ri.Urls, uid, recordId)
	memos := extractMemos(ri.Memos, uid, recordId)

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
		ContainerId:     containerId,
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
// コンテナ(収録物)は、他に論理削除されていない書誌が残っていない場合のみ削除する。
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

			// 収録物(他の書誌のコンテナになっている書誌)は単体の削除対象にしない
			//
			// 収録物のIDはAPIレスポンスに露出しないため、指定された場合は
			// 不正なリクエストとして記録を残す(このエラーはリクエストログのerrorsに出力される)。
			// レスポンスは存在を漏らさないため404に寄せる。
			articleCount, err := u.recordRepo.CountActiveArticles(ctx, tx, target.RecordEntryId)
			if err != nil {
				return fmt.Errorf("count articles of target error: %w", err)
			}

			if articleCount > 0 {
				return fmt.Errorf(
					"収録物の単体削除が要求されました。収録物は削除できません。record id=%s, internal uid=%d, article count=%d :%w",
					di.RecordId,
					di.InternalUserId,
					articleCount,
					ErrRecordNotFound,
				)
			}

			// 構成要素の論理削除処理
			if err := u.recordRepo.LogicalDelete(ctx, tx, target.RecordEntryId); err != nil {
				return fmt.Errorf("delete article record error: %w", err)
			}

			if target.ContainerId == nil {
				return nil
			}

			// コンテナがある場合、コンテナの論理削除処理に進む
			count, err := u.recordRepo.CountActiveArticles(ctx, tx, *target.ContainerId)
			if err != nil {
				return fmt.Errorf("count container articles error: %w", err)
			}

			// コンテナに紐づく生きている書誌が残っていれば削除しない
			if count > 0 {
				return nil
			}

			if err := u.recordRepo.LogicalDelete(ctx, tx, *target.ContainerId); err != nil {
				return fmt.Errorf("delete container record error: %w", err)
			}

			return nil
		},
	)
}

// Get 書誌情報を1件取得する
//
// 本体・収録物・構成要素を複数回のSELECTで読むため、一貫したスナップショットを得る目的で
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

			// 収録物(他の書誌のコンテナになっている書誌)は単体の参照対象にしない
			//
			// 収録物のIDはAPIレスポンスに露出しないため、指定された場合は
			// 不正なリクエストとして記録を残す(このエラーはリクエストログのerrorsに出力される)。
			// レスポンスは存在を漏らさないため404に寄せる。
			articleCount, err := u.recordRepo.CountActiveArticles(ctx, tx, article.RecordEntryId)
			if err != nil {
				return fmt.Errorf("count articles of target error: %w", err)
			}

			if articleCount > 0 {
				return fmt.Errorf(
					"収録物の単体参照が要求されました。収録物は参照できません。record id=%s, internal uid=%d, article count=%d :%w",
					gi.RecordId,
					gi.InternalUserId,
					articleCount,
					ErrRecordNotFound,
				)
			}

			var container *record.UserRecordEntry
			if article.ContainerId != nil {
				container, err = u.recordRepo.SelectByPk(ctx, tx, *article.ContainerId, gi.InternalUserId)
				if err != nil {
					if !errors.Is(err, repository.ErrNotFound) {
						return fmt.Errorf("select container record error: %w", err)
					}

					// 収録物が論理削除済みの場合は、収録物を持たない書誌として返す
					container = nil
				}
			}

			// containerがnilの場合も含めて渡す(repository側でnilを読み飛ばす)
			if err := u.recordRepo.SelectComponents(
				ctx, tx, []*record.UserRecordEntry{article, container},
			); err != nil {
				return fmt.Errorf("select components error: %w", err)
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

// Update 書誌情報を更新する
//
// リクエストの内容でそのまま上書きする(未指定の項目はNULLで上書きされる)。
// 書誌タイプの変更は許容しない(コンテナを持てるか否かが変わり、整合が取れなくなるため)。
// 収録物は、リクエストに応じて新規登録・更新・削除のいずれかを行う。
func (u *RecordUsecaseImpl) Update(
	ctx context.Context,
	ui dto.UpdateRecordInput,
) error {

	if err := chkInsertLimit(ui.RecordInput, ui.ContainerInput); err != nil {
		return err
	}

	return u.txManager.WithTransaction(
		ctx,
		func(ctx context.Context, tx repository.DBTX) error {
			// 存在確認と所有権確認を兼ねる
			target, err := u.recordRepo.SelectByPk(ctx, tx, ui.RecordId, ui.InternalUserId)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return fmt.Errorf(
						"更新対象の書誌が存在しません。record id=%s :%w",
						ui.RecordId,
						ErrRecordNotFound,
					)
				}
				return fmt.Errorf("select record error: %w", err)
			}

			// 収録物(他の書誌のコンテナになっている書誌)は単体の更新対象にしない
			//
			// 収録物のIDはAPIレスポンスに露出しないため、指定された場合は
			// 不正なリクエストとして記録を残す。レスポンスは存在を漏らさないため404に寄せる。
			articleCount, err := u.recordRepo.CountActiveArticles(ctx, tx, target.RecordEntryId)
			if err != nil {
				return fmt.Errorf("count articles of target error: %w", err)
			}

			if articleCount > 0 {
				return fmt.Errorf(
					"収録物の単体更新が要求されました。収録物は更新できません。record id=%s, internal uid=%d, article count=%d :%w",
					ui.RecordId,
					ui.InternalUserId,
					articleCount,
					ErrRecordNotFound,
				)
			}

			// 書誌タイプの変更は親子関係の変更を伴うため許容しない
			if target.RecordType != ui.RecordInput.Type {
				return fmt.Errorf(
					"書誌タイプの変更が要求されました。record id=%s, before=%s, after=%s :%w",
					ui.RecordId,
					target.RecordType,
					ui.RecordInput.Type,
					ErrRecordTypeNotChangeable,
				)
			}

			containerId, err := u.applyContainer(ctx, tx, ui, target)
			if err != nil {
				return err
			}

			article := extractArticle(ui.RecordInput, ui.InternalUserId, ui.RecordId, containerId)
			if err := u.recordRepo.Update(ctx, tx, *article); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					// 同一トランザクション内でSelectByPkが成功しているため、ここには到達しない。
					// 到達した場合はデータ不整合であり、404(存在しない)ではなく500に寄せる。
					return fmt.Errorf(
						"更新対象の書誌が取得後に失われました。record id=%s :%w",
						ui.RecordId,
						ErrRecordUpdateFailed,
					)
				}
				return fmt.Errorf("update article record error: %w", err)
			}

			// 収録物を外した場合は、書誌のCONTAINER_IDを外した後に論理削除する
			if target.ContainerId != nil && ui.ContainerInput == nil {
				return u.deleteDetachedContainer(ctx, tx, *target.ContainerId)
			}

			return nil
		},
	)
}

// applyContainer 収録物の登録・更新を行い、書誌が持つべきコンテナIDを返す
//
// 返り値がnilの場合、書誌は収録物を持たない状態になる。
// 収録物の論理削除は、書誌のCONTAINER_IDを外した後に行う必要があるためここでは扱わない。
//
// repositoryはコンテナIDの所有権を検証しないため、返すIDは必ず所有権確認済みにする。
// 新規追加はサーバ側で生成したIDを自ユーザで登録し、既存は SelectByPk(id, uid) で確認する。
// リクエストでコンテナIDを受け取る仕様を追加する場合も、ここで同じ確認を通すこと。
func (u *RecordUsecaseImpl) applyContainer(
	ctx context.Context,
	tx repository.DBTX,
	ui dto.UpdateRecordInput,
	target *record.UserRecordEntry,
) (*string, error) {

	// 収録物を持たない状態にする(既存の収録物があれば呼び出し側が論理削除する)
	if ui.ContainerInput == nil {
		return nil, nil
	}

	// 収録物を新規に追加する場合
	if target.ContainerId == nil {
		containerId := u.idGen.Generate()
		container := extractContainer(
			ui.RecordInput, *ui.ContainerInput, ui.InternalUserId, containerId,
		)

		// 書誌のCONTAINER_IDが参照できるよう、先に登録する(FK順)
		if err := u.recordRepo.Create(ctx, tx, *container); err != nil {
			return nil, fmt.Errorf("create container record error: %w", err)
		}

		return &containerId, nil
	}

	// 既存の収録物を更新する場合
	containerId := *target.ContainerId
	if _, err := u.recordRepo.SelectByPk(ctx, tx, containerId, ui.InternalUserId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 書誌がCONTAINER_IDを持つのに収録物が存在しないか、ユーザのものでない。データ不整合として扱う
			return nil, fmt.Errorf(
				"収録物が存在しません。record id=%s, container id=%s :%w",
				ui.RecordId,
				containerId,
				ErrContainerNotFound,
			)
		}
		return nil, fmt.Errorf("select container record error: %w", err)
	}

	container := extractContainer(
		ui.RecordInput, *ui.ContainerInput, ui.InternalUserId, containerId,
	)
	if err := u.recordRepo.Update(ctx, tx, *container); err != nil {
		return nil, fmt.Errorf("update container record error: %w", err)
	}

	return &containerId, nil
}

// deleteDetachedContainer 書誌から外された収録物を論理削除する
//
// 削除APIと同じく、他に論理削除されていない書誌が残っている場合は削除しない。
func (u *RecordUsecaseImpl) deleteDetachedContainer(
	ctx context.Context,
	tx repository.DBTX,
	containerId string,
) error {

	articleCount, err := u.recordRepo.CountActiveArticles(ctx, tx, containerId)
	if err != nil {
		return fmt.Errorf("count container articles error: %w", err)
	}

	if articleCount > 0 {
		return nil
	}

	if err := u.recordRepo.LogicalDelete(ctx, tx, containerId); err != nil {
		return fmt.Errorf("delete container record error: %w", err)
	}

	return nil
}
