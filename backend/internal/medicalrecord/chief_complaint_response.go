package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type chiefComplaintResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toChiefComplaintResponse(cc *model.ChiefComplaintType) chiefComplaintResponse {
	return chiefComplaintResponse{
		ID:          cc.ID,
		ClinicID:    cc.ClinicID,
		Name:        cc.Name,
		Description: cc.Description,
		IsActive:    cc.IsActive,
		SortOrder:   cc.SortOrder,
		CreatedAt:   httpapi.LocalTime(cc.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(cc.UpdatedAt),
	}
}
