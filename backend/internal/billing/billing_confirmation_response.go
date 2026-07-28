package billing

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type billingConfirmationResponse struct {
	ID              string     `json:"id"`
	MedicalRecordID string     `json:"medical_record_id"`
	Status          string     `json:"status"`
	ConfirmedBy     *string    `json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	ReturnedBy      *string    `json:"returned_by,omitempty"`
	ReturnedAt      *time.Time `json:"returned_at,omitempty"`
	ReturnReason    string     `json:"return_reason"`
	Memo            string     `json:"memo"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func toBillingConfirmationResponse(r *model.BillingConfirmation) billingConfirmationResponse {
	resp := billingConfirmationResponse{
		ID:              strconv.FormatUint(r.ID, 10),
		MedicalRecordID: strconv.FormatUint(r.MedicalRecordID, 10),
		Status:          string(r.Status),
		ConfirmedAt:     httpapi.LocalTimePtr(r.ConfirmedAt),
		ReturnedAt:      httpapi.LocalTimePtr(r.ReturnedAt),
		ReturnReason:    r.ReturnReason,
		Memo:            r.Memo,
		CreatedAt:       httpapi.LocalTime(r.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(r.UpdatedAt),
	}
	if r.ConfirmedBy != nil {
		s := strconv.FormatUint(*r.ConfirmedBy, 10)
		resp.ConfirmedBy = &s
	}
	if r.ReturnedBy != nil {
		s := strconv.FormatUint(*r.ReturnedBy, 10)
		resp.ReturnedBy = &s
	}
	return resp
}
