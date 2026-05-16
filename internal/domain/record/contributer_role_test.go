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
		name          string
		input         ContributerRole
		expected      string
		expectedError bool
	}{
		{
			name:          "正常_著者",
			input:         ContributerRoleAuthor,
			expected:      ContributerRoleIdAuthor,
			expectedError: false,
		},
		{
			name:          "正常_翻訳者",
			input:         ContributerRoleInterpreter,
			expected:      ContributerRoleIdInterpreter,
			expectedError: false,
		},
		{
			name:          "正常_編者",
			input:         ContributerRoleCompiler,
			expected:      ContributerRoleIdCompiler,
			expectedError: false,
		},
		{
			name:          "正常_編著者",
			input:         ContributerRoleCompilerAuthor,
			expected:      ContributerRoleIdCompilerAuthor,
			expectedError: false,
		},
		{
			name:          "異常_不正な役割",
			input:         ContributerRole("invalid"),
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			actual, err := tt.input.GetId()
			assert.Equal(t, tt.expected, actual)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_ContributerRole_RoleFromId(t *testing.T) {

	tests := []struct {
		name          string
		input         string
		expected      ContributerRole
		expectedError bool
	}{
		{
			name:          "正常_著者",
			input:         ContributerRoleIdAuthor,
			expected:      ContributerRoleAuthor,
			expectedError: false,
		},
		{
			name:          "正常_翻訳者",
			input:         ContributerRoleIdInterpreter,
			expected:      ContributerRoleInterpreter,
			expectedError: false,
		},
		{
			name:          "正常_編者",
			input:         ContributerRoleIdCompiler,
			expected:      ContributerRoleCompiler,
			expectedError: false,
		},
		{
			name:          "正常_編著者",
			input:         ContributerRoleIdCompilerAuthor,
			expected:      ContributerRoleCompilerAuthor,
			expectedError: false,
		},
		{
			name:          "異常_不正な役割",
			input:         "invalid",
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := RoleFromId(tt.input)
			assert.Equal(t, tt.expected, actual)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
