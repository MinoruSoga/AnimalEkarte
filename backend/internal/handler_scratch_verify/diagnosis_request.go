package handler

import "github.com/animal-ekarte/backend/internal/service"

type createDiagnosisTypeRequest struct {
	Name        string `json:"name"        binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func (r createDiagnosisTypeRequest) toServiceInput() *service.CreateDiagnosisTypeInput {
	return &service.CreateDiagnosisTypeInput{
		Name:        r.Name,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

type updateDiagnosisTypeRequest struct {
	Name        *string `json:"name"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateDiagnosisTypeRequest) toServiceInput() *service.UpdateDiagnosisTypeInput {
	return &service.UpdateDiagnosisTypeInput{
		Name:        r.Name,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

type createDiagnosisNameRequest struct {
	Name            string `json:"name"                  binding:"required"`
	DiagnosisTypeID uint64 `json:"diagnosis_type_id" binding:"required"`
	IsActive        bool   `json:"is_active"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
}

func (r createDiagnosisNameRequest) toServiceInput() *service.CreateDiagnosisNameInput {
	return &service.CreateDiagnosisNameInput{
		Name:            r.Name,
		DiagnosisTypeID: r.DiagnosisTypeID,
		IsActive:        r.IsActive,
		Description:     r.Description,
		SortOrder:       r.SortOrder,
	}
}

type updateDiagnosisNameRequest struct {
	Name            *string `json:"name"`
	DiagnosisTypeID *uint64 `json:"diagnosis_type_id"`
	IsActive        *bool   `json:"is_active"`
	Description     *string `json:"description"`
	SortOrder       *int    `json:"sort_order"`
}

func (r updateDiagnosisNameRequest) toServiceInput() *service.UpdateDiagnosisNameInput {
	return &service.UpdateDiagnosisNameInput{
		Name:            r.Name,
		DiagnosisTypeID: r.DiagnosisTypeID,
		IsActive:        r.IsActive,
		Description:     r.Description,
		SortOrder:       r.SortOrder,
	}
}
