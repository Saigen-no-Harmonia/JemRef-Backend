package record

import (
	"fmt"
	"regexp"
	"time"
)

type DatePrecision int

const (
	DatePrecisionDay   = DatePrecision(0)
	DatePrecisionMonth = DatePrecision(1)
	DatePrecisionYear  = DatePrecision(2)
)

var (
	datePattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	monthPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)
	yearPattern  = regexp.MustCompile(`^\d{4}$`)
)

// GetDateAndPrecision 入力された文字列を日時と日時制度に変換して返却する
func GetDateAndPrecision(publishDate *string) (*time.Time, *DatePrecision, error) {
	if publishDate == nil {
		return nil, nil, nil
	}

	var (
		layout    string
		precision DatePrecision
	)

	publishDateStr := *publishDate
	switch {
	case datePattern.MatchString(publishDateStr):
		layout = "2006-01-02"
		precision = DatePrecisionDay

	case monthPattern.MatchString(publishDateStr):
		layout = "2006-01"
		precision = DatePrecisionMonth

	case yearPattern.MatchString(publishDateStr):
		layout = "2006"
		precision = DatePrecisionYear

	default:
		return nil, nil, fmt.Errorf(
			"invalid publication date format: %q",
			publishDateStr,
		)
	}

	date, err := time.Parse(layout, publishDateStr)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"invalid publication date: %w",
			err,
		)
	}

	return &date, &precision, nil
}

// FormatPublicationDate 出版日と精度から、登録時の入力と同じ形式の文字列を返す（GetDateAndPrecision の逆変換）。
// 出版日と精度はDB上で別カラム(どちらもNULL許容)なので、片方だけ設定された不整合データが
// 理論上あり得る。精度が不明なまま日付だけを出力すると月日を捏造することになるため、
// どちらか一方でも未設定、または未知の精度の場合は nil を返す。
func FormatPublicationDate(date *time.Time, precision *DatePrecision) *string {
	if date == nil || precision == nil {
		return nil
	}

	var layout string
	switch *precision {
	case DatePrecisionDay:
		layout = "2006-01-02"

	case DatePrecisionMonth:
		layout = "2006-01"

	case DatePrecisionYear:
		layout = "2006"

	default:
		return nil
	}

	formatted := date.Format(layout)

	return &formatted
}
