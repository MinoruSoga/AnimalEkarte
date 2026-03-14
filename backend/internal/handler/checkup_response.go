package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type checkupResponse struct {
	ID              string     `json:"id"`
	MedicalRecordID string     `json:"medical_record_id"`
	CheckupTypeID   string     `json:"checkup_type_id"`
	PetID           *string    `json:"pet_id,omitempty"`
	Date            time.Time  `json:"date"`
	NextDate        *time.Time `json:"next_date,omitempty"`
	DoctorID        *string    `json:"doctor_id,omitempty"`
	Result          string     `json:"result"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Nested
	CheckupType *checkupTypeResponse  `json:"checkup_type,omitempty"`
	Doctor      *staffSummaryResponse `json:"doctor,omitempty"`
}

func toCheckupResponse(c *model.Checkup) checkupResponse {
	r := checkupResponse{
		ID:              strconv.FormatUint(c.ID, 10),
		MedicalRecordID: strconv.FormatUint(c.MedicalRecordID, 10),
		CheckupTypeID:   strconv.FormatUint(c.CheckupTypeID, 10),
		Date:            c.Date,
		NextDate:        c.NextDate,
		Result:          c.Result,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
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

func toCheckupResponseList(items []model.Checkup) []checkupResponse {
	list := make([]checkupResponse, 0, len(items))
	for i := range items {
		list = append(list, toCheckupResponse(&items[i]))
	}
	return list
}
