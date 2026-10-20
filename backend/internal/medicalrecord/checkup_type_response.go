package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type checkupTypeResponse struct {
	ID          uint64  `json:"id"`
	ClinicID    uint64  `json:"clinic_id"`
	Name        string  `json:"name"`
	Price       *int64  `json:"price,omitempty"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Interval    string  `json:"interval"`
	TargetAge   string  `json:"target_age"`
	ParentID    *uint64 `json:"parent_id,omitempty"`
	SortOrder   int     `json:"sort_order"`
	// EMR-456: null=未保存（旧欠落）、true=保険対象外、false=保険対象。欠落を false/true に補完しない。
	IsNonInsurance *bool     `json:"is_non_insurance"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toCheckupTypeResponse(ct *model.CheckupType) checkupTypeResponse {
	sortOrder := 0
	if ct.SortOrder != nil {
		sortOrder = *ct.SortOrder
	}
	return checkupTypeResponse{
		ID:             ct.ID,
		ClinicID:       ct.ClinicID,
		Name:           ct.Name,
		Price:          ct.Price,
		IsActive:       ct.IsActive,
		Description:    ct.Description,
		Interval:       ct.Interval,
		TargetAge:      ct.TargetAge,
		ParentID:       ct.ParentID,
		SortOrder:      sortOrder,
		IsNonInsurance: ct.IsNonInsurance,
		Version:        ct.Version,
		CreatedAt:      httpapi.LocalTime(ct.CreatedAt),
		UpdatedAt:      httpapi.LocalTime(ct.UpdatedAt),
	}
}
