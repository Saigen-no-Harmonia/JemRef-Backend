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

// -------------UpdateRecord(*gin.Context)のテスト-------------
//
// リクエストDTOはCreateと同一構造で、変換も toRecordInput / toContainerInput /
// toReadStatus / toContributorsInput / nilIfEmpty / GetDateAndPrecision を
// CreateRecordと共用している。このため出版日の精度・任意スカラの空文字と未指定・
// option_valuesの各パターン・貢献者やメモURLの必須エラーといった変換の網羅は
// TestRecordHandler_CreateRecord_正常 / _異常 で担保する。
//
// ここで確認するのはUpdate固有の差分に絞る。
//   - パスパラメータ id を RecordId へ渡すこと(Createにはない)
//   - publicUid を参照しないこと(Createとの差分)
//   - レスポンスが 200 + StatusResponse{"OK"} であること(Createは201)
//   - API IDが REF-API-004 であること
//   - chkUpdateParam の2つの検証
//   - toUpdateRecordInput が持つ3つのエラー返却経路が握りつぶされないこと
//   - usecaseのエラーが変換されずそのまま c.Errors に乗ること

// recordUpdateTargetId 更新対象として指定する書誌ID
const recordUpdateTargetId = "UPDATETARGET00000000000001"

// newRecordUpdateUsecaseMock 指定したエラーを返すMockRecordUsecaseを返す(nilなら成功)
func newRecordUpdateUsecaseMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		UpdateFunc: func(ctx context.Context, ui usecaseDto.UpdateRecordInput) error {
			return err
		},
	}
}

// serveUpdateRecord UpdateRecordを実行する(recordIdが空文字ならパスパラメータを設定しない)
func serveUpdateRecord(
	uc *mock.MockRecordUsecase,
	recordId string,
	body string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := newJsonRequest(http.MethodPut, "/api/v0/records/"+recordUpdateTargetId, body)
	return serveRecord(uc, (*RecordHandler).UpdateRecord, req, recordId, rc)
}

// TestRecordHandler_UpdateRecord_正常 パスパラメータとボディがusecase入力に変換され、200を返すこと
func TestRecordHandler_UpdateRecord_正常(t *testing.T) {

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		ctx  recordCtx
		body string
		// expectedInput usecaseへ渡される変換結果
		expectedInput usecaseDto.UpdateRecordInput
	}{
		{
			name: "正常_最小項目",
			body: `{"type":"authored_book","main_title":"更新後の単著書"}`,
			// パスパラメータがRecordIdへ、contextの値がInternalUserIdへ渡る
			expectedInput: usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "更新後の単著書",
				},
			},
		},
		{
			// 共有している変換のスモーク。各項目の網羅はCreateのテストで担保する
			name: "正常_全項目指定",
			body: `{
				"type":"journal_article",
				"main_title":"第1章 「学校」制度の境界線",
				"sub_title":"その形成と展開",
				"page_range":"15-46",
				"publisher":"東京大学出版会",
				"publication_date":"2022-11-30",
				"read_status":"reading",
				"contributors":[
					{"name":"木村元","role":"author"},
					{"name":"山田太郎","role":"interpreter"}
				],
				"memos":[
					{"label":"要約","content":"学校の戦後史。"},
					{"content":"ラベルなしのメモ"}
				],
				"urls":[
					{"label":"オンラインPDF","value":"https://example.com/1"},
					{"value":"https://example.com/2"}
				],
				"option_values":{"isbn":"9784130513555","edition":"第2版","original_title":"School History"},
				"container":{
					"main_title":"境界線の学校史",
					"sub_title":"戦後日本の学校化社会の周辺",
					"volume":"日本教育史講座第1巻",
					"issue":"第11号",
					"read_status":"read",
					"contributors":[{"name":"編者A","role":"compiler"}],
					"option_values":{"isbn":"9784130513556"}
				}
			}`,
			expectedInput: usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeJournalArticle,
					MainTitle:       "第1章 「学校」制度の境界線",
					SubTitle:        testutil.Ptr("その形成と展開"),
					PageRange:       testutil.Ptr("15-46"),
					Publisher:       testutil.Ptr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   testutil.Ptr(record.DatePrecisionDay),
					ReadStatus:      testutil.Ptr(record.ReadStatusReading),
					Contributors: []usecaseDto.ContributorInput{
						{Name: "木村元", Role: record.ContributorRoleAuthor},
						{Name: "山田太郎", Role: record.ContributorRoleInterpreter},
					},
					Memos: []usecaseDto.MemoInput{
						{Label: testutil.Ptr("要約"), Content: "学校の戦後史。"},
						{Content: "ラベルなしのメモ"},
					},
					Urls: []usecaseDto.UrlInput{
						{Label: testutil.Ptr("オンラインPDF"), Value: "https://example.com/1"},
						{Value: "https://example.com/2"},
					},
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn:          testutil.Ptr("9784130513555"),
						Edition:       testutil.Ptr("第2版"),
						OriginalTitle: testutil.Ptr("School History"),
					},
				},
				ContainerInput: &usecaseDto.ContainerInput{
					MainTitle:  "境界線の学校史",
					SubTitle:   testutil.Ptr("戦後日本の学校化社会の周辺"),
					Volume:     testutil.Ptr("日本教育史講座第1巻"),
					Issue:      testutil.Ptr("第11号"),
					ReadStatus: testutil.Ptr(record.ReadStatusRead),
					Contributors: []usecaseDto.ContributorInput{
						{Name: "編者A", Role: record.ContributorRoleCompiler},
					},
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn: testutil.Ptr("9784130513556"),
					},
				},
			},
		},
		{
			name: "正常_コンテナあり",
			body: `{"type":"journal_article","main_title":"論文","container":{"main_title":"雑誌"}}`,
			expectedInput: usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				ContainerInput: &usecaseDto.ContainerInput{
					MainTitle: "雑誌",
				},
			},
		},
		{
			// containerを省略した場合はContainerInputがnilになり、
			// usecase側で「収録物を外す」経路に入る
			name: "正常_コンテナを持てるtypeでcontainer省略",
			body: `{"type":"journal_article","main_title":"論文"}`,
			expectedInput: usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				ContainerInput: nil,
			},
		},
		{
			// UpdateRecordはpublicUidを参照しない(レスポンスがStatusResponseのみのため)。
			// 将来publicUidを参照する変更が入った場合にここで気づける。
			name: "正常_publicUidが未設定でも成功",
			ctx:  recordCtx{omitPublicUid: true},
			body: `{"type":"authored_book","main_title":"単著書"}`,
			expectedInput: usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordUpdateUsecaseMock(nil)

			rec, c := serveUpdateRecord(uc, recordUpdateTargetId, tt.body, tt.ctx)

			require.True(t, uc.UpdateCalled)
			assert.Equal(t, tt.expectedInput, uc.LastUpdateInput)

			require.Len(t, c.Errors, 0)
			// Createは201だが、更新はIDを返さないため200
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, ctxutil.REF_API_004, ctxutil.GetApiId(uc.LastCtx))
			testutil.AssertResponse(t, rec, handlerDto.StatusResponse{Status: statusOk})
		})
	}
}

// TestRecordHandler_UpdateRecord_異常 contextの値・パスパラメータ・ボディが不正なら、usecaseを呼ばずにエラーを返すこと
func TestRecordHandler_UpdateRecord_異常(t *testing.T) {

	tests := []struct {
		name string
		ctx  recordCtx
		// omitRecordId パスパラメータ id を設定しない
		omitRecordId      bool
		body              string
		expectedErrorBody error
	}{
		{
			name:              "異常_internalUid取得失敗",
			ctx:               recordCtx{omitInternalUid: true},
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_書誌IDが空",
			omitRecordId:      true,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			expectedErrorBody: ErrRecordIdRequired,
		},
		{
			name:              "異常_JSONが不正",
			body:              `{"type":`,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			// updateRecordRequiredErrorsが対応表として使われていることの確認
			name:              "異常_type未指定",
			body:              `{"main_title":"単著書"}`,
			expectedErrorBody: ErrRecordTypeRequired,
		},
		{
			name:              "異常_main_title未指定",
			body:              `{"type":"authored_book"}`,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			// chkUpdateParamの1つ目の検証(IsValid)
			name:              "異常_typeが不正値",
			body:              `{"type":"unknown_type","main_title":"単著書"}`,
			expectedErrorBody: ErrInvalidRecordType,
		},
		{
			// chkUpdateParamの2つ目の検証(IsArticle)
			name:              "異常_コンテナを持てないtypeにcontainer同送",
			body:              `{"type":"authored_book","main_title":"単著書","container":{"main_title":"雑誌"}}`,
			expectedErrorBody: ErrContainerNotAllowed,
		},
		{
			// toUpdateRecordInputがtoRecordInputのエラーを握りつぶさないことの確認。
			// 出版日以外(read_status・貢献者)の網羅はCreateのテストで担保する。
			name:              "異常_出版日フォーマット不正",
			body:              `{"type":"authored_book","main_title":"単著書","publication_date":"2022/11/30"}`,
			expectedErrorBody: ErrInvalidPublicationDate,
		},
		{
			// toUpdateRecordInputがtoContainerInputのエラーを握りつぶさないことの確認
			name: "異常_containerのread_statusが不正値",
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","read_status":"unknown"}}`,
			expectedErrorBody: ErrInvalidReadStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordUpdateUsecaseMock(nil)

			recordId := recordUpdateTargetId
			if tt.omitRecordId {
				recordId = ""
			}
			_, c := serveUpdateRecord(uc, recordId, tt.body, tt.ctx)

			assert.False(t, uc.UpdateCalled)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestRecordHandler_UpdateRecord_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestRecordHandler_UpdateRecord_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		body       string
		usecaseErr error
	}{
		{
			name:       "異常_usecaseが書誌なしエラー",
			body:       `{"type":"authored_book","main_title":"単著書"}`,
			usecaseErr: usecase.ErrRecordNotFound,
		},
		{
			// REF-API-004で新設した書誌タイプ変更の拒否(E0025)
			name:       "異常_usecaseが書誌タイプ変更エラー",
			body:       `{"type":"authored_book","main_title":"単著書"}`,
			usecaseErr: usecase.ErrRecordTypeNotChangeable,
		},
		{
			// 収録物のデータ不整合(500)。handlerは変換せず素通しする
			name:       "異常_usecaseが収録物不整合エラー",
			body:       `{"type":"journal_article","main_title":"論文","container":{"main_title":"雑誌"}}`,
			usecaseErr: usecase.ErrContainerNotFound,
		},
		{
			name:       "異常_usecaseが予期しないエラー",
			body:       `{"type":"authored_book","main_title":"単著書"}`,
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordUpdateUsecaseMock(tt.usecaseErr)

			_, c := serveUpdateRecord(uc, recordUpdateTargetId, tt.body, recordCtx{})

			assert.True(t, uc.UpdateCalled)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
