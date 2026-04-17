package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type chiefComplaintResponse struct {
	ID          string    `json:"id"`
	ClinicID    string    `json:"clinic_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toChiefComplaintResponse(cc *model.ChiefComplaintType) chiefComplaintResponse {
	return chiefComplaintResponse{
		ID:          strconv.FormatUint(cc.ID, 10),
		ClinicID:    strconv.FormatUint(cc.ClinicID, 10),
		Name:        cc.Name,
		Description: cc.Description,
		IsActive:    cc.IsActive,
		SortOrder:   cc.SortOrder,
		CreatedAt:   cc.CreatedAt,
		UpdatedAt:   cc.UpdatedAt,
	}
}
