package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type checkupResponse struct {
	ID              string  `json:"id"`
	MedicalRecordID string  `json:"medical_record_id"`
	CheckupTypeID   string  `json:"checkup_type_id"`
	PetID           *string `json:"pet_id,omitempty"`
	Date            string  `json:"date"`
	NextDate        *string `json:"next_date,omitempty"`
	DoctorID        *string `json:"doctor_id,omitempty"`
	Result          string  `json:"result"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`

	// Nested
	CheckupType *checkupTypeResponse  `json:"checkup_type,omitempty"`
	Doctor      *StaffSummaryResponse `json:"doctor,omitempty"`
}

func toCheckupResponse(c *model.Checkup) checkupResponse {
	r := checkupResponse{
		ID:              strconv.FormatUint(c.ID, 10),
		MedicalRecordID: strconv.FormatUint(c.MedicalRecordID, 10),
		CheckupTypeID:   strconv.FormatUint(c.CheckupTypeID, 10),
		Date:            c.Date.In(time.Local).Format(time.DateOnly),
		Result:          c.Result,
		CreatedAt:       httpapi.LocalTimeRFC3339(c.CreatedAt),
		UpdatedAt:       httpapi.LocalTimeRFC3339(c.UpdatedAt),
	}
	if c.NextDate != nil {
		nd := c.NextDate.In(time.Local).Format(time.DateOnly)
		r.NextDate = &nd
	}
	if c.PetID != nil {
		s := strconv.FormatUint(*c.PetID, 10)
		r.PetID = &s
	}
	if c.DoctorID != nil {
		s := strconv.FormatUint(*c.DoctorID, 10)
		r.DoctorID = &s
	}
	if c.CheckupType != nil {
		ct := toCheckupTypeResponse(c.CheckupType)
		r.CheckupType = &ct
	}
	r.Doctor = toStaffSummary(c.Doctor)
	return r
}

// checkupGlobalResponse はクリニック横断一覧用レスポンス（MedicalRecord.Pet.Owner を含む）
type checkupGlobalResponse struct {
	ID              string  `json:"id"`
	MedicalRecordID string  `json:"medical_record_id"`
	CheckupTypeID   string  `json:"checkup_type_id"`
	PetID           *string `json:"pet_id,omitempty"`
	Date            string  `json:"date"`
	NextDate        *string `json:"next_date,omitempty"`
	DoctorID        *string `json:"doctor_id,omitempty"`
	Result          string  `json:"result"`

	// Nested
	CheckupType *checkupTypeResponse  `json:"checkup_type,omitempty"`
	Doctor      *StaffSummaryResponse `json:"doctor,omitempty"`
	PetName     string                `json:"pet_name"`
	OwnerName   string                `json:"owner_name"`
	OwnerID     *string               `json:"owner_id,omitempty"`
}

func toCheckupGlobalResponse(c *model.Checkup) checkupGlobalResponse {
	r := checkupGlobalResponse{
		ID:              strconv.FormatUint(c.ID, 10),
		MedicalRecordID: strconv.FormatUint(c.MedicalRecordID, 10),
		CheckupTypeID:   strconv.FormatUint(c.CheckupTypeID, 10),
		Date:            c.Date.In(time.Local).Format(time.DateOnly),
		Result:          c.Result,
	}
	if c.NextDate != nil {
		nd := c.NextDate.In(time.Local).Format(time.DateOnly)
		r.NextDate = &nd
	}
	if c.PetID != nil {
		s := strconv.FormatUint(*c.PetID, 10)
		r.PetID = &s
	}
	if c.DoctorID != nil {
		s := strconv.FormatUint(*c.DoctorID, 10)
		r.DoctorID = &s
	}
	if c.CheckupType != nil {
		ct := toCheckupTypeResponse(c.CheckupType)
		r.CheckupType = &ct
	}
	r.Doctor = toStaffSummary(c.Doctor)
	// Pet/Owner from MedicalRecord relation
	if c.MedicalRecord != nil && c.MedicalRecord.Pet != nil {
		r.PetName = c.MedicalRecord.Pet.Name
		if c.MedicalRecord.Pet.Owner != nil {
			r.OwnerName = c.MedicalRecord.Pet.Owner.Name
			s := strconv.FormatUint(c.MedicalRecord.Pet.Owner.ID, 10)
			r.OwnerID = &s
		}
	}
	return r
}
