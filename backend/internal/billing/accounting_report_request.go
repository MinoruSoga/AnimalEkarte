package billing

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

type monthlyReportQuery struct {
	Year      string
	Month     string
	StartDate string
	EndDate   string
}

func newMonthlyReportQuery(values url.Values) monthlyReportQuery {
	return monthlyReportQuery{
		Year:      values.Get("year"),
		Month:     values.Get("month"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

// hasPeriod は期間モード（start_date/end_date 指定）かを判定する。
// OR 条件は意図的: いずれか一方でも指定されたら期間モードとみなし、toPeriod() で
// 両方必須を強制して 400 を返す。AND にすると片方欠落時に年月モードへ silent fallback し、
// ユーザーが指定した日付が黙殺される（silent failure）ため OR が正しい設計。
func (q monthlyReportQuery) hasPeriod() bool {
	return q.StartDate != "" || q.EndDate != ""
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

func (q monthlyReportQuery) toPeriod() (startDate, endDate time.Time, err error) {
	if q.StartDate == "" || q.EndDate == "" {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("start_date と end_date は両方指定してください")
	}
	startDate, err = time.ParseInLocation(time.DateOnly, q.StartDate, config.JST)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
	}
	endDate, err = time.ParseInLocation(time.DateOnly, q.EndDate, config.JST)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
	}
	return startDate, endDate, nil
}
