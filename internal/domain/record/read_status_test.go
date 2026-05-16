package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStatus_IsValid(t *testing.T) {

	tests := []struct {
		name     string
		input    ReadStatus
		expected bool
	}{
		{
			name:     "正常_ステータス既読",
			input:    ReadStatusRead,
			expected: true,
		},
		{
			name:     "正常_ステータス未読",
			input:    ReadStatusUnread,
			expected: true,
		},
		{
			name:     "正常_ステータス読書中",
			input:    ReadStatusReading,
			expected: true,
		},
		{
			name:     "正常_ステータス一部既読",
			input:    ReadStatusPertialRead,
			expected: true,
		},
		{
			name:     "異常_ステータス不正",
			input:    ReadStatus("invalid"),
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

func TestReadStatus_GetId(t *testing.T) {
	tests := []struct {
		name          string
		input         ReadStatus
		expected      int
		expectedError bool
	}{
		{
			name:          "正常_ステータス既読",
			input:         ReadStatusRead,
			expected:      1,
			expectedError: false,
		},
		{
			name:          "正常_ステータス未読",
			input:         ReadStatusUnread,
			expected:      2,
			expectedError: false,
		},
		{
			name:          "正常_ステータス読書中",
			input:         ReadStatusReading,
			expected:      3,
			expectedError: false,
		},
		{
			name:          "正常_ステータス一部既読",
			input:         ReadStatusPertialRead,
			expected:      4,
			expectedError: false,
		},
		{
			name:          "異常_ステータス不正",
			input:         ReadStatus("invalid"),
			expected:      0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		actual, err := tt.input.GetId()
		if tt.expectedError {
			assert.Error(t, err)
			return
		}

		assert.NoError(t, err)
		assert.Equal(t, tt.expected, actual)
	}
}

func TestReadStatus_ReadStatusFromId(t *testing.T) {

	tests := []struct {
		name          string
		input         int
		expected      ReadStatus
		expectedError bool
	}{
		{
			name:          "正常_ステータス既読",
			input:         ReadStatusIdRead,
			expected:      ReadStatusRead,
			expectedError: false,
		},
		{
			name:          "正常_ステータス未読",
			input:         ReadStatusIdUnread,
			expected:      ReadStatusUnread,
			expectedError: false,
		},
		{
			name:          "正常_ステータス読書中",
			input:         ReadStatusIdReading,
			expected:      ReadStatusReading,
			expectedError: false,
		},
		{
			name:          "正常_ステータス一部既読",
			input:         ReadStatusIdPertialRead,
			expected:      ReadStatusPertialRead,
			expectedError: false,
		},
		{
			name:          "異常_ステータス不正",
			input:         0,
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ReadStatusFromId(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
