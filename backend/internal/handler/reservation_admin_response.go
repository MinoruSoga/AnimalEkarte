package handler

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// reservationSummaryResponse は月表示用サマリ
type reservationSummaryResponse struct {
	ID               uint64    `json:"id"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	CustomerName     string    `json:"customer_name"`
	CourseShortName  string    `json:"course_short_name"`
	StaffName        string    `json:"staff_name"`
	IsStaffDelegated bool      `json:"is_staff_delegated"`
	Source           string    `json:"source"`
	Status           string    `json:"status"`
}

// reservationDetailResponse は日表示用詳細
type reservationDetailResponse struct {
	ID               uint64          `json:"id"`
	StartTime        time.Time       `json:"start_time"`
	EndTime          time.Time       `json:"end_time"`
	OwnerID          *uint64         `json:"owner_id,omitempty"`
	PetID            *uint64         `json:"pet_id,omitempty"`
	VisitType        string          `json:"visit_type"`
	ServiceTypeID    uint64          `json:"service_type_id"`
	DoctorID         *uint64         `json:"doctor_id,omitempty"`
	IsDesignated     bool            `json:"is_designated"`
	IsStaffDelegated bool            `json:"is_staff_delegated"`
	Source           string          `json:"source"`
	Status           string          `json:"status"`
	Notes            string          `json:"notes"`
	CustomerFields   json.RawMessage `json:"customer_fields"`
	LineCustomerID   *uint64         `json:"line_customer_id,omitempty"`
	CustomerName     string          `json:"customer_name"`
	CourseShortName  string          `json:"course_short_name"`
	StaffName        string          `json:"staff_name"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func toReservationSummaryResponse(ra *model.ReservationAppointment) reservationSummaryResponse {
	customerName := ""
	if ra.LineCustomer != nil {
		customerName = ra.LineCustomer.DisplayName
		if ra.LineCustomer.RealName != "" {
			customerName = ra.LineCustomer.RealName
		}
	} else if ra.Owner != nil {
		customerName = ra.Owner.OwnerName
	}

	courseShortName := ""
	if ra.ServiceType != nil {
		courseShortName = ra.ServiceType.ShortName
		if courseShortName == "" {
			courseShortName = ra.ServiceType.Name
		}
	}

	staffName := ""
	if ra.Doctor != nil {
		staffName = ra.Doctor.Name
	}

	return reservationSummaryResponse{
		ID:               ra.ID,
		StartTime:        ra.StartTime,
		EndTime:          ra.EndTime,
		CustomerName:     customerName,
		CourseShortName:  courseShortName,
		StaffName:        staffName,
		IsStaffDelegated: ra.IsStaffDelegated,
		Source:           string(ra.Source),
		Status:           string(ra.Status),
	}
}

func toReservationDetailResponse(ra *model.ReservationAppointment) reservationDetailResponse {
	customerName := ""
	if ra.LineCustomer != nil {
		customerName = ra.LineCustomer.DisplayName
		if ra.LineCustomer.RealName != "" {
			customerName = ra.LineCustomer.RealName
		}
	} else if ra.Owner != nil {
		customerName = ra.Owner.OwnerName
	}

	courseShortName := ""
	if ra.ServiceType != nil {
		courseShortName = ra.ServiceType.ShortName
		if courseShortName == "" {
			courseShortName = ra.ServiceType.Name
		}
	}

	staffName := ""
	if ra.Doctor != nil {
		staffName = ra.Doctor.Name
	}

	return reservationDetailResponse{
		ID:               ra.ID,
		StartTime:        ra.StartTime,
		EndTime:          ra.EndTime,
		OwnerID:          ra.OwnerID,
		PetID:            ra.PetID,
		VisitType:        string(ra.VisitType),
		ServiceTypeID:    ra.ServiceTypeID,
		DoctorID:         ra.DoctorID,
		IsDesignated:     ra.IsDesignated,
		IsStaffDelegated: ra.IsStaffDelegated,
		Source:           string(ra.Source),
		Status:           string(ra.Status),
		Notes:            ra.Notes,
		CustomerFields:   json.RawMessage(ra.CustomerFields),
		LineCustomerID:   ra.LineCustomerID,
		CustomerName:     customerName,
		CourseShortName:  courseShortName,
		StaffName:        staffName,
		CreatedAt:        ra.CreatedAt,
		UpdatedAt:        ra.UpdatedAt,
	}
}
