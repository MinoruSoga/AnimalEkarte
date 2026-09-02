package reservation

// reservation_request.go — 予約 CRUD 用 query parser。

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type listReservationQuery struct {
	Date      string
	StartDate string
	EndDate   string
	Status    string
	PetID     string
	OwnerID   string
	Source    string
}

type reservationAvailableTimesQuery struct {
	ReservationTypeID string
	StaffID           string
	Date              string
}

func newListReservationQuery(values url.Values) listReservationQuery {
	return listReservationQuery{
		Date:      values.Get("date"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
		Status:    values.Get("status"),
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		Source:    values.Get("source"),
	}
}

func newReservationAvailableTimesQuery(values url.Values) reservationAvailableTimesQuery {
	return reservationAvailableTimesQuery{
		ReservationTypeID: values.Get("reservation_type_id"),
		StaffID:           values.Get("staff_id"),
		Date:              values.Get("date"),
	}
}

type listReservationFilters struct {
	Date      *time.Time
	StartDate *time.Time
	EndDate   *time.Time
	Status    *string
	Source    *string
	PetID     *uint64
	OwnerID   *uint64
}

func (q *listReservationQuery) toServiceFilters() (listReservationFilters, error) {
	var filters listReservationFilters

	if q.Date != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.Date, time.Local)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
		}
		filters.Date = &t
	}
	if q.StartDate != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.StartDate, time.Local)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid start_date format, use YYYY-MM-DD")
		}
		filters.StartDate = &t
	}
	if q.EndDate != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.EndDate, time.Local)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid end_date format, use YYYY-MM-DD")
		}
		// end_date は当日を含めるため翌日 0 時を排他的上限とする
		end := t.AddDate(0, 0, 1)
		filters.EndDate = &end
	}
	if q.Status != "" {
		filters.Status = &q.Status
	}
	if q.PetID != "" {
		id, err := strconv.ParseUint(q.PetID, 10, 64)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid pet_id")
		}
		filters.PetID = &id
	}
	if q.OwnerID != "" {
		id, err := strconv.ParseUint(q.OwnerID, 10, 64)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid owner_id")
		}
		filters.OwnerID = &id
	}
	if q.Source != "" {
		filters.Source = &q.Source
	}

	return filters, nil
}

type reservationAvailableTimesFilters struct {
	ReservationTypeID uint64
	StaffID           uint64
	Date              time.Time
}

func (q reservationAvailableTimesQuery) toServiceFilters() (reservationAvailableTimesFilters, error) {
	reservationTypeID, err := parseRequiredUintQueryFilter(q.ReservationTypeID, "reservation_type_id")
	if err != nil {
		return reservationAvailableTimesFilters{}, err
	}
	date, err := time.ParseInLocation(time.DateOnly, q.Date, time.Local)
	if err != nil {
		return reservationAvailableTimesFilters{}, apperrors.WrapInvalidInput("invalid date: must be YYYY-MM-DD")
	}
	return reservationAvailableTimesFilters{
		ReservationTypeID: reservationTypeID,
		StaffID:           parseOptionalUintQueryValue(q.StaffID),
		Date:              date,
	}, nil
}

// createReservationRequest は予約作成のバインド struct
type createReservationRequest struct {
	StartTime         time.Time `json:"start_time"      binding:"required"`
	EndTime           time.Time `json:"end_time"        binding:"required"`
	OwnerID           *uint64   `json:"owner_id"`
	PetID             *uint64   `json:"pet_id"`
	VisitType         string    `json:"visit_type"          binding:"omitempty,oneof=first revisit"`
	ReservationTypeID uint64    `json:"reservation_type_id" binding:"required"`
	DoctorID          *uint64   `json:"doctor_id"`
	IsDesignated      bool      `json:"is_designated"`
	Status            string    `json:"status"              binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	Notes             string    `json:"notes"`
	Source            string    `json:"source"              binding:"omitempty,oneof=manual line"`
	ReservationRoute  string    `json:"reservation_route"   binding:"omitempty,oneof=line phone reception exam_room record_shortcut"`
}

func (r *createReservationRequest) toServiceInput(clinicID, staffID uint64) (*CreateManualReservationInput, error) {
	source := model.ReservationSourceManual
	if r.Source == string(model.ReservationSourceLine) {
		source = model.ReservationSourceLine
	}

	input := &CreateManualReservationInput{
		ClinicID:          clinicID,
		StartTime:         r.StartTime,
		EndTime:           r.EndTime,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		ReservationTypeID: r.ReservationTypeID,
		DoctorID:          r.DoctorID,
		IsDesignated:      r.IsDesignated,
		Notes:             r.Notes,
		Source:            source,
		CreatedBy:         &staffID,
	}
	if r.ReservationRoute != "" {
		input.ReservationRoute = &r.ReservationRoute
	}
	if r.VisitType != "" {
		vt, err := httpapi.ValidateEnum(r.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid visit_type: %w", err)
		}
		input.VisitType = vt
	}
	if r.Status != "" {
		status, err := httpapi.ValidateEnum(r.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		input.Status = status
	}
	return input, nil
}

type createReservationBatchPetRequest struct {
	OwnerID uint64 `json:"owner_id" binding:"required"`
	PetID   uint64 `json:"pet_id" binding:"required"`
}

type createReservationBatchRequest struct {
	StartTime         time.Time                          `json:"start_time" binding:"required"`
	EndTime           time.Time                          `json:"end_time" binding:"required"`
	ReservationTypeID uint64                             `json:"reservation_type_id" binding:"required"`
	VisitType         string                             `json:"visit_type" binding:"required"`
	DoctorID          *uint64                            `json:"doctor_id"`
	IsDesignated      bool                               `json:"is_designated"`
	Status            string                             `json:"status" binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	Notes             string                             `json:"notes"`
	Source            string                             `json:"source" binding:"omitempty,oneof=manual line"`
	ReservationRoute  string                             `json:"reservation_route" binding:"omitempty,oneof=line phone reception exam_room record_shortcut"`
	Pets              []createReservationBatchPetRequest `json:"pets" binding:"required,min=2"`
}

func (r *createReservationBatchRequest) toServiceInput(clinicID, staffID uint64) (*CreateManualReservationInput, []ReservationBatchPet, error) {
	input, err := (&createReservationRequest{StartTime: r.StartTime, EndTime: r.EndTime, ReservationTypeID: r.ReservationTypeID, VisitType: r.VisitType, DoctorID: r.DoctorID, IsDesignated: r.IsDesignated, Status: r.Status, Notes: r.Notes, Source: r.Source, ReservationRoute: r.ReservationRoute}).toServiceInput(clinicID, staffID)
	if err != nil {
		return nil, nil, err
	}
	pets := make([]ReservationBatchPet, len(r.Pets))
	for i, pet := range r.Pets {
		pets[i] = ReservationBatchPet(pet)
	}
	return input, pets, nil
}

// patchReservationReservationRouteRequest は予約経路更新リクエスト（FEAT-381-2）。
type patchReservationReservationRouteRequest struct {
	Route string `json:"route" binding:"max=20"`
}

func (r patchReservationReservationRouteRequest) toServiceInput() UpdateReservationRouteInput {
	return UpdateReservationRouteInput(r)
}

// updateReservationRequest は予約更新のバインド struct
type updateReservationRequest struct {
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	OwnerID           *uint64    `json:"owner_id"`
	PetID             *uint64    `json:"pet_id"`
	VisitType         *string    `json:"visit_type"           binding:"omitempty,oneof=first revisit"`
	ReservationTypeID *uint64    `json:"reservation_type_id"`
	DoctorID          *uint64    `json:"doctor_id"`
	IsDesignated      *bool      `json:"is_designated"`
	Status            *string    `json:"status"               binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting completed"`
	Notes             *string    `json:"notes"`
}

func (r *updateReservationRequest) toServiceInput() (UpdateReservationInput, error) {
	input := UpdateReservationInput{
		StartTime:         r.StartTime,
		EndTime:           r.EndTime,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		ReservationTypeID: r.ReservationTypeID,
		DoctorID:          r.DoctorID,
		IsDesignated:      r.IsDesignated,
		Notes:             r.Notes,
	}
	if r.VisitType != nil {
		vt, err := httpapi.ValidateEnum(*r.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			return UpdateReservationInput{}, fmt.Errorf("invalid visit_type: %w", err)
		}
		input.VisitType = &vt
	}
	if r.Status != nil {
		status, err := httpapi.ValidateEnum(*r.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			return UpdateReservationInput{}, fmt.Errorf("invalid status: %w", err)
		}
		input.Status = &status
	}
	return input, nil
}

// parseRequiredUintQueryFilter / parseOptionalUintQueryValue parse reservation query IDs.
func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func parseRequiredUintQueryFilter(value, field string) (uint64, error) {
	id, err := parseOptionalUintQueryFilter(value, field)
	if err != nil || id == nil || *id == 0 {
		return 0, apperrors.WrapInvalidInput("invalid " + field)
	}
	return *id, nil
}

func parseOptionalUintQueryValue(value string) uint64 {
	if value == "" {
		return 0
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
