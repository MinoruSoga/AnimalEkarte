package handler

import "github.com/animal-ekarte/backend/internal/service"

type createExaminationTypeRequest struct {
	Name           string  `json:"name"             binding:"required"`
	Price          *int64  `json:"price"`
	IsActive       bool    `json:"is_active"`
	Description    string  `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	SortOrder      int     `json:"sort_order"`
	IsNonInsurance bool    `json:"is_non_insurance"`
}

func (r createExaminationTypeRequest) toServiceInput() *service.CreateExamTypeInput {
	return &service.CreateExamTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		ParentID:       r.ParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
	}
}

type updateExaminationTypeRequest struct {
	Name           *string `json:"name"`
	Price          *int64  `json:"price"`
	IsActive       *bool   `json:"is_active"`
	Description    *string `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	ClearParentID  bool    `json:"clear_parent_id"`
	SortOrder      *int    `json:"sort_order"`
	IsNonInsurance *bool   `json:"is_non_insurance"`
}

func (r updateExaminationTypeRequest) toServiceInput() *service.UpdateExamTypeInput {
	return &service.UpdateExamTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		ParentID:       r.ParentID,
		ClearParentID:  r.ClearParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
	}
}
