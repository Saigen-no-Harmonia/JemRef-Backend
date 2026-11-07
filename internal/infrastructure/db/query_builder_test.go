package db

import (
	"testing"

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
