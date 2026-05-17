package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ContributerRole_IsValid(t *testing.T) {

	tests := []struct {
		name     string
		input    ContributerRole
		expected bool
	}{
		{
			name:     "正常_著者",
			input:    ContributerRoleAuthor,
			expected: true,
		},
		{
			name:     "正常_翻訳者",
			input:    ContributerRoleInterpreter,
			expected: true,
		},
		{
			name:     "正常_編者",
			input:    ContributerRoleCompiler,
			expected: true,
		},
		{
			name:     "正常_編著者",
			input:    ContributerRoleCompilerAuthor,
			expected: true,
		},
		{
			name:     "異常_不正な役割",
			input:    ContributerRole("invalid"),
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

func Test_ContributerRole_GetId(t *testing.T) {

	tests := []struct {
		name        string
		input       ContributerRole
		expected    string
		expectPanic bool
	}{
		{
			name:        "正常_著者",
			input:       ContributerRoleAuthor,
			expected:    ContributerRoleIdAuthor,
			expectPanic: false,
		},
		{
			name:        "正常_翻訳者",
			input:       ContributerRoleInterpreter,
			expected:    ContributerRoleIdInterpreter,
			expectPanic: false,
		},
		{
			name:        "正常_編者",
			input:       ContributerRoleCompiler,
			expected:    ContributerRoleIdCompiler,
			expectPanic: false,
		},
		{
			name:        "正常_編著者",
			input:       ContributerRoleCompilerAuthor,
			expected:    ContributerRoleIdCompilerAuthor,
			expectPanic: false,
		},
		{
			name:        "異常_不正な役割",
			input:       ContributerRole("invalid"),
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

func Test_ContributerRole_RoleFromId(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		expected    ContributerRole
		expectPanic bool
	}{
		{
			name:        "正常_著者",
			input:       ContributerRoleIdAuthor,
			expected:    ContributerRoleAuthor,
			expectPanic: false,
		},
		{
			name:        "正常_翻訳者",
			input:       ContributerRoleIdInterpreter,
			expected:    ContributerRoleInterpreter,
			expectPanic: false,
		},
		{
			name:        "正常_編者",
			input:       ContributerRoleIdCompiler,
			expected:    ContributerRoleCompiler,
			expectPanic: false,
		},
		{
			name:        "正常_編著者",
			input:       ContributerRoleIdCompilerAuthor,
			expected:    ContributerRoleCompilerAuthor,
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
