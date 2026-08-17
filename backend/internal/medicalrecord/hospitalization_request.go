package medicalrecord

import (
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validateHospitalizationDateRange enforces end_date >= start_date at the application
// boundary (MRB-06). Schema CHECK alone only yields a generic constraint message.
func validateHospitalizationDateRange(start, end time.Time) error {
	// Compare calendar dates in the values' locations after truncating to date-only.
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	if endDay.Before(startDay) {
		return apperrors.WrapInvalidInput("end_date は start_date 以降である必要があります")
	}
	return nil
}

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

func (r dischargeWithBillingRequest) toServiceInput(actorID uint64) DischargeWithBillingInput {
	return DischargeWithBillingInput{
		DischargeDate:    r.DischargeDate,
		CreateAccounting: r.CreateAccounting,
		ActorID:          &actorID,
	}
}

// createHospitalizationRequest は入院作成のバインド struct
type createHospitalizationRequest struct {
	OwnerID              uint64                        `json:"owner_id"               binding:"required"`
	PetID                uint64                        `json:"pet_id"                 binding:"required"`
	HospitalizationType  string                        `json:"hospitalization_type"   binding:"required,oneof=hospitalization hotel"`
	StartDate            time.Time                     `json:"start_date"             binding:"required"`
	EndDate              time.Time                     `json:"end_date"               binding:"required"`
	Status               string                        `json:"status"                 binding:"omitempty,oneof=admitted discharged reserved"`
	CageID               *uint64                       `json:"cage_id"`
	DoctorID             *uint64                       `json:"doctor_id"`
	Memo                 string                        `json:"memo"`
	OwnerRequest         string                        `json:"owner_request"`
	StaffNotes           string                        `json:"staff_notes"`
	IsInsurance          bool                          `json:"is_insurance"`
	InsuranceCompanyName *string                       `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string                       `json:"insurance_number,omitempty"`
	TreatmentPlans       []createTreatmentPlanRequest  `json:"treatment_plans"`
}

func (r *createHospitalizationRequest) toServiceInput() (*CreateHospitalizationInput, error) {
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

	// MRB-06: application-level end_date >= start_date (schema CHECK is not a field-level error).
	if err := validateHospitalizationDateRange(r.StartDate, r.EndDate); err != nil {
		return nil, err
	}

	// BUG-037: new hospitalization must assign a cage (DB column remains nullable for legacy rows).
	if r.CageID == nil || *r.CageID == 0 {
		return nil, apperrors.WrapInvalidInput("cage_id is required")
	}

	plans := make([]CreateTreatmentPlanInput, 0, len(r.TreatmentPlans))
	for i := range r.TreatmentPlans {
		plan := r.TreatmentPlans[i].toServiceInput()
		if plan != nil {
			plans = append(plans, *plan)
		}
	}

	return &CreateHospitalizationInput{
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
		TreatmentPlans:       plans,
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

func (r *updateHospitalizationRequest) toServiceInput() (UpdateHospitalizationInput, error) {
	input := UpdateHospitalizationInput{
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
			return UpdateHospitalizationInput{}, fmt.Errorf("invalid hospitalization_type: %w", err)
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
			return UpdateHospitalizationInput{}, fmt.Errorf("invalid status: %w", err)
		}
		input.Status = &status
	}
	// MRB-06: when both dates are present on the patch, reject inverted ranges early.
	if r.StartDate != nil && r.EndDate != nil {
		if err := validateHospitalizationDateRange(*r.StartDate, *r.EndDate); err != nil {
			return UpdateHospitalizationInput{}, err
		}
	}
	return input, nil
}

// validateEnum は internal/handler/validation.go の同名 generic の最小複製（BE9-2D ⑤:
// 原本は旧 package の appointment/medical_record request が引き続き使うため移動不可。純関数）。
func validateEnum[T ~string](v string, allowed ...T) (T, error) {
	if slices.Contains(allowed, T(v)) {
		return T(v), nil
	}
	var zero T
	return zero, fmt.Errorf("invalid value %q", v)
}
