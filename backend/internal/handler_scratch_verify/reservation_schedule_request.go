package handler

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/service"
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

func toReservationScheduleBreakInputs(reqs []breakInputRequest) []service.ReservationScheduleBreakInput {
	breaks := make([]service.ReservationScheduleBreakInput, 0, len(reqs))
	for _, req := range reqs {
		breaks = append(breaks, service.ReservationScheduleBreakInput{
			Start: req.Start,
			End:   req.End,
		})
	}
	return breaks
}

type upsertReservationScheduleRequest struct {
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	WorkStart *string             `json:"work_start"`
	WorkEnd   *string             `json:"work_end"`
	Breaks    []breakInputRequest `json:"breaks"`
}

func (r upsertReservationScheduleRequest) toServiceInput() *service.CreateReservationScheduleInput {
	return &service.CreateReservationScheduleInput{
		ShiftType: r.ShiftType,
		WorkStart: r.WorkStart,
		WorkEnd:   r.WorkEnd,
		Breaks:    toReservationScheduleBreakInputs(r.Breaks),
	}
}
