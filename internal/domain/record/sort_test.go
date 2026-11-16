package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortKey_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		input    SortKey
		expected bool
	}{
		{
			name:     "正常_出版年",
			input:    SortKeyPublishYear,
			expected: true,
		},
		{
			name:     "正常_登録日",
			input:    SortKeyInsertDate,
			expected: true,
		},
		{
			name:     "異常_未知の値",
			input:    SortKey("main_title"),
			expected: false,
		},
		{
			// 大文字小文字を区別する(リクエストの値をそのまま判定するため)
			name:     "異常_大文字",
			input:    SortKey("PUBLISH_YEAR"),
			expected: false,
		},
		{
			// 空白を含む値も補正せず不正値とする(handlerも並び替えの値はtrimしないため、400になる)
			name:     "異常_前後に空白",
			input:    SortKey(" publish_year "),
			expected: false,
		},
		{
			// 未指定時のデフォルト補完はhandlerの責務。空文字は不正値として扱う
			name:     "異常_空文字",
			input:    SortKey(""),
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

func TestSortOrder_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		input    SortOrder
		expected bool
	}{
		{
			name:     "正常_昇順",
			input:    SortOrderAsc,
			expected: true,
		},
		{
			name:     "正常_降順",
			input:    SortOrderDesc,
			expected: true,
		},
		{
			name:     "異常_未知の値",
			input:    SortOrder("random"),
			expected: false,
		},
		{
			// 大文字小文字を区別する(リクエストの値をそのまま判定するため)
			name:     "異常_大文字",
			input:    SortOrder("ASC"),
			expected: false,
		},
		{
			// 空白を含む値も補正せず不正値とする(handlerも並び替えの値はtrimしないため、400になる)
			name:     "異常_前後に空白",
			input:    SortOrder(" asc "),
			expected: false,
		},
		{
			// 未指定時のデフォルト補完はhandlerの責務。空文字は不正値として扱う
			name:     "異常_空文字",
			input:    SortOrder(""),
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
