package pet

import (
	"fmt"
	"time"
)

type createChronicConditionRequest struct {
	ConditionCode string  `json:"condition_code" binding:"required"`
	ConditionName string  `json:"condition_name" binding:"required,max=255"`
	DiagnosedAt   string  `json:"diagnosed_at"   binding:"required"`
	Notes         *string `json:"notes"`
	IsActive      *bool   `json:"is_active"`
}

func (r createChronicConditionRequest) toServiceInput() (CreateChronicConditionInput, error) {
	diagnosedAt, err := time.ParseInLocation(time.DateOnly, r.DiagnosedAt, time.Local)
	if err != nil {
		return CreateChronicConditionInput{}, fmt.Errorf("diagnosed_at must be YYYY-MM-DD")
	}

	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}

	return CreateChronicConditionInput{
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		DiagnosedAt:   diagnosedAt,
		Notes:         r.Notes,
		IsActive:      isActive,
	}, nil
}

type updateChronicConditionRequest struct {
	ConditionCode *string `json:"condition_code"`
	ConditionName *string `json:"condition_name" binding:"omitempty,max=255"`
	DiagnosedAt   *string `json:"diagnosed_at"`
	Notes         *string `json:"notes"`
	IsActive      *bool   `json:"is_active"`
}

func (r updateChronicConditionRequest) toServiceInput() (UpdateChronicConditionInput, error) {
	input := UpdateChronicConditionInput{
		ConditionCode: r.ConditionCode,
		ConditionName: r.ConditionName,
		Notes:         r.Notes,
		IsActive:      r.IsActive,
	}

	if r.DiagnosedAt != nil {
		diagnosedAt, err := time.ParseInLocation(time.DateOnly, *r.DiagnosedAt, time.Local)
		if err != nil {
			return UpdateChronicConditionInput{}, fmt.Errorf("diagnosed_at must be YYYY-MM-DD")
		}
		input.DiagnosedAt = &diagnosedAt
	}

	return input, nil
}
