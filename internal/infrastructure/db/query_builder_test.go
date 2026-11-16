package db

import (
	"testing"

	"jemref_go/internal/domain/record"

	"github.com/stretchr/testify/assert"
)

// -------------SQL組み立てヘルパのテスト-------------
//
// 実DBを使わない純粋な文字列組み立てのため、テーブル駆動とする。

func TestBuildPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:     "正常_1件",
			count:    1,
			expected: "?",
		},
		{
			name:     "正常_2件",
			count:    2,
			expected: "?, ?",
		},
		{
			name:     "正常_3件",
			count:    3,
			expected: "?, ?, ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildPlaceholders(tt.count)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestBuildPlaceholders_異常 IN ()は妥当なSQLではないため、0件はバグとして落とす
func TestBuildPlaceholders_異常(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "異常_0件",
			count: 0,
		},
		{
			name:  "異常_負数",
			count: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				buildPlaceholders(tt.count)
			})
		})
	}
}

func TestBuildBulkInsertQuery(t *testing.T) {
	const (
		base   = "INSERT INTO t (A, B) VALUES "
		values = "(?, ?)"
	)

	tests := []struct {
		name     string
		rows     int
		expected string
	}{
		{
			name:     "正常_1行",
			rows:     1,
			expected: base + "(?, ?)",
		},
		{
			name:     "正常_2行",
			rows:     2,
			expected: base + "(?, ?), (?, ?)",
		},
		{
			name:     "正常_3行",
			rows:     3,
			expected: base + "(?, ?), (?, ?), (?, ?)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildBulkInsertQuery(base, values, tt.rows)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestBuildBulkInsertQuery_異常 VALUES句のないINSERTは妥当なSQLではないため、0件はバグとして落とす
func TestBuildBulkInsertQuery_異常(t *testing.T) {
	tests := []struct {
		name string
		rows int
	}{
		{
			name: "異常_0行",
			rows: 0,
		},
		{
			name: "異常_負数",
			rows: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				buildBulkInsertQuery("INSERT INTO t (A) VALUES ", "(?)", tt.rows)
			})
		})
	}
}

// TestBuildRecordListOrderBy 並び替えキー×並び順から、ホワイトリストの固定句を引けること
func TestBuildRecordListOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		key      record.SortKey
		order    record.SortOrder
		expected string
	}{
		{
			name:     "正常_出版年昇順",
			key:      record.SortKeyPublishYear,
			order:    record.SortOrderAsc,
			expected: recordListOrderByPublishYearAsc,
		},
		{
			name:     "正常_出版年降順",
			key:      record.SortKeyPublishYear,
			order:    record.SortOrderDesc,
			expected: recordListOrderByPublishYearDesc,
		},
		{
			name:     "正常_登録日昇順",
			key:      record.SortKeyInsertDate,
			order:    record.SortOrderAsc,
			expected: recordListOrderByInsertDateAsc,
		},
		{
			name:     "正常_登録日降順",
			key:      record.SortKeyInsertDate,
			order:    record.SortOrderDesc,
			expected: recordListOrderByInsertDateDesc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildRecordListOrderBy(tt.key, tt.order)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestBuildRecordListOrderBy_異常 ホワイトリストにない値はSQLに連結せず、バグとして落とす
func TestBuildRecordListOrderBy_異常(t *testing.T) {
	tests := []struct {
		name  string
		key   record.SortKey
		order record.SortOrder
	}{
		{
			name:  "異常_未知のSortKey",
			key:   record.SortKey("MAIN_TITLE; drop table e_records"),
			order: record.SortOrderAsc,
		},
		{
			name:  "異常_未知のSortOrder",
			key:   record.SortKeyPublishYear,
			order: record.SortOrder("ASC"),
		},
		{
			name:  "異常_SortKeyが空",
			key:   record.SortKey(""),
			order: record.SortOrderAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				buildRecordListOrderBy(tt.key, tt.order)
			})
		})
	}
}

// TestEscapeLike LIKEの特殊文字(! % _)を、エスケープ文字 '!' でエスケープすること
func TestEscapeLike(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "正常_特殊文字なし",
			input:    "学校",
			expected: "学校",
		},
		{
			name:     "正常_パーセント",
			input:    "100%",
			expected: "100!%",
		},
		{
			name:     "正常_アンダースコア",
			input:    "a_b",
			expected: "a!_b",
		},
		{
			name:     "正常_エスケープ文字",
			input:    "注意!",
			expected: "注意!!",
		},
		{
			// 置換で付けた '!' が再び置換されないこと(!!!%!_ が正。順序の誤りがあると !!!!% 等になる)
			name:     "正常_混在",
			input:    "!%_",
			expected: "!!!%!_",
		},
		{
			name:     "正常_空文字",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := escapeLike(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
