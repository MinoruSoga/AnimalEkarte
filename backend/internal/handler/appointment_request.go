package handler

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type listReservationQuery struct {
	Date    string
	Status  string
	PetID   string
	OwnerID string
	Source  string
}

func newListReservationQuery(values url.Values) listReservationQuery {
	return listReservationQuery{
		Date:    values.Get("date"),
		Status:  values.Get("status"),
		PetID:   values.Get("pet_id"),
		OwnerID: values.Get("owner_id"),
		Source:  values.Get("source"),
	}
}

type listReservationFilters struct {
	Date    *time.Time
	Status  *string
	Source  *string
	PetID   *uint64
	OwnerID *uint64
}

func (q listReservationQuery) toServiceFilters() (listReservationFilters, error) {
	var filters listReservationFilters

	if q.Date != "" {
		t, err := time.Parse("2006-01-02", q.Date)
		if err != nil {
			return listReservationFilters{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
		}
		filters.Date = &t
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

type reservationAvailableTimesQuery struct {
	ReservationTypeID string
	StaffID           string
	Date              string
}

func newReservationAvailableTimesQuery(values url.Values) reservationAvailableTimesQuery {
	return reservationAvailableTimesQuery{
		ReservationTypeID: values.Get("reservation_type_id"),
		StaffID:           values.Get("staff_id"),
		Date:              values.Get("date"),
	}
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
	date, err := time.ParseInLocation("2006-01-02", q.Date, time.Local)
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
}

func (r createReservationRequest) toServiceInput(clinicID, staffID uint64) (*service.CreateManualReservationInput, error) {
	source := model.ReservationSourceManual
	if r.Source == string(model.ReservationSourceLine) {
		source = model.ReservationSourceLine
	}

	input := &service.CreateManualReservationInput{
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
	if r.VisitType != "" {
		vt, err := validateEnum(r.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid visit_type: %w", err)
		}
		input.VisitType = vt
	}
	if r.Status != "" {
		status, err := validateEnum(r.Status,
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

// patchReservationReservationRouteRequest は予約経路更新リクエスト（FEAT-381-2）。
type patchReservationReservationRouteRequest struct {
	Route string `json:"route" binding:"max=20"`
}

func (r patchReservationReservationRouteRequest) toServiceInput() service.UpdateReservationRouteInput {
	return service.UpdateReservationRouteInput{Route: r.Route}
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

func (r updateReservationRequest) toServiceInput() (service.UpdateReservationInput, error) {
	input := service.UpdateReservationInput{
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
		vt, err := validateEnum(*r.VisitType,
			model.VisitTypeFirst,
			model.VisitTypeRevisit,
		)
		if err != nil {
			return service.UpdateReservationInput{}, fmt.Errorf("invalid visit_type: %w", err)
		}
		input.VisitType = &vt
	}
	if r.Status != nil {
		status, err := validateEnum(*r.Status,
			model.ReservationStatusConfirmed,
			model.ReservationStatusPending,
			model.ReservationStatusCancelled,
			model.ReservationStatusCheckedIn,
			model.ReservationStatusInConsultation,
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
		)
		if err != nil {
			return service.UpdateReservationInput{}, fmt.Errorf("invalid status: %w", err)
		}
		input.Status = &status
	}
	return input, nil
}
