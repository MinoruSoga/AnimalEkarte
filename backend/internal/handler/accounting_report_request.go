package handler

import (
	"net/url"
	"strconv"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

type monthlyReportQuery struct {
	Year  string
	Month string
}

func newMonthlyReportQuery(values url.Values) monthlyReportQuery {
	return monthlyReportQuery{
		Year:  values.Get("year"),
		Month: values.Get("month"),
	}
}

func (q monthlyReportQuery) toYearMonth(now time.Time) (year, month int, err error) {
	yearStr := q.Year
	if yearStr == "" {
		yearStr = strconv.Itoa(now.Year())
	}
	monthStr := q.Month
	if monthStr == "" {
		monthStr = strconv.Itoa(int(now.Month()))
	}

	year, err = strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, apperrors.WrapInvalidInput("year は 2000〜2100 の整数で指定してください")
	}
	month, err = strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, apperrors.WrapInvalidInput("month は 1〜12 の整数で指定してください")
	}
	return year, month, nil
}
