package handler

import (
	"context"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/record"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// recordTestInternalUid テスト用の内部ユーザID
	recordTestInternalUid = int64(1001)
	// recordTestPublicUid テスト用の公開用ユーザID
	recordTestPublicUid = "PUBLICUID00000000000000001"
	// recordTestRecordId usecaseが返す書誌ID
	recordTestRecordId = "RECORDID000000000000000001"
)

// recordStrPtr string型のポインタを返却する
func recordStrPtr(s string) *string {
	return &s
}

// newRecordUsecaseMock 常に成功するMockRecordUsecaseを返す
func newRecordUsecaseMock() *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		CreateFunc: func(
			ctx context.Context,
			ci usecaseDto.CreateRecordInput,
		) (*usecaseDto.CreateRecordOutput, error) {
			return &usecaseDto.CreateRecordOutput{RecordId: recordTestRecordId}, nil
		},
	}
}

// newRecordUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		CreateFunc: func(
			ctx context.Context,
			ci usecaseDto.CreateRecordInput,
		) (*usecaseDto.CreateRecordOutput, error) {
			return nil, err
		},
	}
}

func TestRecordHandler_RegisterRoutes(t *testing.T) {
	r := gin.New()
	group := r.Group("/api/v0")

	h := NewRecordHandler(nil)
	h.RegisterRoutes(group)

	routes := r.Routes()

	expected := map[string]bool{
		"GET /api/v0/records":        false,
		"GET /api/v0/records/:id":    false,
		"POST /api/v0/records":       false,
		"PUT /api/v0/records/:id":    false,
		"DELETE /api/v0/records/:id": false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}

	for key, found := range expected {
		assert.True(t, found, "%s is not registered", key)
	}
}

// -------------CreateRecord(*gin.Context)のテスト-------------

func TestRecordHandler_CreateRecord(t *testing.T) {

	path := "/api/v0/records"
	gin.SetMode(gin.TestMode)

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	publicationMonth := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	publicationYear := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	datePrecisionDay := record.DatePrecisionDay
	datePrecisionMonth := record.DatePrecisionMonth
	datePrecisionYear := record.DatePrecisionYear
	readStatusReading := record.ReadStatusReading
	readStatusRead := record.ReadStatusRead

	tests := []struct {
		name string
		// ctxInternalUid 0の場合はcontextから取得できない状態を表す
		ctxInternalUid int64
		// ctxPublicUid 空文字の場合はcontextから取得できない状態を表す
		ctxPublicUid string
		// body リクエストボディ(生のJSON。json tagとbindingの挙動をそのまま検証する)
		body        string
		mockUsecase *mock.MockRecordUsecase
		// expectedInput usecaseへ渡される変換結果(nilの場合は検証しない)
		expectedInput     *usecaseDto.CreateRecordInput
		usecaseCalled     bool
		expectedBody      handlerDto.CreateRecordResponse
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_最小項目",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書のタイトル"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書のタイトル",
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_全項目指定",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
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
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeJournalArticle,
					MainTitle:       "第1章 「学校」制度の境界線",
					SubTitle:        recordStrPtr("その形成と展開"),
					PageRange:       recordStrPtr("15-46"),
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecisionDay,
					ReadStatus:      &readStatusReading,
					Contributors: []usecaseDto.ContributorInput{
						{Name: "木村元", Role: record.ContributorRoleAuthor},
						{Name: "山田太郎", Role: record.ContributorRoleInterpreter},
					},
					Memos: []usecaseDto.MemoInput{
						{Label: recordStrPtr("要約"), Content: "学校の戦後史。"},
						{Label: recordStrPtr("所感"), Content: "少し難しい。"},
					},
					Urls: []usecaseDto.UrlInput{
						{Label: recordStrPtr("オンラインPDF"), Value: "https://example.com/1"},
						{Label: recordStrPtr("書影"), Value: "https://example.com/2"},
					},
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn:          recordStrPtr("9784130513555"),
						Edition:       recordStrPtr("第2版"),
						OriginalTitle: recordStrPtr("School History"),
					},
				},
				ContainerInput: &usecaseDto.ContainerInput{
					MainTitle:  "境界線の学校史",
					SubTitle:   recordStrPtr("戦後日本の学校化社会の周辺"),
					Volume:     recordStrPtr("日本教育史講座第1巻"),
					Issue:      recordStrPtr("第11号"),
					ReadStatus: &readStatusRead,
					Contributors: []usecaseDto.ContributorInput{
						{Name: "編者A", Role: record.ContributorRoleCompiler},
					},
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn: recordStrPtr("9784130513556"),
					},
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_コンテナあり",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"journal_article","main_title":"論文","container":{"main_title":"雑誌"}}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				ContainerInput: &usecaseDto.ContainerInput{
					MainTitle: "雑誌",
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_親を持てるtypeでcontainer省略",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"journal_article","main_title":"論文"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				// コンテナ省略は許容する
				ContainerInput: nil,
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_出版日_年月日",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書","publication_date":"2022-11-30"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationDate,
					DatePrecision:   &datePrecisionDay,
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_出版日_年月",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書","publication_date":"2022-11"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationMonth,
					DatePrecision:   &datePrecisionMonth,
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_出版日_年のみ",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書","publication_date":"2022"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:            record.RecordTypeAuthoredBook,
					MainTitle:       "単著書",
					PublicationDate: &publicationYear,
					DatePrecision:   &datePrecisionYear,
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_出版日_空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書","publication_date":""}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 空文字は未入力として扱う
					PublicationDate: nil,
					DatePrecision:   nil,
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_任意スカラが空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"sub_title":"","page_range":"","publisher":"","read_status":""}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
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
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_任意スカラ未指定",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:    newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
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
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_option_valuesが全て空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"option_values":{"isbn":"","edition":"","original_title":""}}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					// 全項目が空文字なら構造体ごとnilになる
					OptionValues: nil,
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_option_valuesが一部のみ",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"option_values":{"isbn":"9784130513555"}}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "単著書",
					OptionValues: &usecaseDto.OptionValuesInput{
						Isbn:          recordStrPtr("9784130513555"),
						Edition:       nil,
						OriginalTitle: nil,
					},
				},
			},
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_メモとURLのラベルが空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"memos":[{"label":"","content":"メモの内容"}],
				"urls":[{"label":"","value":"https://example.com/1"}]}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
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
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_子要素が空配列",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[],"memos":[],"urls":[]}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
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
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:           "正常_コンテナの任意項目が空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","sub_title":"","volume":"","issue":"","read_status":""}}`,
			mockUsecase: newRecordUsecaseMock(),
			expectedInput: &usecaseDto.CreateRecordInput{
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
			usecaseCalled: true,
			expectedBody: handlerDto.CreateRecordResponse{
				PublicUserId:   recordTestPublicUid,
				PublicRecordId: recordTestRecordId,
			},
		},
		{
			name:              "異常_internalUid取得失敗",
			ctxInternalUid:    0,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_publicUid取得失敗",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      "",
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrPublicUidNotFound,
		},
		{
			name:              "異常_JSONが不正",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:           "異常_type未指定",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body:           `{"main_title":"単著書"}`,
			mockUsecase:    newRecordUsecaseMock(),
			usecaseCalled:  false,
			expectError:    true,
			// bindingタグの必須違反が専用エラーに翻訳される
			expectedErrorBody: ErrRecordTypeRequired,
		},
		{
			name:              "異常_main_title未指定",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			name:              "異常_typeが不正値",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"foo","main_title":"単著書"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRecordType,
		},
		{
			name:              "異常_親なしtypeにcontainer同送",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書","container":{"main_title":"雑誌"}}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContainerNotAllowed,
		},
		{
			name:              "異常_containerの主題が空",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"journal_article","main_title":"論文","container":{"main_title":""}}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			name:              "異常_出版日フォーマット不正",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書","publication_date":"2022/11/30"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublicationDate,
		},
		{
			name:              "異常_read_statusが不正値",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書","read_status":"foo"}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidReadStatus,
		},
		{
			name:           "異常_containerのread_statusが不正値",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","read_status":"foo"}}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidReadStatus,
		},
		{
			name:           "異常_貢献者の役割が空文字",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":""}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContributorRoleRequired,
		},
		{
			name:           "異常_貢献者の役割が不正値",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":"foo"}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidContributorRole,
		},
		{
			name:           "異常_貢献者の氏名が空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"","role":"author"}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			// validatorはフィールド宣言順(Name -> Role)で報告するため氏名側が返る
			name:           "異常_貢献者の氏名と役割が両方空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"","role":""}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			// 2件目以降の要素も検証されること
			name:           "異常_2件目の貢献者の氏名が空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":"author"},{"name":"","role":"author"}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			name:           "異常_containerの貢献者の役割が不正値",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","contributors":[{"name":"編者A","role":"foo"}]}}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidContributorRole,
		},
		{
			name:           "異常_containerの貢献者の氏名が空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","contributors":[{"name":"","role":"author"}]}}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			name:           "異常_メモの内容が空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"memos":[{"label":"要約","content":""}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrMemoContentRequired,
		},
		{
			name:           "異常_URLの値が空",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			body: `{"type":"authored_book","main_title":"単著書",
				"urls":[{"label":"PDF","value":""}]}`,
			mockUsecase:       newRecordUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrUrlValueRequired,
		},
		{
			name:              "異常_usecaseが登録上限エラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUsecaseErrMock(usecase.ErrReachRegistrationLimit),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrReachRegistrationLimit,
		},
		{
			name:              "異常_usecaseが予期しないエラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUsecaseErrMock(usecase.ErrTestUnexpected),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)
			c.Set(ctxutil.CtxKeyPublicUid, tt.ctxPublicUid)

			req := httptest.NewRequest(
				http.MethodPost,
				path,
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			h := NewRecordHandler(tt.mockUsecase)
			h.CreateRecord(c)

			// 呼び出しチェック
			assert.Equal(t, tt.usecaseCalled, tt.mockUsecase.CreateCalled)
			if tt.expectedInput != nil {
				assert.Equal(t, *tt.expectedInput, tt.mockUsecase.LastCreateInput)
			}

			// 異常系の場合のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}

			// 正常系の場合のassertion
			require.Len(t, c.Errors, 0)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Equal(t, ctxutil.REF_API_003, ctxutil.GetApiId(tt.mockUsecase.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}

// -------------DeleteRecord(*gin.Context)のテスト-------------

// recordDeleteTargetId 削除対象として指定する書誌ID
const recordDeleteTargetId = "DELETETARGET00000000000001"

// newRecordDeleteUsecaseMock 削除が成功するMockRecordUsecaseを返す
func newRecordDeleteUsecaseMock() *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		DeleteFunc: func(
			ctx context.Context,
			di usecaseDto.DeleteRecordInput,
		) error {
			return nil
		},
	}
}

// newRecordDeleteUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordDeleteUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		DeleteFunc: func(
			ctx context.Context,
			di usecaseDto.DeleteRecordInput,
		) error {
			return err
		},
	}
}

func TestRecordHandler_DeleteRecord(t *testing.T) {

	path := "/api/v0/records/" + recordDeleteTargetId
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		// ctxInternalUid 0の場合はcontextから取得できない状態を表す
		ctxInternalUid int64
		// recordId パスパラメータ。空文字の場合はパラメータ未設定を表す
		recordId    string
		mockUsecase *mock.MockRecordUsecase
		// expectedInput usecaseへ渡される入力(nilの場合は検証しない)
		expectedInput     *usecaseDto.DeleteRecordInput
		usecaseCalled     bool
		expectedBody      handlerDto.StatusResponse
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_書誌を削除",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordDeleteTargetId,
			mockUsecase:    newRecordDeleteUsecaseMock(),
			// パスパラメータと内部ユーザIDがそのままusecaseへ渡る
			expectedInput: &usecaseDto.DeleteRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordDeleteTargetId,
			},
			usecaseCalled: true,
			expectedBody:  handlerDto.StatusResponse{Status: statusOk},
		},
		{
			name:              "異常_internalUid取得失敗",
			ctxInternalUid:    0,
			recordId:          recordDeleteTargetId,
			mockUsecase:       newRecordDeleteUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_書誌IDが空",
			ctxInternalUid:    recordTestInternalUid,
			recordId:          "",
			mockUsecase:       newRecordDeleteUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrRecordIdRequired,
		},
		{
			name:              "異常_usecaseが書誌なしエラー",
			ctxInternalUid:    recordTestInternalUid,
			recordId:          recordDeleteTargetId,
			mockUsecase:       newRecordDeleteUsecaseErrMock(usecase.ErrRecordNotFound),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrRecordNotFound,
		},
		{
			name:              "異常_usecaseが予期しないエラー",
			ctxInternalUid:    recordTestInternalUid,
			recordId:          recordDeleteTargetId,
			mockUsecase:       newRecordDeleteUsecaseErrMock(usecase.ErrTestUnexpected),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)

			if tt.recordId != "" {
				c.Params = gin.Params{{Key: ParamRecordId, Value: tt.recordId}}
			}

			c.Request = httptest.NewRequest(http.MethodDelete, path, nil)

			h := NewRecordHandler(tt.mockUsecase)
			h.DeleteRecord(c)

			// 呼び出しチェック
			assert.Equal(t, tt.usecaseCalled, tt.mockUsecase.DeleteCalled)
			if tt.expectedInput != nil {
				assert.Equal(t, *tt.expectedInput, tt.mockUsecase.LastDeleteInput)
			}

			// 異常系の場合のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}

			// 正常系の場合のassertion
			require.Len(t, c.Errors, 0)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, ctxutil.REF_API_005, ctxutil.GetApiId(tt.mockUsecase.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}
