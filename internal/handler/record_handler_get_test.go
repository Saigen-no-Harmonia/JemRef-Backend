package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------GetRecord(*gin.Context)のテスト-------------

const (
	// recordGetTargetId 参照対象の書誌ID
	recordGetTargetId = "GETTARGET000000000000000001"
	// recordGetContainerId 収録物の書誌ID。レスポンスには露出しない。
	recordGetContainerId = "GETCONTAINER00000000000001"
)

// newRecordGetUsecaseMock 指定したOutputを返すMockRecordUsecaseを返す
func newRecordGetUsecaseMock(output *usecaseDto.GetRecordOutput) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		GetFunc: func(ctx context.Context, gi usecaseDto.GetRecordInput) (*usecaseDto.GetRecordOutput, error) {
			return output, nil
		},
	}
}

// newRecordGetUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordGetUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		GetFunc: func(ctx context.Context, gi usecaseDto.GetRecordInput) (*usecaseDto.GetRecordOutput, error) {
			return nil, err
		},
	}
}

// serveGetRecord GetRecordを実行する(recordIdが空文字ならパスパラメータを設定しない)
func serveGetRecord(
	uc *mock.MockRecordUsecase,
	recordId string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodGet, "/api/v0/records/"+recordGetTargetId, nil)
	return serveRecord(uc, (*RecordHandler).GetRecord, req, recordId, rc)
}

// TestRecordHandler_GetRecord_正常 usecaseの出力からレスポンスを組み立てること
func TestRecordHandler_GetRecord_正常(t *testing.T) {

	dayDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	monthDate := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	yearDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	// minimalRecord 必須項目のみを持つ書誌を返す
	// テストケースごとに書き換えるため、共有せず都度生成する。
	minimalRecord := func() record.UserRecordEntry {
		return record.UserRecordEntry{
			RecordEntryId: recordGetTargetId,
			InternalUid:   recordTestInternalUid,
			RecordType:    record.RecordTypeAuthoredBook,
			MainTitle:     "単著書のタイトル",
		}
	}

	// minimalResponse 必須項目のみのレスポンスを返す
	minimalResponse := func() handlerDto.GetRecordResponse {
		return handlerDto.GetRecordResponse{
			RecordId: recordGetTargetId,
			RecordDetailItem: handlerDto.RecordDetailItem{
				Type:      "authored_book",
				MainTitle: "単著書のタイトル",
			},
		}
	}

	// withPublicationDate 出版日と精度を設定した書誌を返す
	withPublicationDate := func(date *time.Time, precision *record.DatePrecision) record.UserRecordEntry {
		article := minimalRecord()
		article.PublicationDate = date
		article.DatePrecision = precision
		return article
	}

	// withPublicationDateResponse 出版日を設定したレスポンスを返す
	withPublicationDateResponse := func(publicationDate string) handlerDto.GetRecordResponse {
		res := minimalResponse()
		res.PublicationDate = testutil.Ptr(publicationDate)
		return res
	}

	// withOptionValues 項目値を設定した書誌を返す
	withOptionValues := func(values ...record.OptionValue) record.UserRecordEntry {
		article := minimalRecord()
		article.OptionValues = values
		return article
	}

	// 全項目と構成要素4種を持つ書誌
	fullRecord := record.UserRecordEntry{
		RecordEntryId:   recordGetTargetId,
		InternalUid:     recordTestInternalUid,
		RecordType:      record.RecordTypeJournalArticle,
		MainTitle:       "第1章 「学校」制度の境界線",
		SubTitle:        testutil.Ptr("その形成と展開"),
		PageRange:       testutil.Ptr("15-46"),
		Publisher:       testutil.Ptr("東京大学出版会"),
		PublicationDate: &dayDate,
		DatePrecision:   testutil.Ptr(record.DatePrecisionDay),
		ReadStatus:      testutil.Ptr(record.ReadStatusReading),
		ContainerId:     testutil.Ptr(recordGetContainerId),
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
		Memos: []record.MemoEntry{
			{Title: testutil.Ptr("要約"), Content: "学校の戦後史。"},
			{Title: nil, Content: "見出しなしのメモ"},
		},
		Urls: []record.UrlEntry{
			{Label: testutil.Ptr("オンラインPDF"), Url: "https://example.com/pdf"},
			{Label: nil, Url: "https://example.com/info"},
		},
		OptionValues: []record.OptionValue{
			{RecordField: record.RecordFieldIsbn, RawValue: "9784130513555"},
			{RecordField: record.RecordFieldEdition, RawValue: "第2版"},
			{RecordField: record.RecordFieldOriginalTitle, RawValue: "School History of Boundary"},
		},
	}

	// 全項目と構成要素を持つ収録物
	fullContainer := record.UserRecordEntry{
		RecordEntryId: recordGetContainerId,
		InternalUid:   recordTestInternalUid,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "境界線の学校史",
		SubTitle:      testutil.Ptr("収録物の副題"),
		Volume:        testutil.Ptr("第1巻"),
		Issue:         testutil.Ptr("第11号"),
		ReadStatus:    testutil.Ptr(record.ReadStatusRead),
		ContributorRelations: []record.ContributorRelation{
			{
				Contributor:  record.ContributorEntry{RawName: "編者A"},
				Contribution: record.ContributionEntry{Role: record.ContributorRoleCompiler},
			},
		},
		OptionValues: []record.OptionValue{
			{RecordField: record.RecordFieldIsbn, RawValue: "9784130519999"},
		},
	}

	// 空文字を持つ書誌と収録物(NULL化の検証用)
	emptyStrRecord := minimalRecord()
	emptyStrRecord.SubTitle = testutil.Ptr("")
	emptyStrRecord.PageRange = testutil.Ptr("")
	emptyStrRecord.Publisher = testutil.Ptr("")
	emptyStrRecord.Memos = []record.MemoEntry{{Title: testutil.Ptr(""), Content: "メモ内容"}}
	emptyStrRecord.Urls = []record.UrlEntry{{Label: testutil.Ptr(""), Url: "https://example.com/1"}}

	emptyStrContainer := record.UserRecordEntry{
		RecordEntryId: recordGetContainerId,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "収録物のタイトル",
		SubTitle:      testutil.Ptr(""),
		Volume:        testutil.Ptr(""),
		Issue:         testutil.Ptr(""),
	}

	tests := []struct {
		name string
		// article / container usecaseが返す書誌と収録物
		article      record.UserRecordEntry
		container    *record.UserRecordEntry
		expectedBody handlerDto.GetRecordResponse
		// notContains レスポンスのJSONに含まれてはいけない文字列
		notContains []string
	}{
		{
			name:      "正常_全項目を含む書誌を取得",
			article:   fullRecord,
			container: &fullContainer,
			expectedBody: handlerDto.GetRecordResponse{
				RecordId: recordGetTargetId,
				RecordDetailItem: handlerDto.RecordDetailItem{
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
					Memos: []handlerDto.MemoInfo{
						{Label: testutil.Ptr("要約"), Content: "学校の戦後史。"},
						{Label: nil, Content: "見出しなしのメモ"},
					},
					Urls: []handlerDto.UrlInfo{
						{Label: testutil.Ptr("オンラインPDF"), Value: "https://example.com/pdf"},
						{Label: nil, Value: "https://example.com/info"},
					},
					OptionValues: &handlerDto.OptionValues{
						Isbn:          testutil.Ptr("9784130513555"),
						Edition:       testutil.Ptr("第2版"),
						OriginalTitle: testutil.Ptr("School History of Boundary"),
					},
				},
				ContainerDetailItem: &handlerDto.ContainerDetailItem{
					MainTitle:  "境界線の学校史",
					SubTitle:   testutil.Ptr("収録物の副題"),
					Volume:     testutil.Ptr("第1巻"),
					Issue:      testutil.Ptr("第11号"),
					ReadStatus: testutil.Ptr("read"),
					Contributors: []handlerDto.ContributorInfo{
						{Name: "編者A", Role: "compiler"},
					},
					OptionValues: &handlerDto.OptionValues{
						Isbn: testutil.Ptr("9784130519999"),
					},
				},
			},
			// 収録物のIDは削除・参照のガードの前提となるため露出させない
			notContains: []string{recordGetContainerId},
		},
		{
			name:         "正常_収録物なし",
			article:      minimalRecord(),
			expectedBody: minimalResponse(),
		},
		{
			name:    "正常_任意項目がnil",
			article: minimalRecord(),
			expectedBody: handlerDto.GetRecordResponse{
				RecordId: recordGetTargetId,
				RecordDetailItem: handlerDto.RecordDetailItem{
					Type:      "authored_book",
					MainTitle: "単著書のタイトル",
					// 任意項目と構成要素はすべてnil(空スライスではない)
					SubTitle:        nil,
					PageRange:       nil,
					Publisher:       nil,
					PublicationDate: nil,
					ReadStatus:      nil,
					Contributors:    nil,
					Memos:           nil,
					Urls:            nil,
					OptionValues:    nil,
				},
				ContainerDetailItem: nil,
			},
		},
		{
			name:      "正常_空文字はnullになる",
			article:   emptyStrRecord,
			container: &emptyStrContainer,
			expectedBody: handlerDto.GetRecordResponse{
				RecordId: recordGetTargetId,
				RecordDetailItem: handlerDto.RecordDetailItem{
					Type:      "authored_book",
					MainTitle: "単著書のタイトル",
					SubTitle:  nil,
					PageRange: nil,
					Publisher: nil,
					Memos:     []handlerDto.MemoInfo{{Label: nil, Content: "メモ内容"}},
					Urls:      []handlerDto.UrlInfo{{Label: nil, Value: "https://example.com/1"}},
				},
				ContainerDetailItem: &handlerDto.ContainerDetailItem{
					MainTitle: "収録物のタイトル",
					SubTitle:  nil,
					Volume:    nil,
					Issue:     nil,
				},
			},
		},
		{
			name:         "正常_出版日が日精度",
			article:      withPublicationDate(&dayDate, testutil.Ptr(record.DatePrecisionDay)),
			expectedBody: withPublicationDateResponse("2022-11-30"),
		},
		{
			name:         "正常_出版日が月精度",
			article:      withPublicationDate(&monthDate, testutil.Ptr(record.DatePrecisionMonth)),
			expectedBody: withPublicationDateResponse("2022-11"),
		},
		{
			name:         "正常_出版日が年精度",
			article:      withPublicationDate(&yearDate, testutil.Ptr(record.DatePrecisionYear)),
			expectedBody: withPublicationDateResponse("2022"),
		},
		{
			// 精度が不明なまま日付を出すと月日を捏造することになるためnullを返す
			name:         "正常_出版日の精度がnil",
			article:      withPublicationDate(&dayDate, nil),
			expectedBody: minimalResponse(),
		},
		{
			name: "正常_項目値が一部のみ",
			article: withOptionValues(
				record.OptionValue{RecordField: record.RecordFieldIsbn, RawValue: "9784130513555"},
			),
			expectedBody: func() handlerDto.GetRecordResponse {
				res := minimalResponse()
				res.OptionValues = &handlerDto.OptionValues{
					Isbn:          testutil.Ptr("9784130513555"),
					Edition:       nil,
					OriginalTitle: nil,
				}
				return res
			}(),
		},
		{
			// 空文字の項目値はスキップされ、結果が空になるためoption_values自体がnullになる
			name: "正常_項目値が空文字のみならnull",
			article: withOptionValues(
				record.OptionValue{RecordField: record.RecordFieldIsbn, RawValue: ""},
			),
			expectedBody: minimalResponse(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordGetUsecaseMock(&usecaseDto.GetRecordOutput{
				Record:    tt.article,
				Container: tt.container,
			})

			rec, c := serveGetRecord(uc, recordGetTargetId, recordCtx{})

			// パスパラメータと内部ユーザIDがそのままusecaseへ渡る
			require.True(t, uc.GetCalled)
			assert.Equal(t, usecaseDto.GetRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordGetTargetId,
			}, uc.LastGetInput)

			require.Len(t, c.Errors, 0)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, ctxutil.REF_API_002, ctxutil.GetApiId(uc.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)

			// 露出してはいけない値が含まれないこと
			for _, ng := range tt.notContains {
				assert.NotContains(t, rec.Body.String(), ng)
			}
		})
	}
}

// TestRecordHandler_GetRecord_異常 contextの値やパスパラメータが不正なら、usecaseを呼ばずにエラーを返すこと
func TestRecordHandler_GetRecord_異常(t *testing.T) {

	tests := []struct {
		name string
		ctx  recordCtx
		// omitRecordId パスパラメータ id を設定しない
		omitRecordId      bool
		expectedErrorBody error
	}{
		{
			name:              "異常_internalUid取得失敗",
			ctx:               recordCtx{omitInternalUid: true},
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_書誌IDが空",
			omitRecordId:      true,
			expectedErrorBody: ErrRecordIdRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordGetUsecaseErrMock(ErrTestUnexpected)

			recordId := recordGetTargetId
			if tt.omitRecordId {
				recordId = ""
			}
			_, c := serveGetRecord(uc, recordId, tt.ctx)

			assert.False(t, uc.GetCalled)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestRecordHandler_GetRecord_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestRecordHandler_GetRecord_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		usecaseErr error
	}{
		{
			name:       "異常_usecaseが書誌なしエラー",
			usecaseErr: usecase.ErrRecordNotFound,
		},
		{
			name:       "異常_usecaseが予期しないエラー",
			usecaseErr: ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordGetUsecaseErrMock(tt.usecaseErr)

			_, c := serveGetRecord(uc, recordGetTargetId, recordCtx{})

			assert.True(t, uc.GetCalled)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
