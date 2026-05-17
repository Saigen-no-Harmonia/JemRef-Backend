package usecase

import (
	"context"
	"errors"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/usecase/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeneralUsecaseImpl_GetPolicies(t *testing.T) {

	tests := []struct {
		name          string
		mockRepo      *MockGeneralRepository
		inputDto      dto.GetPoliciesInput
		expected      dto.GetPoliciesOutput
		expectedError bool
	}{
		{
			name: "正常_ユーザ利用規約取得",
			mockRepo: &MockGeneralRepository{
				selectLatestByIdFunc: func(
					ctx context.Context,
					id string,
				) (*policy.Policy, error) {
					return &policy.Policy{
						Id:            policy.PolicyIdTermsOfService,
						Version:       "sample_version_01",
						Name:          "ユーザ利用規約",
						Content:       "sample_content_terms_01",
						EffectiveDate: mustParseDate("2025-01-31"),
					}, nil
				},
			},
			inputDto: dto.GetPoliciesInput{
				PolicyId: policy.PolicyIdTermsOfService,
			},
			expected: dto.GetPoliciesOutput{
				PolicyId:      policy.PolicyIdTermsOfService,
				Label:         "ユーザ利用規約",
				LatestVersion: "sample_version_01",
				EffectiveDate: mustParseDate("2025-01-31"),
				Content:       "sample_content_terms_01",
			},
			expectedError: false,
		},
		{
			name: "正常_プライバシーポリシー取得",
			mockRepo: &MockGeneralRepository{
				selectLatestByIdFunc: func(
					ctx context.Context,
					id string,
				) (*policy.Policy, error) {
					return &policy.Policy{
						Id:            policy.PolicyIdPrivacyPolicy,
						Version:       "sample_version_02",
						Name:          "プライバシーポリシー",
						Content:       "sample_content_privacy_policy_01",
						EffectiveDate: mustParseDate("2025-02-28"),
					}, nil
				},
			},
			inputDto: dto.GetPoliciesInput{
				PolicyId: policy.PolicyIdPrivacyPolicy,
			},
			expected: dto.GetPoliciesOutput{
				PolicyId:      policy.PolicyIdPrivacyPolicy,
				Label:         "プライバシーポリシー",
				LatestVersion: "sample_version_02",
				EffectiveDate: mustParseDate("2025-02-28"),
				Content:       "sample_content_privacy_policy_01",
			},
			expectedError: false,
		},
		{
			name: "異常_規約が存在しない",
			mockRepo: &MockGeneralRepository{
				selectLatestByIdFunc: func(
					ctx context.Context,
					id string,
				) (*policy.Policy, error) {
					return nil, errors.New("not found")
				},
			},
			inputDto: dto.GetPoliciesInput{
				PolicyId: policy.PolicyIdTermsOfService,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewGeneralUsecaseImpl(tt.mockRepo)

			actual, err := u.GetPolicies(
				context.Background(),
				tt.inputDto,
			)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, actual)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.PolicyId, actual.PolicyId)
			assert.Equal(t, tt.expected.Label, actual.Label)
			assert.Equal(t, tt.expected.LatestVersion, actual.LatestVersion)
			assert.Equal(t, tt.expected.EffectiveDate, actual.EffectiveDate)
			assert.Equal(t, tt.expected.Content, actual.Content)
		})
	}
}

type MockGeneralRepository struct {
	selectLatestByIdFunc func(
		ctx context.Context,
		id string,
	) (*policy.Policy, error)
}

func (m *MockGeneralRepository) SelectLatestById(
	ctx context.Context,
	id string,
) (*policy.Policy, error) {
	return m.selectLatestByIdFunc(ctx, id)
}

func mustParseDate(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)

	if err != nil {
		panic(err)
	}

	return t
}
