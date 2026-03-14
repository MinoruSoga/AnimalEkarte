package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type vaccineResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toVaccineResponse(v *model.Vaccine) vaccineResponse {
	return vaccineResponse{
		ID:          v.ID,
		ClinicID:    v.ClinicID,
		Name:        v.Name,
		Price:       v.Price,
		IsActive:    v.IsActive,
		Description: v.Description,
		SortOrder:   v.SortOrder,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

// vaccineSummaryResponse はネストされたレスポンスで使用するワクチンの要約型
type vaccineSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toVaccineSummary は *model.Vaccine を *vaccineSummaryResponse に変換する。nilの場合はnilを返す。
func toVaccineSummary(v *model.Vaccine) *vaccineSummaryResponse {
	if v == nil {
		return nil
	}
	return &vaccineSummaryResponse{
		ID:   v.ID,
		Name: v.Name,
	}
}

func toVaccineResponseList(vaccines []model.Vaccine) []vaccineResponse {
	list := make([]vaccineResponse, 0, len(vaccines))
	for i := range vaccines {
		list = append(list, toVaccineResponse(&vaccines[i]))
	}
	return list
}
