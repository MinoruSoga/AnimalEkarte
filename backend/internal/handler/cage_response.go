package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type cageResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	CageType    string    `json:"cage_type"`
	CageSize    string    `json:"cage_size"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCageResponse(c *model.Cage) cageResponse {
	return cageResponse{
		ID:          c.ID,
		ClinicID:    c.ClinicID,
		Name:        c.Name,
		CageType:    string(c.CageType),
		CageSize:    string(c.CageSize),
		Price:       c.Price,
		IsActive:    c.IsActive,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// cageSummaryResponse はネストされたレスポンスで使用するケージの要約型
type cageSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toCageSummary は *model.Cage を *cageSummaryResponse に変換する。nilの場合はnilを返す。
func toCageSummary(c *model.Cage) *cageSummaryResponse {
	if c == nil {
		return nil
	}
	return &cageSummaryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}

func toCageResponseList(cages []model.Cage) []cageResponse {
	list := make([]cageResponse, 0, len(cages))
	for i := range cages {
		list = append(list, toCageResponse(&cages[i]))
	}
	return list
}
