package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type createPrescriptionRequest struct {
	Date         string `json:"date"          binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required"`
}

func (r createPrescriptionRequest) toServiceInput() (*CreatePrescriptionInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}

	return &CreatePrescriptionInput{
		PrescribedAt: date,
		DurationDays: r.DurationDays,
	}, nil
}

type updatePrescriptionRequest struct {
	Date         *string `json:"date"`
	DurationDays *int    `json:"duration_days"`
}

func (r updatePrescriptionRequest) toServiceInput() (*UpdatePrescriptionInput, error) {
	var updateDate *time.Time
	if r.Date != nil && *r.Date != "" {
		date, err := time.ParseInLocation(time.DateOnly, *r.Date, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
		}
		updateDate = &date
	}

	return &UpdatePrescriptionInput{
		PrescribedAt: updateDate,
		DurationDays: r.DurationDays,
	}, nil
}
