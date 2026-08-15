package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// vaccineSummaryResponse は接種記録要約内で使用するワクチンの要約型
type vaccineSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type vaccinationResponse struct {
	ID               uint64     `json:"id"`
	ClinicID         uint64     `json:"clinic_id"`
	MedicalRecordID  *uint64    `json:"medical_record_id,omitempty"`
	PetID            *uint64    `json:"pet_id,omitempty"`
	VaccineID        uint64     `json:"vaccine_id"`
	Date             time.Time  `json:"date"`
	DoctorID         *uint64    `json:"doctor_id,omitempty"`
	NextDate         *time.Time `json:"next_date,omitempty"`
	NextScheduleType *string    `json:"next_schedule_type,omitempty"`
	Supplemental     string     `json:"supplemental"`
	Lot1             string     `json:"lot1"`
	Lot2             string     `json:"lot2"`
	Lot3             string     `json:"lot3"`
	Lot4             string     `json:"lot4"`
	Remarks          string     `json:"remarks"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	// リレーション: 一覧で飼主名/ペット名/ワクチン名/担当医を表示するため。Preload 時のみ埋まる。
	Pet     *PetSummaryResponse     `json:"pet,omitempty"`
	Doctor  *StaffSummaryResponse   `json:"doctor,omitempty"`
	Vaccine *vaccineSummaryResponse `json:"vaccine,omitempty"`
}

func toVaccinationResponse(v *model.Vaccination) vaccinationResponse {
	var nextScheduleType *string
	if v.NextScheduleType != nil {
		s := string(*v.NextScheduleType)
		nextScheduleType = &s
	}
	resp := vaccinationResponse{
		ID:               v.ID,
		ClinicID:         v.ClinicID,
		MedicalRecordID:  v.MedicalRecordID,
		PetID:            v.PetID,
		VaccineID:        v.VaccineID,
		Date:             httpapi.LocalTime(v.Date),
		NextDate:         httpapi.LocalTimePtr(v.NextDate),
		NextScheduleType: nextScheduleType,
		Supplemental:     v.Supplemental,
		Lot1:             v.Lot1,
		Lot2:             v.Lot2,
		Lot3:             v.Lot3,
		Lot4:             v.Lot4,
		Remarks:          v.Remarks,
		CreatedAt:        httpapi.LocalTime(v.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(v.UpdatedAt),
		Pet:              toPetSummary(v.Pet),
		Doctor:           toStaffSummary(v.Doctor),
	}
	if v.DoctorID != nil && v.Doctor != nil && v.Doctor.ID == *v.DoctorID {
		resp.DoctorID = v.DoctorID
	}
	if v.Vaccine != nil {
		resp.Vaccine = &vaccineSummaryResponse{
			ID:   v.Vaccine.ID,
			Name: v.Vaccine.Name,
		}
	}
	return resp
}
