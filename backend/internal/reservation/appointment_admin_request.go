package reservation

import (
	"net/url"
	"time"
)

type listReservationsAdminQuery struct {
	View string
	Date string
}

func newListReservationsAdminQuery(values url.Values, now time.Time) listReservationsAdminQuery {
	view := values.Get("view")
	if view == "" {
		view = "month"
	}
	date := values.Get("date")
	if date == "" {
		date = now.Format("2006-01")
	}
	return listReservationsAdminQuery{View: view, Date: date}
}

type createReservationAdminRequest struct {
	StartTime         time.Time      `json:"start_time"         binding:"required"`
	EndTime           time.Time      `json:"end_time"           binding:"required"`
	OwnerID           *uint64        `json:"owner_id"`
	PetID             *uint64        `json:"pet_id"`
	VisitType         string         `json:"visit_type"`
	ReservationTypeID uint64         `json:"reservation_type_id"    binding:"required"`
	DoctorID          *uint64        `json:"doctor_id"`
	IsDesignated      bool           `json:"is_designated"`
	Notes             string         `json:"notes"`
	LineCustomerID    *uint64        `json:"line_customer_id"`
	IsStaffDelegated  bool           `json:"is_staff_delegated"`
	CustomerFields    jsonRawOrEmpty `json:"customer_fields"`
}

func (r *createReservationAdminRequest) toServiceInput(createdBy uint64) *CreateReservationAdminInput {
	return &CreateReservationAdminInput{
		StartTime:         r.StartTime,
		EndTime:           r.EndTime,
		OwnerID:           r.OwnerID,
		PetID:             r.PetID,
		VisitType:         r.VisitType,
		ReservationTypeID: r.ReservationTypeID,
		DoctorID:          r.DoctorID,
		IsDesignated:      r.IsDesignated,
		Notes:             r.Notes,
		LineCustomerID:    r.LineCustomerID,
		IsStaffDelegated:  r.IsStaffDelegated,
		CustomerFields:    r.CustomerFields,
		CreatedBy:         &createdBy,
	}
}
