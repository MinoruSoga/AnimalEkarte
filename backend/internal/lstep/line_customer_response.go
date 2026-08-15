package lstep

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type lineCustomerResponse struct {
	ID               uint64          `json:"id"`
	ClinicID         uint64          `json:"clinic_id"`
	LineUserID       string          `json:"line_user_id"`
	DisplayName      string          `json:"display_name"`
	RealName         string          `json:"real_name"`
	AdditionalFields json.RawMessage `json:"additional_fields"`
	OwnerID          *uint64         `json:"owner_id,omitempty"`
	OwnerName        string          `json:"owner_name,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func toLineCustomerResponse(rc *model.LineCustomer) lineCustomerResponse {
	ownerName := ""
	if rc.Owner != nil {
		ownerName = rc.Owner.Name
	}
	return lineCustomerResponse{
		ID:               rc.ID,
		ClinicID:         rc.ClinicID,
		LineUserID:       rc.LineUserID,
		DisplayName:      rc.DisplayName,
		RealName:         rc.RealName,
		AdditionalFields: json.RawMessage(rc.AdditionalFields),
		OwnerID:          rc.OwnerID,
		OwnerName:        ownerName,
		CreatedAt:        httpapi.LocalTime(rc.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(rc.UpdatedAt),
	}
}
