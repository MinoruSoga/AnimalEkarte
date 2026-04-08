package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type shiftEntryBreakResponse struct {
	ID         uint64 `json:"id"`
	BreakStart string `json:"break_start"`
	BreakEnd   string `json:"break_end"`
}

type scheduleEntryResponse struct {
	ID        uint64                    `json:"id"`
	ClinicID  uint64                    `json:"clinic_id"`
	StaffID   uint64                    `json:"staff_id"`
	Date      string                    `json:"date"`
	ShiftType string                    `json:"shift_type"`
	StartTime *string                   `json:"start_time,omitempty"`
	EndTime   *string                   `json:"end_time,omitempty"`
	Note      string                    `json:"note"`
	Breaks    []shiftEntryBreakResponse `json:"breaks"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

func toScheduleEntryResponse(e service.ScheduleEntry) scheduleEntryResponse {
	breaks := make([]shiftEntryBreakResponse, 0, len(e.Breaks))
	for _, b := range e.Breaks {
		breaks = append(breaks, toShiftEntryBreakResponse(b))
	}
	return scheduleEntryResponse{
		ID:        e.Entry.ID,
		ClinicID:  e.Entry.ClinicID,
		StaffID:   e.Entry.StaffID,
		Date:      e.Entry.Date.Format("2006-01-02"),
		ShiftType: string(e.Entry.ShiftType),
		StartTime: e.Entry.StartTime,
		EndTime:   e.Entry.EndTime,
		Note:      e.Entry.Note,
		Breaks:    breaks,
		CreatedAt: e.Entry.CreatedAt,
		UpdatedAt: e.Entry.UpdatedAt,
	}
}

func toShiftEntryBreakResponse(b model.ShiftEntryBreak) shiftEntryBreakResponse {
	return shiftEntryBreakResponse{
		ID:         b.ID,
		BreakStart: b.BreakStart,
		BreakEnd:   b.BreakEnd,
	}
}
