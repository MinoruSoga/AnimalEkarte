package handler

import (
	"fmt"
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/service"
)

const shiftDateLayout = "2006-01-02"

type listShiftEntriesQuery struct {
	Date    string
	StaffID string
}

func newListShiftEntriesQuery(values url.Values) listShiftEntriesQuery {
	return listShiftEntriesQuery{
		Date:    values.Get("date"),
		StaffID: values.Get("staff_id"),
	}
}

type listShiftEntriesFilters struct {
	YearMonth string
	StaffID   *uint64
}

type onDutyStaffsQuery struct {
	Date string
}

func newOnDutyStaffsQuery(values url.Values) onDutyStaffsQuery {
	return onDutyStaffsQuery{Date: values.Get("date")}
}

func (q onDutyStaffsQuery) toDate() (time.Time, error) {
	if q.Date == "" {
		return time.Time{}, fmt.Errorf("date query parameter is required (YYYY-MM-DD)")
	}
	date, err := time.Parse(shiftDateLayout, q.Date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD")
	}
	return date, nil
}

func (q listShiftEntriesQuery) toServiceFilters() (listShiftEntriesFilters, error) {
	staffID, err := parseOptionalUintQueryFilter(q.StaffID, "staff_id")
	if err != nil {
		return listShiftEntriesFilters{}, err
	}
	return listShiftEntriesFilters{
		YearMonth: q.Date,
		StaffID:   staffID,
	}, nil
}

// shiftBreakRequest は休憩時間リクエスト
type shiftBreakRequest struct {
	BreakStart string `json:"break_start" binding:"required"`
	BreakEnd   string `json:"break_end"   binding:"required"`
}

func toShiftBreakInputs(reqs []shiftBreakRequest) []service.ShiftBreakInput {
	breaks := make([]service.ShiftBreakInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, service.ShiftBreakInput{
			BreakStart: req.BreakStart,
			BreakEnd:   req.BreakEnd,
		})
	}
	return breaks
}

// createShiftRequest はシフト登録リクエスト
type createShiftRequest struct {
	StaffID   uint64              `json:"staff_id"   binding:"required"`
	Date      string              `json:"date"       binding:"required"` // YYYY-MM-DD
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	StartTime string              `json:"start_time"`
	EndTime   string              `json:"end_time"`
	Notes     string              `json:"notes"`
	Breaks    []shiftBreakRequest `json:"breaks"`
}

func (r *createShiftRequest) toServiceInput() (*service.CreateShiftEntryInput, error) {
	date, err := time.Parse(shiftDateLayout, r.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: use YYYY-MM-DD")
	}

	var startTime, endTime *string
	if r.StartTime != "" {
		startTime = &r.StartTime
	}
	if r.EndTime != "" {
		endTime = &r.EndTime
	}

	return &service.CreateShiftEntryInput{
		StaffID:   r.StaffID,
		Date:      date,
		ShiftType: r.ShiftType,
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     r.Notes,
		Breaks:    toShiftBreakInputs(r.Breaks),
	}, nil
}

// updateShiftRequest はシフト更新リクエスト（PATCH）
type updateShiftRequest struct {
	ShiftType *string              `json:"shift_type" binding:"omitempty,oneof=full morning afternoon off paid_leave"`
	StartTime *string              `json:"start_time"`
	EndTime   *string              `json:"end_time"`
	Notes     *string              `json:"notes"`
	Breaks    *[]shiftBreakRequest `json:"breaks"`
}

func (r updateShiftRequest) toServiceInput() *service.UpdateShiftEntryInput {
	input := &service.UpdateShiftEntryInput{
		ShiftType: r.ShiftType,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
		Notes:     r.Notes,
	}
	if r.Breaks != nil {
		breaks := toShiftBreakInputs(*r.Breaks)
		input.Breaks = &breaks
	}
	return input
}
