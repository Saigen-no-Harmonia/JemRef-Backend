package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------GetUserAgreements(*gin.Context)のテスト-------------

// newGetUserAgreementsUsecaseMock 指定したOutputとエラーを返すMockUserUsecaseを返す
func newGetUserAgreementsUsecaseMock(
	output *usecaseDto.GetUserAgreementsOutput,
	err error,
) *mock.MockUserUsecase {
	return &mock.MockUserUsecase{
		GetUserAgreementsFunc: func(ctx context.Context, uid int64) (*usecaseDto.GetUserAgreementsOutput, error) {
			return output, err
		},
	}
}

// serveGetUserAgreements GetUserAgreementsを実行する
func serveGetUserAgreements(uc *mock.MockUserUsecase, uctx userCtx) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodGet, "/api/v0/users/agreements", nil)
	return serveUser(uc, (*UserHandler).GetUserAgreements, req, uctx)
}

func TestUserHandler_GetUserAgreements_正常(t *testing.T) {
	t.Run("正常_ユーザ規約同意状況参照", func(t *testing.T) {
		uc := newGetUserAgreementsUsecaseMock(&usecaseDto.GetUserAgreementsOutput{
			Agreements: []usecaseDto.GetUserAgreement{
				{
					PolicyType:    policy.PolicyTypeTermsOfService,
					Label:         "_terms_",
					LatestVersion: "_1.0_",
					AgreedVersion: "_1.0_",
					Status:        user.PolicyAgreementStatusAgreed,
				},
				{
					PolicyType:    policy.PolicyTypePrivacyPolicy,
					Label:         "_プラポリ_",
					LatestVersion: "_2.0_",
					AgreedVersion: "_1.0_",
					Status:        user.PolicyAgreementStatusUpdateRequired,
				},
			},
		}, nil)

		rec, c := serveGetUserAgreements(uc, userCtx{})

		assertUserUsecaseCalled(t, uc, "GetUserAgreements", true)

		require.Len(t, c.Errors, 0)
		assert.Equal(t, ctxutil.MEM_API_005, ctxutil.GetApiId(uc.LastCtx))
		testutil.AssertResponse(t, rec, handlerDto.GetUserAgreementsResponse{
			Agreements: []handlerDto.UserAgreementResponse{
				{
					PolicyType:    string(policy.PolicyTypeTermsOfService),
					Label:         "_terms_",
					LatestVersion: "_1.0_",
					AgreedVersion: "_1.0_",
					Status:        user.PolicyAgreementStatusAgreed,
				},
				{
					PolicyType:    string(policy.PolicyTypePrivacyPolicy),
					Label:         "_プラポリ_",
					LatestVersion: "_2.0_",
					AgreedVersion: "_1.0_",
					Status:        user.PolicyAgreementStatusUpdateRequired,
				},
			},
		})
	})
}

func TestUserHandler_GetUserAgreements_異常(t *testing.T) {
	t.Run("異常_internal uid取得失敗", func(t *testing.T) {
		uc := newGetUserAgreementsUsecaseMock(nil, ErrTestUnexpected)

		_, c := serveGetUserAgreements(uc, userCtx{omitInternalUid: true})

		assertUserUsecaseCalled(t, uc, "GetUserAgreements", false)
		assertHandlerError(t, c, ErrInternalUidNotFound)
	})
}

// TestUserHandler_GetUserAgreements_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestUserHandler_GetUserAgreements_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		usecaseErr error
	}{
		{
			name:       "異常_ユーザ非存在",
			usecaseErr: usecase.ErrUserNotFound,
		},
		{
			name:       "異常_ユーザ情報取得時に予期せぬエラー",
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newGetUserAgreementsUsecaseMock(nil, tt.usecaseErr)

			_, c := serveGetUserAgreements(uc, userCtx{})

			assertUserUsecaseCalled(t, uc, "GetUserAgreements", true)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}

// -------------UpdateUserAgreements(*gin.Context)のテスト-------------

// validUpdateUserAgreementsRequest 規約・プライバシーポリシーの両方に同意する正しいリクエスト
var validUpdateUserAgreementsRequest = handlerDto.UpdateUserAgreementsRequest{
	Policies: []handlerDto.PolicyAgreementRequest{
		{
			PolicyType: string(policy.PolicyTypeTermsOfService),
			Version:    "0.1",
		},
		{
			PolicyType: string(policy.PolicyTypePrivacyPolicy),
			Version:    "0.2",
		},
	},
}

// newUpdateUserAgreementsUsecaseMock 指定したエラーを返すMockUserUsecaseを返す(nilなら成功)
func newUpdateUserAgreementsUsecaseMock(err error) *mock.MockUserUsecase {
	return &mock.MockUserUsecase{
		UpdateUserAgreementsFunc: func(ctx context.Context, uuai usecaseDto.UpdateUserAgreementsInput) error {
			return err
		},
	}
}

// serveUpdateUserAgreements UpdateUserAgreementsを実行する
func serveUpdateUserAgreements(
	t *testing.T,
	uc *mock.MockUserUsecase,
	body handlerDto.UpdateUserAgreementsRequest,
	uctx userCtx,
) (*httptest.ResponseRecorder, *gin.Context) {
	req := httptest.NewRequest(http.MethodPut, "/api/v0/user/agreements", testutil.NewJsonReader(t, body))
	return serveUser(uc, (*UserHandler).UpdateUserAgreements, req, uctx)
}

// assertUpdateUserAgreementsInput 内部ユーザIDと、リクエストの規約タイプ・バージョンが順にusecaseへ渡ったこと
func assertUpdateUserAgreementsInput(
	t *testing.T,
	uc *mock.MockUserUsecase,
	body handlerDto.UpdateUserAgreementsRequest,
) {
	t.Helper()

	assert.Equal(t, userTestInternalUid, uc.LastUpdateUserInput.InternalUid)
	require.Len(t, uc.LastUpdateUserInput.Agreements, len(body.Policies))
	for i, p := range body.Policies {
		arg := uc.LastUpdateUserInput.Agreements[i]
		assert.Equal(t, p.PolicyType, string(arg.PolicyType))
		assert.Equal(t, p.Version, string(arg.AgreedVersion))
	}
}

func TestUserHandler_UpdateUserAgreements_正常(t *testing.T) {
	t.Run("正常_ユーザ規約同意状況更新", func(t *testing.T) {
		uc := newUpdateUserAgreementsUsecaseMock(nil)

		_, c := serveUpdateUserAgreements(t, uc, validUpdateUserAgreementsRequest, userCtx{})

		assertUserUsecaseCalled(t, uc, "UpdateUserAgreements", true)
		assertUpdateUserAgreementsInput(t, uc, validUpdateUserAgreementsRequest)

		require.Len(t, c.Errors, 0)
		assert.Equal(t, ctxutil.MEM_API_006, ctxutil.GetApiId(uc.LastCtx))
	})
}

// TestUserHandler_UpdateUserAgreements_異常 contextの値やリクエストボディが不正なら、usecaseを呼ばずにエラーを返すこと
func TestUserHandler_UpdateUserAgreements_異常(t *testing.T) {

	tests := []struct {
		name              string
		ctx               userCtx
		body              handlerDto.UpdateUserAgreementsRequest
		expectedErrorBody error
	}{
		{
			name:              "異常_ユーザID取得失敗",
			ctx:               userCtx{omitInternalUid: true},
			body:              handlerDto.UpdateUserAgreementsRequest{},
			expectedErrorBody: ErrInternalUidNotFound,
		},
		{
			name:              "異常_規約情報がnull",
			body:              handlerDto.UpdateUserAgreementsRequest{},
			expectedErrorBody: ErrPolicyRequired,
		},
		{
			// Policiesのdiveタグにより各要素のbindingタグが検証される
			name: "異常_要素の規約タイプが未指定",
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{PolicyType: "", Version: "0.1"},
				},
			},
			expectedErrorBody: ErrInvalidPolicyType,
		},
		{
			name: "異常_要素の規約バージョンが未指定",
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{PolicyType: "terms", Version: ""},
				},
			},
			expectedErrorBody: ErrAgreedVersionRequired,
		},
		{
			// 2件目以降の要素も検証されること
			name: "異常_2件目の要素が不正",
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{
					{PolicyType: "terms", Version: "0.1"},
					{PolicyType: "", Version: "0.1"},
				},
			},
			expectedErrorBody: ErrInvalidPolicyType,
		},
		{
			name: "異常_リクエストが空の配列",
			body: handlerDto.UpdateUserAgreementsRequest{
				Policies: []handlerDto.PolicyAgreementRequest{},
			},
			expectedErrorBody: ErrPolicyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newUpdateUserAgreementsUsecaseMock(nil)

			_, c := serveUpdateUserAgreements(t, uc, tt.body, tt.ctx)

			assertUserUsecaseCalled(t, uc, "UpdateUserAgreements", false)
			assertHandlerError(t, c, tt.expectedErrorBody)
		})
	}
}

// TestUserHandler_UpdateUserAgreements_usecaseエラー usecaseのエラーを包み直さずに伝播すること
func TestUserHandler_UpdateUserAgreements_usecaseエラー(t *testing.T) {

	tests := []struct {
		name       string
		usecaseErr error
	}{
		{
			name:       "異常_ユーザ取得処理異常",
			usecaseErr: usecase.ErrUserNotFound,
		},
		{
			name:       "異常_規約タイプ指定不正",
			usecaseErr: usecase.ErrInvalidPolicyType,
		},
		{
			name:       "異常_規約バージョン指定不正",
			usecaseErr: usecase.ErrInvalidPolicyVersion,
		},
		{
			name:       "異常_予期せぬエラー",
			usecaseErr: usecase.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newUpdateUserAgreementsUsecaseMock(tt.usecaseErr)

			_, c := serveUpdateUserAgreements(t, uc, validUpdateUserAgreementsRequest, userCtx{})

			assertUserUsecaseCalled(t, uc, "UpdateUserAgreements", true)
			assertUpdateUserAgreementsInput(t, uc, validUpdateUserAgreementsRequest)
			assertHandlerError(t, c, tt.usecaseErr)
		})
	}
}
