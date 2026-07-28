package reservation

import (
	"net/url"
	"time"
)

type listReservationSchedulesQuery struct {
	Month string
}

func newListReservationSchedulesQuery(values url.Values, now time.Time) listReservationSchedulesQuery {
	month := values.Get("month")
	if month == "" {
		month = now.Format("2006-01")
	}
	return listReservationSchedulesQuery{Month: month}
}

type breakInputRequest struct {
	Start string `json:"start" binding:"required"`
	End   string `json:"end"   binding:"required"`
}

func toReservationScheduleBreakInputs(reqs []breakInputRequest) []ReservationScheduleBreakInput {
	breaks := make([]ReservationScheduleBreakInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, ReservationScheduleBreakInput(req))
	}
	return breaks
}

type upsertReservationScheduleRequest struct {
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	WorkStart *string             `json:"work_start"`
	WorkEnd   *string             `json:"work_end"`
	Breaks    []breakInputRequest `json:"breaks"`
}

func (r upsertReservationScheduleRequest) toServiceInput() *CreateReservationScheduleInput {
	return &CreateReservationScheduleInput{
		ShiftType: r.ShiftType,
		WorkStart: r.WorkStart,
		WorkEnd:   r.WorkEnd,
		Breaks:    toReservationScheduleBreakInputs(r.Breaks),
	}
}
