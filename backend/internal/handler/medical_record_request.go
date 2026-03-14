package handler

import "time"

// createMedicalRecordRequest はカルテ作成のバインド struct
type createMedicalRecordRequest struct {
	RecordNo                 string    `json:"record_no"                   binding:"required"`
	Date                     time.Time `json:"date"                        binding:"required"`
	OwnerID                  *uint64   `json:"owner_id"`
	PetID                    *uint64   `json:"pet_id"`
	DoctorID                 *uint64   `json:"doctor_id"`
	ReservationAppointmentID *uint64   `json:"reservation_appointment_id"`
	Status                   string    `json:"status"`
}

// updateMedicalRecordRequest はカルテ更新のバインド struct
type updateMedicalRecordRequest struct {
	RecordNo                 string     `json:"record_no"`
	Date                     *time.Time `json:"date"`
	OwnerID                  *uint64    `json:"owner_id"`
	PetID                    *uint64    `json:"pet_id"`
	DoctorID                 *uint64    `json:"doctor_id"`
	ReservationAppointmentID *uint64    `json:"reservation_appointment_id"`
	Status                   string     `json:"status"`
}
