package reservation

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type liffCourseQuery struct {
	CourseID string
}

func newLiffCourseQuery(values url.Values) liffCourseQuery {
	return liffCourseQuery{CourseID: values.Get("courseId")}
}

func (q liffCourseQuery) toCourseID() (uint64, error) {
	return parseRequiredUintQueryFilter(q.CourseID, "courseId")
}

type liffAvailableDatesQuery struct {
	CourseID string
	StaffID  string
}

func newLiffAvailableDatesQuery(values url.Values) liffAvailableDatesQuery {
	return liffAvailableDatesQuery{
		CourseID: values.Get("courseId"),
		StaffID:  values.Get("staffId"),
	}
}

type liffAvailableDatesFilters struct {
	CourseID uint64
	StaffID  uint64
}

func (q liffAvailableDatesQuery) toServiceFilters() (liffAvailableDatesFilters, error) {
	courseID, err := parseRequiredUintQueryFilter(q.CourseID, "courseId")
	if err != nil {
		return liffAvailableDatesFilters{}, err
	}
	return liffAvailableDatesFilters{
		CourseID: courseID,
		StaffID:  parseOptionalUintQueryValue(q.StaffID),
	}, nil
}

type liffAvailableTimesQuery struct {
	CourseID string
	StaffID  string
	Date     string
}

func newLiffAvailableTimesQuery(values url.Values) liffAvailableTimesQuery {
	return liffAvailableTimesQuery{
		CourseID: values.Get("courseId"),
		StaffID:  values.Get("staffId"),
		Date:     values.Get("date"),
	}
}

type liffAvailableTimesFilters struct {
	CourseID uint64
	StaffID  uint64
	Date     time.Time
}

func (q liffAvailableTimesQuery) toServiceFilters() (liffAvailableTimesFilters, error) {
	courseID, err := parseRequiredUintQueryFilter(q.CourseID, "courseId")
	if err != nil {
		return liffAvailableTimesFilters{}, err
	}
	date, err := time.ParseInLocation(time.DateOnly, q.Date, time.Local)
	if err != nil {
		return liffAvailableTimesFilters{}, apperrors.WrapInvalidInput("invalid date: must be YYYY-MM-DD")
	}
	return liffAvailableTimesFilters{
		CourseID: courseID,
		StaffID:  parseOptionalUintQueryValue(q.StaffID),
		Date:     date,
	}, nil
}

// liffCreateReservationRequest は予約確定リクエスト。
type liffCreateReservationRequest struct {
	TypeID               uint64          `json:"course_id"        binding:"required"`
	StaffID              uint64          `json:"staff_id"`                            // 0 = 指名なし
	Date                 string          `json:"date"             binding:"required"` // "YYYY-MM-DD"
	StartTime            string          `json:"start_time"       binding:"required"` // "HHMM"
	EndTime              string          `json:"end_time"         binding:"required"` // "HHMM"
	CustomerFields       json.RawMessage `json:"customer_fields"`
	RequestText          string          `json:"request_text"`
	TrimmingCourseID     *uint64         `json:"trimming_course_id"`     // BE-120
	TrimmingOptionIDs    []uint64        `json:"trimming_option_ids"`    // BE-120
	TrimmingStyleRequest string          `json:"trimming_style_request"` // BE-120
}

func (r *liffCreateReservationRequest) toServiceInput() (*CreateReservationInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid date: must be YYYY-MM-DD")
	}

	return &CreateReservationInput{
		ReservationTypeID:    r.TypeID,
		StaffID:              r.StaffID,
		Date:                 date,
		StartTime:            r.StartTime,
		EndTime:              r.EndTime,
		CustomerFields:       r.CustomerFields,
		RequestText:          r.RequestText,
		TrimmingCourseID:     r.TrimmingCourseID,
		TrimmingOptionIDs:    r.TrimmingOptionIDs,
		TrimmingStyleRequest: r.TrimmingStyleRequest,
	}, nil
}
