package handler

import (
	"testing"
	"time"

	"jemref_go/internal/domain/record"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
)

// -------------GetRecords(*gin.Context)のテスト(レスポンスの組み立て)-------------

// TestRecordHandler_GetRecords_レスポンス usecaseの出力からレスポンスを組み立てること
func TestRecordHandler_GetRecords_レスポンス検証(t *testing.T) {

	dayDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	monthDate := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	yearDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	// newOutput 指定した総件数と書誌からusecaseの出力を組み立てる
	newOutput := func(total int, items ...usecaseDto.RecordListItem) *usecaseDto.GetRecordsOutput {
		return &usecaseDto.GetRecordsOutput{TotalCount: total, Records: items}
	}

	// minimalItem 必須項目のみを持つ、収録物なしの書誌を返す
	minimalItem := func(recordId string) usecaseDto.RecordListItem {
		return usecaseDto.RecordListItem{
			Record: record.UserRecordEntry{
				RecordEntryId: recordId,
				InternalUid:   recordTestInternalUid,
				RecordType:    record.RecordTypeAuthoredBook,
				MainTitle:     "単著書のタイトル",
			},
		}
	}

	// minimalResponseItem minimalItemに対応するレスポンスを返す
	minimalResponseItem := func(recordId string) handlerDto.RecordListItemResponse {
		return handlerDto.RecordListItemResponse{
			RecordId:  recordId,
			Type:      "authored_book",
			MainTitle: "単著書のタイトル",
		}
	}

	// withPublicationDate 出版日と精度を設定した書誌を返す
	withPublicationDate := func(date *time.Time, precision *record.DatePrecision) usecaseDto.RecordListItem {
		item := minimalItem(recordListArticleId1)
		item.Record.PublicationDate = date
		item.Record.DatePrecision = precision
		return item
	}

	// withPublicationDateResponse 出版日を設定したレスポンスを返す
	withPublicationDateResponse := func(publicationDate *string) handlerDto.GetRecordsResponse {
		item := minimalResponseItem(recordListArticleId1)
		item.PublicationDate = publicationDate
		return newRecordListResponse(1, item)
	}

	// 全項目と貢献者を持つ、収録物ありの書誌
	fullItem := usecaseDto.RecordListItem{
		Record: record.UserRecordEntry{
			RecordEntryId:   recordListArticleId1,
			InternalUid:     recordTestInternalUid,
			RecordType:      record.RecordTypeJournalArticle,
			MainTitle:       "第1章 「学校」制度の境界線",
			SubTitle:        testutil.Ptr("その形成と展開"),
			PageRange:       testutil.Ptr("15-46"),
			Publisher:       testutil.Ptr("東京大学出版会"),
			PublicationDate: &dayDate,
			DatePrecision:   testutil.Ptr(record.DatePrecisionDay),
			ReadStatus:      testutil.Ptr(record.ReadStatusReading),
			ContainerId:     testutil.Ptr(recordListContainerId),
			ContributorRelations: []record.ContributorRelation{
				{
					Contributor:  record.ContributorEntry{RawName: "木村元"},
					Contribution: record.ContributionEntry{Role: record.ContributorRoleAuthor},
				},
				{
					Contributor:  record.ContributorEntry{RawName: "訳者B"},
					Contribution: record.ContributionEntry{Role: record.ContributorRoleInterpreter},
				},
			},
		},
		Container: &record.UserRecordEntry{
			RecordEntryId: recordListContainerId,
			InternalUid:   recordTestInternalUid,
			RecordType:    record.RecordTypeJournal,
			MainTitle:     "境界線の学校史",
			SubTitle:      testutil.Ptr("収録物の副題"),
			Volume:        testutil.Ptr("第1巻"),
			Issue:         testutil.Ptr("第11号"),
			ContributorRelations: []record.ContributorRelation{
				{
					Contributor:  record.ContributorEntry{RawName: "編者A"},
					Contribution: record.ContributionEntry{Role: record.ContributorRoleCompiler},
				},
			},
		},
	}

	fullResponseItem := handlerDto.RecordListItemResponse{
		RecordId:        recordListArticleId1,
		Type:            "journal_article",
		MainTitle:       "第1章 「学校」制度の境界線",
		SubTitle:        testutil.Ptr("その形成と展開"),
		PageRange:       testutil.Ptr("15-46"),
		Publisher:       testutil.Ptr("東京大学出版会"),
		PublicationDate: testutil.Ptr("2022-11-30"),
		ReadStatus:      testutil.Ptr("reading"),
		Contributors: []handlerDto.ContributorInfo{
			{Name: "木村元", Role: "author"},
			{Name: "訳者B", Role: "interpreter"},
		},
		ContainerListItem: &handlerDto.ContainerListItemResponse{
			MainTitle: "境界線の学校史",
			SubTitle:  testutil.Ptr("収録物の副題"),
			Volume:    testutil.Ptr("第1巻"),
			Issue:     testutil.Ptr("第11号"),
			Contributors: []handlerDto.ContributorInfo{
				{Name: "編者A", Role: "compiler"},
			},
		},
	}

	// 任意項目が空文字の書誌と収録物(NULL化の検証用)
	emptyStrItem := usecaseDto.RecordListItem{
		Record: record.UserRecordEntry{
			RecordEntryId: recordListArticleId1,
			RecordType:    record.RecordTypeJournalArticle,
			MainTitle:     "論文のタイトル",
			SubTitle:      testutil.Ptr(""),
			PageRange:     testutil.Ptr(""),
			Publisher:     testutil.Ptr(""),
			ContainerId:   testutil.Ptr(recordListContainerId),
		},
		Container: &record.UserRecordEntry{
			RecordEntryId: recordListContainerId,
			RecordType:    record.RecordTypeJournal,
			MainTitle:     "雑誌のタイトル",
			SubTitle:      testutil.Ptr(""),
			Volume:        testutil.Ptr(""),
			Issue:         testutil.Ptr(""),
		},
	}

	tests := []struct {
		name string
		// output usecaseが返す出力
		output       *usecaseDto.GetRecordsOutput
		expectedBody handlerDto.GetRecordsResponse
		// bodyContains / notContains レスポンスのJSONに含まれる/含まれてはいけない文字列
		bodyContains []string
		notContains  []string
	}{
		{
			name: "正常_全項目を含む書誌一覧",
			output: newOutput(2,
				fullItem,
				minimalItem(recordListArticleId2),
			),
			expectedBody: newRecordListResponse(2,
				fullResponseItem,
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			name:         "正常_収録物なしはcontainerがnull",
			output:       newOutput(1, minimalItem(recordListArticleId1)),
			expectedBody: newRecordListResponse(1, minimalResponseItem(recordListArticleId1)),
			bodyContains: []string{`"container":null`},
		},
		{
			name:   "正常_任意項目がnil",
			output: newOutput(1, minimalItem(recordListArticleId1)),
			expectedBody: newRecordListResponse(1, handlerDto.RecordListItemResponse{
				RecordId:  recordListArticleId1,
				Type:      "authored_book",
				MainTitle: "単著書のタイトル",
				// 任意項目と貢献者はすべてnil(空スライスではない)
				SubTitle:          nil,
				PageRange:         nil,
				Publisher:         nil,
				PublicationDate:   nil,
				ReadStatus:        nil,
				Contributors:      nil,
				ContainerListItem: nil,
			}),
		},
		{
			name:   "正常_任意項目の空文字はnull",
			output: newOutput(1, emptyStrItem),
			expectedBody: newRecordListResponse(1, handlerDto.RecordListItemResponse{
				RecordId:  recordListArticleId1,
				Type:      "journal_article",
				MainTitle: "論文のタイトル",
				SubTitle:  nil,
				PageRange: nil,
				Publisher: nil,
				ContainerListItem: &handlerDto.ContainerListItemResponse{
					MainTitle: "雑誌のタイトル",
					SubTitle:  nil,
					Volume:    nil,
					Issue:     nil,
				},
			}),
		},
		{
			name:         "正常_出版日が日精度",
			output:       newOutput(1, withPublicationDate(&dayDate, testutil.Ptr(record.DatePrecisionDay))),
			expectedBody: withPublicationDateResponse(testutil.Ptr("2022-11-30")),
		},
		{
			name:         "正常_出版日が月精度",
			output:       newOutput(1, withPublicationDate(&monthDate, testutil.Ptr(record.DatePrecisionMonth))),
			expectedBody: withPublicationDateResponse(testutil.Ptr("2022-11")),
		},
		{
			name:         "正常_出版日が年精度",
			output:       newOutput(1, withPublicationDate(&yearDate, testutil.Ptr(record.DatePrecisionYear))),
			expectedBody: withPublicationDateResponse(testutil.Ptr("2022")),
		},
		{
			// 精度が不明なまま日付を出すと月日を捏造することになるためnullを返す
			name:         "正常_出版日の精度がnil",
			output:       newOutput(1, withPublicationDate(&dayDate, nil)),
			expectedBody: withPublicationDateResponse(nil),
		},
		{
			// countはrecordsの件数ではなく、limit/offset適用前の総件数
			name: "正常_countは総件数",
			output: newOutput(25,
				minimalItem(recordListArticleId1),
				minimalItem(recordListArticleId2),
			),
			expectedBody: newRecordListResponse(25,
				minimalResponseItem(recordListArticleId1),
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			// usecaseは該当なしの場合にRecordsをnilで返すが、レスポンスはnullにしない
			name:         "正常_0件はrecordsが空配列",
			output:       &usecaseDto.GetRecordsOutput{TotalCount: 0, Records: nil},
			expectedBody: newRecordListResponse(0),
			bodyContains: []string{`"records":[]`},
		},
		{
			name: "正常_一覧の順序を保つ",
			output: newOutput(3,
				minimalItem(recordListArticleId3),
				minimalItem(recordListArticleId1),
				minimalItem(recordListArticleId2),
			),
			expectedBody: newRecordListResponse(3,
				minimalResponseItem(recordListArticleId3),
				minimalResponseItem(recordListArticleId1),
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			// 収録物のIDは参照・削除のガードの前提となるため露出させない
			name:         "正常_収録物IDを露出しない",
			output:       newOutput(1, fullItem),
			expectedBody: newRecordListResponse(1, fullResponseItem),
			notContains:  []string{recordListContainerId},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordListUsecaseMock(tt.output)

			rec, c := serveGetRecords(uc, "", recordCtx{})

			assertGetRecordsSucceeded(t, uc, rec, c, newRecordListInput(nil), tt.expectedBody)

			for _, s := range tt.bodyContains {
				assert.Contains(t, rec.Body.String(), s)
			}

			// 露出してはいけない値が含まれないこと
			for _, ng := range tt.notContains {
				assert.NotContains(t, rec.Body.String(), ng)
			}
		})
	}
}
