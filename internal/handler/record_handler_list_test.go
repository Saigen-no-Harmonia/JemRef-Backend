package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------GetRecords(*gin.Context)のテスト-------------

const (
	// recordListArticleId1〜3 usecaseが返す書誌のID
	recordListArticleId1 = "LISTARTICLE000000000000001"
	recordListArticleId2 = "LISTARTICLE000000000000002"
	recordListArticleId3 = "LISTARTICLE000000000000003"
	// recordListContainerId 収録物の書誌ID。レスポンスには露出しない。
	recordListContainerId = "LISTCONTAINER0000000000001"
)

// newRecordListUsecaseMock 指定したOutputを返すMockRecordUsecaseを返す
func newRecordListUsecaseMock(output *usecaseDto.GetRecordsOutput) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		ListFunc: func(ctx context.Context, li usecaseDto.GetRecordsInput) (*usecaseDto.GetRecordsOutput, error) {
			return output, nil
		},
	}
}

// newRecordListUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordListUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		ListFunc: func(ctx context.Context, li usecaseDto.GetRecordsInput) (*usecaseDto.GetRecordsOutput, error) {
			return nil, err
		},
	}
}

// serveGetRecords GetRecordsを実行する(queryは"?"を含めないクエリ文字列)
func serveGetRecords(
	uc *mock.MockRecordUsecase,
	query string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodGet, "/api/v0/records?"+query, nil)
	return serveRecord(uc, (*RecordHandler).GetRecords, req, "", rc)
}

// newRecordListInput デフォルト値の入力を書き換えたusecase入力を返す
//
// パラメータ未指定時は publish_year / asc / 20 / 0 になる。
func newRecordListInput(modify func(in *usecaseDto.GetRecordsInput)) usecaseDto.GetRecordsInput {
	in := usecaseDto.GetRecordsInput{
		InternalUserId: recordTestInternalUid,
		SortKey:        record.SortKeyPublishYear,
		SortOrder:      record.SortOrderAsc,
		Limit:          20,
		Offset:         0,
	}
	if modify != nil {
		modify(&in)
	}
	return in
}

// newRecordListResponse 指定した件数と書誌からレスポンスを組み立てる(書誌なしは空配列)
func newRecordListResponse(count int, items ...handlerDto.RecordListItemResponse) handlerDto.GetRecordsResponse {
	records := []handlerDto.RecordListItemResponse{}
	records = append(records, items...)
	return handlerDto.GetRecordsResponse{Count: count, Records: &records}
}

// assertGetRecordsSucceeded usecaseが期待どおりの入力で呼ばれ、200で期待どおりのレスポンスを返したこと
func assertGetRecordsSucceeded(
	t *testing.T,
	uc *mock.MockRecordUsecase,
	rec *httptest.ResponseRecorder,
	c *gin.Context,
	expectedInput usecaseDto.GetRecordsInput,
	expectedBody handlerDto.GetRecordsResponse,
) {
	t.Helper()

	require.True(t, uc.ListCalled)
	assert.Equal(t, expectedInput, uc.LastListInput)

	require.Len(t, c.Errors, 0)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ctxutil.REF_API_001, ctxutil.GetApiId(uc.LastCtx))
	testutil.AssertResponse(t, rec, expectedBody)
}

// TestRecordHandler_GetRecords_入力変換 クエリパラメータがusecase入力に変換されること
func TestRecordHandler_GetRecords_入力変換(t *testing.T) {

	tests := []struct {
		name string
		// query URLのクエリ文字列("?"は含めない)
		query         string
		expectedInput usecaseDto.GetRecordsInput
	}{
		{
			name:          "正常_パラメータ未指定はデフォルト値",
			query:         "",
			expectedInput: newRecordListInput(nil),
		},
		{
			name: "正常_全パラメータを指定",
			query: "contributor_name=" + url.QueryEscape("木村") +
				"&title=" + url.QueryEscape("学校") +
				"&publish_year=2022&sort_key=insert_date&sort_type=desc&limit=50&offset=20",
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.ContributorName = testutil.Ptr("木村")
				in.Title = testutil.Ptr("学校")
				in.PublishYear = testutil.Ptr(2022)
				in.SortKey = record.SortKeyInsertDate
				in.SortOrder = record.SortOrderDesc
				in.Limit = 50
				in.Offset = 20
			}),
		},
		{
			name:          "正常_空文字の検索条件は未指定扱い",
			query:         "contributor_name=&title=&publish_year=&sort_key=&sort_type=",
			expectedInput: newRecordListInput(nil),
		},
		{
			name: "正常_検索語の前後空白を除去",
			// 貢献者名は半角空白、タイトルは全角空白で囲む
			query: "contributor_name=" + url.QueryEscape(" 木村 ") +
				"&title=" + url.QueryEscape("　学校　"),
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.ContributorName = testutil.Ptr("木村")
				in.Title = testutil.Ptr("学校")
			}),
		},
		{
			name: "正常_空白のみの検索語は未指定扱い",
			query: "contributor_name=" + url.QueryEscape("  ") +
				"&title=" + url.QueryEscape(" 　"),
			expectedInput: newRecordListInput(nil),
		},
		{
			name:  "正常_語中の空白は残す",
			query: "title=" + url.QueryEscape("学校 教育"),
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.Title = testutil.Ptr("学校 教育")
			}),
		},
		{
			name:  "正常_limitの下限",
			query: "limit=1",
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.Limit = 1
			}),
		},
		{
			name:  "正常_limitの上限",
			query: "limit=100",
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.Limit = 100
			}),
		},
		{
			// 明示した0はデフォルト値の0と区別されず、そのまま渡る
			name:          "正常_offsetが0",
			query:         "offset=0",
			expectedInput: newRecordListInput(nil),
		},
		{
			name:  "正常_sort_keyとsort_typeの各値",
			query: "sort_key=publish_year&sort_type=desc",
			expectedInput: newRecordListInput(func(in *usecaseDto.GetRecordsInput) {
				in.SortKey = record.SortKeyPublishYear
				in.SortOrder = record.SortOrderDesc
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordListUsecaseMock(&usecaseDto.GetRecordsOutput{TotalCount: 0})

			rec, c := serveGetRecords(uc, tt.query, recordCtx{})

			assertGetRecordsSucceeded(t, uc, rec, c, tt.expectedInput, newRecordListResponse(0))
		})
	}
}

// TestRecordHandler_GetRecords_異常 contextの値やクエリパラメータが不正なら、usecaseを呼ばずにエラーを返すこと
func TestRecordHandler_GetRecords_異常(t *testing.T) {

	tests := []struct {
		name string
		ctx  recordCtx
		// query URLのクエリ文字列("?"は含めない)
		query             string
		expectedErrorBody error
	}{
		{
			name:              "異常_internalUid取得失敗",
			ctx:               recordCtx{omitInternalUid: true},
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_limitが数値でない",
			query:             "limit=abc",
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:              "異常_offsetが数値でない",
			query:             "offset=abc",
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:              "異常_publish_yearが数字以外を含む",
			query:             "publish_year=20xx",
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_publish_yearが3桁",
			query:             "publish_year=202",
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_publish_yearが5桁",
			query:             "publish_year=20201",
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			// \d は半角数字にのみ一致する
			name:              "異常_publish_yearが全角数字",
			query:             "publish_year=" + url.QueryEscape("２０２２"),
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_sort_keyが不正",
			query:             "sort_key=title",
			expectedErrorBody: ErrInvalidSortKey,
		},
		{
			// 大文字小文字は区別する
			name:              "異常_sort_typeが不正",
			query:             "sort_type=ASC",
			expectedErrorBody: ErrInvalidSortOrder,
		},
		{
			name:              "異常_limitが0",
			query:             "limit=0",
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_limitが上限超過",
			query:             "limit=101",
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_limitが負",
			query:             "limit=-1",
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			// キーがあるためginのbindで0になり、未指定(デフォルト値)にはならない
			name:              "異常_limitが空文字",
			query:             "limit=",
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_offsetが負",
			query:             "offset=-1",
			expectedErrorBody: ErrInvalidOffset,
		},
		{
			// DTOの宣言順(publish_year → sort_key)で先の項目のエラーを返す
			name:              "異常_複数項目が不正なら宣言順で先の項目",
			query:             "sort_key=title&publish_year=20xx",
			expectedErrorBody: ErrInvalidPublishYear,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordListUsecaseErrMock(ErrTestUnexpected)

			_, c := serveGetRecords(uc, tt.query, tt.ctx)

			assert.False(t, uc.ListCalled)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestRecordHandler_GetRecords_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestRecordHandler_GetRecords_usecaseエラー(t *testing.T) {
	t.Run("異常_usecaseが予期しないエラー", func(t *testing.T) {
		uc := newRecordListUsecaseErrMock(ErrTestUnexpected)

		_, c := serveGetRecords(uc, "", recordCtx{})

		// エラーになる場合も、入力はデフォルト値で渡っていること
		require.True(t, uc.ListCalled)
		assert.Equal(t, newRecordListInput(nil), uc.LastListInput)
		assertHandlerError(t, c, ErrTestUnexpected)
	})
}
