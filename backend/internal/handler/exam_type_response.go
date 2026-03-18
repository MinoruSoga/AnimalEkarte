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
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
}

type examTypeResponse struct {
	ID          uint64                 `json:"id"`
	ClinicID    uint64                 `json:"clinic_id"`
	Name        string                 `json:"name"`
	Price       *float64               `json:"price,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Description string                 `json:"description"`
	SortOrder   int                    `json:"sort_order"`
	Items       []examTypeItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func toExamTypeItemResponse(item *model.ExaminationTypeItem) examTypeItemResponse {
	return examTypeItemResponse{
		ID:              item.ID,
		ExamTypeID:      item.ExamTypeID,
		Name:            item.Name,
		InspectionValue: item.InspectionValue,
		NormalValue:     item.NormalValue,
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
		SortOrder:   et.SortOrder,
		Items:       items,
		CreatedAt:   et.CreatedAt,
		UpdatedAt:   et.UpdatedAt,
	}
}

func toExamTypeResponseList(items []model.ExaminationType) []examTypeResponse {
	list := make([]examTypeResponse, 0, len(items))
	for i := range items {
		list = append(list, toExamTypeResponse(&items[i]))
	}
	return list
}
