package handler

import (
	"testing"

	"jemref_go/internal/usecase"

	"github.com/stretchr/testify/assert"
)

// -------------CreateRecord(*gin.Context)のテスト(異常系)-------------

// TestRecordHandler_CreateRecord_異常 contextの値やリクエストボディが不正なら、usecaseを呼ばずにエラーを返すこと
func TestRecordHandler_CreateRecord_異常(t *testing.T) {

	tests := []struct {
		name              string
		ctx               recordCtx
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
			name:              "異常_publicUid取得失敗",
			ctx:               recordCtx{omitPublicUid: true},
			body:              `{"type":"authored_book","main_title":"単著書"}`,
			expectedErrorBody: ErrPublicUidNotFound,
		},
		{
			name:              "異常_JSONが不正",
			body:              `{"type":`,
			expectedErrorBody: ErrInvalidRequestBody,
		},
		{
			name: "異常_type未指定",
			body: `{"main_title":"単著書"}`,
			// bindingタグの必須違反が専用エラーに翻訳される
			expectedErrorBody: ErrRecordTypeRequired,
		},
		{
			name:              "異常_main_title未指定",
			body:              `{"type":"authored_book"}`,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			name:              "異常_typeが不正値",
			body:              `{"type":"foo","main_title":"単著書"}`,
			expectedErrorBody: ErrInvalidRecordType,
		},
		{
			name:              "異常_コンテナを持てないtypeにcontainer同送",
			body:              `{"type":"authored_book","main_title":"単著書","container":{"main_title":"雑誌"}}`,
			expectedErrorBody: ErrContainerNotAllowed,
		},
		{
			name:              "異常_containerの主題が空",
			body:              `{"type":"journal_article","main_title":"論文","container":{"main_title":""}}`,
			expectedErrorBody: ErrMainTitleRequired,
		},
		{
			name:              "異常_出版日フォーマット不正",
			body:              `{"type":"authored_book","main_title":"単著書","publication_date":"2022/11/30"}`,
			expectedErrorBody: ErrInvalidPublicationDate,
		},
		{
			name:              "異常_read_statusが不正値",
			body:              `{"type":"authored_book","main_title":"単著書","read_status":"foo"}`,
			expectedErrorBody: ErrInvalidReadStatus,
		},
		{
			name: "異常_containerのread_statusが不正値",
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","read_status":"foo"}}`,
			expectedErrorBody: ErrInvalidReadStatus,
		},
		{
			name: "異常_貢献者の役割が空文字",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":""}]}`,
			expectedErrorBody: ErrContributorRoleRequired,
		},
		{
			name: "異常_貢献者の役割が不正値",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":"foo"}]}`,
			expectedErrorBody: ErrInvalidContributorRole,
		},
		{
			name: "異常_貢献者の氏名が空",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"","role":"author"}]}`,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			// validatorはフィールド宣言順(Name -> Role)で報告するため氏名側が返る
			name: "異常_貢献者の氏名と役割が両方空",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"","role":""}]}`,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			// 2件目以降の要素も検証されること
			name: "異常_2件目の貢献者の氏名が空",
			body: `{"type":"authored_book","main_title":"単著書",
				"contributors":[{"name":"木村元","role":"author"},{"name":"","role":"author"}]}`,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			name: "異常_containerの貢献者の役割が不正値",
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","contributors":[{"name":"編者A","role":"foo"}]}}`,
			expectedErrorBody: ErrInvalidContributorRole,
		},
		{
			name: "異常_containerの貢献者の氏名が空",
			body: `{"type":"journal_article","main_title":"論文",
				"container":{"main_title":"雑誌","contributors":[{"name":"","role":"author"}]}}`,
			expectedErrorBody: ErrContributorNameRequired,
		},
		{
			name: "異常_メモの内容が空",
			body: `{"type":"authored_book","main_title":"単著書",
				"memos":[{"label":"要約","content":""}]}`,
			expectedErrorBody: ErrMemoContentRequired,
		},
		{
			name: "異常_URLの値が空",
			body: `{"type":"authored_book","main_title":"単著書",
				"urls":[{"label":"PDF","value":""}]}`,
			expectedErrorBody: ErrUrlValueRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordCreateUsecaseMock(nil)

			_, c := serveCreateRecord(uc, tt.body, tt.ctx)

			assert.False(t, uc.CreateCalled)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestRecordHandler_CreateRecord_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestRecordHandler_CreateRecord_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		usecaseErr error
	}{
		{
			name:       "異常_usecaseが登録上限エラー",
			usecaseErr: usecase.ErrReachRegistrationLimit,
		},
		{
			name:       "異常_usecaseが予期しないエラー",
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordCreateUsecaseMock(tt.usecaseErr)

			_, c := serveCreateRecord(uc, `{"type":"authored_book","main_title":"単著書"}`, recordCtx{})

			assert.True(t, uc.CreateCalled)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
