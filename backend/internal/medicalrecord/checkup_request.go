package medicalrecord

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type createCheckupRequest struct {
	CheckupTypeID uint64  `json:"checkup_type_id" binding:"required"`
	PetID         *uint64 `json:"pet_id"`
	Date          string  `json:"date"            binding:"required"`
	NextDate      *string `json:"next_date"`
	DoctorID      *uint64 `json:"doctor_id"`
	Result        string  `json:"result"`
}

func (r createCheckupRequest) toServiceInput(clinicID uint64) (*CreateCheckupInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}

	var nextDate *time.Time
	if r.NextDate != nil && *r.NextDate != "" {
		nd, err := time.ParseInLocation(time.DateOnly, *r.NextDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("next_date must be YYYY-MM-DD format")
		}
		nextDate = &nd
	}

	return &CreateCheckupInput{
		ClinicID:      clinicID,
		CheckupTypeID: r.CheckupTypeID,
		PetID:         r.PetID,
		Date:          date,
		NextDate:      nextDate,
		DoctorID:      r.DoctorID,
		Result:        r.Result,
	}, nil
}

type updateCheckupRequest struct {
	CheckupTypeID *uint64 `json:"checkup_type_id"`
	PetID         *uint64 `json:"pet_id"`
	Date          *string `json:"date"`
	NextDate      *string `json:"next_date"`
	DoctorID      *uint64 `json:"doctor_id"`
	DoctorIDClear *bool   `json:"doctor_id_clear"`
	Result        *string `json:"result"`
}

func (r updateCheckupRequest) toServiceInput() (*UpdateCheckupInput, error) {
	var updateDate *time.Time
	if r.Date != nil && *r.Date != "" {
		d, err := time.ParseInLocation(time.DateOnly, *r.Date, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
		}
		updateDate = &d
	}

	var updateNextDate *time.Time
	if r.NextDate != nil && *r.NextDate != "" {
		nd, err := time.ParseInLocation(time.DateOnly, *r.NextDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("next_date must be YYYY-MM-DD format")
		}
		updateNextDate = &nd
	}

	return &UpdateCheckupInput{
		CheckupTypeID: r.CheckupTypeID,
		PetID:         r.PetID,
		Date:          updateDate,
		NextDate:      updateNextDate,
		DoctorID:      r.DoctorID,
		DoctorIDClear: r.DoctorIDClear,
		Result:        r.Result,
	}, nil
}

type listGlobalCheckupsQuery struct {
	PetID        string
	ClinicID      uint64
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

func newListGlobalCheckupsQuery(clinicID uint64, values url.Values) listGlobalCheckupsQuery {
	return listGlobalCheckupsQuery{
		PetID:        values.Get("pet_id"),
		ClinicID:      clinicID,
		StartDate:     optionalStringQueryFilter(values.Get("start_date")),
		EndDate:       optionalStringQueryFilter(values.Get("end_date")),
		NextStartDate: optionalStringQueryFilter(values.Get("next_start_date")),
		NextEndDate:   optionalStringQueryFilter(values.Get("next_end_date")),
	}
}

func (q listGlobalCheckupsQuery) toServiceInput() (ListCheckupsByClinicInput, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return ListCheckupsByClinicInput{}, err
	}
	return ListCheckupsByClinicInput{
		ClinicID:      q.ClinicID,
		PetID:         petID,
		StartDate:     q.StartDate,
		EndDate:       q.EndDate,
		NextStartDate: q.NextStartDate,
		NextEndDate:   q.NextEndDate,
	}, nil
}
