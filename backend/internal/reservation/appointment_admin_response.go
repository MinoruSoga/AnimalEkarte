package reservation

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
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
	ID                uint64          `json:"id"`
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time"`
	OwnerID           *uint64         `json:"owner_id,omitempty"`
	PetID             *uint64         `json:"pet_id,omitempty"`
	VisitType         string          `json:"visit_type"`
	ReservationTypeID uint64          `json:"reservation_type_id"`
	DoctorID          *uint64         `json:"doctor_id,omitempty"`
	IsDesignated      bool            `json:"is_designated"`
	IsStaffDelegated  bool            `json:"is_staff_delegated"`
	Source            string          `json:"source"`
	Status            string          `json:"status"`
	Notes             string          `json:"notes"`
	CustomerFields    json.RawMessage `json:"customer_fields"`
	LineCustomerID    *uint64         `json:"line_customer_id,omitempty"`
	CustomerName      string          `json:"customer_name"`
	CourseShortName   string          `json:"course_short_name"`
	StaffName         string          `json:"staff_name"`
	CreatedBy         *uint64         `json:"created_by,omitempty"`
	CreatedByName     string          `json:"created_by_name"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toReservationSummaryResponse(ra *model.Reservation) reservationSummaryResponse {
	customerName := ""
	if ra.LineCustomer != nil {
		customerName = ra.LineCustomer.DisplayName
		if ra.LineCustomer.RealName != "" {
			customerName = ra.LineCustomer.RealName
		}
	} else if ra.Owner != nil {
		customerName = ra.Owner.Name
	}

	courseShortName := ""
	if ra.ReservationType != nil {
		courseShortName = ra.ReservationType.ShortName
		if courseShortName == "" {
			courseShortName = ra.ReservationType.Name
		}
	}

	staffName := ""
	if ra.Doctor != nil {
		staffName = ra.Doctor.Name
	}

	return reservationSummaryResponse{
		ID:               ra.ID,
		StartTime:        httpapi.LocalTime(ra.StartTime),
		EndTime:          httpapi.LocalTime(ra.EndTime),
		CustomerName:     customerName,
		CourseShortName:  courseShortName,
		StaffName:        staffName,
		IsStaffDelegated: ra.IsStaffDelegated,
		Source:           string(ra.Source),
		Status:           string(ra.Status),
	}
}

func toReservationDetailResponse(ra *model.Reservation) reservationDetailResponse {
	customerName := ""
	if ra.LineCustomer != nil {
		customerName = ra.LineCustomer.DisplayName
		if ra.LineCustomer.RealName != "" {
			customerName = ra.LineCustomer.RealName
		}
	} else if ra.Owner != nil {
		customerName = ra.Owner.Name
	}

	courseShortName := ""
	if ra.ReservationType != nil {
		courseShortName = ra.ReservationType.ShortName
		if courseShortName == "" {
			courseShortName = ra.ReservationType.Name
		}
	}

	staffName := ""
	if ra.Doctor != nil {
		staffName = ra.Doctor.Name
	}

	createdByName := ""
	if ra.CreatedByStaff != nil {
		createdByName = ra.CreatedByStaff.Name
	}

	return reservationDetailResponse{
		ID:                ra.ID,
		StartTime:         httpapi.LocalTime(ra.StartTime),
		EndTime:           httpapi.LocalTime(ra.EndTime),
		OwnerID:           ra.OwnerID,
		PetID:             ra.PetID,
		VisitType:         string(ra.VisitType),
		ReservationTypeID: ra.ReservationTypeID,
		DoctorID:          ra.DoctorID,
		IsDesignated:      ra.IsDesignated,
		IsStaffDelegated:  ra.IsStaffDelegated,
		Source:            string(ra.Source),
		Status:            string(ra.Status),
		Notes:             ra.Notes,
		CustomerFields:    ra.CustomerFields,
		LineCustomerID:    ra.LineCustomerID,
		CustomerName:      customerName,
		CourseShortName:   courseShortName,
		StaffName:         staffName,
		CreatedBy:         ra.CreatedBy,
		CreatedByName:     createdByName,
		CreatedAt:         httpapi.LocalTime(ra.CreatedAt),
		UpdatedAt:         httpapi.LocalTime(ra.UpdatedAt),
	}
}
