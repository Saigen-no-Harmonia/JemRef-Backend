package record

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatePrecision_GetDateAndPrecision(t *testing.T) {

	datePrecisionDay := DatePrecisionDay
	datePrecisionMonth := DatePrecisionMonth
	datePrecisionYear := DatePrecisionYear

	tests := []struct {
		name              string
		input             *string
		expectedDate      *time.Time
		expectedPrecision *DatePrecision
		expectError       bool
		// expectedErrorMsg エラーメッセージの部分一致(フォーマット不正とパース失敗の分岐を区別する)
		expectedErrorMsg string
	}{
		{
			name:              "正常_未入力",
			input:             nil,
			expectedDate:      nil,
			expectedPrecision: nil,
			expectError:       false,
		},
		{
			name:              "正常_年月日",
			input:             dateStrPtr("2022-11-30"),
			expectedDate:      timePtr(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
			expectedPrecision: &datePrecisionDay,
			expectError:       false,
		},
		{
			name:              "正常_年月",
			input:             dateStrPtr("2022-11"),
			expectedDate:      timePtr(time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)),
			expectedPrecision: &datePrecisionMonth,
			expectError:       false,
		},
		{
			name:              "正常_年のみ",
			input:             dateStrPtr("2022"),
			expectedDate:      timePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			expectedPrecision: &datePrecisionYear,
			expectError:       false,
		},
		{
			name:             "異常_区切り文字がスラッシュ",
			input:            dateStrPtr("2022/11/30"),
			expectError:      true,
			expectedErrorMsg: "invalid publication date format",
		},
		{
			name:             "異常_空文字",
			input:            dateStrPtr(""),
			expectError:      true,
			expectedErrorMsg: "invalid publication date format",
		},
		{
			name:             "異常_桁数不足",
			input:            dateStrPtr("202"),
			expectError:      true,
			expectedErrorMsg: "invalid publication date format",
		},
		{
			name:             "異常_桁数超過",
			input:            dateStrPtr("2022-11-30-01"),
			expectError:      true,
			expectedErrorMsg: "invalid publication date format",
		},
		{
			// 書式は一致するが日付として成立しないため、time.Parseで失敗する
			name:             "異常_日付として不正",
			input:            dateStrPtr("2022-13-45"),
			expectError:      true,
			expectedErrorMsg: "invalid publication date:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualDate, actualPrecision, err := GetDateAndPrecision(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Nil(t, actualDate)
				assert.Nil(t, actualPrecision)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedDate, actualDate)
			assert.Equal(t, tt.expectedPrecision, actualPrecision)
		})
	}
}

// dateStrPtr string型のポインタを返却する
func dateStrPtr(s string) *string {
	return &s
}

// timePtr time.Time型のポインタを返却する
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestDatePrecision_FormatPublicationDate(t *testing.T) {

	datePrecisionDay := DatePrecisionDay
	datePrecisionMonth := DatePrecisionMonth
	datePrecisionYear := DatePrecisionYear
	// 未知の精度。DATE_PRECISIONはFKを持たないTINYINTのため不整合データが入り得る
	datePrecisionUnknown := DatePrecision(99)

	tests := []struct {
		name      string
		date      *time.Time
		precision *DatePrecision
		expected  *string
	}{
		{
			name:      "正常_日精度",
			date:      timePtr(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionDay,
			expected:  dateStrPtr("2022-11-30"),
		},
		{
			name:      "正常_月精度",
			date:      timePtr(time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionMonth,
			expected:  dateStrPtr("2022-11"),
		},
		{
			name:      "正常_年精度",
			date:      timePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionYear,
			expected:  dateStrPtr("2022"),
		},
		{
			// 登録時の入力と同じ形式で返すため、月日は2桁にゼロ埋めする
			name:      "正常_月日が1桁ならゼロ埋めする",
			date:      timePtr(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionDay,
			expected:  dateStrPtr("2022-01-05"),
		},
		{
			// 日精度でも月日を切り捨てる(精度が優先される)
			name:      "正常_月精度なら日を切り捨てる",
			date:      timePtr(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionMonth,
			expected:  dateStrPtr("2022-11"),
		},
		{
			name:      "正常_日付がnil",
			date:      nil,
			precision: &datePrecisionDay,
			expected:  nil,
		},
		{
			// 精度が分からないまま日付を出すと月日を捏造することになるためnilを返す
			name:      "正常_精度がnil",
			date:      timePtr(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
			precision: nil,
			expected:  nil,
		},
		{
			name:      "正常_日付と精度の両方がnil",
			date:      nil,
			precision: nil,
			expected:  nil,
		},
		{
			name:      "正常_未知の精度",
			date:      timePtr(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
			precision: &datePrecisionUnknown,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := FormatPublicationDate(tt.date, tt.precision)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestDatePrecision_FormatPublicationDate_ラウンドトリップ
// 詳細参照のレスポンスは更新リクエストのベースに使われるため、
// GetDateAndPrecision と FormatPublicationDate で入力が復元できる必要がある。
func TestDatePrecision_FormatPublicationDate_ラウンドトリップ(t *testing.T) {

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "正常_年月日",
			input: "2022-11-30",
		},
		{
			name:  "正常_年月",
			input: "2022-11",
		},
		{
			name:  "正常_年のみ",
			input: "2022",
		},
		{
			name:  "正常_月日が1桁",
			input: "2022-01-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			date, precision, err := GetDateAndPrecision(dateStrPtr(tt.input))
			require.NoError(t, err)

			actual := FormatPublicationDate(date, precision)

			require.NotNil(t, actual)
			assert.Equal(t, tt.input, *actual)
		})
	}
}
