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
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------CreateRecord(*gin.Context)のテスト-------------

const recordCreatePath = "/api/v0/records"

// newRecordCreateUsecaseMock 指定したエラーを返すMockRecordUsecaseを返す(nilなら書誌IDを返して成功)
func newRecordCreateUsecaseMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		CreateFunc: func(ctx context.Context, ci usecaseDto.CreateRecordInput) (*usecaseDto.CreateRecordOutput, error) {
			if err != nil {
				return nil, err
			}
			return &usecaseDto.CreateRecordOutput{RecordId: recordTestRecordId}, nil
		},
	}
}

// serveCreateRecord CreateRecordを実行する
func serveCreateRecord(
	uc *mock.MockRecordUsecase,
	body string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	return serveRecord(uc, (*RecordHandler).CreateRecord,
		newJsonRequest(http.MethodPost, recordCreatePath, body), "", rc)
}

// TestRecordHandler_CreateRecord_正常 リクエストボディがusecase入力に変換され、201で書誌IDを返すこと
func TestRecordHandler_CreateRecord_正常(t *testing.T) {

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	publicationMonth := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	publicationYear := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		body string
		// expectedInput usecaseへ渡される変換結果
		expectedInput usecaseDto.CreateRecordInput
	}{
		{
			name: "正常_最小項目",
			body: `{"type":"authored_book","main_title":"単著書のタイトル"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
				},
			},
		},
		{
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
					{"label":"所感","content":"少し難しい。"}
				],
				"urls":[
					{"label":"オンラインPDF","value":"https://example.com/1"},
					{"label":"書影","value":"https://example.com/2"}
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
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
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
						{Label: testutil.Ptr("所感"), Content: "少し難しい。"},
					},
					Urls: []usecaseDto.UrlInput{
						{Label: testutil.Ptr("オンラインPDF"), Value: "https://example.com/1"},
						{Label: testutil.Ptr("書影"), Value: "https://example.com/2"},
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
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
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
			name: "正常_コンテナを持てるtypeでcontainer省略",
			body: `{"type":"journal_article","main_title":"論文"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				// コンテナ省略は許容する
				ContainerInput: nil,
			},
		},
		{
			name: "正常_出版日_年月日",
			body: `{"type":"authored_book","main_title":"単著書","publication_date":"2022-11-30"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationDate,
					DatePrecision:   testutil.Ptr(record.DatePrecisionDay),
				},
			},
		},
		{
			name: "正常_出版日_年月",
			body: `{"type":"authored_book","main_title":"単著書","publication_date":"2022-11"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationMonth,
					DatePrecision:   testutil.Ptr(record.DatePrecisionMonth),
				},
			},
		},
		{
			name: "正常_出版日_年のみ",
			body: `{"type":"authored_book","main_title":"単著書","publication_date":"2022"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationYear,
					DatePrecision:   testutil.Ptr(record.DatePrecisionYear),
				},
			},
		},
		{
			name: "正常_出版日_空文字",
			body: `{"type":"authored_book","main_title":"単著書","publication_date":""}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 空文字は未入力として扱う
					PublicationDate: nil,
					DatePrecision:   nil,
				},
			},
		},
		{
			name: "正常_任意スカラが空文字",
			body: `{"type":"authored_book","main_title":"単著書",
				"sub_title":"","page_range":"","publisher":"","read_status":""}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 空文字はすべてnilに変換される
					SubTitle:   nil,
					PageRange:  nil,
					Publisher:  nil,
					ReadStatus: nil,
				},
			},
		},
		{
			name: "正常_任意スカラ未指定",
			body: `{"type":"authored_book","main_title":"単著書"}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:       record.RecordTypeAuthoredBook,
					MainTitle:  "単著書",
					SubTitle:   nil,
					PageRange:  nil,
					Publisher:  nil,
					ReadStatus: nil,
				},
			},
		},
		{
			name: "正常_option_valuesが全て空文字",
			body: `{"type":"authored_book","main_title":"単著書",
				"option_values":{"isbn":"","edition":"","original_title":""}}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 全項目が空文字なら構造体ごとnilになる
					OptionValues: nil,
				},
			},
		},
		{
			name: "正常_option_valuesが一部のみ",
			body: `{"type":"authored_book","main_title":"単著書",
				"option_values":{"isbn":"9784130513555"}}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn:          testutil.Ptr("9784130513555"),
						Edition:       nil,
						OriginalTitle: nil,
					},
				},
			},
		},
		{
			name: "正常_メモとURLのラベルが空文字",
			body: `{"type":"authored_book","main_title":"単著書",
				"memos":[{"label":"","content":"メモの内容"}],
				"urls":[{"label":"","value":"https://example.com/1"}]}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// ラベルの空文字はnilになるが、内容と値はそのまま渡る
					Memos: []usecaseDto.MemoInput{
						{Label: nil, Content: "メモの内容"},
					},
					Urls: []usecaseDto.UrlInput{
						{Label: nil, Value: "https://example.com/1"},
					},
				},
			},
		},
		{
			name: "正常_構成要素が空配列",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[],"memos":[],"urls":[]}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 空配列はnilとして渡される
					Contributors: nil,
					Memos:        nil,
					Urls:         nil,
				},
			},
		},
		{
			name: "正常_コンテナの任意項目が空文字",
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","sub_title":"","volume":"","issue":"","read_status":""}}`,
			expectedInput: usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				ContainerInput: &usecaseDto.ContainerInput{
					MainTitle: "雑誌",
					// 空文字はすべてnilに変換される
					SubTitle:   nil,
					Volume:     nil,
					Issue:      nil,
					ReadStatus: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordCreateUsecaseMock(nil)

			rec, c := serveCreateRecord(uc, tt.body, recordCtx{})

			require.True(t, uc.CreateCalled)
			assert.Equal(t, tt.expectedInput, uc.LastCreateInput)

			require.Len(t, c.Errors, 0)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, ctxutil.REF_API_003, ctxutil.GetApiId(uc.LastCtx))
			testutil.AssertResponse(t, rec, handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			})
		})
	}
}
