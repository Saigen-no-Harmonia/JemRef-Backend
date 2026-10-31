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
