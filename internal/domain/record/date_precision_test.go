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
