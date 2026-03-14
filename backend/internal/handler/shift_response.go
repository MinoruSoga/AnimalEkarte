package handler

import (
	"strconv"

	"github.com/animal-ekarte/backend/internal/model"
)

// shiftResponse はシフトエントリのレスポンス
type shiftResponse struct {
	ID        string               `json:"id"`
	ClinicID  string               `json:"clinic_id"`
	StaffID   string               `json:"staff_id"`
	StaffName string               `json:"staff_name,omitempty"`
	Date      string               `json:"date"` // YYYY-MM-DD
	ShiftType model.ShiftType      `json:"shift_type"`
	StartTime string               `json:"start_time"`
	EndTime   string               `json:"end_time"`
	Note      string               `json:"note"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
	Staff     *staffSummaryResponse `json:"staff,omitempty"`
}

func toShiftResponse(s *model.ShiftEntry) shiftResponse {
	r := shiftResponse{
		ID:        strconv.FormatUint(s.ID, 10),
		ClinicID:  strconv.FormatUint(s.ClinicID, 10),
		StaffID:   strconv.FormatUint(s.StaffID, 10),
		Date:      s.Date.Format("2006-01-02"),
		ShiftType: s.ShiftType,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Note:      s.Note,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.Staff.ID != 0 {
		r.StaffName = s.Staff.Name
		r.Staff = &staffSummaryResponse{ID: s.Staff.ID, Name: s.Staff.Name}
	}
	return r
}

func toShiftResponseList(shifts []model.ShiftEntry) []shiftResponse {
	list := make([]shiftResponse, 0, len(shifts))
	for i := range shifts {
		list = append(list, toShiftResponse(&shifts[i]))
	}
	return list
}
