package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"jemref_go/internal/domain/policy"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	usecasedto "jemref_go/internal/usecase/dto"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGeneralHandler_GetPolicies(t *testing.T) {

	basePath := "/api/v0/policies/"
	effectiveDate1, _ := time.Parse("2006-01-02", "2025-01-31")
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		path              string
		mockUsecase       *MockGeneralUsecase
		expectedStatus    int
		expectedBody      handlerdto.GetPoliciesResponse
		expectedErrorBody handlerdto.ErrorResponse
	}{
		{
			name: "正常_ユーザ規約参照",
			path: basePath + string(policy.PolicyTypeTermsOfService),
			mockUsecase: &MockGeneralUsecase{
				getPoliciesFunc: func(
					ctx context.Context,
					input usecasedto.GetPoliciesInput,
				) (*usecasedto.GetPoliciesOutput, error) {
					return &usecasedto.GetPoliciesOutput{
						PolicyId:      string(policy.PolicyIdTermsOfService),
						Label:         "ユーザ利用規約",
						LatestVersion: "1.0",
						EffectiveDate: effectiveDate1,
						Content:       "sample_terms_content",
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody: handlerdto.GetPoliciesResponse{
				PolicyType:    string(policy.PolicyTypeTermsOfService),
				Label:         "ユーザ利用規約",
				LatestVersion: "1.0",
				EffectiveDate: effectiveDate1.Format("2006-01-02"),
				Content:       "sample_terms_content",
			},
		},
		{
			name:           "異常_パラメータ不正",
			path:           basePath + "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code: "E0003",
				Message: fmt.Sprintf(
					"規約IDが不正です。policy_id=%s",
					"invalid",
				),
			},
		},
		{
			name: "異常_リクエストされた規約がDBに存在しない",
			path: basePath + string(policy.PolicyTypeTermsOfService),
			mockUsecase: &MockGeneralUsecase{
				getPoliciesFunc: func(
					ctx context.Context,
					gpi usecasedto.GetPoliciesInput,
				) (*usecasedto.GetPoliciesOutput, error) {
					return nil, usecase.ErrPolicyNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code: "E0006",
				Message: fmt.Sprintf("リクエストされた規約が存在しません。policy_type=%s",
					string(policy.PolicyTypeTermsOfService),
				),
			},
		},
		{
			name: "異常_意図しないエラー",
			path: basePath + string(policy.PolicyTypeTermsOfService),
			mockUsecase: &MockGeneralUsecase{
				getPoliciesFunc: func(
					ctx context.Context,
					gpi usecasedto.GetPoliciesInput,
				) (*usecasedto.GetPoliciesOutput, error) {
					return nil, fmt.Errorf("意図しないエラー")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedErrorBody: handlerdto.ErrorResponse{
				Code:    "A0001",
				Message: "Fatal: internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := gin.New()
			h := NewGeneralHandler(tt.mockUsecase)

			r.GET(
				"/api/v0/policies/:"+ParamPolicyType,
				h.GetPolicies,
			)

			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			// 異常系の場合のassertion
			if tt.expectedStatus != http.StatusOK {
				var actual handlerdto.ErrorResponse
				err := json.Unmarshal(
					rec.Body.Bytes(),
					&actual,
				)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedErrorBody, actual)
				return
			}

			// 正常系の場合のassertion
			var actual handlerdto.GetPoliciesResponse
			err := json.Unmarshal(
				rec.Body.Bytes(),
				&actual,
			)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, actual)
		})
	}
}

type MockGeneralUsecase struct {
	getPoliciesFunc func(
		context.Context,
		usecasedto.GetPoliciesInput,
	) (*usecasedto.GetPoliciesOutput, error)
}

func (m *MockGeneralUsecase) GetPolicies(
	ctx context.Context,
	input usecasedto.GetPoliciesInput,
) (*usecasedto.GetPoliciesOutput, error) {
	return m.getPoliciesFunc(ctx, input)
}
