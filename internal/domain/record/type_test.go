package record

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordType_IsValid(t *testing.T) {

	tests := []struct {
		name     string
		input    RecordType
		expected bool
	}{
		{
			name:     "正常_単著書",
			input:    RecordTypeAuthoredBook,
			expected: true,
		},
		{
			name:     "正常_雑誌",
			input:    RecordTypeJournal,
			expected: true,
		},
		{
			name:     "正常_雑誌記事",
			input:    RecordTypeJournalArticle,
			expected: true,
		},
		{
			name:     "正常_編著書",
			input:    RecordTypeContributedBook,
			expected: true,
		},
		{
			name:     "正常_編著書記事",
			input:    RecordTypeContributedBookArticle,
			expected: true,
		},
		{
			name:     "異常_書誌タイプ不正",
			input:    RecordType("invalid"),
			expected: false,
		},
		{
			name:     "異常_書誌タイプ空文字",
			input:    RecordType(""),
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

func TestRecordType_GetId(t *testing.T) {

	tests := []struct {
		name        string
		input       RecordType
		expected    string
		expectPanic bool
	}{
		{
			name:        "正常_単著書",
			input:       RecordTypeAuthoredBook,
			expected:    RecordTypeIdAuthoredBook,
			expectPanic: false,
		},
		{
			name:        "正常_雑誌",
			input:       RecordTypeJournal,
			expected:    RecordTypeIdJournal,
			expectPanic: false,
		},
		{
			name:        "正常_雑誌記事",
			input:       RecordTypeJournalArticle,
			expected:    RecordTypeIdJournalArticle,
			expectPanic: false,
		},
		{
			name:        "正常_編著書",
			input:       RecordTypeContributedBook,
			expected:    RecordTypeIdContributedBook,
			expectPanic: false,
		},
		{
			name:        "正常_編著書記事",
			input:       RecordTypeContributedBookArticle,
			expected:    RecordTypeIdContributedBookArticle,
			expectPanic: false,
		},
		{
			name:        "異常_書誌タイプ不正",
			input:       RecordType("invalid"),
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

func TestRecordType_HasParent(t *testing.T) {

	tests := []struct {
		name     string
		input    RecordType
		expected bool
	}{
		{
			name:     "正常_雑誌記事は親を持つ",
			input:    RecordTypeJournalArticle,
			expected: true,
		},
		{
			name:     "正常_編著書記事は親を持つ",
			input:    RecordTypeContributedBookArticle,
			expected: true,
		},
		{
			name:     "正常_単著書は親を持たない",
			input:    RecordTypeAuthoredBook,
			expected: false,
		},
		{
			name:     "正常_雑誌は親を持たない",
			input:    RecordTypeJournal,
			expected: false,
		},
		{
			name:     "正常_編著書は親を持たない",
			input:    RecordTypeContributedBook,
			expected: false,
		},
		{
			name:     "異常_書誌タイプ不正",
			input:    RecordType("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.input.HasParent()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRecordType_GetParent(t *testing.T) {

	tests := []struct {
		name        string
		input       RecordType
		expected    RecordType
		expectPanic bool
	}{
		{
			name:        "正常_雑誌記事の親は雑誌",
			input:       RecordTypeJournalArticle,
			expected:    RecordTypeJournal,
			expectPanic: false,
		},
		{
			name:        "正常_編著書記事の親は編著書",
			input:       RecordTypeContributedBookArticle,
			expected:    RecordTypeContributedBook,
			expectPanic: false,
		},
		{
			name:        "異常_親を持たない書誌タイプ",
			input:       RecordTypeAuthoredBook,
			expectPanic: true,
		},
		{
			name:        "異常_書誌タイプ不正",
			input:       RecordType("invalid"),
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					tt.input.GetParent()
				})
				return
			}

			actual := tt.input.GetParent()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRecordType_RecordTypeFromId(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		expected    RecordType
		expectPanic bool
	}{
		{
			name:        "正常_単著書",
			input:       RecordTypeIdAuthoredBook,
			expected:    RecordTypeAuthoredBook,
			expectPanic: false,
		},
		{
			name:        "正常_雑誌",
			input:       RecordTypeIdJournal,
			expected:    RecordTypeJournal,
			expectPanic: false,
		},
		{
			name:        "正常_雑誌記事",
			input:       RecordTypeIdJournalArticle,
			expected:    RecordTypeJournalArticle,
			expectPanic: false,
		},
		{
			name:        "正常_編著書",
			input:       RecordTypeIdContributedBook,
			expected:    RecordTypeContributedBook,
			expectPanic: false,
		},
		{
			name:        "正常_編著書記事",
			input:       RecordTypeIdContributedBookArticle,
			expected:    RecordTypeContributedBookArticle,
			expectPanic: false,
		},
		{
			name:        "異常_想定外のID",
			input:       "RT999999",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.expectPanic {
				assert.Panics(t, func() {
					RecordTypeFromId(tt.input)
				})
				return
			}

			actual := RecordTypeFromId(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
