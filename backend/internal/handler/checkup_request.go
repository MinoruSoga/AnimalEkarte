package handler

type createCheckupRequest struct {
	CheckupTypeID uint64  `json:"checkup_type_id" binding:"required"`
	PetID         *uint64 `json:"pet_id"`
	Date          string  `json:"date"            binding:"required"`
	NextDate      *string `json:"next_date"`
	DoctorID      *uint64 `json:"doctor_id"`
	Result        string  `json:"result"`
}

type updateCheckupRequest struct {
	CheckupTypeID *uint64 `json:"checkup_type_id"`
	PetID         *uint64 `json:"pet_id"`
	Date          *string `json:"date"`
	NextDate      *string `json:"next_date"`
	DoctorID      *uint64 `json:"doctor_id"`
	Result        *string `json:"result"`
}
