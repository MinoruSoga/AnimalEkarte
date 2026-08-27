package billing

import (
	"fmt"
	"net/url"
	"strings"
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

// voidCashRegisterCloseRequest はレジ締め特権取消リクエスト。
type voidCashRegisterCloseRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (r voidCashRegisterCloseRequest) toServiceInput(id, staffID uint64) (VoidReopenInput, error) {
	reason := strings.TrimSpace(r.Reason)
	if reason == "" {
		return VoidReopenInput{}, fmt.Errorf("reason は必須です")
	}
	if id == 0 {
		return VoidReopenInput{}, fmt.Errorf("id は必須です")
	}
	if staffID == 0 {
		return VoidReopenInput{}, fmt.Errorf("authenticated staff is required")
	}
	return VoidReopenInput{
		ID:      id,
		Reason:  reason,
		ActorID: staffID,
	}, nil
}
