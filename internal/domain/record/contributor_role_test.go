package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ContributorRole_IsValid(t *testing.T) {

	tests := []struct {
		name     string
		input    ContributorRole
		expected bool
	}{
		{
			name:     "正常_著者",
			input:    ContributorRoleAuthor,
			expected: true,
		},
		{
			name:     "正常_翻訳者",
			input:    ContributorRoleInterpreter,
			expected: true,
		},
		{
			name:     "正常_編者",
			input:    ContributorRoleCompiler,
			expected: true,
		},
		{
			name:     "正常_編著者",
			input:    ContributorRoleCompilerAuthor,
			expected: true,
		},
		{
			name:     "異常_不正な役割",
			input:    ContributorRole("invalid"),
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

func Test_ContributorRole_GetId(t *testing.T) {

	tests := []struct {
		name        string
		input       ContributorRole
		expected    string
		expectPanic bool
	}{
		{
			name:        "正常_著者",
			input:       ContributorRoleAuthor,
			expected:    ContributorRoleIdAuthor,
			expectPanic: false,
		},
		{
			name:        "正常_翻訳者",
			input:       ContributorRoleInterpreter,
			expected:    ContributorRoleIdInterpreter,
			expectPanic: false,
		},
		{
			name:        "正常_編者",
			input:       ContributorRoleCompiler,
			expected:    ContributorRoleIdCompiler,
			expectPanic: false,
		},
		{
			name:        "正常_編著者",
			input:       ContributorRoleCompilerAuthor,
			expected:    ContributorRoleIdCompilerAuthor,
			expectPanic: false,
		},
		{
			name:        "異常_不正な役割",
			input:       ContributorRole("invalid"),
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
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_ContributorRole_RoleFromId(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		expected    ContributorRole
		expectPanic bool
	}{
		{
			name:        "正常_著者",
			input:       ContributorRoleIdAuthor,
			expected:    ContributorRoleAuthor,
			expectPanic: false,
		},
		{
			name:        "正常_翻訳者",
			input:       ContributorRoleIdInterpreter,
			expected:    ContributorRoleInterpreter,
			expectPanic: false,
		},
		{
			name:        "正常_編者",
			input:       ContributorRoleIdCompiler,
			expected:    ContributorRoleCompiler,
			expectPanic: false,
		},
		{
			name:        "正常_編著者",
			input:       ContributorRoleIdCompilerAuthor,
			expected:    ContributorRoleCompilerAuthor,
			expectPanic: false,
		},
		{
			name:        "異常_不正な役割",
			input:       "invalid",
			expected:    "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					RoleFromId(tt.input)
				})
				return
			}

			actual := RoleFromId(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
