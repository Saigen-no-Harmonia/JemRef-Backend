package handler

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
)

func TestGeneralHandler_GetPolicies(t *testing.T) {

	basePath := "/api/v0/policies/"
	effectiveDate1, _ := time.Parse("2006-01-02", "2025-01-31")
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		policyType        string
		mockUsecase       *MockGeneralUsecase
		expectedBody      handlerdto.GetPoliciesResponse
		expectError       bool
		expectedErrorBody error
	}{
		{
			name:       "正常_ユーザ規約参照",
			policyType: string(policy.PolicyTypeTermsOfService),
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
			expectedBody: handlerdto.GetPoliciesResponse{
				PolicyType:    string(policy.PolicyTypeTermsOfService),
				Label:         "ユーザ利用規約",
				LatestVersion: "1.0",
				EffectiveDate: effectiveDate1.Format("2006-01-02"),
				Content:       "sample_terms_content",
			},
			expectError: false,
		},
		{
			name:              "異常_パラメータ不正",
			policyType:        "invalid",
			expectError:       true,
			expectedErrorBody: ErrPolicyTypeInvalid,
		},
		{
			name:       "異常_リクエストされた規約がDBに存在しない",
			policyType: string(policy.PolicyTypeTermsOfService),
			mockUsecase: &MockGeneralUsecase{
				getPoliciesFunc: func(
					ctx context.Context,
					gpi usecasedto.GetPoliciesInput,
				) (*usecasedto.GetPoliciesOutput, error) {
					return nil, usecase.ErrPolicyNotFound
				},
			},
			expectError:       true,
			expectedErrorBody: usecase.ErrPolicyNotFound,
		},
		{
			name:       "異常_意図しないエラー",
			policyType: string(policy.PolicyTypeTermsOfService),
			mockUsecase: &MockGeneralUsecase{
				getPoliciesFunc: func(
					ctx context.Context,
					gpi usecasedto.GetPoliciesInput,
				) (*usecasedto.GetPoliciesOutput, error) {
					return nil, ErrTestUnexpected
				},
			},
			expectError:       true,
			expectedErrorBody: ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gin.SetMode(gin.TestMode)

			h := NewGeneralHandler(tt.mockUsecase)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			req := httptest.NewRequest(
				http.MethodGet,
				basePath+tt.policyType,
				nil,
			)

			c.Request = req

			c.Params = gin.Params{
				{
					Key:   ParamPolicyType,
					Value: tt.policyType,
				},
			}

			h.GetPolicies(c)

			// 異常系の場合のassertion
			if tt.expectError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
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
			assert.Empty(t, c.Errors)
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
