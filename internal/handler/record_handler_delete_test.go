package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------DeleteRecord(*gin.Context)のテスト-------------

// recordDeleteTargetId 削除対象として指定する書誌ID
const recordDeleteTargetId = "DELETETARGET00000000000001"

// newRecordDeleteUsecaseMock 指定したエラーを返すMockRecordUsecaseを返す(nilなら成功)
func newRecordDeleteUsecaseMock(err error) *mock.MockRecordUsecase {
	return &mock.MockRecordUsecase{
		DeleteFunc: func(ctx context.Context, di usecaseDto.DeleteRecordInput) error {
			return err
		},
	}
}

// serveDeleteRecord DeleteRecordを実行する(recordIdが空文字ならパスパラメータを設定しない)
func serveDeleteRecord(
	uc *mock.MockRecordUsecase,
	recordId string,
	rc recordCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodDelete, "/api/v0/records/"+recordDeleteTargetId, nil)
	return serveRecord(uc, (*RecordHandler).DeleteRecord, req, recordId, rc)
}

func TestRecordHandler_DeleteRecord_正常(t *testing.T) {
	t.Run("正常_書誌を削除", func(t *testing.T) {
		uc := newRecordDeleteUsecaseMock(nil)

		rec, c := serveDeleteRecord(uc, recordDeleteTargetId, recordCtx{})

		// パスパラメータと内部ユーザIDがそのままusecaseへ渡る
		require.True(t, uc.DeleteCalled)
		assert.Equal(t, usecaseDto.DeleteRecordInput{
			InternalUserId: recordTestInternalUid,
			RecordId:       recordDeleteTargetId,
		}, uc.LastDeleteInput)

		require.Len(t, c.Errors, 0)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, ctxutil.REF_API_005, ctxutil.GetApiId(uc.LastCtx))
		testutil.AssertResponse(t, rec, handlerDto.StatusResponse{Status: statusOk})
	})
}

// TestRecordHandler_DeleteRecord_異常 contextの値やパスパラメータが不正なら、usecaseを呼ばずにエラーを返すこと
func TestRecordHandler_DeleteRecord_異常(t *testing.T) {

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
			uc := newRecordDeleteUsecaseMock(nil)

			recordId := recordDeleteTargetId
			if tt.omitRecordId {
				recordId = ""
			}
			_, c := serveDeleteRecord(uc, recordId, tt.ctx)

			assert.False(t, uc.DeleteCalled)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestRecordHandler_DeleteRecord_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestRecordHandler_DeleteRecord_usecaseエラー(t *testing.T) {

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
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newRecordDeleteUsecaseMock(tt.usecaseErr)

			_, c := serveDeleteRecord(uc, recordDeleteTargetId, recordCtx{})

			assert.True(t, uc.DeleteCalled)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
