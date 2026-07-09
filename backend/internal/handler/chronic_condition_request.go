package handler

import (
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/service"
)

type createChronicConditionRequest struct {
	ConditionCode string  `json:"condition_code" binding:"required"`
	ConditionName string  `json:"condition_name" binding:"required"`
	DiagnosedAt   string  `json:"diagnosed_at"   binding:"required"`
	Notes         *string `json:"notes"`
	IsActive      *bool   `json:"is_active"`
}

func (r createChronicConditionRequest) toServiceInput() (service.CreateChronicConditionInput, error) {
	diagnosedAt, err := time.ParseInLocation(time.DateOnly, r.DiagnosedAt, time.Local)
	if err != nil {
		return service.CreateChronicConditionInput{}, fmt.Errorf("diagnosed_at must be YYYY-MM-DD")
	}

	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}

	return service.CreateChronicConditionInput{
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		DiagnosedAt:   diagnosedAt,
		Notes:         r.Notes,
		IsActive:      isActive,
	}, nil
}

type updateChronicConditionRequest struct {
	ConditionCode *string `json:"condition_code"`
	ConditionName *string `json:"condition_name"`
	DiagnosedAt   *string `json:"diagnosed_at"`
	Notes         *string `json:"notes"`
	IsActive      *bool   `json:"is_active"`
}

func (r updateChronicConditionRequest) toServiceInput() (service.UpdateChronicConditionInput, error) {
	input := service.UpdateChronicConditionInput{
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		Notes:         r.Notes,
		IsActive:      r.IsActive,
	}

	if r.DiagnosedAt != nil {
		diagnosedAt, err := time.ParseInLocation(time.DateOnly, *r.DiagnosedAt, time.Local)
		if err != nil {
			return service.UpdateChronicConditionInput{}, fmt.Errorf("diagnosed_at must be YYYY-MM-DD")
		}
		input.DiagnosedAt = &diagnosedAt
	}

	return input, nil
}
