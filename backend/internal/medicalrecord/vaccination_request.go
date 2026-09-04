package medicalrecord

import (
	"net/url"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type listVaccinationQuery struct {
	PetID     string
	OwnerID   string
	StartDate string
	EndDate   string
	Search    string
}

func newListVaccinationQuery(values url.Values) listVaccinationQuery {
	return listVaccinationQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
		Search:    values.Get("search"),
	}
}

type listVaccinationFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	StartDate *string
	EndDate   *string
	Search    string
}

func (q listVaccinationQuery) toServiceFilters() (listVaccinationFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listVaccinationFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listVaccinationFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listVaccinationFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listVaccinationFilters{}, err
	}
	return listVaccinationFilters{
		PetID:     petID,
		OwnerID:   ownerID,
		StartDate: startDate,
		EndDate:   endDate,
		Search:    strings.TrimSpace(q.Search),
	}, nil
}

// createVaccinationRequest はワクチン接種作成のバインド struct
type createVaccinationRequest struct {
	MedicalRecordID  *uint64 `json:"medical_record_id"`
	PetID            *uint64 `json:"pet_id"`
	VaccineID        uint64  `json:"vaccine_id"          binding:"required"`
	Date             *string `json:"date"                binding:"required"`
	DoctorID         *uint64 `json:"doctor_id"`
	NextDate         *string `json:"next_date"`
	NextScheduleType string  `json:"next_schedule_type"`
	Supplemental     string  `json:"supplemental"`
	Lot1             string  `json:"lot1"`
	Lot2             string  `json:"lot2"`
	Lot3             string  `json:"lot3"`
	Lot4             string  `json:"lot4"`
	Remarks          string  `json:"remarks"`
}

func (r *createVaccinationRequest) toServiceInput() (*CreateVaccinationInput, error) {
	date, err := httpapi.ParseDate(r.Date)
	if err != nil {
		return nil, err
	}
	if date == nil {
		return nil, apperrors.WrapInvalidInput("date is required")
	}

	nextDate, err := httpapi.ParseDate(r.NextDate)
	if err != nil {
		return nil, err
	}

	var nextScheduleType *model.NextScheduleType
	if r.NextScheduleType != "" {
		nst := model.NextScheduleType(r.NextScheduleType)
		nextScheduleType = &nst
	}

	return &CreateVaccinationInput{
		MedicalRecordID:  r.MedicalRecordID,
		PetID:            r.PetID,
		VaccineID:        r.VaccineID,
		Date:             *date,
		DoctorID:         r.DoctorID,
		NextDate:         nextDate,
		NextScheduleType: nextScheduleType,
		Supplemental:     r.Supplemental,
		Lot1:             r.Lot1,
		Lot2:             r.Lot2,
		Lot3:             r.Lot3,
		Lot4:             r.Lot4,
		Remarks:          r.Remarks,
	}, nil
}

// updateVaccinationRequest はワクチン接種更新のバインド struct
// ポインタ型を使用することで、未送信フィールド（nil）と明示的な空文字（""）を区別する
type updateVaccinationRequest struct {
	MedicalRecordID  *uint64 `json:"medical_record_id"`
	PetID            *uint64 `json:"pet_id"`
	VaccineID        *uint64 `json:"vaccine_id"`
	Date             *string `json:"date"`
	DoctorID         *uint64 `json:"doctor_id"`
	NextDate         *string `json:"next_date"`
	NextScheduleType *string `json:"next_schedule_type"`
	Supplemental     *string `json:"supplemental"`
	Lot1             *string `json:"lot1"`
	Lot2             *string `json:"lot2"`
	Lot3             *string `json:"lot3"`
	Lot4             *string `json:"lot4"`
	Remarks          *string `json:"remarks"`
}

func (r *updateVaccinationRequest) toServiceInput() (*UpdateVaccinationInput, error) {
	date, err := httpapi.ParseDate(r.Date)
	if err != nil {
		return nil, err
	}

	nextDate, err := httpapi.ParseDate(r.NextDate)
	if err != nil {
		return nil, err
	}

	input := &UpdateVaccinationInput{
		MedicalRecordID: r.MedicalRecordID,
		PetID:           r.PetID,
		VaccineID:       r.VaccineID,
		Date:            date,
		DoctorID:        r.DoctorID,
		NextDate:        nextDate,
		Supplemental:    r.Supplemental,
		Lot1:            r.Lot1,
		Lot2:            r.Lot2,
		Lot3:            r.Lot3,
		Lot4:            r.Lot4,
		Remarks:         r.Remarks,
	}
	if r.NextScheduleType != nil {
		nst := model.NextScheduleType(*r.NextScheduleType)
		input.NextScheduleType = &nst
	}

	return input, nil
}
