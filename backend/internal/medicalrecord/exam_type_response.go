package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type examTypeItemResponse struct {
	ID              uint64                       `json:"id"`
	ExamTypeID      uint64                       `json:"exam_type_id"`
	Name            string                       `json:"name"`
	InspectionValue string                       `json:"inspection_value"`
	NormalValue     string                       `json:"normal_value"`
	Unit            string                       `json:"unit"`
	SortOrder       int                          `json:"sort_order"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	ReferenceRanges []examReferenceRangeResponse `json:"reference_ranges"`
}

type examReferenceRangeResponse struct {
	ID              uint64   `json:"id"`
	ExamTypeFieldID uint64   `json:"exam_type_field_id"`
	AnimalSpeciesID uint64   `json:"animal_species_id"`
	RefMin          *float64 `json:"ref_min,omitempty"`
	RefMax          *float64 `json:"ref_max,omitempty"`
	QualitativeMin  *string  `json:"qualitative_min,omitempty"`
	QualitativeMax  *string  `json:"qualitative_max,omitempty"`
}

type examTypeResponse struct {
	ID             uint64                 `json:"id"`
	ClinicID       uint64                 `json:"clinic_id"`
	Name           string                 `json:"name"`
	Price          *int64                 `json:"price,omitempty"`
	IsActive       bool                   `json:"is_active"`
	Description    string                 `json:"description"`
	ParentID       *uint64                `json:"parent_id,omitempty"`
	SortOrder      int                    `json:"sort_order"`
	IsNonInsurance bool                   `json:"is_non_insurance"`
	Items          []examTypeItemResponse `json:"items,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func toExamTypeItemResponse(item *model.ExamTypeField) examTypeItemResponse {
	return toExamTypeItemResponseWithRanges(item, nil)
}

func toExamTypeItemResponseWithRanges(
	item *model.ExamTypeField,
	ranges []model.ExamReferenceRange,
) examTypeItemResponse {
	rangeResponses := make([]examReferenceRangeResponse, 0, len(ranges))
	for i := range ranges {
		rangeResponses = append(rangeResponses, examReferenceRangeResponse{
			ID:              ranges[i].ID,
			ExamTypeFieldID: ranges[i].ExamTypeFieldID,
			AnimalSpeciesID: ranges[i].AnimalSpeciesID,
			RefMin:          ranges[i].RefMin,
			RefMax:          ranges[i].RefMax,
			QualitativeMin:  ranges[i].QualitativeMin,
			QualitativeMax:  ranges[i].QualitativeMax,
		})
	}
	return examTypeItemResponse{
		ID:              item.ID,
		ExamTypeID:      item.ExamTypeID,
		Name:            item.Name,
		InspectionValue: item.InspectionValue,
		NormalValue:     item.NormalValue,
		Unit:            item.Unit,
		SortOrder:       item.SortOrder,
		CreatedAt:       httpapi.LocalTime(item.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(item.UpdatedAt),
		ReferenceRanges: rangeResponses,
	}
}

func toExaminationTypeResponse(et *model.ExaminationType) examTypeResponse {
	return toExaminationTypeResponseWithRanges(et, nil)
}

func toExaminationTypeResponseWithRanges(
	et *model.ExaminationType,
	ranges map[uint64][]model.ExamReferenceRange,
) examTypeResponse {
	items := make([]examTypeItemResponse, 0, len(et.Items))
	for i := range et.Items {
		items = append(items, toExamTypeItemResponseWithRanges(&et.Items[i], ranges[et.Items[i].ID]))
	}
	return examTypeResponse{
		ID:             et.ID,
		ClinicID:       et.ClinicID,
		Name:           et.Name,
		Price:          et.Price,
		IsActive:       et.IsActive,
		Description:    et.Description,
		ParentID:       et.ParentID,
		SortOrder:      et.SortOrder,
		IsNonInsurance: et.IsNonInsurance,
		Items:          items,
		CreatedAt:      httpapi.LocalTime(et.CreatedAt),
		UpdatedAt:      httpapi.LocalTime(et.UpdatedAt),
	}
}
