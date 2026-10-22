package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type vaccineResponse struct {
	Version     int64                 `json:"version"`
	ID          uint64                `json:"id"`
	ClinicID    uint64                `json:"clinic_id"`
	Name        string                `json:"name"`
	Price       *int64                `json:"price,omitempty"`
	IsActive    bool                  `json:"is_active"`
	Description string                `json:"description"`
	Species     *model.VaccineSpecies `json:"species,omitempty"`
	Interval    string                `json:"interval"`
	InventoryID *uint64               `json:"inventory_id,omitempty"`
	ParentID    *uint64               `json:"parent_id,omitempty"`
	SortOrder   int                   `json:"sort_order"`
	// IsNonInsurance は null=未保存/旧欠落、true=対象外、false=対象。欠落を false へ補完しない。
	IsNonInsurance *bool `json:"is_non_insurance"`
	// EMR-447 A01: 全体割引除外。null=未保存/旧欠落を raw のまま返し false へ書換えない。
	IsBillDiscountExcluded *bool     `json:"is_bill_discount_excluded"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func toVaccineResponse(v *model.Vaccine) vaccineResponse {
	return vaccineResponse{
		Version:        v.Version,
		ID:             v.ID,
		ClinicID:       v.ClinicID,
		Name:           v.Name,
		Price:          v.Price,
		IsActive:       v.IsActive,
		Description:    v.Description,
		Species:        v.Species,
		Interval:       v.Interval,
		InventoryID:    v.InventoryID,
		ParentID:       v.ParentID,
		SortOrder:      vaccineSortOrderValue(v.SortOrder),
		IsNonInsurance: v.IsNonInsurance,

		IsBillDiscountExcluded: v.IsBillDiscountExcluded,
		CreatedAt:              httpapi.LocalTime(v.CreatedAt),
		UpdatedAt:              httpapi.LocalTime(v.UpdatedAt),
	}
}

// vaccineSortOrderValue は nullable な保存列 sort_order を wire の int へ写す（NULL=0。DB default 0 と同値）。
func vaccineSortOrderValue(sortOrder *int) int {
	if sortOrder == nil {
		return 0
	}
	return *sortOrder
}
