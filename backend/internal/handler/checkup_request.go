package handler

import "time"

type createCheckupRequest struct {
	CheckupTypeID uint64     `json:"checkup_type_id" binding:"required"`
	PetID         *uint64    `json:"pet_id"`
	Date          time.Time  `json:"date"            binding:"required"`
	NextDate      *time.Time `json:"next_date"`
	DoctorID      *uint64    `json:"doctor_id"`
	Result        string     `json:"result"`
}

type updateCheckupRequest struct {
	CheckupTypeID *uint64    `json:"checkup_type_id"`
	PetID         *uint64    `json:"pet_id"`
	Date          *time.Time `json:"date"`
	NextDate      *time.Time `json:"next_date"`
	DoctorID      *uint64    `json:"doctor_id"`
	Result        *string    `json:"result"`
}
