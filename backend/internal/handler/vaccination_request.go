package handler

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
