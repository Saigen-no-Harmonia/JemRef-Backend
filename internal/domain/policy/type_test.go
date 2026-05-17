package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_Isvalid(t *testing.T) {

	tests := []struct {
		name     string
		input    PolicyType
		expected bool
	}{
		{
			name:     "正常_ユーザ利用規約",
			input:    PolicyTypeTermsOfService,
			expected: true,
		},
		{
			name:     "正常_プライバシーポリシー",
			input:    PolicyTypeTermsOfService,
			expected: true,
		},
		{
			name:     "異常_不正なポリシー",
			input:    PolicyType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			actual := tt.input.IsValid()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestType_GetId(t *testing.T) {

	tests := []struct {
		name        string
		input       PolicyType
		expected    string
		expectPanic bool
	}{
		{
			name:        "正常_ユーザ利用規約をIDに変換",
			input:       PolicyTypeTermsOfService,
			expected:    string(PolicyIdTermsOfService),
			expectPanic: false,
		},
		{
			name:        "正常_プライバシーポリシーをIDに変換",
			input:       PolicyTypePrivacyPolicy,
			expected:    string(PolicyIdPrivacyPolicy),
			expectPanic: false,
		},
		{
			name:        "異常_不正値を変換",
			input:       PolicyType("invalid"),
			expected:    "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					tt.input.GetId()
				})
				return
			}

			actual := tt.input.GetId()
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestType_PoliicyTypeFromCode(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		expected    PolicyType
		expectPanic bool
	}{
		{
			name:        "正常_ユーザ利用規約",
			input:       PolicyIdTermsOfService,
			expected:    PolicyTypeTermsOfService,
			expectPanic: false,
		},
		{
			name:        "正常_プライバシーポリシー",
			input:       PolicyIdPrivacyPolicy,
			expected:    PolicyTypePrivacyPolicy,
			expectPanic: false,
		},
		{
			name:        "異常_想定外の値",
			input:       "invavlid",
			expected:    "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					PolicyTypeFromCode(tt.input)
				})
				return
			}

			actual := PolicyTypeFromCode(tt.input)
			assert.Equal(t, actual, tt.expected)
		})
	}
}
