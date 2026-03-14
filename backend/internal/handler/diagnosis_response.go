package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type diagnosisCategoryResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toDiagnosisCategoryResponse(c *model.DiagnosisCategory) diagnosisCategoryResponse {
	return diagnosisCategoryResponse{
		ID:          c.ID,
		ClinicID:    c.ClinicID,
		Name:        c.Name,
		IsActive:    c.IsActive,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toDiagnosisCategoryResponseList(categories []model.DiagnosisCategory) []diagnosisCategoryResponse {
	list := make([]diagnosisCategoryResponse, 0, len(categories))
	for i := range categories {
		list = append(list, toDiagnosisCategoryResponse(&categories[i]))
	}
	return list
}

type diagnosisNameResponse struct {
	ID                  uint64    `json:"id"`
	ClinicID            uint64    `json:"clinic_id"`
	Name                string    `json:"name"`
	IsActive            bool      `json:"is_active"`
	Description         string    `json:"description"`
	DiagnosisCategoryID uint64    `json:"diagnosis_category_id"`
	SortOrder           int       `json:"sort_order"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func toDiagnosisNameResponse(n *model.DiagnosisName) diagnosisNameResponse {
	return diagnosisNameResponse{
		ID:                  n.ID,
		ClinicID:            n.ClinicID,
		Name:                n.Name,
		IsActive:            n.IsActive,
		Description:         n.Description,
		DiagnosisCategoryID: n.DiagnosisCategoryID,
		SortOrder:           n.SortOrder,
		CreatedAt:           n.CreatedAt,
		UpdatedAt:           n.UpdatedAt,
	}
}

func toDiagnosisNameResponseList(names []model.DiagnosisName) []diagnosisNameResponse {
	list := make([]diagnosisNameResponse, 0, len(names))
	for i := range names {
		list = append(list, toDiagnosisNameResponse(&names[i]))
	}
	return list
}
