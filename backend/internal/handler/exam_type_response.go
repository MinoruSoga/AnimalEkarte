package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type examTypeItemResponse struct {
	ID              uint64    `json:"id"`
	ExamTypeID      uint64    `json:"exam_type_id"`
	Name            string    `json:"name"`
	InspectionValue string    `json:"inspection_value"`
	NormalValue     string    `json:"normal_value"`
	Unit            string    `json:"unit"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
}

type examTypeResponse struct {
	ID          uint64                 `json:"id"`
	ClinicID    uint64                 `json:"clinic_id"`
	Name        string                 `json:"name"`
	Price       *int64                 `json:"price,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Description string                 `json:"description"`
	ParentID    *uint64                `json:"parent_id,omitempty"`
	SortOrder   int                    `json:"sort_order"`
	Items       []examTypeItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func toExamTypeItemResponse(item *model.ExamTypeField) examTypeItemResponse {
	return examTypeItemResponse{
		ID:              item.ID,
		ExamTypeID:      item.ExamTypeID,
		Name:            item.Name,
		InspectionValue: item.InspectionValue,
		NormalValue:     item.NormalValue,
		Unit:            item.Unit,
		SortOrder:       item.SortOrder,
		CreatedAt:       item.CreatedAt,
	}
}

func toExamTypeResponse(et *model.ExaminationType) examTypeResponse {
	items := make([]examTypeItemResponse, 0, len(et.Items))
	for i := range et.Items {
		items = append(items, toExamTypeItemResponse(&et.Items[i]))
	}
	return examTypeResponse{
		ID:          et.ID,
		ClinicID:    et.ClinicID,
		Name:        et.Name,
		Price:       et.Price,
		IsActive:    et.IsActive,
		Description: et.Description,
		ParentID:    et.ParentID,
		SortOrder:   et.SortOrder,
		Items:       items,
		CreatedAt:   et.CreatedAt,
		UpdatedAt:   et.UpdatedAt,
	}
}
