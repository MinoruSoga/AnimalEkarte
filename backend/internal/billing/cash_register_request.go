package billing

import (
	"fmt"
	"net/url"

	"time"
)

type listCashRegisterClosesQuery struct {
	StartDate string
	EndDate   string
}

func newListCashRegisterClosesQuery(values url.Values) listCashRegisterClosesQuery {
	return listCashRegisterClosesQuery{
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

type listCashRegisterClosesFilters struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type cashRegisterPreviewQuery struct {
	Date   string
	Period string
}

func newCashRegisterPreviewQuery(values url.Values) cashRegisterPreviewQuery {
	return cashRegisterPreviewQuery{
		Date:   values.Get("date"),
		Period: values.Get("period"),
	}
}

func (q listCashRegisterClosesQuery) toServiceFilters() (listCashRegisterClosesFilters, error) {
	startDate, err := parseOptionalDateTimeQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listCashRegisterClosesFilters{}, err
	}
	endDate, err := parseOptionalDateTimeQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listCashRegisterClosesFilters{}, err
	}
	return listCashRegisterClosesFilters{
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// closeCashRegisterRequest はレジ締め実行リクエスト。
type closeCashRegisterRequest struct {
	Date       string `json:"date"        binding:"required"` // YYYY-MM-DD
	Period     string `json:"period"      binding:"required"` // "am", "pm", or "emg"
	ActualCash int64  `json:"actual_cash"`
	Memo       string `json:"memo"`
}

func (r closeCashRegisterRequest) toServiceInput(staffID uint64) (CloseRegisterInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
	if err != nil {
		return CloseRegisterInput{}, fmt.Errorf("date は YYYY-MM-DD 形式で指定してください")
	}
	// FE CashRegisterClosePage min=0 と整合。負の実査現金は無効（V02-F04）。
	if r.ActualCash < 0 {
		return CloseRegisterInput{}, fmt.Errorf("actual_cash は 0 以上で指定してください")
	}

	return CloseRegisterInput{
		Date:       date,
		Period:     r.Period,
		ActualCash: r.ActualCash,
		Memo:       r.Memo,
		ClosedBy:   &staffID,
	}, nil
}
