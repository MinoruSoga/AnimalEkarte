package staff

import (
	"fmt"
	"net/url"
	"time"
)

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
	date, err := time.ParseInLocation(time.DateOnly, q.Date, time.Local)
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
	BreakStart string `json:"break_start" binding:"required,max=8"`
	BreakEnd   string `json:"break_end"   binding:"required,max=8"`
}

func toShiftBreakInputs(reqs []shiftBreakRequest) []ShiftBreakInput {
	breaks := make([]ShiftBreakInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, ShiftBreakInput{
			BreakStart: req.BreakStart,
			BreakEnd:   req.BreakEnd,
		})
	}
	return breaks
}

// createShiftRequest はシフト登録リクエスト
type createShiftRequest struct {
	StaffID   uint64              `json:"staff_id"   binding:"required"`
	Date      string              `json:"date"       binding:"required,max=10"` // YYYY-MM-DD
	ShiftType string              `json:"shift_type" binding:"required,max=10,oneof=full morning afternoon off paid_leave"`
	StartTime string              `json:"start_time" binding:"max=8"`
	EndTime   string              `json:"end_time"   binding:"max=8"`
	Notes     string              `json:"notes"      binding:"max=2000"`
	Breaks    []shiftBreakRequest `json:"breaks"     binding:"max=50,dive"`
}

func (r *createShiftRequest) toServiceInput() (*CreateShiftEntryInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
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

	return &CreateShiftEntryInput{
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
	ShiftType *string              `json:"shift_type" binding:"omitempty,max=10,oneof=full morning afternoon off paid_leave"`
	StartTime *string              `json:"start_time" binding:"omitempty,max=8"`
	EndTime   *string              `json:"end_time"   binding:"omitempty,max=8"`
	Notes     *string              `json:"notes"      binding:"omitempty,max=2000"`
	Breaks    *[]shiftBreakRequest `json:"breaks"     binding:"omitempty,max=50,dive"`
}

func (r updateShiftRequest) toServiceInput() *UpdateShiftEntryInput {
	input := &UpdateShiftEntryInput{
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
