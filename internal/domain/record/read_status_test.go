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
			input:    ReadStatusPartiallyRead,
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
		name        string
		input       ReadStatus
		expected    int
		expectPanic bool
	}{
		{
			name:        "正常_ステータス既読",
			input:       ReadStatusRead,
			expected:    1,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス未読",
			input:       ReadStatusUnread,
			expected:    2,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス読書中",
			input:       ReadStatusReading,
			expected:    3,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス一部既読",
			input:       ReadStatusPartiallyRead,
			expected:    4,
			expectPanic: false,
		},
		{
			name:        "異常_ステータス不正",
			input:       ReadStatus("invalid"),
			expected:    0,
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

func TestReadStatus_GetNullableId(t *testing.T) {

	readStatusRead := ReadStatusRead
	readStatusPartiallyRead := ReadStatusPartiallyRead
	readStatusInvalid := ReadStatus("invalid")

	tests := []struct {
		name        string
		input       *ReadStatus
		expected    any
		expectPanic bool
	}{
		{
			name:        "正常_未設定はnil",
			input:       nil,
			expected:    nil,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス既読",
			input:       &readStatusRead,
			expected:    ReadStatusIdRead,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス一部既読",
			input:       &readStatusPartiallyRead,
			expected:    ReadStatusIdPartiallyRead,
			expectPanic: false,
		},
		{
			name:        "異常_ステータス不正",
			input:       &readStatusInvalid,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					tt.input.GetNullableId()
				})
				return
			}

			actual := tt.input.GetNullableId()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestReadStatus_ReadStatusFromId(t *testing.T) {

	tests := []struct {
		name        string
		input       int
		expected    ReadStatus
		expectPanic bool
	}{
		{
			name:        "正常_ステータス既読",
			input:       ReadStatusIdRead,
			expected:    ReadStatusRead,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス未読",
			input:       ReadStatusIdUnread,
			expected:    ReadStatusUnread,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス読書中",
			input:       ReadStatusIdReading,
			expected:    ReadStatusReading,
			expectPanic: false,
		},
		{
			name:        "正常_ステータス一部既読",
			input:       ReadStatusIdPartiallyRead,
			expected:    ReadStatusPartiallyRead,
			expectPanic: false,
		},
		{
			name:        "異常_ステータス不正",
			input:       0,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					ReadStatusFromId(tt.input)
				})
				return
			}

			actual := ReadStatusFromId(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
