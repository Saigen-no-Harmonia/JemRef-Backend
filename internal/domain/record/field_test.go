package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordField_IsValid(t *testing.T) {

	tests := []struct {
		name     string
		input    RecordField
		expected bool
	}{
		{
			name:     "正常_ISBN",
			input:    RecordFieldIsbn,
			expected: true,
		},
		{
			name:     "正常_版",
			input:    RecordFieldEdition,
			expected: true,
		},
		{
			name:     "正常_原題",
			input:    RecordFieldOriginalTitle,
			expected: true,
		},
		{
			name:     "異常_入力項目不正",
			input:    RecordField("invalid"),
			expected: false,
		},
		{
			name:     "異常_入力項目空文字",
			input:    RecordField(""),
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

func TestRecordField_GetId(t *testing.T) {

	tests := []struct {
		name        string
		input       RecordField
		expected    string
		expectPanic bool
	}{
		{
			name:        "正常_ISBN",
			input:       RecordFieldIsbn,
			expected:    RecordFieldIdIsbn,
			expectPanic: false,
		},
		{
			name:        "正常_版",
			input:       RecordFieldEdition,
			expected:    RecordFieldIdEdition,
			expectPanic: false,
		},
		{
			name:        "正常_原題",
			input:       RecordFieldOriginalTitle,
			expected:    RecordFieldIdOriginalTitle,
			expectPanic: false,
		},
		{
			name:        "異常_入力項目不正",
			input:       RecordField("invalid"),
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

func TestRecordField_RecordFieldFromId(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		expected    RecordField
		expectPanic bool
	}{
		{
			name:        "正常_ISBN",
			input:       RecordFieldIdIsbn,
			expected:    RecordFieldIsbn,
			expectPanic: false,
		},
		{
			name:        "正常_版",
			input:       RecordFieldIdEdition,
			expected:    RecordFieldEdition,
			expectPanic: false,
		},
		{
			name:        "正常_原題",
			input:       RecordFieldIdOriginalTitle,
			expected:    RecordFieldOriginalTitle,
			expectPanic: false,
		},
		{
			name:        "異常_想定外のID",
			input:       "RF999999",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					RecordFieldFromId(tt.input)
				})
				return
			}

			actual := RecordFieldFromId(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
