package handler

type createPrescriptionRequest struct {
	Date         string `json:"date"          binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required"`
}

type updatePrescriptionRequest struct {
	Date         *string `json:"date"`
	DurationDays *int    `json:"duration_days"`
}
