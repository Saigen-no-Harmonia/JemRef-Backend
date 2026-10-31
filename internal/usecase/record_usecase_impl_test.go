package usecase

import (
	"context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// recordTestUid テスト用の内部ユーザID
	recordTestUid = int64(1001)
	// recordArticleId idGenが2回目(コンテナなしの場合は1回目)に生成するID
	recordArticleId = "ARTICLEID00000000000000001"
	// recordContainerId idGenが1回目に生成するID
	recordContainerId = "CONTAINERID000000000000001"
)

// recordStrPtr string型のポインタを返却する
func recordStrPtr(s string) *string {
	return &s
}

// -------------Create(context.Context, dto.CreateRecordInput)のテスト-------------

func TestRecordUsecase_Create(t *testing.T) {

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	datePrecision := record.DatePrecisionDay
	articleReadStatus := record.ReadStatusReading
	containerReadStatus := record.ReadStatusRead

	// 上限件数ちょうどの貢献者情報とその期待値
	limitContributors := make([]dto.ContributorInput, 0, recordInsertLimit)
	expectedLimitRelations := make([]record.ContributorRelation, 0, recordInsertLimit)
	for i := 0; i < recordInsertLimit; i++ {
		name := "貢献者" + strconv.Itoa(i)
		limitContributors = append(limitContributors, dto.ContributorInput{
			Name: name,
			Role: record.ContributorRoleAuthor,
		})
		expectedLimitRelations = append(expectedLimitRelations, record.ContributorRelation{
			Contributor:  record.ContributorEntry{RawName: name, DelFlg: 0},
			Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor, DelFlg: 0},
		})
	}

	tests := []struct {
		name  string
		input dto.CreateRecordInput
		// generatedIds idGenが呼び出し順に返すID
		generatedIds []string
		// createErrs repository.Createが呼び出し回数ごとに返すエラー(nilは成功)
		createErrs []error
		// withTransactionFunc nilの場合は更新処理をそのまま実行する
		withTransactionFunc func(
			ctx context.Context,
			fn func(ctx context.Context, tx repository.DBTX) error,
		) error
		expectedEntries       []record.UserRecordEntry
		expectedOutput        *dto.CreateRecordOutput
		expectedCreateCalled  int
		expectedGenerateCalls int
		expectedTxCalled      int
		expectError           bool
		expectedErrorBody     error
		expectedErrorMsg      string
	}{
		{
			name: "正常_コンテナなし_必須項目のみ",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_コンテナあり",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:            record.RecordTypeJournalArticle,
					MainTitle:       "第1章 「学校」制度の境界線",
					SubTitle:        recordStrPtr("その形成と展開"),
					PageRange:       recordStrPtr("15-46"),
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecision,
					ReadStatus:      &articleReadStatus,
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle:  "境界線の学校史",
					SubTitle:   recordStrPtr("戦後日本の学校化社会の周辺"),
					Volume:     recordStrPtr("日本教育史講座第1巻"),
					Issue:      recordStrPtr("第11号"),
					ReadStatus: &containerReadStatus,
				},
			},
			generatedIds: []string{recordContainerId, recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				// コンテナが先に登録される
				{
					RecordEntryId: recordContainerId,
					InternalUid:   recordTestUid,
					// 記事の書誌タイプの親が設定される
					RecordType: record.RecordTypeJournal,
					MainTitle:  "境界線の学校史",
					SubTitle:   recordStrPtr("戦後日本の学校化社会の周辺"),
					Volume:     recordStrPtr("日本教育史講座第1巻"),
					Issue:      recordStrPtr("第11号"),
					// 出版情報は記事から引き継ぐ
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecision,
					ReadStatus:      &containerReadStatus,
					DelFlg:          0,
				},
				// 記事はコンテナのIDを親として登録される
				{
					RecordEntryId:   recordArticleId,
					InternalUid:     recordTestUid,
					RecordType:      record.RecordTypeJournalArticle,
					MainTitle:       "第1章 「学校」制度の境界線",
					SubTitle:        recordStrPtr("その形成と展開"),
					PageRange:       recordStrPtr("15-46"),
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecision,
					ReadStatus:      &articleReadStatus,
					ParentId:        recordStrPtr(recordContainerId),
					DelFlg:          0,
				},
			},
			// 返却するIDはコンテナではなく記事のID
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  2,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
		},
		{
			name: "正常_貢献者複数",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					Contributors: []dto.ContributorInput{
						{Name: "貢献者A", Role: record.ContributorRoleAuthor},
						{Name: "貢献者B", Role: record.ContributorRoleInterpreter},
						{Name: "貢献者C", Role: record.ContributorRoleCompiler},
					},
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					// 書誌ID・ユーザID・表示順はrepositoryが設定するためここでは空
					ContributorRelations: []record.ContributorRelation{
						{
							Contributor:  record.ContributorEntry{RawName: "貢献者A", DelFlg: 0},
							Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor, DelFlg: 0},
						},
						{
							Contributor:  record.ContributorEntry{RawName: "貢献者B", DelFlg: 0},
							Contribution: record.ContributionEntry{Role: record.ContributorRoleInterpreter, DelFlg: 0},
						},
						{
							Contributor:  record.ContributorEntry{RawName: "貢献者C", DelFlg: 0},
							Contribution: record.ContributionEntry{Role: record.ContributorRoleCompiler, DelFlg: 0},
						},
					},
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_コンテナの貢献者",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文のタイトル",
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle: "雑誌のタイトル",
					Contributors: []dto.ContributorInput{
						{Name: "編者A", Role: record.ContributorRoleCompiler},
						{Name: "編者B", Role: record.ContributorRoleCompilerAuthor},
					},
				},
			},
			generatedIds: []string{recordContainerId, recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordContainerId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournal,
					MainTitle:     "雑誌のタイトル",
					DelFlg:        0,
					ContributorRelations: []record.ContributorRelation{
						{
							Contributor:  record.ContributorEntry{RawName: "編者A", DelFlg: 0},
							Contribution: record.ContributionEntry{Role: record.ContributorRoleCompiler, DelFlg: 0},
						},
						{
							Contributor:  record.ContributorEntry{RawName: "編者B", DelFlg: 0},
							Contribution: record.ContributionEntry{Role: record.ContributorRoleCompilerAuthor, DelFlg: 0},
						},
					},
				},
				// 記事側に貢献者は設定されない
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournalArticle,
					MainTitle:     "論文のタイトル",
					ParentId:      recordStrPtr(recordContainerId),
					DelFlg:        0,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  2,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
		},
		{
			name: "正常_メモとURL複数",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					Memos: []dto.MemoInput{
						{Label: recordStrPtr("要約"), Content: "学校の戦後史。"},
						{Label: nil, Content: "ラベルなしのメモ。"},
					},
					Urls: []dto.UrlInput{
						{Label: recordStrPtr("オンラインPDF"), Value: "https://example.com/1"},
						{Label: nil, Value: "https://example.com/2"},
					},
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					Urls: []record.UrlEntry{
						{RecordEntryId: recordArticleId, DisplayNumber: 0, Label: recordStrPtr("オンラインPDF"), Url: "https://example.com/1", InternalUid: recordTestUid, DelFlg: 0},
						{RecordEntryId: recordArticleId, DisplayNumber: 1, Label: nil, Url: "https://example.com/2", InternalUid: recordTestUid, DelFlg: 0},
					},
					Memos: []record.MemoEntry{
						{RecordEntryId: recordArticleId, DisplayNumber: 0, Title: recordStrPtr("要約"), Content: "学校の戦後史。", InternalUid: recordTestUid, DelFlg: 0},
						{RecordEntryId: recordArticleId, DisplayNumber: 1, Title: nil, Content: "ラベルなしのメモ。", InternalUid: recordTestUid, DelFlg: 0},
					},
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_オプション値_全指定",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					OptionValues: &dto.OptionValuesInput{
						Isbn:          recordStrPtr("9784130513555"),
						Edition:       recordStrPtr("第2版"),
						OriginalTitle: recordStrPtr("School History of Boundary"),
					},
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					OptionValues: []record.OptionValue{
						{RecordEntryId: recordArticleId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "9784130513555", InternalUid: recordTestUid, DelFlg: 0},
						{RecordEntryId: recordArticleId, RecordField: record.RecordFieldEdition, DisplayNumber: 20, RawValue: "第2版", InternalUid: recordTestUid, DelFlg: 0},
						{RecordEntryId: recordArticleId, RecordField: record.RecordFieldOriginalTitle, DisplayNumber: 30, RawValue: "School History of Boundary", InternalUid: recordTestUid, DelFlg: 0},
					},
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_オプション値_一部のみ",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					OptionValues: &dto.OptionValuesInput{
						Isbn: recordStrPtr("9784130513555"),
					},
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					// 未設定(nil)の項目はスキップされる
					OptionValues: []record.OptionValue{
						{RecordEntryId: recordArticleId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "9784130513555", InternalUid: recordTestUid, DelFlg: 0},
					},
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_オプション値_nil",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:         record.RecordTypeAuthoredBook,
					MainTitle:    "単著書のタイトル",
					OptionValues: nil,
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					OptionValues:  nil,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_コンテナのオプション値",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeContributedBookArticle,
					MainTitle: "論文のタイトル",
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle: "編著書のタイトル",
					OptionValues: &dto.OptionValuesInput{
						Isbn: recordStrPtr("9784130513555"),
					},
				},
			},
			generatedIds: []string{recordContainerId, recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordContainerId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeContributedBook,
					MainTitle:     "編著書のタイトル",
					DelFlg:        0,
					// オプション値はコンテナのIDに紐づく
					OptionValues: []record.OptionValue{
						{RecordEntryId: recordContainerId, RecordField: record.RecordFieldIsbn, DisplayNumber: 10, RawValue: "9784130513555", InternalUid: recordTestUid, DelFlg: 0},
					},
				},
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeContributedBookArticle,
					MainTitle:     "論文のタイトル",
					ParentId:      recordStrPtr(recordContainerId),
					DelFlg:        0,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  2,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
		},
		{
			name: "正常_子要素が空スライス",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:         record.RecordTypeAuthoredBook,
					MainTitle:    "単著書のタイトル",
					Contributors: []dto.ContributorInput{},
					Memos:        []dto.MemoInput{},
					Urls:         []dto.UrlInput{},
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
					// 長さ0のスライスはnilとして渡される
					ContributorRelations: nil,
					Memos:                nil,
					Urls:                 nil,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "正常_子要素が上限件数",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:         record.RecordTypeAuthoredBook,
					MainTitle:    "単著書のタイトル",
					Contributors: limitContributors,
				},
			},
			generatedIds: []string{recordArticleId},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId:        recordArticleId,
					InternalUid:          recordTestUid,
					RecordType:           record.RecordTypeAuthoredBook,
					MainTitle:            "単著書のタイトル",
					DelFlg:               0,
					ContributorRelations: expectedLimitRelations,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
		},
		{
			name: "異常_貢献者が上限超過",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:         record.RecordTypeAuthoredBook,
					MainTitle:    "単著書のタイトル",
					Contributors: make([]dto.ContributorInput, recordInsertLimit+1),
				},
			},
			expectedCreateCalled:  0,
			expectedGenerateCalls: 0,
			expectedTxCalled:      0,
			expectError:           true,
			expectedErrorBody:     ErrReachRegistrationLimit,
			expectedErrorMsg:      "貢献者情報の件数が登録上限を超えています。",
		},
		{
			name: "異常_コンテナの貢献者が上限超過",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文のタイトル",
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle:    "雑誌のタイトル",
					Contributors: make([]dto.ContributorInput, recordInsertLimit+1),
				},
			},
			expectedCreateCalled:  0,
			expectedGenerateCalls: 0,
			expectedTxCalled:      0,
			expectError:           true,
			expectedErrorBody:     ErrReachRegistrationLimit,
			expectedErrorMsg:      "収録物の貢献者情報の件数が登録上限を超えています。",
		},
		{
			name: "異常_メモが上限超過",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					Memos:     make([]dto.MemoInput, recordInsertLimit+1),
				},
			},
			expectedCreateCalled:  0,
			expectedGenerateCalls: 0,
			expectedTxCalled:      0,
			expectError:           true,
			expectedErrorBody:     ErrReachRegistrationLimit,
			expectedErrorMsg:      "メモ情報の件数が登録上限を超えています。",
		},
		{
			name: "異常_URLが上限超過",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
					Urls:      make([]dto.UrlInput, recordInsertLimit+1),
				},
			},
			expectedCreateCalled:  0,
			expectedGenerateCalls: 0,
			expectedTxCalled:      0,
			expectError:           true,
			expectedErrorBody:     ErrReachRegistrationLimit,
			expectedErrorMsg:      "URL情報の件数が登録上限を超えています。",
		},
		{
			name: "異常_コンテナ登録失敗",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文のタイトル",
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle: "雑誌のタイトル",
				},
			},
			generatedIds: []string{recordContainerId, recordArticleId},
			createErrs:   []error{repository.ErrTestUnexpected},
			// コンテナの登録のみが試行され、記事の登録は行われない
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordContainerId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournal,
					MainTitle:     "雑誌のタイトル",
					DelFlg:        0,
				},
			},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
			expectError:           true,
			expectedErrorBody:     repository.ErrTestUnexpected,
			expectedErrorMsg:      "create container record error",
		},
		{
			name: "異常_記事登録失敗",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
				},
			},
			generatedIds: []string{recordArticleId},
			createErrs:   []error{repository.ErrTestUnexpected},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeAuthoredBook,
					MainTitle:     "単著書のタイトル",
					DelFlg:        0,
				},
			},
			expectedCreateCalled:  1,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
			expectError:           true,
			expectedErrorBody:     repository.ErrTestUnexpected,
			expectedErrorMsg:      "create article record error",
		},
		{
			name: "異常_コンテナ成功後に記事登録失敗",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文のタイトル",
				},
				ContainerInput: &dto.ContainerInput{
					MainTitle: "雑誌のタイトル",
				},
			},
			generatedIds: []string{recordContainerId, recordArticleId},
			createErrs:   []error{nil, repository.ErrTestUnexpected},
			expectedEntries: []record.UserRecordEntry{
				{
					RecordEntryId: recordContainerId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournal,
					MainTitle:     "雑誌のタイトル",
					DelFlg:        0,
				},
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournalArticle,
					MainTitle:     "論文のタイトル",
					ParentId:      recordStrPtr(recordContainerId),
					DelFlg:        0,
				},
			},
			expectedCreateCalled:  2,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
			expectError:           true,
			expectedErrorBody:     repository.ErrTestUnexpected,
			expectedErrorMsg:      "create article record error",
		},
		{
			name: "異常_トランザクション開始失敗",
			input: dto.CreateRecordInput{
				InternalUserId: recordTestUid,
				RecordInput: dto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
				},
			},
			generatedIds: []string{recordArticleId},
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				// 更新処理を実行せずにエラーを返す
				return repository.ErrTestUnexpected
			},
			// ID生成はトランザクション開始より前に行われる
			expectedCreateCalled:  0,
			expectedGenerateCalls: 1,
			expectedTxCalled:      1,
			expectError:           true,
			expectedErrorBody:     repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
			repo.CreateFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				ur record.UserRecordEntry,
			) error {
				// Createの呼び出し回数に応じたエラーを返す
				idx := repo.CreateCalled - 1
				if idx < len(tt.createErrs) {
					return tt.createErrs[idx]
				}
				return nil
			}

			txManager := mock.NewMockTxManager()
			if tt.withTransactionFunc != nil {
				txManager.WithTransactionFunc = tt.withTransactionFunc
			}

			uc := NewRecordUsecaseImpl(repo, txManager, idGen)

			output, err := uc.Create(context.Background(), tt.input)

			// 呼び出しチェック
			assert.Equal(t, tt.expectedCreateCalled, repo.CreateCalled)
			assert.Equal(t, tt.expectedGenerateCalls, idGen.GenerateCalls)
			assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

			// repositoryに渡されたユーザ入力書誌の検証
			assert.Equal(t, tt.expectedEntries, repo.CreatedEntries)

			// 異常系の場合のassertion
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorBody)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
				assert.Nil(t, output)
				return
			}

			// 正常系の場合のassertion
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}
