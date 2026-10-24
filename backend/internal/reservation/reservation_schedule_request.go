package reservation

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// parseScheduleTimestamp は work_start / work_end / breaks[].start / end の RFC 3339 date-time（オフセット付き）を解析する。
// 同日の時刻文字列（HH:MM）は受け付けない。時刻から日付や翌日を推測しない。
func parseScheduleTimestamp(raw string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("work_start, work_end and breaks must be RFC 3339 date-time with an offset")
	}
	return parsed, nil
}

// parseOptionalScheduleTimestamp は nil / 空文字を「指定なし」として nil を返す。
func parseOptionalScheduleTimestamp(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := parseScheduleTimestamp(*raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toReservationScheduleBreakInputs(reqs []breakInputRequest) ([]ReservationScheduleBreakInput, error) {
	breaks := make([]ReservationScheduleBreakInput, 0, len(reqs))
	for _, req := range reqs {
		startsAt, err := parseScheduleTimestamp(req.Start)
		if err != nil {
			return nil, err
		}
		endsAt, err := parseScheduleTimestamp(req.End)
		if err != nil {
			return nil, err
		}
		breaks = append(breaks, ReservationScheduleBreakInput{StartsAt: startsAt, EndsAt: endsAt})
	}
	return breaks, nil
}

type upsertReservationScheduleRequest struct {
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	WorkStart *string             `json:"work_start"`
	WorkEnd   *string             `json:"work_end"`
	Breaks    []breakInputRequest `json:"breaks"`
}

// toServiceInput は work_start / work_end / breaks の RFC 3339 date-time を明示日時として service へ渡す。
func (r upsertReservationScheduleRequest) toServiceInput() (*CreateReservationScheduleInput, error) {
	startsAt, err := parseOptionalScheduleTimestamp(r.WorkStart)
	if err != nil {
		return nil, err
	}
	endsAt, err := parseOptionalScheduleTimestamp(r.WorkEnd)
	if err != nil {
		return nil, err
	}
	breaks, err := toReservationScheduleBreakInputs(r.Breaks)
	if err != nil {
		return nil, err
	}
	return &CreateReservationScheduleInput{
		ShiftType: r.ShiftType,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Breaks:    breaks,
	}, nil
}
