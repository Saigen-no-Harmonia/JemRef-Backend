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
