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
					// 書誌に対応するコンテナの書誌タイプが設定される
					RecordType: record.RecordTypeJournal,
					MainTitle:  "境界線の学校史",
					SubTitle:   recordStrPtr("戦後日本の学校化社会の周辺"),
					Volume:     recordStrPtr("日本教育史講座第1巻"),
					Issue:      recordStrPtr("第11号"),
					// 出版情報は書誌から引き継ぐ
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecision,
					ReadStatus:      &containerReadStatus,
					DelFlg:          0,
				},
				// 書誌はコンテナのIDを保持して登録される
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
					ContainerId:     recordStrPtr(recordContainerId),
					DelFlg:          0,
				},
			},
			// 返却するIDはコンテナではなく書誌のID
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
				// 書誌側に貢献者は設定されない
				{
					RecordEntryId: recordArticleId,
					InternalUid:   recordTestUid,
					RecordType:    record.RecordTypeJournalArticle,
					MainTitle:     "論文のタイトル",
					ContainerId:   recordStrPtr(recordContainerId),
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
					ContainerId:   recordStrPtr(recordContainerId),
					DelFlg:        0,
				},
			},
			expectedOutput:        &dto.CreateRecordOutput{RecordId: recordArticleId},
			expectedCreateCalled:  2,
			expectedGenerateCalls: 2,
			expectedTxCalled:      1,
		},
		{
			name: "正常_構成要素が空スライス",
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
			name: "正常_構成要素が上限件数",
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
			// コンテナの登録のみが試行され、書誌の登録は行われない
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
			name: "異常_書誌登録失敗",
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
			name: "異常_コンテナ成功後に書誌登録失敗",
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
					ContainerId:   recordStrPtr(recordContainerId),
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

// -------------Delete(context.Context, dto.DeleteRecordInput)のテスト-------------

func TestRecordUsecase_Delete(t *testing.T) {

	// SelectByPkが返す書誌(RecordEntryIdとContainerIdのみ設定される軽量SELECT)
	soloTarget := &record.UserRecordEntry{RecordEntryId: recordArticleId}
	articleTarget := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		ContainerId:   recordStrPtr(recordContainerId),
	}

	input := dto.DeleteRecordInput{
		InternalUserId: recordTestUid,
		RecordId:       recordArticleId,
	}

	tests := []struct {
		name  string
		input dto.DeleteRecordInput
		// selectResult / selectErr はSelectByPkの戻り値
		selectResult *record.UserRecordEntry
		selectErr    error
		// countResults / countErrs はCountActiveArticlesが呼び出し順に返す値
		countResults []int
		countErrs    []error
		// deleteErrs はLogicalDeleteが呼び出し順に返すエラー(nilは成功)
		deleteErrs []error
		// withTransactionFunc nilの場合は更新処理をそのまま実行する
		withTransactionFunc func(
			ctx context.Context,
			fn func(ctx context.Context, tx repository.DBTX) error,
		) error
		// expectedCountedIds CountActiveArticlesに渡された書誌IDの呼び出し順
		expectedCountedIds []string
		// expectedDeletedIds LogicalDeleteに渡された書誌IDの呼び出し順
		expectedDeletedIds   []string
		expectedSelectCalled int
		expectedCountCalled  int
		expectedDeleteCalled int
		expectedTxCalled     int
		expectError          bool
		expectedErrorBody    error
		expectedErrorMsgs    []string
	}{
		{
			name:                 "正常_コンテナなしの書誌を削除",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			// コンテナがないため書誌件数の確認は1回(自身が収録物かの判定)のみ
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
		{
			name:         "正常_コンテナありで他に書誌がない場合は収録物も削除",
			input:        input,
			selectResult: articleTarget,
			// 1回目=自身の書誌件数, 2回目=収録物に残る書誌件数
			countResults: []int{0, 0},
			// 書誌 -> 収録物の順に削除される
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
		},
		{
			name:         "正常_コンテナありで他に書誌が残っている場合は収録物を残す",
			input:        input,
			selectResult: articleTarget,
			countResults: []int{0, 1},
			// 収録物は削除されない
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
		{
			name:                 "正常_入力がそのままrepositoryへ渡る",
			input:                input,
			selectResult:         articleTarget,
			countResults:         []int{0, 0},
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
		},
		{
			// 収録物ガードが書誌を持たない書誌を誤って弾かないことの確認
			name:                 "正常_書誌を持たない書誌は削除できる",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
		},
		{
			name:      "異常_書誌が存在しない",
			input:     input,
			selectErr: repository.ErrNotFound,
			// 存在確認で中断するため以降は呼ばれない
			expectedSelectCalled: 1,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    ErrRecordNotFound,
			expectedErrorMsgs:    []string{"削除対象の書誌が存在しません"},
		},
		{
			name:         "異常_収録物を指定した場合は削除されない",
			input:        input,
			selectResult: soloTarget,
			// 自身が3件の書誌を持つ=収録物
			countResults:         []int{3},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    ErrRecordNotFound,
			// ログに残す必要がある情報がメッセージに含まれること
			expectedErrorMsgs: []string{
				"収録物の単体削除が要求されました",
				"record id=" + recordArticleId,
				"internal uid=1001",
				"article count=3",
			},
		},
		{
			name:                 "異常_SelectByPkが予期しないエラー",
			input:                input,
			selectErr:            repository.ErrTestUnexpected,
			expectedSelectCalled: 1,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"select record error"},
		},
		{
			name:                 "異常_対象の書誌件数取得に失敗",
			input:                input,
			selectResult:         soloTarget,
			countErrs:            []error{repository.ErrTestUnexpected},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"count articles of target error"},
		},
		{
			name:                 "異常_書誌の論理削除に失敗",
			input:                input,
			selectResult:         soloTarget,
			countResults:         []int{0},
			deleteErrs:           []error{repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId},
			expectedSelectCalled: 1,
			expectedCountCalled:  1,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"delete article record error"},
		},
		{
			name:         "異常_収録物の書誌件数取得に失敗",
			input:        input,
			selectResult: articleTarget,
			// 1回目は成功、2回目(収録物の書誌件数)で失敗
			countResults:         []int{0, 0},
			countErrs:            []error{nil, repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 1,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"count container articles error"},
		},
		{
			name:         "異常_収録物の論理削除に失敗",
			input:        input,
			selectResult: articleTarget,
			countResults: []int{0, 0},
			// 2回目(収録物)の削除で失敗
			deleteErrs:           []error{nil, repository.ErrTestUnexpected},
			expectedDeletedIds:   []string{recordArticleId, recordContainerId},
			expectedCountedIds:   []string{recordArticleId, recordContainerId},
			expectedSelectCalled: 1,
			expectedCountCalled:  2,
			expectedDeleteCalled: 2,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
			expectedErrorMsgs:    []string{"delete container record error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: input,
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				// 更新処理を実行せずにエラーを返す
				return repository.ErrTestUnexpected
			},
			expectedSelectCalled: 0,
			expectedCountCalled:  0,
			expectedDeleteCalled: 0,
			expectedTxCalled:     1,
			expectError:          true,
			expectedErrorBody:    repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := &mock.MockRecordRepository{}

			repo.SelectByPkFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				recordId string,
				internalUid int64,
			) (*record.UserRecordEntry, error) {
				if tt.selectErr != nil {
					return nil, tt.selectErr
				}
				return tt.selectResult, nil
			}

			repo.CountActiveArticlesFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				containerRecordId string,
			) (int, error) {
				// CountActiveArticlesの呼び出し回数に応じた戻り値を返す
				idx := repo.CountActiveArticlesCalled - 1
				if idx < len(tt.countErrs) && tt.countErrs[idx] != nil {
					return 0, tt.countErrs[idx]
				}
				require.Less(t, idx, len(tt.countResults), "想定外のCountActiveArticles呼び出し")
				return tt.countResults[idx], nil
			}

			repo.LogicalDeleteFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				recordId string,
			) error {
				// LogicalDeleteの呼び出し回数に応じたエラーを返す
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

			uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

			err := uc.Delete(context.Background(), tt.input)

			// 呼び出しチェック
			assert.Equal(t, tt.expectedSelectCalled, repo.SelectByPkCalled)
			assert.Equal(t, tt.expectedCountCalled, repo.CountActiveArticlesCalled)
			assert.Equal(t, tt.expectedDeleteCalled, repo.LogicalDeleteCalled)
			assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

			// 論理削除に渡された書誌IDと順序の検証
			assert.Equal(t, tt.expectedDeletedIds, repo.DeletedRecordIds)

			// 入力がそのままrepositoryへ渡ること
			if tt.expectedSelectCalled > 0 {
				assert.Equal(t, tt.input.RecordId, repo.LastSelectByPkRecordId)
				assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPkInternalUid)
			}
			// 書誌件数の確認は1回目が対象自身(収録物かの判定)、2回目が収録物に対して行われる
			assert.Equal(t, tt.expectedCountedIds, repo.CountedRecordIds)

			// 異常系の場合のassertion
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorBody)
				for _, msg := range tt.expectedErrorMsgs {
					assert.Contains(t, err.Error(), msg)
				}
				return
			}

			// 正常系の場合のassertion
			require.NoError(t, err)
		})
	}
}

// -------------Get(context.Context, dto.GetRecordInput)のテスト-------------

func TestRecordUsecase_Get(t *testing.T) {

	// SelectByPkが返す書誌(構成要素はSelectComponentsが設定するためここでは持たない)
	soloArticle := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeAuthoredBook,
		MainTitle:     "単著書のタイトル",
	}
	containedArticle := &record.UserRecordEntry{
		RecordEntryId: recordArticleId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeJournalArticle,
		MainTitle:     "論文のタイトル",
		ContainerId:   recordStrPtr(recordContainerId),
	}
	container := &record.UserRecordEntry{
		RecordEntryId: recordContainerId,
		InternalUid:   recordTestUid,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "雑誌のタイトル",
	}

	// setComponents 渡されたentryに構成要素を設定する(repositoryの振る舞いの模倣)
	// 書誌と収録物を区別できるよう、書誌IDを貢献者名に含める。
	setComponents := func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		for _, e := range entries {
			if e == nil {
				continue
			}
			e.ContributorRelations = []record.ContributorRelation{
				{
					Contributor:  record.ContributorEntry{RawName: "貢献者_" + e.RecordEntryId},
					Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor},
				},
			}
			e.Urls = []record.UrlEntry{{Url: "https://example.com/" + e.RecordEntryId}}
			e.Memos = []record.MemoEntry{{Content: "メモ_" + e.RecordEntryId}}
			e.OptionValues = []record.OptionValue{
				{RecordField: record.RecordFieldIsbn, RawValue: "9784130513555"},
			}
		}
		return nil
	}

	tests := []struct {
		name  string
		input dto.GetRecordInput
		// selectResults / selectErrs はSelectByPkが呼び出し順に返す値(最大2回)
		selectResults []*record.UserRecordEntry
		selectErrs    []error
		// countResult / countErr はCountActiveArticlesの戻り値(必ず1回のみ呼ばれる)
		countResult int
		countErr    error
		// selectComponentsFunc nilの場合は構成要素を設定せず成功させる
		selectComponentsFunc func(
			ctx context.Context,
			tx repository.DBTX,
			entries []*record.UserRecordEntry,
		) error
		// withTransactionFunc nilの場合は処理をそのまま実行する
		withTransactionFunc func(
			ctx context.Context,
			fn func(ctx context.Context, tx repository.DBTX) error,
		) error
		// expectedSelectedIds SelectByPkに渡された書誌IDの呼び出し順
		expectedSelectedIds []string
		// expectedCountedIds CountActiveArticlesに渡された書誌IDの呼び出し順
		expectedCountedIds []string
		// expectedComponentIds SelectComponentsに渡された書誌IDの組(nilは空文字で記録される)
		expectedComponentIds     [][]string
		expectedSelectCalled     int
		expectedCountCalled      int
		expectedComponentsCalled int
		expectedTxCalled         int
		// expectedContainerId 収録物の書誌ID。空文字の場合はContainerがnilであることを検証する
		expectedContainerId string
		// expectComponentsSet 構成要素が設定されていることを検証するか
		expectComponentsSet bool
		expectError         bool
		expectedErrorBody   error
		expectedErrorMsgs   []string
	}{
		{
			name:  "正常_収録物なしの書誌を取得",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      "",
		},
		{
			name:  "正常_収録物ありの書誌を取得",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{containedArticle, container},
			countResult:   0,

			// 1回目が書誌、2回目が収録物の順で引かれる
			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, recordContainerId}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      recordContainerId,
		},
		{
			name:  "正常_書誌と収録物に構成要素が設定される",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults:        []*record.UserRecordEntry{containedArticle, container},
			countResult:          0,
			selectComponentsFunc: setComponents,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, recordContainerId}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      recordContainerId,
			expectComponentsSet:      true,
		},
		{
			name:  "正常_収録物が論理削除済み",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			// 2回目(収録物)のSelectByPkがErrNotFoundを返す
			selectResults: []*record.UserRecordEntry{containedArticle, nil},
			selectErrs:    []error{nil, repository.ErrNotFound},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			// 収録物を持たない書誌として返す(エラーにしない)
			expectedContainerId: "",
		},
		{
			name:  "正常_入力がそのままrepositoryへ渡る",
			input: dto.GetRecordInput{InternalUserId: 2002, RecordId: "OTHERRECORDID0000000000001"},

			selectResults: []*record.UserRecordEntry{
				{RecordEntryId: "OTHERRECORDID0000000000001", InternalUid: 2002},
			},
			countResult: 0,

			expectedSelectedIds:      []string{"OTHERRECORDID0000000000001"},
			expectedCountedIds:       []string{"OTHERRECORDID0000000000001"},
			expectedComponentIds:     [][]string{{"OTHERRECORDID0000000000001", ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectedContainerId:      "",
		},
		{
			name:  "異常_書誌が存在しない",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{nil},
			selectErrs:    []error{repository.ErrNotFound},

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       nil,
			expectedComponentIds:     nil,
			expectedSelectCalled:     1,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrRecordNotFound,
			expectedErrorMsgs:        []string{"参照対象の書誌が存在しません", recordArticleId},
		},
		{
			name:  "異常_収録物を指定した場合は参照できない",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordContainerId},

			selectResults: []*record.UserRecordEntry{container},
			// 生きている書誌が1件いる = 収録物である
			countResult: 1,

			expectedSelectedIds:      []string{recordContainerId},
			expectedCountedIds:       []string{recordContainerId},
			expectedComponentIds:     nil,
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrRecordNotFound,
			// リクエストログに残すため、書誌ID・ユーザID・書誌件数をメッセージに含める
			expectedErrorMsgs: []string{
				"収録物の単体参照が要求されました",
				recordContainerId,
				strconv.FormatInt(recordTestUid, 10),
				"article count=1",
			},
		},
		{
			name:  "異常_SelectByPkが予期しないエラー",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{nil},
			selectErrs:    []error{ErrTestUnexpected},

			expectedSelectedIds:      []string{recordArticleId},
			expectedSelectCalled:     1,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select record error"},
		},
		{
			name:  "異常_書誌件数の取得に失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countErr:      ErrTestUnexpected,

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"count articles of target error"},
		},
		{
			name:  "異常_収録物のSelectByPkが予期しないエラー",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			// ErrNotFound以外は収録物なしとして扱わず、エラーとして返す
			selectResults: []*record.UserRecordEntry{containedArticle, nil},
			selectErrs:    []error{nil, ErrTestUnexpected},
			countResult:   0,

			expectedSelectedIds:      []string{recordArticleId, recordContainerId},
			expectedCountedIds:       []string{recordArticleId},
			expectedSelectCalled:     2,
			expectedCountCalled:      1,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select container record error"},
		},
		{
			name:  "異常_SelectComponentsが失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			selectResults: []*record.UserRecordEntry{soloArticle},
			countResult:   0,
			selectComponentsFunc: func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				return ErrTestUnexpected
			},

			expectedSelectedIds:      []string{recordArticleId},
			expectedCountedIds:       []string{recordArticleId},
			expectedComponentIds:     [][]string{{recordArticleId, ""}},
			expectedSelectCalled:     1,
			expectedCountCalled:      1,
			expectedComponentsCalled: 1,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrTestUnexpected,
			expectedErrorMsgs:        []string{"select components error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: dto.GetRecordInput{InternalUserId: recordTestUid, RecordId: recordArticleId},

			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				return ErrTestUnexpected
			},

			// クロージャが実行されないため、repositoryは一度も呼ばれない
			expectedSelectCalled:     0,
			expectedCountCalled:      0,
			expectedComponentsCalled: 0,
			expectedTxCalled:         1,
			expectError:              true,
			expectedErrorBody:        ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockRecordRepository{}

			repo.SelectByPkFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				recordId string,
				internalUid int64,
			) (*record.UserRecordEntry, error) {
				// SelectByPkの呼び出し回数に応じた戻り値を返す
				idx := repo.SelectByPkCalled - 1
				if idx < len(tt.selectErrs) && tt.selectErrs[idx] != nil {
					return nil, tt.selectErrs[idx]
				}
				require.Less(t, idx, len(tt.selectResults), "想定外のSelectByPk呼び出し")

				// 書誌はテストケース間で共有するため、複製を返して副作用を防ぐ
				src := tt.selectResults[idx]
				if src == nil {
					return nil, nil
				}
				copied := *src
				return &copied, nil
			}

			repo.CountActiveArticlesFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				containerRecordId string,
			) (int, error) {
				if tt.countErr != nil {
					return 0, tt.countErr
				}
				return tt.countResult, nil
			}

			repo.SelectComponentsFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				if tt.selectComponentsFunc != nil {
					return tt.selectComponentsFunc(ctx, tx, entries)
				}
				return nil
			}

			txManager := mock.NewMockTxManager()
			if tt.withTransactionFunc != nil {
				txManager.WithTransactionFunc = tt.withTransactionFunc
			}

			uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

			actual, err := uc.Get(context.Background(), tt.input)

			// 呼び出しチェック
			assert.Equal(t, tt.expectedSelectCalled, repo.SelectByPkCalled)
			assert.Equal(t, tt.expectedCountCalled, repo.CountActiveArticlesCalled)
			assert.Equal(t, tt.expectedComponentsCalled, repo.SelectComponentsCalled)
			assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

			// repositoryへ渡された書誌IDと順序の検証
			assert.Equal(t, tt.expectedSelectedIds, repo.SelectedRecordIds)
			assert.Equal(t, tt.expectedCountedIds, repo.CountedRecordIds)

			// SelectComponentsには常に[対象書誌, 収録物]の2要素を渡す(収録物なしはnil)
			assert.Equal(t, tt.expectedComponentIds, repo.ComponentTargetIds)

			// 入力がそのままrepositoryへ渡ること
			if tt.expectedSelectCalled > 0 {
				assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPkInternalUid)
			}

			// 異常系の場合のassertion
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorBody)
				for _, msg := range tt.expectedErrorMsgs {
					assert.Contains(t, err.Error(), msg)
				}
				assert.Nil(t, actual)
				return
			}

			// 正常系の場合のassertion
			require.NoError(t, err)
			require.NotNil(t, actual)

			assert.Equal(t, tt.input.RecordId, actual.Record.RecordEntryId)

			if tt.expectedContainerId == "" {
				assert.Nil(t, actual.Container)
			} else {
				require.NotNil(t, actual.Container)
				assert.Equal(t, tt.expectedContainerId, actual.Container.RecordEntryId)
			}

			if !tt.expectComponentsSet {
				return
			}

			// SelectComponentsが設定した構成要素がOutputに反映されていること
			// (Outputへのコピーが SelectComponents の後に行われることの検証)
			assertComponentsSet(t, actual.Record)
			require.NotNil(t, actual.Container)
			assertComponentsSet(t, *actual.Container)
		})
	}
}

// assertComponentsSet 構成要素が設定され、書誌IDに対応した値になっていることを検証する
func assertComponentsSet(t *testing.T, e record.UserRecordEntry) {
	t.Helper()

	require.Len(t, e.ContributorRelations, 1)
	assert.Equal(t, "貢献者_"+e.RecordEntryId, e.ContributorRelations[0].Contributor.RawName)

	require.Len(t, e.Urls, 1)
	assert.Equal(t, "https://example.com/"+e.RecordEntryId, e.Urls[0].Url)

	require.Len(t, e.Memos, 1)
	assert.Equal(t, "メモ_"+e.RecordEntryId, e.Memos[0].Content)

	require.Len(t, e.OptionValues, 1)
	assert.Equal(t, record.RecordFieldIsbn, e.OptionValues[0].RecordField)
}

// -------------Update(context.Context, dto.UpdateRecordInput)のテスト-------------

// entryIds Create/Updateに渡された書誌のIDを呼び出し順に返す
func entryIds(entries []record.UserRecordEntry) []string {
	if entries == nil {
		return nil
	}

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.RecordEntryId)
	}

	return ids
}

func TestRecordUsecase_Update(t *testing.T) {

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
		ContainerId:   recordStrPtr(recordContainerId),
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
		Volume:    recordStrPtr("第1巻"),
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
			SubTitle:        recordStrPtr("更新後の副題"),
			PageRange:       recordStrPtr("15-46"),
			Publisher:       recordStrPtr("更新後の出版者"),
			PublicationDate: &publicationDate,
			DatePrecision:   &datePrecision,
			ReadStatus:      &articleStatus,
			Contributors: []dto.ContributorInput{
				{Name: "著者A", Role: record.ContributorRoleAuthor},
				{Name: "訳者B", Role: record.ContributorRoleInterpreter},
			},
			Memos: []dto.MemoInput{
				{Label: recordStrPtr("要約"), Content: "メモ1"},
				{Content: "メモ2"},
			},
			Urls: []dto.UrlInput{
				{Label: recordStrPtr("PDF"), Value: "https://example.com/1"},
				{Value: "https://example.com/2"},
			},
			OptionValues: &dto.OptionValuesInput{
				Isbn:    recordStrPtr("9784130513555"),
				Edition: recordStrPtr("第2版"),
			},
		},
		ContainerInput: &dto.ContainerInput{
			MainTitle:  "更新後の雑誌",
			SubTitle:   recordStrPtr("雑誌の副題"),
			Volume:     recordStrPtr("第1巻"),
			Issue:      recordStrPtr("第11号"),
			ReadStatus: &containerStatus,
			Contributors: []dto.ContributorInput{
				{Name: "編者C", Role: record.ContributorRoleCompiler},
			},
			OptionValues: &dto.OptionValuesInput{
				Isbn: recordStrPtr("9784130519999"),
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

	tests := []struct {
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

		expectError          bool
		expectedErrorBody    error
		notExpectedErrorBody error
		expectedErrorMsgs    []string
	}{
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
				assert.Equal(t, recordStrPtr(recordContainerId), repo.UpdatedEntries[0].ContainerId)
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
				assert.Equal(t, recordStrPtr(recordContainerId), repo.UpdatedEntries[1].ContainerId)
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
				assert.Equal(t, recordStrPtr(recordContainerId), article.ContainerId)

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
		{
			name:                "異常_書誌が存在しない",
			input:               soloInput,
			selectErrs:          map[string]error{recordArticleId: repository.ErrNotFound},
			expectedSelectedIds: []string{recordArticleId},
			// 存在確認で中断するため以降は呼ばれない
			expectedCountedIds: nil,
			expectedUpdatedIds: nil,
			expectedTxCalled:   1,
			expectError:        true,
			expectedErrorBody:  ErrRecordNotFound,
			expectedErrorMsgs:  []string{"更新対象の書誌が存在しません", recordArticleId},
		},
		{
			name:                "異常_書誌の取得が想定外のエラー",
			input:               soloInput,
			selectErrs:          map[string]error{recordArticleId: ErrTestUnexpected},
			expectedSelectedIds: []string{recordArticleId},
			expectedTxCalled:    1,
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:        true,
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
			expectError:        true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:        true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
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
			expectError:         true,
			expectedErrorBody:   ErrTestUnexpected,
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
			if tt.expectError {
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
		})
	}
}

// -------------List(context.Context, dto.GetRecordsInput)のテスト-------------

const (
	// listSoloId1/2 収録物を持たない書誌のID
	listSoloId1 = "LISTSOLO000000000000000001"
	listSoloId2 = "LISTSOLO000000000000000002"
	// listArticleAId/BId 収録物1を持つ書誌のID(同じ収録物を共有する)
	listArticleAId = "LISTARTICLEA00000000000001"
	listArticleBId = "LISTARTICLEB00000000000001"
	// listArticleCId 収録物2を持つ書誌のID
	listArticleCId = "LISTARTICLEC00000000000001"
	// listContainerId1/2 収録物のID
	listContainerId1 = "LISTCONTAINER0000000000001"
	listContainerId2 = "LISTCONTAINER0000000000002"
)

// copyEntries 書誌の複製を返す
// mockの戻り値はテストケース間で共有するため、usecaseによる書き換え(貢献者の設定)が漏れないようにする。
func copyEntries(src []*record.UserRecordEntry) []*record.UserRecordEntry {
	if src == nil {
		return nil
	}

	copied := make([]*record.UserRecordEntry, 0, len(src))
	for _, e := range src {
		c := *e
		copied = append(copied, &c)
	}

	return copied
}

func TestRecordUsecase_List(t *testing.T) {

	newSolo := func(recordId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeAuthoredBook,
			MainTitle:     "単著書のタイトル",
		}
	}
	newContained := func(recordId string, containerId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeJournalArticle,
			MainTitle:     "論文のタイトル",
			ContainerId:   recordStrPtr(containerId),
		}
	}
	newContainer := func(recordId string) *record.UserRecordEntry {
		return &record.UserRecordEntry{
			RecordEntryId: recordId,
			InternalUid:   recordTestUid,
			RecordType:    record.RecordTypeJournal,
			MainTitle:     "雑誌のタイトル",
		}
	}

	soloArticle1 := newSolo(listSoloId1)
	soloArticle2 := newSolo(listSoloId2)
	containedArticleA := newContained(listArticleAId, listContainerId1)
	containedArticleB := newContained(listArticleBId, listContainerId1)
	containedArticleC := newContained(listArticleCId, listContainerId2)
	container1 := newContainer(listContainerId1)
	container2 := newContainer(listContainerId2)

	// baseInput デフォルト値相当の入力(handlerで補完済みの想定)
	baseInput := dto.GetRecordsInput{
		InternalUserId: recordTestUid,
		SortKey:        record.SortKeyPublishYear,
		SortOrder:      record.SortOrderAsc,
		Limit:          20,
		Offset:         0,
	}
	withOffset := func(offset int) dto.GetRecordsInput {
		in := baseInput
		in.Offset = offset
		return in
	}

	// fullInput 全項目を指定した入力(詰め替えの検証用)
	publishYear := 2022
	fullInput := dto.GetRecordsInput{
		InternalUserId:  2002,
		ContributorName: recordStrPtr("木村"),
		Title:           recordStrPtr("学校"),
		PublishYear:     &publishYear,
		SortKey:         record.SortKeyInsertDate,
		SortOrder:       record.SortOrderDesc,
		Limit:           50,
		Offset:          20,
	}

	// setContributors 渡されたentryに貢献者を1人追加する(repositoryの振る舞いの模倣)
	// どのentryに付いたかを判別できるよう、書誌IDを貢献者名に含める。
	// appendで追加するため、同じentryが2回渡されると貢献者が2人になり検出できる。
	setContributors := func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error {
		for _, e := range entries {
			if e == nil {
				continue
			}
			e.ContributorRelations = append(e.ContributorRelations, record.ContributorRelation{
				Contributor:  record.ContributorEntry{RawName: "貢献者_" + e.RecordEntryId},
				Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor},
			})
		}
		return nil
	}

	tests := []struct {
		name  string
		input dto.GetRecordsInput

		// countResult / countErr CountByConditionの戻り値
		countResult int
		countErr    error
		// selectResults / selectErr SelectByConditionの戻り値
		selectResults []*record.UserRecordEntry
		selectErr     error
		// pkResults / pkErr SelectByPksの戻り値
		pkResults []*record.UserRecordEntry
		pkErr     error
		// contributorsFunc nilの場合は貢献者を設定せず成功させる
		contributorsFunc func(
			ctx context.Context,
			tx repository.DBTX,
			entries []*record.UserRecordEntry,
		) error
		// withTransactionFunc nilの場合は処理をそのまま実行する
		withTransactionFunc func(
			ctx context.Context,
			fn func(ctx context.Context, tx repository.DBTX) error,
		) error

		expectedCountCalled        int
		expectedSelectCalled       int
		expectedPkCalled           int
		expectedContributorsCalled int
		expectedTxCalled           int
		// expectedCondition repositoryへ渡る検索条件(nilの場合は検証しない)
		expectedCondition *repository.RecordListCondition
		// expectedPkIds SelectByPksに渡された書誌IDの組
		expectedPkIds [][]string
		// expectedContributorIds SelectContributorRelationsに渡された書誌IDの組
		expectedContributorIds [][]string

		expectedTotal int
		// expectedRecordIds 出力の書誌IDの順。nilの場合はRecordsが空であることを検証する
		expectedRecordIds []string
		// expectedContainerIds 出力の各書誌に付いた収録物のID(収録物なしは "")
		expectedContainerIds []string
		// expectContributorsSet 書誌と収録物に、自身のIDの貢献者が1人ずつ付いていることを検証するか
		expectContributorsSet bool

		expectError       bool
		expectedErrorBody error
		expectedErrorMsgs []string
	}{
		{
			name:          "正常_収録物なしの書誌のみ",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{soloArticle1, soloArticle2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              nil,
			expectedContributorIds:     [][]string{{listSoloId1, listSoloId2}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listSoloId1, listSoloId2},
			expectedContainerIds:       []string{"", ""},
		},
		{
			name:          "正常_収録物ありの書誌",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			// 書誌の後に収録物を渡す
			expectedContributorIds: [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:          1,
			expectedRecordIds:      []string{listArticleAId},
			expectedContainerIds:   []string{listContainerId1},
		},
		{
			name:          "正常_収録物あり・なしの混在",
			input:         baseInput,
			countResult:   3,
			selectResults: []*record.UserRecordEntry{soloArticle1, containedArticleA, containedArticleC},
			pkResults:     []*record.UserRecordEntry{container1, container2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			// 収録物IDは書誌の出現順に集める
			expectedPkIds: [][]string{{listContainerId1, listContainerId2}},
			expectedContributorIds: [][]string{{
				listSoloId1, listArticleAId, listArticleCId, listContainerId1, listContainerId2,
			}},
			expectedTotal:        3,
			expectedRecordIds:    []string{listSoloId1, listArticleAId, listArticleCId},
			expectedContainerIds: []string{"", listContainerId1, listContainerId2},
		},
		{
			name:          "正常_同じ収録物を持つ書誌はIDを1回だけ渡す",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{containedArticleA, containedArticleB},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			// 収録物は1回だけ渡す(同じIDの別entryを渡すと貢献者が片方にしか付かないため)
			expectedContributorIds: [][]string{{listArticleAId, listArticleBId, listContainerId1}},
			expectedTotal:          2,
			expectedRecordIds:      []string{listArticleAId, listArticleBId},
			expectedContainerIds:   []string{listContainerId1, listContainerId1},
		},
		{
			name:          "正常_収録物が見つからなければ収録物なし",
			input:         baseInput,
			countResult:   2,
			selectResults: []*record.UserRecordEntry{containedArticleA, containedArticleC},
			// 収録物1は論理削除済みで返らない想定
			pkResults: []*record.UserRecordEntry{container2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1, listContainerId2}},
			expectedContributorIds:     [][]string{{listArticleAId, listArticleCId, listContainerId2}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listArticleAId, listArticleCId},
			// エラーにせず、収録物を持たない書誌として返す(詳細参照と同じ扱い)
			expectedContainerIds: []string{"", listContainerId2},
		},
		{
			name:             "正常_書誌と収録物に貢献者が設定される",
			input:            baseInput,
			countResult:      1,
			selectResults:    []*record.UserRecordEntry{containedArticleA},
			pkResults:        []*record.UserRecordEntry{container1},
			contributorsFunc: setContributors,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:              1,
			expectedRecordIds:          []string{listArticleAId},
			expectedContainerIds:       []string{listContainerId1},
			// 出力へのコピーが貢献者の取得より後に行われていること
			expectContributorsSet: true,
		},
		{
			name:             "正常_同じ収録物を共有する書誌に同じ貢献者が付く",
			input:            baseInput,
			countResult:      2,
			selectResults:    []*record.UserRecordEntry{containedArticleA, containedArticleB},
			pkResults:        []*record.UserRecordEntry{container1},
			contributorsFunc: setContributors,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listArticleBId, listContainerId1}},
			expectedTotal:              2,
			expectedRecordIds:          []string{listArticleAId, listArticleBId},
			expectedContainerIds:       []string{listContainerId1, listContainerId1},
			// 両方の書誌の収録物に、貢献者がちょうど1人ずつ付いていること
			expectContributorsSet: true,
		},
		{
			name:        "正常_総件数が0なら一覧を取得しない",
			input:       baseInput,
			countResult: 0,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedTotal:              0,
			expectedRecordIds:          nil,
		},
		{
			name:        "正常_offsetが総件数以上なら一覧を取得しない",
			input:       withOffset(40),
			countResult: 25,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			// 総件数はそのまま返す
			expectedTotal:     25,
			expectedRecordIds: nil,
		},
		{
			name:        "正常_offsetが総件数と同じなら一覧を取得しない",
			input:       withOffset(20),
			countResult: 20,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedTotal:              20,
			expectedRecordIds:          nil,
		},
		{
			name:          "正常_offsetが総件数未満なら一覧を取得する",
			input:         withOffset(20),
			countResult:   21,
			selectResults: []*record.UserRecordEntry{soloArticle1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedContributorIds:     [][]string{{listSoloId1}},
			expectedTotal:              21,
			expectedRecordIds:          []string{listSoloId1},
			expectedContainerIds:       []string{""},
		},
		{
			name:          "正常_総件数はlimitで切り詰めない",
			input:         baseInput,
			countResult:   25,
			selectResults: []*record.UserRecordEntry{soloArticle1, soloArticle2},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedContributorIds:     [][]string{{listSoloId1, listSoloId2}},
			expectedTotal:              25,
			expectedRecordIds:          []string{listSoloId1, listSoloId2},
			expectedContainerIds:       []string{"", ""},
		},
		{
			name:          "正常_入力が変換されてrepositoryへ渡る",
			input:         fullInput,
			countResult:   21,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedCondition: &repository.RecordListCondition{
				InternalUid:     2002,
				ContributorName: recordStrPtr("木村"),
				Title:           recordStrPtr("学校"),
				PublishYear:     &publishYear,
				SortKey:         record.SortKeyInsertDate,
				SortOrder:       record.SortOrderDesc,
				Limit:           50,
				Offset:          20,
			},
			expectedPkIds:          [][]string{{listContainerId1}},
			expectedContributorIds: [][]string{{listArticleAId, listContainerId1}},
			expectedTotal:          21,
			expectedRecordIds:      []string{listArticleAId},
			expectedContainerIds:   []string{listContainerId1},
		},
		{
			name:     "異常_件数取得に失敗",
			input:    baseInput,
			countErr: ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectError:                true,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"count records error"},
		},
		{
			name:        "異常_一覧取得に失敗",
			input:       baseInput,
			countResult: 1,
			selectErr:   ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectError:                true,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select records error"},
		},
		{
			name:          "異常_収録物の取得に失敗",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkErr:         ErrTestUnexpected,

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectError:                true,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select container records error"},
		},
		{
			name:          "異常_貢献者の取得に失敗",
			input:         baseInput,
			countResult:   1,
			selectResults: []*record.UserRecordEntry{containedArticleA},
			pkResults:     []*record.UserRecordEntry{container1},
			contributorsFunc: func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				return ErrTestUnexpected
			},

			expectedCountCalled:        1,
			expectedSelectCalled:       1,
			expectedPkCalled:           1,
			expectedContributorsCalled: 1,
			expectedTxCalled:           1,
			expectedPkIds:              [][]string{{listContainerId1}},
			expectedContributorIds:     [][]string{{listArticleAId, listContainerId1}},
			expectError:                true,
			expectedErrorBody:          ErrTestUnexpected,
			expectedErrorMsgs:          []string{"select contributor relations error"},
		},
		{
			name:  "異常_トランザクション開始失敗",
			input: baseInput,
			withTransactionFunc: func(
				ctx context.Context,
				fn func(ctx context.Context, tx repository.DBTX) error,
			) error {
				return ErrTestUnexpected
			},

			// クロージャが実行されないため、repositoryは一度も呼ばれない
			expectedCountCalled:        0,
			expectedSelectCalled:       0,
			expectedPkCalled:           0,
			expectedContributorsCalled: 0,
			expectedTxCalled:           1,
			expectError:                true,
			expectedErrorBody:          ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockRecordRepository{}

			repo.CountByConditionFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				cond repository.RecordListCondition,
			) (int, error) {
				if tt.countErr != nil {
					return 0, tt.countErr
				}
				return tt.countResult, nil
			}

			repo.SelectByConditionFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				cond repository.RecordListCondition,
			) ([]*record.UserRecordEntry, error) {
				if tt.selectErr != nil {
					return nil, tt.selectErr
				}
				return copyEntries(tt.selectResults), nil
			}

			repo.SelectByPksFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				recordIds []string,
				internalUid int64,
			) ([]*record.UserRecordEntry, error) {
				if tt.pkErr != nil {
					return nil, tt.pkErr
				}
				return copyEntries(tt.pkResults), nil
			}

			repo.SelectContributorRelationsFunc = func(
				ctx context.Context,
				tx repository.DBTX,
				entries []*record.UserRecordEntry,
			) error {
				if tt.contributorsFunc != nil {
					return tt.contributorsFunc(ctx, tx, entries)
				}
				return nil
			}

			txManager := mock.NewMockTxManager()
			if tt.withTransactionFunc != nil {
				txManager.WithTransactionFunc = tt.withTransactionFunc
			}

			uc := NewRecordUsecaseImpl(repo, txManager, &mock.MockIdGen{})

			actual, err := uc.List(context.Background(), tt.input)

			// 呼び出しチェック
			assert.Equal(t, tt.expectedCountCalled, repo.CountByConditionCalled)
			assert.Equal(t, tt.expectedSelectCalled, repo.SelectByConditionCalled)
			assert.Equal(t, tt.expectedPkCalled, repo.SelectByPksCalled)
			assert.Equal(t, tt.expectedContributorsCalled, repo.SelectContributorRelationsCalled)
			assert.Equal(t, tt.expectedTxCalled, txManager.WithTransactionCalled)

			// repositoryへ渡された値の検証
			assert.Equal(t, tt.expectedPkIds, repo.SelectedPkIds)
			assert.Equal(t, tt.expectedContributorIds, repo.ContributorTargetIds)

			if tt.expectedPkCalled > 0 {
				assert.Equal(t, tt.input.InternalUserId, repo.LastSelectByPksInternalUid)
			}

			if tt.expectedCondition != nil {
				assert.Equal(t, []repository.RecordListCondition{*tt.expectedCondition}, repo.CountedConditions)
				assert.Equal(t, []repository.RecordListCondition{*tt.expectedCondition}, repo.SelectedConditions)
			}

			// 件数と一覧は同じ条件で引く
			if tt.expectedSelectCalled > 0 {
				assert.Equal(t, repo.CountedConditions, repo.SelectedConditions)
			}

			// 異常系の場合のassertion
			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorBody)
				for _, msg := range tt.expectedErrorMsgs {
					assert.Contains(t, err.Error(), msg)
				}
				assert.Nil(t, actual)
				return
			}

			// 正常系の場合のassertion
			require.NoError(t, err)
			require.NotNil(t, actual)

			assert.Equal(t, tt.expectedTotal, actual.TotalCount)

			if tt.expectedRecordIds == nil {
				assert.Empty(t, actual.Records)
				return
			}

			recordIds := make([]string, 0, len(actual.Records))
			containerIds := make([]string, 0, len(actual.Records))
			for _, item := range actual.Records {
				recordIds = append(recordIds, item.Record.RecordEntryId)
				if item.Container == nil {
					containerIds = append(containerIds, "")
					continue
				}
				containerIds = append(containerIds, item.Container.RecordEntryId)
			}
			assert.Equal(t, tt.expectedRecordIds, recordIds)
			assert.Equal(t, tt.expectedContainerIds, containerIds)

			if !tt.expectContributorsSet {
				return
			}

			// 書誌と収録物に、自身のIDを含む貢献者がちょうど1人ずつ付いていること
			for _, item := range actual.Records {
				assertListContributorSet(t, item.Record)
				if item.Container != nil {
					assertListContributorSet(t, *item.Container)
				}
			}
		})
	}
}

// assertListContributorSet 書誌IDに対応した貢献者がちょうど1人付いていることを検証する
func assertListContributorSet(t *testing.T, e record.UserRecordEntry) {
	t.Helper()

	require.Len(t, e.ContributorRelations, 1, "record id=%s", e.RecordEntryId)
	assert.Equal(t, "貢献者_"+e.RecordEntryId, e.ContributorRelations[0].Contributor.RawName)
}
