package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestToExamTypeItemResponse(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 9, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		item *model.ExamTypeField
		want examTypeItemResponse
	}{
		{
			name: "converts populated field",
			item: &model.ExamTypeField{
				ID:              1,
				ExamTypeID:      2,
				Name:            "WBC",
				InspectionValue: "8000",
				NormalValue:     "4000-9000",
				Unit:            "/uL",
				SortOrder:       3,
				CreatedAt:       createdAt,
			},
			want: examTypeItemResponse{
				ID:              1,
				ExamTypeID:      2,
				Name:            "WBC",
				InspectionValue: "8000",
				NormalValue:     "4000-9000",
				Unit:            "/uL",
				SortOrder:       3,
				CreatedAt:       httpapi.LocalTime(createdAt),
			},
		},
		{
			name: "converts zero-value field",
			item: &model.ExamTypeField{},
			want: examTypeItemResponse{
				CreatedAt: httpapi.LocalTime(time.Time{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toExamTypeItemResponse(tt.item)
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.ExamTypeID != tt.want.ExamTypeID {
				t.Errorf("ExamTypeID = %v, want %v", got.ExamTypeID, tt.want.ExamTypeID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.InspectionValue != tt.want.InspectionValue {
				t.Errorf("InspectionValue = %q, want %q", got.InspectionValue, tt.want.InspectionValue)
			}
			if got.NormalValue != tt.want.NormalValue {
				t.Errorf("NormalValue = %q, want %q", got.NormalValue, tt.want.NormalValue)
			}
			if got.Unit != tt.want.Unit {
				t.Errorf("Unit = %q, want %q", got.Unit, tt.want.Unit)
			}
			if got.SortOrder != tt.want.SortOrder {
				t.Errorf("SortOrder = %v, want %v", got.SortOrder, tt.want.SortOrder)
			}
			if !got.CreatedAt.Equal(tt.want.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
			}
		})
	}
}

func TestToExaminationTypeResponse(t *testing.T) {
	createdAt := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 2, 2, 11, 0, 0, 0, time.UTC)
	price := int64(3300)
	parentID := uint64(9)

	tests := []struct {
		name      string
		et        *model.ExaminationType
		wantItems int
		wantPrice *int64
		wantParen *uint64
	}{
		{
			name: "converts exam type with items and optional fields",
			et: &model.ExaminationType{
				ID:             1,
				ClinicID:       2,
				Name:           "血液検査",
				Price:          &price,
				IsActive:       true,
				Description:    "CBC",
				ParentID:       &parentID,
				SortOrder:      5,
				IsNonInsurance: true,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				Items: []model.ExamTypeField{
					{ID: 10, ExamTypeID: 1, Name: "WBC"},
					{ID: 11, ExamTypeID: 1, Name: "RBC"},
				},
			},
			wantItems: 2,
			wantPrice: &price,
			wantParen: &parentID,
		},
		{
			name: "converts exam type with nil items and no optional fields",
			et: &model.ExaminationType{
				ID:        3,
				ClinicID:  4,
				Name:      "検査",
				IsActive:  false,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Items:     nil,
			},
			wantItems: 0,
			wantPrice: nil,
			wantParen: nil,
		},
		{
			name:      "converts zero-value exam type",
			et:        &model.ExaminationType{},
			wantItems: 0,
			wantPrice: nil,
			wantParen: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toExaminationTypeResponse(tt.et)

			if got.ID != tt.et.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.et.ID)
			}
			if got.ClinicID != tt.et.ClinicID {
				t.Errorf("ClinicID = %v, want %v", got.ClinicID, tt.et.ClinicID)
			}
			if got.Name != tt.et.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.et.Name)
			}
			if got.IsActive != tt.et.IsActive {
				t.Errorf("IsActive = %v, want %v", got.IsActive, tt.et.IsActive)
			}
			if got.Description != tt.et.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.et.Description)
			}
			if got.SortOrder != tt.et.SortOrder {
				t.Errorf("SortOrder = %v, want %v", got.SortOrder, tt.et.SortOrder)
			}
			if got.IsNonInsurance != tt.et.IsNonInsurance {
				t.Errorf("IsNonInsurance = %v, want %v", got.IsNonInsurance, tt.et.IsNonInsurance)
			}
			if got.Items == nil {
				t.Errorf("Items = nil, want non-nil empty slice when no items")
			}
			if len(got.Items) != tt.wantItems {
				t.Errorf("len(Items) = %d, want %d", len(got.Items), tt.wantItems)
			}
			if (got.Price == nil) != (tt.wantPrice == nil) {
				t.Errorf("Price nilness = %v, want %v", got.Price == nil, tt.wantPrice == nil)
			}
			if got.Price != nil && tt.wantPrice != nil && *got.Price != *tt.wantPrice {
				t.Errorf("Price = %v, want %v", *got.Price, *tt.wantPrice)
			}
			if (got.ParentID == nil) != (tt.wantParen == nil) {
				t.Errorf("ParentID nilness = %v, want %v", got.ParentID == nil, tt.wantParen == nil)
			}
			if got.ParentID != nil && tt.wantParen != nil && *got.ParentID != *tt.wantParen {
				t.Errorf("ParentID = %v, want %v", *got.ParentID, *tt.wantParen)
			}
			if !got.CreatedAt.Equal(httpapi.LocalTime(tt.et.CreatedAt)) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, httpapi.LocalTime(tt.et.CreatedAt))
			}
			if !got.UpdatedAt.Equal(httpapi.LocalTime(tt.et.UpdatedAt)) {
				t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, httpapi.LocalTime(tt.et.UpdatedAt))
			}
		})
	}
}

func TestToExaminationTypeResponseWithRanges_AttachesRangesToMatchingField(t *testing.T) {
	min := 1.5
	examType := &model.ExaminationType{
		ID: 10,
		Items: []model.ExamTypeField{
			{ID: 20, ExamTypeID: 10, Name: "WBC"},
			{ID: 21, ExamTypeID: 10, Name: "RBC"},
		},
	}
	response := toExaminationTypeResponseWithRanges(examType, map[uint64][]model.ExamReferenceRange{
		20: {{ID: 30, ExamTypeFieldID: 20, AnimalSpeciesID: 40, RefMin: &min}},
	})

	require.Len(t, response.Items, 2)
	require.Len(t, response.Items[0].ReferenceRanges, 1)
	assert.Equal(t, uint64(40), response.Items[0].ReferenceRanges[0].AnimalSpeciesID)
	assert.Empty(t, response.Items[1].ReferenceRanges)
}
