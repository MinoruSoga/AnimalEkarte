package handler

import (
	"fmt"
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type listHospitalizationQuery struct {
	PetID     string
	OwnerID   string
	Status    string
	StartDate string
	EndDate   string
}

func newListHospitalizationQuery(values url.Values) listHospitalizationQuery {
	return listHospitalizationQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		Status:    values.Get("status"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

type listHospitalizationFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	Status    *string
	StartDate *string
	EndDate   *string
}

func (q *listHospitalizationQuery) toServiceFilters() (listHospitalizationFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	return listHospitalizationFilters{
		PetID:     petID,
		OwnerID:   ownerID,
		Status:    optionalStringQueryFilter(q.Status),
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// dischargeWithBillingRequest は退院+会計作成のバインド struct
type dischargeWithBillingRequest struct {
	DischargeDate    time.Time `json:"discharge_date" binding:"required"`
	CreateAccounting bool      `json:"create_accounting"`
}

func (r dischargeWithBillingRequest) toServiceInput() service.DischargeWithBillingInput {
	return service.DischargeWithBillingInput{
		DischargeDate:    r.DischargeDate,
		CreateAccounting: r.CreateAccounting,
	}
}

// createHospitalizationRequest は入院作成のバインド struct
type createHospitalizationRequest struct {
	OwnerID              uint64    `json:"owner_id"               binding:"required"`
	PetID                uint64    `json:"pet_id"                 binding:"required"`
	HospitalizationType  string    `json:"hospitalization_type"   binding:"required,oneof=hospitalization hotel"`
	StartDate            time.Time `json:"start_date"             binding:"required"`
	EndDate              time.Time `json:"end_date"               binding:"required"`
	Status               string    `json:"status"                 binding:"omitempty,oneof=admitted discharged reserved"`
	CageID               *uint64   `json:"cage_id"`
	DoctorID             *uint64   `json:"doctor_id"`
	Memo                 string    `json:"memo"`
	OwnerRequest         string    `json:"owner_request"`
	StaffNotes           string    `json:"staff_notes"`
	IsInsurance          bool      `json:"is_insurance"`
	InsuranceCompanyName *string   `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string   `json:"insurance_number,omitempty"`
}

func (r *createHospitalizationRequest) toServiceInput() (*service.CreateHospitalizationInput, error) {
	hospType, err := validateEnum(r.HospitalizationType,
		model.HospitalizationTypeInpatient,
		model.HospitalizationTypeHotel,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid hospitalization_type: %w", err)
	}

	var status model.HospitalizationStatus
	if r.Status != "" {
		s, err := validateEnum(r.Status,
			model.HospitalizationStatusAdmitted,
			model.HospitalizationStatusDischarged,
			model.HospitalizationStatusReserved,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		status = s
	}

	return &service.CreateHospitalizationInput{
		OwnerID:              r.OwnerID,
		PetID:                r.PetID,
		HospitalizationType:  hospType,
		StartDate:            r.StartDate,
		EndDate:              r.EndDate,
		Status:               status,
		CageID:               r.CageID,
		DoctorID:             r.DoctorID,
		Memo:                 r.Memo,
		OwnerRequest:         r.OwnerRequest,
		StaffNotes:           r.StaffNotes,
		IsInsurance:          r.IsInsurance,
		InsuranceCompanyName: r.InsuranceCompanyName,
		InsuranceNumber:      r.InsuranceNumber,
	}, nil
}

// updateHospitalizationRequest は入院更新のバインド struct
type updateHospitalizationRequest struct {
	OwnerID              *uint64    `json:"owner_id"`
	PetID                *uint64    `json:"pet_id"`
	HospitalizationType  *string    `json:"hospitalization_type"   binding:"omitempty,oneof=hospitalization hotel"`
	StartDate            *time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date"`
	Status               *string    `json:"status"                 binding:"omitempty,oneof=admitted discharged reserved"`
	CageID               *uint64    `json:"cage_id"`
	DoctorID             *uint64    `json:"doctor_id"`
	Memo                 *string    `json:"memo"`
	OwnerRequest         *string    `json:"owner_request"`
	StaffNotes           *string    `json:"staff_notes"`
	IsInsurance          *bool      `json:"is_insurance,omitempty"`
	InsuranceCompanyName *string    `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string    `json:"insurance_number,omitempty"`
}

func (r *updateHospitalizationRequest) toServiceInput() (service.UpdateHospitalizationInput, error) {
	input := service.UpdateHospitalizationInput{
		OwnerID:              r.OwnerID,
		PetID:                r.PetID,
		StartDate:            r.StartDate,
		EndDate:              r.EndDate,
		CageID:               r.CageID,
		DoctorID:             r.DoctorID,
		Memo:                 r.Memo,
		OwnerRequest:         r.OwnerRequest,
		StaffNotes:           r.StaffNotes,
		IsInsurance:          r.IsInsurance,
		InsuranceCompanyName: r.InsuranceCompanyName,
		InsuranceNumber:      r.InsuranceNumber,
	}
	if r.HospitalizationType != nil {
		hospType, err := validateEnum(*r.HospitalizationType,
			model.HospitalizationTypeInpatient,
			model.HospitalizationTypeHotel,
		)
		if err != nil {
			return service.UpdateHospitalizationInput{}, fmt.Errorf("invalid hospitalization_type: %w", err)
		}
		input.HospitalizationType = &hospType
	}
	if r.Status != nil {
		status, err := validateEnum(*r.Status,
			model.HospitalizationStatusAdmitted,
			model.HospitalizationStatusDischarged,
			model.HospitalizationStatusReserved,
		)
		if err != nil {
			return service.UpdateHospitalizationInput{}, fmt.Errorf("invalid status: %w", err)
		}
		input.Status = &status
	}
	return input, nil
}
