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
	"net/url"
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
			name:           "正常_コンテナを持てるtypeでcontainer省略",
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
			name:           "正常_構成要素が空配列",
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
			name:              "異常_コンテナを持てないtypeにcontainer同送",
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
		GetFunc: func(
			ctx context.Context,
			gi usecaseDto.GetRecordInput,
		) (*usecaseDto.GetRecordOutput, error) {
			return output, nil
		},
	}
}

// newRecordGetUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordGetUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		GetFunc: func(
			ctx context.Context,
			gi usecaseDto.GetRecordInput,
		) (*usecaseDto.GetRecordOutput, error) {
			return nil, err
		},
	}
}

// newRecordGetOutput 指定した書誌と収録物を返すMockRecordUsecaseを返す
func newRecordGetOutput(
	article record.UserRecordEntry,
	container *record.UserRecordEntry,
) *mock.MockRecordUsecase {
	return newRecordGetUsecaseMock(&usecaseDto.GetRecordOutput{
		Record:    article,
		Container: container,
	})
}

func TestRecordHandler_GetRecord(t *testing.T) {

	const path = "/api/v0/records/" + recordGetTargetId

	dayDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	monthDate := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	yearDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	precisionDay := record.DatePrecisionDay
	precisionMonth := record.DatePrecisionMonth
	precisionYear := record.DatePrecisionYear
	readStatusReading := record.ReadStatusReading
	readStatusRead := record.ReadStatusRead

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
		SubTitle:        recordStrPtr("その形成と展開"),
		PageRange:       recordStrPtr("15-46"),
		Publisher:       recordStrPtr("東京大学出版会"),
		PublicationDate: &dayDate,
		DatePrecision:   &precisionDay,
		ReadStatus:      &readStatusReading,
		ContainerId:     recordStrPtr(recordGetContainerId),
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
			{Title: recordStrPtr("要約"), Content: "学校の戦後史。"},
			{Title: nil, Content: "見出しなしのメモ"},
		},
		Urls: []record.UrlEntry{
			{Label: recordStrPtr("オンラインPDF"), Url: "https://example.com/pdf"},
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
		SubTitle:      recordStrPtr("収録物の副題"),
		Volume:        recordStrPtr("第1巻"),
		Issue:         recordStrPtr("第11号"),
		ReadStatus:    &readStatusRead,
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
	emptyStrRecord.SubTitle = recordStrPtr("")
	emptyStrRecord.PageRange = recordStrPtr("")
	emptyStrRecord.Publisher = recordStrPtr("")
	emptyStrRecord.Memos = []record.MemoEntry{{Title: recordStrPtr(""), Content: "メモ内容"}}
	emptyStrRecord.Urls = []record.UrlEntry{{Label: recordStrPtr(""), Url: "https://example.com/1"}}

	emptyStrContainer := record.UserRecordEntry{
		RecordEntryId: recordGetContainerId,
		RecordType:    record.RecordTypeJournal,
		MainTitle:     "収録物のタイトル",
		SubTitle:      recordStrPtr(""),
		Volume:        recordStrPtr(""),
		Issue:         recordStrPtr(""),
	}

	tests := []struct {
		name string
		// ctxInternalUid 0の場合はcontextから取得できない状態を表す
		ctxInternalUid int64
		// recordId パスパラメータ。空文字の場合はパラメータ未設定を表す
		recordId    string
		mockUsecase *mock.MockRecordUsecase
		// expectedInput usecaseへ渡される入力(nilの場合は検証しない)
		expectedInput *usecaseDto.GetRecordInput
		usecaseCalled bool
		expectedBody  handlerDto.GetRecordResponse
		// notContains レスポンスのJSONに含まれてはいけない文字列
		notContains       []string
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_全項目を含む書誌を取得",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(fullRecord, &fullContainer),
			expectedInput: &usecaseDto.GetRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordGetTargetId,
			},
			usecaseCalled: true,
			expectedBody: handlerDto.GetRecordResponse{
				RecordId: recordGetTargetId,
				RecordDetailItem: handlerDto.RecordDetailItem{
					Type:            "journal_article",
					MainTitle:       "第1章 「学校」制度の境界線",
					SubTitle:        recordStrPtr("その形成と展開"),
					PageRange:       recordStrPtr("15-46"),
					Publisher:       recordStrPtr("東京大学出版会"),
					PublicationDate: recordStrPtr("2022-11-30"),
					ReadStatus:      recordStrPtr("reading"),
					Contributors: []handlerDto.ContributorInfo{
						{Name: "木村元", Role: "author"},
						{Name: "訳者B", Role: "interpreter"},
					},
					Memos: []handlerDto.MemoInfo{
						{Label: recordStrPtr("要約"), Content: "学校の戦後史。"},
						{Label: nil, Content: "見出しなしのメモ"},
					},
					Urls: []handlerDto.UrlInfo{
						{Label: recordStrPtr("オンラインPDF"), Value: "https://example.com/pdf"},
						{Label: nil, Value: "https://example.com/info"},
					},
					OptionValues: &handlerDto.OptionValues{
						Isbn:          recordStrPtr("9784130513555"),
						Edition:       recordStrPtr("第2版"),
						OriginalTitle: recordStrPtr("School History of Boundary"),
					},
				},
				ContainerDetailItem: &handlerDto.ContainerDetailItem{
					MainTitle:  "境界線の学校史",
					SubTitle:   recordStrPtr("収録物の副題"),
					Volume:     recordStrPtr("第1巻"),
					Issue:      recordStrPtr("第11号"),
					ReadStatus: recordStrPtr("read"),
					Contributors: []handlerDto.ContributorInfo{
						{Name: "編者A", Role: "compiler"},
					},
					OptionValues: &handlerDto.OptionValues{
						Isbn: recordStrPtr("9784130519999"),
					},
				},
			},
			// 収録物のIDは削除・参照のガードの前提となるため露出させない
			notContains: []string{recordGetContainerId},
		},
		{
			name:           "正常_収録物なし",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(minimalRecord(), nil),
			usecaseCalled:  true,
			expectedBody:   minimalResponse(),
		},
		{
			name:           "正常_任意項目がnil",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(minimalRecord(), nil),
			usecaseCalled:  true,
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
			name:           "正常_空文字はnullになる",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(emptyStrRecord, &emptyStrContainer),
			usecaseCalled:  true,
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
			name:           "正常_出版日が日精度",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(withPublicationDate(&dayDate, &precisionDay), nil),
			usecaseCalled:  true,
			expectedBody: func() handlerDto.GetRecordResponse {
				res := minimalResponse()
				res.PublicationDate = recordStrPtr("2022-11-30")
				return res
			}(),
		},
		{
			name:           "正常_出版日が月精度",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(withPublicationDate(&monthDate, &precisionMonth), nil),
			usecaseCalled:  true,
			expectedBody: func() handlerDto.GetRecordResponse {
				res := minimalResponse()
				res.PublicationDate = recordStrPtr("2022-11")
				return res
			}(),
		},
		{
			name:           "正常_出版日が年精度",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(withPublicationDate(&yearDate, &precisionYear), nil),
			usecaseCalled:  true,
			expectedBody: func() handlerDto.GetRecordResponse {
				res := minimalResponse()
				res.PublicationDate = recordStrPtr("2022")
				return res
			}(),
		},
		{
			// 精度が不明なまま日付を出すと月日を捏造することになるためnullを返す
			name:           "正常_出版日の精度がnil",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetOutput(withPublicationDate(&dayDate, nil), nil),
			usecaseCalled:  true,
			expectedBody:   minimalResponse(),
		},
		{
			name:           "正常_項目値が一部のみ",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase: newRecordGetOutput(withOptionValues(
				record.OptionValue{RecordField: record.RecordFieldIsbn, RawValue: "9784130513555"},
			), nil),
			usecaseCalled: true,
			expectedBody: func() handlerDto.GetRecordResponse {
				res := minimalResponse()
				res.OptionValues = &handlerDto.OptionValues{
					Isbn:          recordStrPtr("9784130513555"),
					Edition:       nil,
					OriginalTitle: nil,
				}
				return res
			}(),
		},
		{
			// 空文字の項目値はスキップされ、結果が空になるためoption_values自体がnullになる
			name:           "正常_項目値が空文字のみならnull",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase: newRecordGetOutput(withOptionValues(
				record.OptionValue{RecordField: record.RecordFieldIsbn, RawValue: ""},
			), nil),
			usecaseCalled: true,
			expectedBody:  minimalResponse(),
		},
		{
			name:           "異常_internalUid取得失敗",
			ctxInternalUid: 0,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:  false,
			expectError:    true,
			// usecaseを呼ばずにエラーとなる
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:           "異常_書誌IDが空",
			ctxInternalUid: recordTestInternalUid,
			recordId:       "",
			mockUsecase:    newRecordGetUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:  false,
			expectError:    true,
			// usecaseを呼ばずにエラーとなる
			expectedErrorBody: ErrRecordIdRequired,
		},
		{
			name:           "異常_usecaseが書誌なしエラー",
			ctxInternalUid: recordTestInternalUid,
			recordId:       recordGetTargetId,
			mockUsecase:    newRecordGetUsecaseErrMock(usecase.ErrRecordNotFound),
			usecaseCalled:  true,
			expectError:    true,
			// usecaseのエラーは包み直さずそのまま伝播する
			expectedErrorBody: usecase.ErrRecordNotFound,
		},
		{
			name:              "異常_usecaseが予期しないエラー",
			ctxInternalUid:    recordTestInternalUid,
			recordId:          recordGetTargetId,
			mockUsecase:       newRecordGetUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: ErrTestUnexpected,
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

			c.Request = httptest.NewRequest(http.MethodGet, path, nil)

			h := NewRecordHandler(tt.mockUsecase)
			h.GetRecord(c)

			// 呼び出しチェック
			assert.Equal(t, tt.usecaseCalled, tt.mockUsecase.GetCalled)
			if tt.expectedInput != nil {
				assert.Equal(t, *tt.expectedInput, tt.mockUsecase.LastGetInput)
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
			assert.Equal(t, ctxutil.REF_API_002, ctxutil.GetApiId(tt.mockUsecase.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)

			// 露出してはいけない値が含まれないこと
			for _, ng := range tt.notContains {
				assert.NotContains(t, rec.Body.String(), ng)
			}
		})
	}
}

// -------------UpdateRecord(*gin.Context)のテスト-------------

// recordUpdateTargetId 更新対象として指定する書誌ID
const recordUpdateTargetId = "UPDATETARGET00000000000001"

// newRecordUpdateUsecaseMock 更新が成功するMockRecordUsecaseを返す
func newRecordUpdateUsecaseMock() *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		UpdateFunc: func(
			ctx context.Context,
			ui usecaseDto.UpdateRecordInput,
		) error {
			return nil
		},
	}
}

// newRecordUpdateUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordUpdateUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		UpdateFunc: func(
			ctx context.Context,
			ui usecaseDto.UpdateRecordInput,
		) error {
			return err
		},
	}
}

// UpdateRecordのテスト方針
//
// リクエストDTOはCreateと同一構造で、変換も toRecordInput / toContainerInput /
// toReadStatus / toContributorsInput / nilIfEmpty / GetDateAndPrecision を
// CreateRecordと共用している。このため出版日の精度・任意スカラの空文字と未指定・
// option_valuesの各パターン・貢献者やメモURLの必須エラーといった変換の網羅は
// TestRecordHandler_CreateRecord 側(37ケース)で担保する。
//
// ここで確認するのはUpdate固有の差分に絞る。
//   - パスパラメータ id を RecordId へ渡すこと(Createにはない)
//   - publicUid を参照しないこと(Createとの差分)
//   - レスポンスが 200 + StatusResponse{"OK"} であること(Createは201)
//   - API IDが REF-API-004 であること
//   - chkUpdateParam の2つの検証
//   - toUpdateRecordInput が持つ3つのエラー返却経路が握りつぶされないこと
//   - usecaseのエラーが変換されずそのまま c.Errors に乗ること
func TestRecordHandler_UpdateRecord(t *testing.T) {

	path := "/api/v0/records/" + recordUpdateTargetId
	gin.SetMode(gin.TestMode)

	publicationDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	datePrecisionDay := record.DatePrecisionDay
	readStatusReading := record.ReadStatusReading
	readStatusRead := record.ReadStatusRead

	tests := []struct {
		name string
		// ctxInternalUid 0の場合はcontextから取得できない状態を表す
		ctxInternalUid int64
		// ctxPublicUid 空文字の場合はcontextへ設定しない(UpdateRecordは参照しないため)
		ctxPublicUid string
		// recordId パスパラメータ。空文字の場合はパラメータ未設定を表す
		recordId string
		// body リクエストボディ(生のJSON。json tagとbindingの挙動をそのまま検証する)
		body        string
		mockUsecase *mock.MockRecordUsecase
		// expectedInput usecaseへ渡される変換結果(nilの場合は検証しない)
		expectedInput     *usecaseDto.UpdateRecordInput
		usecaseCalled     bool
		expectedBody      handlerDto.StatusResponse
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:           "正常_最小項目",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			recordId:       recordUpdateTargetId,
			body:           `{"type":"authored_book","main_title":"更新後の単著書"}`,
			mockUsecase:    newRecordUpdateUsecaseMock(),
			// パスパラメータがRecordIdへ、contextの値がInternalUserIdへ渡る
			expectedInput: &usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeAuthoredBook,
					MainTitle: "更新後の単著書",
				},
			},
			usecaseCalled: true,
			expectedBody:  handlerDto.StatusResponse{Status: statusOk},
		},
		{
			// 共有している変換のスモーク。各項目の網羅はCreateのテストで担保する
			name:           "正常_全項目指定",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			recordId:       recordUpdateTargetId,
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
			mockUsecase: newRecordUpdateUsecaseMock(),
			expectedInput: &usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
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
						{Content: "ラベルなしのメモ"},
					},
					Urls: []usecaseDto.UrlInput{
						{Label: recordStrPtr("オンラインPDF"), Value: "https://example.com/1"},
						{Value: "https://example.com/2"},
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
			expectedBody:  handlerDto.StatusResponse{Status: statusOk},
		},
		{
			name:           "正常_コンテナあり",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			recordId:       recordUpdateTargetId,
			body:           `{"type":"journal_article","main_title":"論文","container":{"main_title":"雑誌"}}`,
			mockUsecase:    newRecordUpdateUsecaseMock(),
			expectedInput: &usecaseDto.UpdateRecordInput{
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
			usecaseCalled: true,
			expectedBody:  handlerDto.StatusResponse{Status: statusOk},
		},
		{
			// containerを省略した場合はContainerInputがnilになり、
			// usecase側で「収録物を外す」経路に入る
			name:           "正常_コンテナを持てるtypeでcontainer省略",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			recordId:       recordUpdateTargetId,
			body:           `{"type":"journal_article","main_title":"論文"}`,
			mockUsecase:    newRecordUpdateUsecaseMock(),
			expectedInput: &usecaseDto.UpdateRecordInput{
				InternalUserId: recordTestInternalUid,
				RecordId:       recordUpdateTargetId,
				RecordInput: usecaseDto.RecordInput{
					Type:      record.RecordTypeJournalArticle,
					MainTitle: "論文",
				},
				ContainerInput: nil,
			},
			usecaseCalled: true,
			expectedBody:  handlerDto.StatusResponse{Status: statusOk},
		},
		{
			// UpdateRecordはpublicUidを参照しない(レスポンスがStatusResponseのみのため)。
			// 将来publicUidを参照する変更が入った場合にここで気づける。
			name:           "正常_publicUidが未設定でも成功",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   "",
			recordId:       recordUpdateTargetId,
			body:           `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:    newRecordUpdateUsecaseMock(),
			usecaseCalled:  true,
			expectedBody:   handlerDto.StatusResponse{Status: statusOk},
		},
		{
			name:              "異常_internalUid取得失敗",
			ctxInternalUid:    0,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_書誌IDが空",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          "",
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrRecordIdRequired,
		},
		{
			name:              "異常_JSONが不正",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			// updateRecordRequiredErrorsが対応表として使われていることの確認
			name:              "異常_type未指定",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrRecordTypeRequired,
		},
		{
			name:              "異常_main_title未指定",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			// chkUpdateParamの1つ目の検証(IsValid)
			name:              "異常_typeが不正値",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"unknown_type","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRecordType,
		},
		{
			// chkUpdateParamの2つ目の検証(IsArticle)
			name:              "異常_コンテナを持てないtypeにcontainer同送",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書","container":{"main_title":"雑誌"}}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrContainerNotAllowed,
		},
		{
			// toUpdateRecordInputがtoRecordInputのエラーを握りつぶさないことの確認。
			// 出版日以外(read_status・貢献者)の網羅はCreateのテストで担保する。
			name:              "異常_出版日フォーマット不正",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書","publication_date":"2022/11/30"}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublicationDate,
		},
		{
			// toUpdateRecordInputがtoContainerInputのエラーを握りつぶさないことの確認
			name:           "異常_containerのread_statusが不正値",
			ctxInternalUid: recordTestInternalUid,
			ctxPublicUid:   recordTestPublicUid,
			recordId:       recordUpdateTargetId,
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","read_status":"unknown"}}`,
			mockUsecase:       newRecordUpdateUsecaseMock(),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidReadStatus,
		},
		{
			name:              "異常_usecaseが書誌なしエラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseErrMock(usecase.ErrRecordNotFound),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrRecordNotFound,
		},
		{
			// REF-API-004で新設した書誌タイプ変更の拒否(E0025)
			name:              "異常_usecaseが書誌タイプ変更エラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseErrMock(usecase.ErrRecordTypeNotChangeable),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrRecordTypeNotChangeable,
		},
		{
			// 収録物のデータ不整合(500)。handlerは変換せず素通しする
			name:              "異常_usecaseが収録物不整合エラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"journal_article","main_title":"論文","container":{"main_title":"雑誌"}}`,
			mockUsecase:       newRecordUpdateUsecaseErrMock(usecase.ErrContainerNotFound),
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: usecase.ErrContainerNotFound,
		},
		{
			name:              "異常_usecaseが予期しないエラー",
			ctxInternalUid:    recordTestInternalUid,
			ctxPublicUid:      recordTestPublicUid,
			recordId:          recordUpdateTargetId,
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			mockUsecase:       newRecordUpdateUsecaseErrMock(usecase.ErrTestUnexpected),
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

			// UpdateRecordはpublicUidを参照しないため、未設定のケースも用意する
			if tt.ctxPublicUid != "" {
				c.Set(ctxutil.CtxKeyPublicUid, tt.ctxPublicUid)
			}

			if tt.recordId != "" {
				c.Params = gin.Params{{Key: ParamRecordId, Value: tt.recordId}}
			}

			req := httptest.NewRequest(
				http.MethodPut,
				path,
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			h := NewRecordHandler(tt.mockUsecase)
			h.UpdateRecord(c)

			// 呼び出しチェック
			assert.Equal(t, tt.usecaseCalled, tt.mockUsecase.UpdateCalled)
			if tt.expectedInput != nil {
				assert.Equal(t, *tt.expectedInput, tt.mockUsecase.LastUpdateInput)
			}

			// 異常系の場合のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
				return
			}

			// 正常系の場合のassertion
			require.Len(t, c.Errors, 0)
			// Createは201だが、更新はIDを返さないため200
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, ctxutil.REF_API_004, ctxutil.GetApiId(tt.mockUsecase.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)
		})
	}
}

// -------------GetRecords(*gin.Context)のテスト-------------

const (
	// recordListArticleId1〜3 usecaseが返す書誌のID
	recordListArticleId1 = "LISTARTICLE000000000000001"
	recordListArticleId2 = "LISTARTICLE000000000000002"
	recordListArticleId3 = "LISTARTICLE000000000000003"
	// recordListContainerId 収録物の書誌ID。レスポンスには露出しない。
	recordListContainerId = "LISTCONTAINER0000000000001"
)

// recordIntPtr int型のポインタを返却する
func recordIntPtr(i int) *int {
	return &i
}

// newRecordListUsecaseMock 指定したOutputを返すMockRecordUsecaseを返す
func newRecordListUsecaseMock(output *usecaseDto.GetRecordsOutput) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		ListFunc: func(
			ctx context.Context,
			li usecaseDto.GetRecordsInput,
		) (*usecaseDto.GetRecordsOutput, error) {
			return output, nil
		},
	}
}

// newRecordListUsecaseErrMock 指定したエラーを返すMockRecordUsecaseを返す
func newRecordListUsecaseErrMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		ListFunc: func(
			ctx context.Context,
			li usecaseDto.GetRecordsInput,
		) (*usecaseDto.GetRecordsOutput, error) {
			return nil, err
		},
	}
}

func TestRecordHandler_GetRecords(t *testing.T) {

	const basePath = "/api/v0/records"

	dayDate := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	monthDate := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	yearDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	precisionDay := record.DatePrecisionDay
	precisionMonth := record.DatePrecisionMonth
	precisionYear := record.DatePrecisionYear
	readStatusReading := record.ReadStatusReading

	// withInput デフォルト値の入力を書き換えたusecase入力を返す
	// パラメータ未指定時は publish_year / asc / 20 / 0 になる。
	withInput := func(modify func(in *usecaseDto.GetRecordsInput)) *usecaseDto.GetRecordsInput {
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
		return &in
	}
	defaultInput := withInput(nil)

	// newOutput 指定した総件数と書誌からusecaseの出力を組み立てる
	newOutput := func(total int, items ...usecaseDto.RecordListItem) *usecaseDto.GetRecordsOutput {
		return &usecaseDto.GetRecordsOutput{TotalCount: total, Records: items}
	}

	// newResponse 指定した件数と書誌からレスポンスを組み立てる(書誌なしは空配列)
	newResponse := func(count int, items ...handlerDto.RecordListItemResponse) handlerDto.GetRecordsResponse {
		records := []handlerDto.RecordListItemResponse{}
		records = append(records, items...)
		return handlerDto.GetRecordsResponse{Count: count, Records: &records}
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
		return newResponse(1, item)
	}

	// 全項目と貢献者を持つ、収録物ありの書誌
	fullItem := usecaseDto.RecordListItem{
		Record: record.UserRecordEntry{
			RecordEntryId:   recordListArticleId1,
			InternalUid:     recordTestInternalUid,
			RecordType:      record.RecordTypeJournalArticle,
			MainTitle:       "第1章 「学校」制度の境界線",
			SubTitle:        recordStrPtr("その形成と展開"),
			PageRange:       recordStrPtr("15-46"),
			Publisher:       recordStrPtr("東京大学出版会"),
			PublicationDate: &dayDate,
			DatePrecision:   &precisionDay,
			ReadStatus:      &readStatusReading,
			ContainerId:     recordStrPtr(recordListContainerId),
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
			SubTitle:      recordStrPtr("収録物の副題"),
			Volume:        recordStrPtr("第1巻"),
			Issue:         recordStrPtr("第11号"),
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
		SubTitle:        recordStrPtr("その形成と展開"),
		PageRange:       recordStrPtr("15-46"),
		Publisher:       recordStrPtr("東京大学出版会"),
		PublicationDate: recordStrPtr("2022-11-30"),
		ReadStatus:      recordStrPtr("reading"),
		Contributors: []handlerDto.ContributorInfo{
			{Name: "木村元", Role: "author"},
			{Name: "訳者B", Role: "interpreter"},
		},
		ContainerListItem: &handlerDto.ContainerListItemResponse{
			MainTitle: "境界線の学校史",
			SubTitle:  recordStrPtr("収録物の副題"),
			Volume:    recordStrPtr("第1巻"),
			Issue:     recordStrPtr("第11号"),
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
			SubTitle:      recordStrPtr(""),
			PageRange:     recordStrPtr(""),
			Publisher:     recordStrPtr(""),
			ContainerId:   recordStrPtr(recordListContainerId),
		},
		Container: &record.UserRecordEntry{
			RecordEntryId: recordListContainerId,
			RecordType:    record.RecordTypeJournal,
			MainTitle:     "雑誌のタイトル",
			SubTitle:      recordStrPtr(""),
			Volume:        recordStrPtr(""),
			Issue:         recordStrPtr(""),
		},
	}

	tests := []struct {
		name string
		// ctxInternalUid 0の場合はcontextから取得できない状態を表す
		ctxInternalUid int64
		// query URLのクエリ文字列("?"は含めない)
		query       string
		mockUsecase *mock.MockRecordUsecase
		// expectedInput usecaseへ渡される入力(nilの場合は検証しない)
		expectedInput *usecaseDto.GetRecordsInput
		usecaseCalled bool
		expectedBody  handlerDto.GetRecordsResponse
		// bodyContains / notContains レスポンスのJSONに含まれる/含まれてはいけない文字列
		bodyContains      []string
		notContains       []string
		expectError       bool
		expectedErrorBody error
	}{
		// ---- 入力の変換 ----
		{
			name:           "正常_パラメータ未指定はデフォルト値",
			ctxInternalUid: recordTestInternalUid,
			query:          "",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(0),
		},
		{
			name:           "正常_全パラメータを指定",
			ctxInternalUid: recordTestInternalUid,
			query: "contributor_name=" + url.QueryEscape("木村") +
				"&title=" + url.QueryEscape("学校") +
				"&publish_year=2022&sort_key=insert_date&sort_type=desc&limit=50&offset=20",
			mockUsecase: newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.ContributorName = recordStrPtr("木村")
				in.Title = recordStrPtr("学校")
				in.PublishYear = recordIntPtr(2022)
				in.SortKey = record.SortKeyInsertDate
				in.SortOrder = record.SortOrderDesc
				in.Limit = 50
				in.Offset = 20
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			name:           "正常_空文字の検索条件は未指定扱い",
			ctxInternalUid: recordTestInternalUid,
			query:          "contributor_name=&title=&publish_year=&sort_key=&sort_type=",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(0),
		},
		{
			name:           "正常_検索語の前後空白を除去",
			ctxInternalUid: recordTestInternalUid,
			// 貢献者名は半角空白、タイトルは全角空白で囲む
			query: "contributor_name=" + url.QueryEscape(" 木村 ") +
				"&title=" + url.QueryEscape("　学校　"),
			mockUsecase: newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.ContributorName = recordStrPtr("木村")
				in.Title = recordStrPtr("学校")
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			name:           "正常_空白のみの検索語は未指定扱い",
			ctxInternalUid: recordTestInternalUid,
			query: "contributor_name=" + url.QueryEscape("  ") +
				"&title=" + url.QueryEscape(" 　"),
			mockUsecase:   newRecordListUsecaseMock(newOutput(0)),
			expectedInput: defaultInput,
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			name:           "正常_語中の空白は残す",
			ctxInternalUid: recordTestInternalUid,
			query:          "title=" + url.QueryEscape("学校 教育"),
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.Title = recordStrPtr("学校 教育")
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			name:           "正常_limitの下限",
			ctxInternalUid: recordTestInternalUid,
			query:          "limit=1",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.Limit = 1
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			name:           "正常_limitの上限",
			ctxInternalUid: recordTestInternalUid,
			query:          "limit=100",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.Limit = 100
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},
		{
			// 明示した0はデフォルト値の0と区別されず、そのまま渡る
			name:           "正常_offsetが0",
			ctxInternalUid: recordTestInternalUid,
			query:          "offset=0",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(0),
		},
		{
			name:           "正常_sort_keyとsort_typeの各値",
			ctxInternalUid: recordTestInternalUid,
			query:          "sort_key=publish_year&sort_type=desc",
			mockUsecase:    newRecordListUsecaseMock(newOutput(0)),
			expectedInput: withInput(func(in *usecaseDto.GetRecordsInput) {
				in.SortKey = record.SortKeyPublishYear
				in.SortOrder = record.SortOrderDesc
			}),
			usecaseCalled: true,
			expectedBody:  newResponse(0),
		},

		// ---- レスポンスの組み立て ----
		{
			name:           "正常_全項目を含む書誌一覧",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase: newRecordListUsecaseMock(newOutput(2,
				fullItem,
				minimalItem(recordListArticleId2),
			)),
			expectedInput: defaultInput,
			usecaseCalled: true,
			expectedBody: newResponse(2,
				fullResponseItem,
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			name:           "正常_収録物なしはcontainerがnull",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, minimalItem(recordListArticleId1))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(1, minimalResponseItem(recordListArticleId1)),
			bodyContains:   []string{`"container":null`},
		},
		{
			name:           "正常_任意項目がnil",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, minimalItem(recordListArticleId1))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody: newResponse(1, handlerDto.RecordListItemResponse{
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
			name:           "正常_任意項目の空文字はnull",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, emptyStrItem)),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody: newResponse(1, handlerDto.RecordListItemResponse{
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
			name:           "正常_出版日が日精度",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, withPublicationDate(&dayDate, &precisionDay))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   withPublicationDateResponse(recordStrPtr("2022-11-30")),
		},
		{
			name:           "正常_出版日が月精度",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, withPublicationDate(&monthDate, &precisionMonth))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   withPublicationDateResponse(recordStrPtr("2022-11")),
		},
		{
			name:           "正常_出版日が年精度",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, withPublicationDate(&yearDate, &precisionYear))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   withPublicationDateResponse(recordStrPtr("2022")),
		},
		{
			// 精度が不明なまま日付を出すと月日を捏造することになるためnullを返す
			name:           "正常_出版日の精度がnil",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, withPublicationDate(&dayDate, nil))),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   withPublicationDateResponse(nil),
		},
		{
			// countはrecordsの件数ではなく、limit/offset適用前の総件数
			name:           "正常_countは総件数",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase: newRecordListUsecaseMock(newOutput(25,
				minimalItem(recordListArticleId1),
				minimalItem(recordListArticleId2),
			)),
			expectedInput: defaultInput,
			usecaseCalled: true,
			expectedBody: newResponse(25,
				minimalResponseItem(recordListArticleId1),
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			// usecaseは該当なしの場合にRecordsをnilで返すが、レスポンスはnullにしない
			name:           "正常_0件はrecordsが空配列",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(&usecaseDto.GetRecordsOutput{TotalCount: 0, Records: nil}),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(0),
			bodyContains:   []string{`"records":[]`},
		},
		{
			name:           "正常_一覧の順序を保つ",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase: newRecordListUsecaseMock(newOutput(3,
				minimalItem(recordListArticleId3),
				minimalItem(recordListArticleId1),
				minimalItem(recordListArticleId2),
			)),
			expectedInput: defaultInput,
			usecaseCalled: true,
			expectedBody: newResponse(3,
				minimalResponseItem(recordListArticleId3),
				minimalResponseItem(recordListArticleId1),
				minimalResponseItem(recordListArticleId2),
			),
		},
		{
			// 収録物のIDは参照・削除のガードの前提となるため露出させない
			name:           "正常_収録物IDを露出しない",
			ctxInternalUid: recordTestInternalUid,
			mockUsecase:    newRecordListUsecaseMock(newOutput(1, fullItem)),
			expectedInput:  defaultInput,
			usecaseCalled:  true,
			expectedBody:   newResponse(1, fullResponseItem),
			notContains:    []string{recordListContainerId},
		},

		// ---- 異常系(usecaseを呼ばずにエラーとなる) ----
		{
			name:              "異常_internalUid取得失敗",
			ctxInternalUid:    0,
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_limitが数値でない",
			ctxInternalUid:    recordTestInternalUid,
			query:             "limit=abc",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:              "異常_offsetが数値でない",
			ctxInternalUid:    recordTestInternalUid,
			query:             "offset=abc",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name:              "異常_publish_yearが数字以外を含む",
			ctxInternalUid:    recordTestInternalUid,
			query:             "publish_year=20xx",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_publish_yearが3桁",
			ctxInternalUid:    recordTestInternalUid,
			query:             "publish_year=202",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_publish_yearが5桁",
			ctxInternalUid:    recordTestInternalUid,
			query:             "publish_year=20201",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			// \d は半角数字にのみ一致する
			name:              "異常_publish_yearが全角数字",
			ctxInternalUid:    recordTestInternalUid,
			query:             "publish_year=" + url.QueryEscape("２０２２"),
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublishYear,
		},
		{
			name:              "異常_sort_keyが不正",
			ctxInternalUid:    recordTestInternalUid,
			query:             "sort_key=title",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidSortKey,
		},
		{
			// 大文字小文字は区別する
			name:              "異常_sort_typeが不正",
			ctxInternalUid:    recordTestInternalUid,
			query:             "sort_type=ASC",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidSortOrder,
		},
		{
			name:              "異常_limitが0",
			ctxInternalUid:    recordTestInternalUid,
			query:             "limit=0",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_limitが上限超過",
			ctxInternalUid:    recordTestInternalUid,
			query:             "limit=101",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_limitが負",
			ctxInternalUid:    recordTestInternalUid,
			query:             "limit=-1",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			// キーがあるためginのbindで0になり、未指定(デフォルト値)にはならない
			name:              "異常_limitが空文字",
			ctxInternalUid:    recordTestInternalUid,
			query:             "limit=",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidLimit,
		},
		{
			name:              "異常_offsetが負",
			ctxInternalUid:    recordTestInternalUid,
			query:             "offset=-1",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidOffset,
		},
		{
			// DTOの宣言順(publish_year → sort_key)で先の項目のエラーを返す
			name:              "異常_複数項目が不正なら宣言順で先の項目",
			ctxInternalUid:    recordTestInternalUid,
			query:             "sort_key=title&publish_year=20xx",
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			usecaseCalled:     false,
			expectError:       true,
			expectedErrorBody: ErrInvalidPublishYear,
		},

		// ---- 異常系(usecaseのエラー) ----
		{
			name:              "異常_usecaseが予期しないエラー",
			ctxInternalUid:    recordTestInternalUid,
			mockUsecase:       newRecordListUsecaseErrMock(ErrTestUnexpected),
			expectedInput:     defaultInput,
			usecaseCalled:     true,
			expectError:       true,
			expectedErrorBody: ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Set(ctxutil.CtxKeyUid, tt.ctxInternalUid)
			c.Request = httptest.NewRequest(http.MethodGet, basePath+"?"+tt.query, nil)

			h := NewRecordHandler(tt.mockUsecase)
			h.GetRecords(c)

			// 呼び出しチェック
			assert.Equal(t, tt.usecaseCalled, tt.mockUsecase.ListCalled)
			if tt.expectedInput != nil {
				assert.Equal(t, *tt.expectedInput, tt.mockUsecase.LastListInput)
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
			assert.Equal(t, ctxutil.REF_API_001, ctxutil.GetApiId(tt.mockUsecase.LastCtx))
			testutil.AssertResponse(t, rec, tt.expectedBody)

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
