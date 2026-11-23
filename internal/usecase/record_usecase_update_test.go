package usecase

import (
	"context"
	"strconv"
	"testing"
	"time"

	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------Update(context.Context, dto.UpdateRecordInput)のテスト-------------

// recordUpdateCase Updateのテストケース
//
// 正常系・異常系は同じフィクスチャと実行部(runRecordUpdateCase)を共有する。
type recordUpdateCase struct {
	name  string
	input dto.UpdateRecordInput
	// generatedIds idGenが呼び出し順に返すID
	generatedIds []string
	// selectResults / selectErrs 書誌IDをキーにSelectByPkが返す値
	selectResults map[string]*record.UserRecordEntry
	selectErrs    map[string]error
	// countResults / countErrs CountActiveArticlesが呼び出し順に返す値
	countResults []int
	countErrs    []error
	// createErrs / updateErrs / deleteErrs 各メソッドが呼び出し順に返すエラー(nilは成功)
	createErrs []error
	updateErrs []error
	deleteErrs []error
	// withTransactionFunc nilの場合は更新処理をそのまま実行する
	withTransactionFunc func(
		ctx context.Context,
		fn func(ctx context.Context, tx repository.DBTX) error,
	) error

	// 各メソッドに渡された書誌IDの呼び出し順(件数と順序を同時に固定する)
	expectedSelectedIds []string
	expectedCountedIds  []string
	expectedCreatedIds  []string
	expectedUpdatedIds  []string
	expectedDeletedIds  []string
	expectedGenCalls    int
	expectedTxCalled    int

	// assertEntries repositoryへ渡された書誌の内容を検証する(nilの場合は検証しない)
	assertEntries func(t *testing.T, repo *mock.MockRecordRepository)

	expectedErrorBody    error
	notExpectedErrorBody error
	expectedErrorMsgs    []string
}

// recordUpdateCases Updateの正常系・異常系のテストケースを返す
//
// フィクスチャを両方で共有するため、ケースの定義はここにまとめる。
func recordUpdateCases() (normal, abnormal []recordUpdateCase) {
	// SelectByPkが返す書誌。更新では書誌タイプとコンテナIDを判定に使う
	soloTarget := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		RecordType:    record.RecordTypeAuthoredBook,
	}
	detachedTarget := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		RecordType:    record.RecordTypeJournalArticle,
	}
	articleTarget := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		RecordType:    record.RecordTypeJournalArticle,
		ContainerId:   testutil.Ptr(recordContainerId),
	}
	containerTarget := &record.UserRecordEntry{
		RecordEntryId: recordContainerId,
		RecordType:    record.RecordTypeJournal,
	}

	// 単著書(収録物を持たない書誌)の更新入力
	soloInput := dto.UpdateRecordInput{
		InternalUserId: recordTestUid,
		RecordId:       recordArticleId,
		RecordInput: dto.RecordInput{
			Type:      record.RecordTypeAuthoredBook,
			MainTitle: "更新後の単著書",
		},
	}

	// 収録物の入力。収録物なし/ありの分岐はこれのnilで切り替える
	containerInput := &dto.ContainerInput{
		MainTitle: "更新後の雑誌",
		Volume:    testutil.Ptr("第1巻"),
	}

	// newArticleInput 雑誌記事の更新入力を返す(収録物の有無を引数で切り替える)
	newArticleInput := func(ci *dto.ContainerInput) dto.UpdateRecordInput {
		return dto.UpdateRecordInput{
			InternalUserId: recordTestUid,
			RecordId:       recordArticleId,
			RecordInput: dto.RecordInput{
				Type:      record.RecordTypeJournalArticle,
				MainTitle: "更新後の雑誌記事",
			},
			ContainerInput: ci,
		}
	}

	// 全項目を設定した更新入力(変換結果の検証用)
	publicationDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionMonth
	articleStatus := record.ReadStatusRead
	containerStatus := record.ReadStatusUnread

	fullInput := dto.UpdateRecordInput{
		InternalUserId: recordTestUid,
		RecordId:       recordArticleId,
		RecordInput: dto.RecordInput{
			Type:            record.RecordTypeJournalArticle,
			MainTitle:       "更新後の主題",
			SubTitle:        testutil.Ptr("更新後の副題"),
			PageRange:       testutil.Ptr("15-46"),
			Publisher:       testutil.Ptr("更新後の出版者"),
			PublicationDate: &publicationDate,
			DatePrecision:   &datePrecision,
			ReadStatus:      &articleStatus,
			Contributors: []dto.ContributorInput{
				{Name: "著者A", Role: record.ContributorRoleAuthor},
				{Name: "訳者B", Role: record.ContributorRoleInterpreter},
			},
			Memos: []dto.MemoInput{
				{Label: testutil.Ptr("要約"), Content: "メモ1"},
				{Content: "メモ2"},
			},
			Urls: []dto.UrlInput{
				{Label: testutil.Ptr("PDF"), Value: "https://example.com/1"},
				{Value: "https://example.com/2"},
			},
			OptionValues: &dto.OptionValuesInput{
				Isbn:    testutil.Ptr("9784130513555"),
				Edition: testutil.Ptr("第2版"),
			},
		},
		ContainerInput: &dto.ContainerInput{
			MainTitle:  "更新後の雑誌",
			SubTitle:   testutil.Ptr("雑誌の副題"),
			Volume:     testutil.Ptr("第1巻"),
			Issue:      testutil.Ptr("第11号"),
			ReadStatus: &containerStatus,
			Contributors: []dto.ContributorInput{
				{Name: "編者C", Role: record.ContributorRoleCompiler},
			},
			OptionValues: &dto.OptionValuesInput{
				Isbn: testutil.Ptr("9784130519999"),
			},
		},
	}

	// 登録上限ちょうどの構成要素
	limitContributorInputs := make([]dto.ContributorInput, 0, recordInsertLimit)
	for i := 0; i < recordInsertLimit; i++ {
		limitContributorInputs = append(
			limitContributorInputs,
			dto.ContributorInput{Name: "貢献者" + strconv.Itoa(i), Role: record.ContributorRoleAuthor},
		)
	}
	limitInput := soloInput
	limitInput.RecordInput.Contributors = limitContributorInputs
	limitInput.RecordInput.Memos = make([]dto.MemoInput, recordInsertLimit)
	limitInput.RecordInput.Urls = make([]dto.UrlInput, recordInsertLimit)

	normal = []recordUpdateCase{
		{
			name:                "正常_収録物なしの書誌を更新",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordArticleId},
			// 収録物がないためID生成も収録物の登録も行われない
			expectedGenCalls: 0,
			expectedTxCalled: 1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 1)
				assert.Nil(t, repo.UpdatedEntries[0].ContainerId)
			},
		},
		{
			name:  "正常_収録物を新規に追加",
			input: newArticleInput(containerInput),
			// 更新前はコンテナIDを持たない
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: detachedTarget},
			generatedIds:        []string{recordContainerId},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			// fk_er_containerのため、収録物のCreateが書誌のUpdateより先に行われる
			expectedCreatedIds: []string{recordContainerId},
			expectedUpdatedIds: []string{recordArticleId},
			expectedGenCalls:   1,
			expectedTxCalled:   1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.CreatedEntries, 1)
				// 収録物の書誌タイプは書誌タイプから変換される
				assert.Equal(t, record.RecordTypeJournal, repo.CreatedEntries[0].RecordType)
				assert.Equal(t, "更新後の雑誌", repo.CreatedEntries[0].MainTitle)

				require.Len(t, repo.UpdatedEntries, 1)
				assert.Equal(t, testutil.Ptr(recordContainerId), repo.UpdatedEntries[0].ContainerId)
			},
		},
		{
			name:  "正常_既存の収録物を更新",
			input: newArticleInput(containerInput),
			selectResults: map[string]*record.UserRecordEntry{
				recordArticleId:   articleTarget,
				recordContainerId: containerTarget,
			},
			countResults: []int{0},
			// 書誌 -> 収録物の順に存在確認する
			expectedSelectedIds: []string{recordArticleId, recordContainerId},
			expectedCountedIds:  []string{recordArticleId},
			// 収録物 -> 書誌の順に更新される
			expectedUpdatedIds: []string{recordContainerId, recordArticleId},
			// 既存のIDを引き回すためID生成は行われない
			expectedGenCalls: 0,
			expectedTxCalled: 1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 2)
				assert.Equal(t, testutil.Ptr(recordContainerId), repo.UpdatedEntries[1].ContainerId)
			},
		},
		{
			name:  "正常_収録物を外し他に書誌がなければ収録物も論理削除",
			input: newArticleInput(nil),
			selectResults: map[string]*record.UserRecordEntry{
				recordArticleId: articleTarget,
			},
			// 1回目=対象自身が収録物かの判定, 2回目=外した収録物に残る書誌件数
			countResults:        []int{0, 0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId, recordContainerId},
			expectedUpdatedIds:  []string{recordArticleId},
			// 書誌のCONTAINER_IDを外した後に収録物を論理削除する
			expectedDeletedIds: []string{recordContainerId},
			expectedTxCalled:   1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 1)
				assert.Nil(t, repo.UpdatedEntries[0].ContainerId)
			},
		},
		{
			name:  "正常_収録物を外すが他に書誌が残れば収録物を残す",
			input: newArticleInput(nil),
			selectResults: map[string]*record.UserRecordEntry{
				recordArticleId: articleTarget,
			},
			countResults:        []int{0, 1},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId, recordContainerId},
			expectedUpdatedIds:  []string{recordArticleId},
			// 収録物は論理削除されない
			expectedDeletedIds: nil,
			expectedTxCalled:   1,
		},
		{
			// 全フィールドの変換はCreateのテストで網羅済み。
			// ここでは更新固有の経路(既存コンテナIDの引き回しと入力RecordIdの反映)と、
			// 構成要素に書誌ID・ユーザID・表示順が設定されることを確認する。
			name:  "正常_入力が変換されてrepositoryへ渡る",
			input: fullInput,
			selectResults: map[string]*record.UserRecordEntry{
				recordArticleId:   articleTarget,
				recordContainerId: containerTarget,
			},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId, recordContainerId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordContainerId, recordArticleId},
			expectedTxCalled:    1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 2)
				container := repo.UpdatedEntries[0]
				article := repo.UpdatedEntries[1]

				// 収録物: 既存IDを引き回し、書誌タイプはコンテナ側へ変換される
				assert.Equal(t, recordContainerId, container.RecordEntryId)
				assert.Equal(t, recordTestUid, container.InternalUid)
				assert.Equal(t, record.RecordTypeJournal, container.RecordType)
				assert.Nil(t, container.ContainerId)
				require.Len(t, container.ContributorRelations, 1)
				assert.Equal(t, "編者C", container.ContributorRelations[0].Contributor.RawName)
				require.Len(t, container.OptionValues, 1)
				assert.Equal(t, recordContainerId, container.OptionValues[0].RecordEntryId)
				assert.Equal(t, recordTestUid, container.OptionValues[0].InternalUid)
				// 収録物はURL・メモを持たない
				assert.Nil(t, container.Urls)
				assert.Nil(t, container.Memos)

				// 書誌: IDは入力のRecordId、コンテナIDは既存の収録物
				assert.Equal(t, recordArticleId, article.RecordEntryId)
				assert.Equal(t, recordTestUid, article.InternalUid)
				assert.Equal(t, record.RecordTypeJournalArticle, article.RecordType)
				assert.Equal(t, testutil.Ptr(recordContainerId), article.ContainerId)

				require.Len(t, article.ContributorRelations, 2)
				assert.Equal(t, "著者A", article.ContributorRelations[0].Contributor.RawName)
				assert.Equal(
					t,
					record.ContributorRoleInterpreter,
					article.ContributorRelations[1].Contribution.Role,
				)

				// 構成要素は表示順が0起点で振られ、書誌IDとユーザIDが設定される
				require.Len(t, article.Urls, 2)
				for i, u := range article.Urls {
					assert.Equal(t, i, u.DisplayNumber)
					assert.Equal(t, recordArticleId, u.RecordEntryId)
					assert.Equal(t, recordTestUid, u.InternalUid)
				}

				require.Len(t, article.Memos, 2)
				for i, m := range article.Memos {
					assert.Equal(t, i, m.DisplayNumber)
					assert.Equal(t, recordArticleId, m.RecordEntryId)
					assert.Equal(t, recordTestUid, m.InternalUid)
				}

				require.Len(t, article.OptionValues, 2)
				for _, o := range article.OptionValues {
					assert.Equal(t, recordArticleId, o.RecordEntryId)
					assert.Equal(t, recordTestUid, o.InternalUid)
				}
			},
		},
		{
			// 構成要素が空の場合はnilのまま渡り、repositoryの洗い替えで全削除される
			name:                "正常_構成要素が空でも更新できる",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordArticleId},
			expectedTxCalled:    1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 1)
				article := repo.UpdatedEntries[0]
				assert.Nil(t, article.ContributorRelations)
				assert.Nil(t, article.Urls)
				assert.Nil(t, article.Memos)
				assert.Nil(t, article.OptionValues)
			},
		},
		{
			name:                "正常_構成要素が上限ちょうどなら更新できる",
			input:               limitInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordArticleId},
			expectedTxCalled:    1,
			assertEntries: func(t *testing.T, repo *mock.MockRecordRepository) {
				require.Len(t, repo.UpdatedEntries, 1)
				assert.Len(t, repo.UpdatedEntries[0].ContributorRelations, recordInsertLimit)
				assert.Len(t, repo.UpdatedEntries[0].Urls, recordInsertLimit)
				assert.Len(t, repo.UpdatedEntries[0].Memos, recordInsertLimit)
			},
		},
	}

	abnormal = []recordUpdateCase{
		{
			name:                "異常_書誌が存在しない",
			input:               soloInput,
			selectErrs:          map[string]error{recordArticleId: repository.ErrNotFound},
			expectedSelectedIds: []string{recordArticleId},
			// 存在確認で中断するため以降は呼ばれない
			expectedCountedIds: nil,
			expectedUpdatedIds: nil,
			expectedTxCalled:   1,
			expectedErrorBody:  ErrRecordNotFound,
			expectedErrorMsgs:  []string{"更新対象の書誌が存在しません", recordArticleId},
		},
		{
			name:                "異常_書誌の取得が想定外のエラー",
			input:               soloInput,
			selectErrs:          map[string]error{recordArticleId: ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
			// 想定外のエラーを404に化けさせない
			notExpectedErrorBody: ErrRecordNotFound,
			expectedErrorMsgs:    []string{"select record error"},
		},
		{
			// 収録物のIDはレスポンスに露出しないため、指定自体が不正リクエスト。
			// ログには件数を残し、存在を漏らさないようレスポンスは404に寄せる。
			name:                "異常_収録物の単体更新は拒否される",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{2},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrRecordNotFound,
			expectedErrorMsgs:   []string{"収録物の単体更新が要求されました", "article count=2"},
		},
		{
			name:                "異常_対象書誌の件数取得に失敗",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countErrs:           []error{ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
			expectedErrorMsgs:   []string{"count articles of target error"},
		},
		{
			// 書誌タイプの変更はコンテナ関係の変更を伴うため許容しない
			name:                "異常_書誌タイプの変更は拒否される",
			input:               newArticleInput(nil),
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrRecordTypeNotChangeable,
			expectedErrorMsgs: []string{
				"書誌タイプの変更が要求されました",
				"before=" + string(record.RecordTypeAuthoredBook),
				"after=" + string(record.RecordTypeJournalArticle),
			},
		},
		{
			// 書誌がCONTAINER_IDを持つのに収録物が存在しない。ユーザが対処できないため500
			name:          "異常_収録物が存在しない_データ不整合",
			input:         newArticleInput(containerInput),
			selectResults: map[string]*record.UserRecordEntry{recordArticleId: articleTarget},
			selectErrs:    map[string]error{recordContainerId: repository.ErrNotFound},
			countResults:  []int{0},
			// 書誌 -> 収録物の順に取得し、収録物で中断する
			expectedSelectedIds: []string{recordArticleId, recordContainerId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrContainerNotFound,
			// データ不整合を404で隠さない
			notExpectedErrorBody: ErrRecordNotFound,
			expectedErrorMsgs:    []string{"収録物が存在しません", recordContainerId},
		},
		{
			name:                "異常_収録物の取得が想定外のエラー",
			input:               newArticleInput(containerInput),
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: articleTarget},
			selectErrs:          map[string]error{recordContainerId: ErrTestUnexpected},
			countResults:        []int{0},
			expectedSelectedIds: []string{recordArticleId, recordContainerId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
			expectedErrorMsgs:   []string{"select container record error"},
		},
		{
			name:                "異常_収録物のCreateに失敗",
			input:               newArticleInput(containerInput),
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: detachedTarget},
			generatedIds:        []string{recordContainerId},
			countResults:        []int{0},
			createErrs:          []error{ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedCreatedIds:  []string{recordContainerId},
			// 収録物の登録に失敗した時点で中断し、書誌は更新しない
			expectedUpdatedIds: nil,
			expectedGenCalls:   1,
			expectedTxCalled:   1,
			expectedErrorBody:  ErrTestUnexpected,
			expectedErrorMsgs:  []string{"create container record error"},
		},
		{
			name:  "異常_収録物のUpdateに失敗",
			input: newArticleInput(containerInput),
			selectResults: map[string]*record.UserRecordEntry{
				recordArticleId:   articleTarget,
				recordContainerId: containerTarget,
			},
			countResults:        []int{0},
			updateErrs:          []error{ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId, recordContainerId},
			expectedCountedIds:  []string{recordArticleId},
			// 1回目(収録物)で失敗するため書誌は更新されない
			expectedUpdatedIds: []string{recordContainerId},
			expectedTxCalled:   1,
			expectedErrorBody:  ErrTestUnexpected,
			expectedErrorMsgs:  []string{"update container record error"},
		},
		{
			// SelectByPkが成功しているため通常は到達しない。
			// 到達した場合はデータ不整合であり、404ではなく500に寄せる。
			name:                "異常_書誌のUpdateで対象が失われた場合はデータ不整合",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			updateErrs:          []error{repository.ErrNotFound},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordArticleId},
			expectedDeletedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrRecordUpdateFailed,
			// 存在確認済みのため「存在しない」とは説明しない
			notExpectedErrorBody: ErrRecordNotFound,
			expectedErrorMsgs:    []string{"更新対象の書誌が取得後に失われました", recordArticleId},
		},
		{
			name:                "異常_書誌のUpdateに失敗",
			input:               soloInput,
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: soloTarget},
			countResults:        []int{0},
			updateErrs:          []error{ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId},
			expectedUpdatedIds:  []string{recordArticleId},
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
			expectedErrorMsgs:   []string{"update article record error"},
		},
		{
			name:                "異常_外した収録物の件数取得に失敗",
			input:               newArticleInput(nil),
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: articleTarget},
			countResults:        []int{0},
			countErrs:           []error{nil, ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId, recordContainerId},
			expectedUpdatedIds:  []string{recordArticleId},
			// 件数が取れないため収録物は論理削除しない
			expectedDeletedIds: nil,
			expectedTxCalled:   1,
			expectedErrorBody:  ErrTestUnexpected,
			expectedErrorMsgs:  []string{"count container articles error"},
		},
		{
			name:                "異常_外した収録物の論理削除に失敗",
			input:               newArticleInput(nil),
			selectResults:       map[string]*record.UserRecordEntry{recordArticleId: articleTarget},
			countResults:        []int{0, 0},
			deleteErrs:          []error{ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedCountedIds:  []string{recordArticleId, recordContainerId},
			expectedUpdatedIds:  []string{recordArticleId},
			expectedDeletedIds:  []string{recordContainerId},
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
			expectedErrorMsgs:   []string{"delete container record error"},
		},
		{
			// chkInsertLimitはCreateと共用で、4種(貢献者/収録物の貢献者/メモ/URL)の網羅は
			// TestRecordUsecase_Create側で担保する。ここではUpdateが呼んでいることと、
			// トランザクション開始前に弾くことを確認する。
			name: "異常_貢献者が登録上限を超過",
			input: func() dto.UpdateRecordInput {
				i := soloInput
				i.RecordInput.Contributors = make([]dto.ContributorInput, recordInsertLimit+1)
				return i
			}(),
			// トランザクションを開始せずに弾く
			expectedTxCalled:    0,
			expectedSelectedIds: nil,
			expectedErrorBody:   ErrReachRegistrationLimit,
			expectedErrorMsgs:   []string{"貢献者情報の件数が登録上限を超えています"},
		},
		{
			name: "異常_収録物の貢献者が登録上限を超過",
			input: func() dto.UpdateRecordInput {
				i := newArticleInput(&dto.ContainerInput{
					MainTitle:    "更新後の雑誌",
					Contributors: make([]dto.ContributorInput, recordInsertLimit+1),
				})
				return i
			}(),
			expectedTxCalled:    0,
			expectedSelectedIds: nil,
			expectedErrorBody:   ErrReachRegistrationLimit,
			expectedErrorMsgs:   []string{"収録物の貢献者情報の件数が登録上限を超えています"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: soloInput,
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				// 更新処理を実行せずにエラーを返す
				return ErrTestUnexpected
			},
			expectedSelectedIds: nil,
			expectedUpdatedIds:  nil,
			expectedTxCalled:    1,
			expectedErrorBody:   ErrTestUnexpected,
		},
	}

	return normal, abnormal
}

func TestRecordUsecase_Update_正常(t *testing.T) {
	normal, _ := recordUpdateCases()
	for _, tt := range normal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordUpdateCase(t, tt)
		})
	}
}

func TestRecordUsecase_Update_異常(t *testing.T) {
	_, abnormal := recordUpdateCases()
	for _, tt := range abnormal {
		t.Run(tt.name, func(t *testing.T) {
			runRecordUpdateCase(t, tt)
		})
	}
}

// runRecordUpdateCase Updateを実行し、repositoryの呼び出しと結果を検証する
func runRecordUpdateCase(t *testing.T, tt recordUpdateCase) {
	t.Helper()

	generateCalls := 0
	idGen := &mock.MockIdGen{
		GenerateFunc: func() string {
			require.Less(t, generateCalls, len(tt.generatedIds), "想定外のID生成が行われました")
			id := tt.generatedIds[generateCalls]
			generateCalls++
			return id
		},
	}

	repo := &mock.MockRecordRepository{}

	repo.SelectByPkFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error) {
		if err, ok := tt.selectErrs[recordId]; ok && err != nil {
			return nil, err
		}
		result, ok := tt.selectResults[recordId]
		require.True(t, ok, "想定外のSelectByPk呼び出し record id=%s", recordId)
		return result, nil
	}

	repo.CountActiveArticlesFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		containerRecordId string,
	) (int, error) {
		// 呼び出し回数に応じた戻り値を返す
		idx := repo.CountActiveArticlesCalled - 1
		if idx < len(tt.countErrs) && tt.countErrs[idx] != nil {
			return 0, tt.countErrs[idx]
		}
		require.Less(t, idx, len(tt.countResults), "想定外のCountActiveArticles呼び出し")
		return tt.countResults[idx], nil
	}

	repo.CreateFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		ur record.UserRecordEntry,
	) error {
		idx := repo.CreateCalled - 1
		if idx < len(tt.createErrs) {
			return tt.createErrs[idx]
		}
		return nil
	}

	repo.UpdateFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		ur record.UserRecordEntry,
	) error {
		idx := repo.UpdateCalled - 1
		if idx < len(tt.updateErrs) {
			return tt.updateErrs[idx]
		}
		return nil
	}

	repo.LogicalDeleteFunc = func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
	) error {
		idx := repo.LogicalDeleteCalled - 1
		if idx < len(tt.deleteErrs) {
			return tt.deleteErrs[idx]
		}
		return nil
	}

	txManager := mock.NewMockTxManager()
	if tt.withTransactionFunc != nil {
		txManager.WithTransactionFunc = tt.withTransactionFunc
	}

	uc := NewRecordUsecaseImpl(repo, txManager, idGen)

	err := uc.Update(context.Background(), tt.input)

	// 各メソッドに渡された書誌IDと呼び出し順
	assert.Equal(t, tt.expectedSelectedIds, repo.SelectedRecordIds)
	assert.Equal(t, tt.expectedCountedIds, repo.CountedRecordIds)
	assert.Equal(t, tt.expectedCreatedIds, entryIds(repo.CreatedEntries))
	assert.Equal(t, tt.expectedUpdatedIds, entryIds(repo.UpdatedEntries))
	assert.Equal(t, tt.expectedDeletedIds, repo.DeletedRecordIds)
	assert.Equal(t, tt.expectedGenCalls, idGen.GenerateCalls)
	assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

	// 更新では既存の構成要素を読まない
	assert.Equal(t, 0, repo.SelectComponentsCalled)

	// 入力がそのままrepositoryへ渡ること
	if len(tt.expectedSelectedIds) > 0 {
		assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPkInternalUid)
	}

	if tt.assertEntries != nil {
		tt.assertEntries(t, repo)
	}

	// 異常系の場合のassertion
	if tt.expectedErrorBody != nil {
		require.Error(t, err)
		assert.ErrorIs(t, err, tt.expectedErrorBody)
		if tt.notExpectedErrorBody != nil {
			assert.NotErrorIs(t, err, tt.notExpectedErrorBody)
		}
		for _, msg := range tt.expectedErrorMsgs {
			assert.Contains(t, err.Error(), msg)
		}
		return
	}

	// 正常系の場合のassertion
	require.NoError(t, err)
}
