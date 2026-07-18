package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/service"
)

type createPrescriptionRequest struct {
	Date         string `json:"date"          binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required"`
}

func (r createPrescriptionRequest) toServiceInput() (*service.CreatePrescriptionInput, error) {
	date, err := time.ParseInLocation(time.DateOnly, r.Date, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}

	return &service.CreatePrescriptionInput{
		PrescribedAt: date,
		DurationDays: r.DurationDays,
	}, nil
}

type updatePrescriptionRequest struct {
	Date         *string `json:"date"`
	DurationDays *int    `json:"duration_days"`
}

func (r updatePrescriptionRequest) toServiceInput() (*service.UpdatePrescriptionInput, error) {
	var updateDate *time.Time
	if r.Date != nil && *r.Date != "" {
		date, err := time.ParseInLocation(time.DateOnly, *r.Date, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
		}
		updateDate = &date
	}

	return &service.UpdatePrescriptionInput{
		PrescribedAt: updateDate,
		DurationDays: r.DurationDays,
	}, nil
}
