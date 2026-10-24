package lstep

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/commonbasic"
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
		// 固定シグネチャのため R-NIL（EMR-491）に従う。取得経路（LINE 顧客 repository）が hydrate すること。
		ownerName = commonbasic.MustOwnerBasic(rc.Owner).Name
	}
	// line_customers.owner_id は drop 済み。飼主との紐付けの正本は owner_line_binding_states（state='linked'）だが、
	// この branch には reader が無いため owner_id は返さない（omitempty で省略する。他列へのフォールバックはしない）。
	return lineCustomerResponse{
		ID:               rc.ID,
		ClinicID:         rc.ClinicID,
		LineUserID:       rc.LineUserID,
		DisplayName:      rc.DisplayName,
		RealName:         rc.RealName,
		AdditionalFields: json.RawMessage(rc.AdditionalFields),
		OwnerName:        ownerName,
		CreatedAt:        httpapi.LocalTime(rc.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(rc.UpdatedAt),
	}
}
