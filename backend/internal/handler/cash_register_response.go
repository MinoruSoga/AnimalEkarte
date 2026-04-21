package handler

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type cashRegisterCloseResponse struct {
	ID                uint64          `json:"id"`
	ClinicID          uint64          `json:"clinic_id"`
	CloseDate         time.Time       `json:"close_date"`
	Period            string          `json:"period"`
	TheoreticalCash   int64           `json:"theoretical_cash"`
	ActualCash        int64           `json:"actual_cash"`
	CashDifference    int64           `json:"cash_difference"`
	CategoryBreakdown json.RawMessage `json:"category_breakdown"`
	Memo              string          `json:"memo"`
	ClosedBy          *uint64         `json:"closed_by,omitempty"`
	ClosedAt          time.Time       `json:"closed_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toCashRegisterCloseResponse(r *model.CashRegisterClose) cashRegisterCloseResponse {
	return cashRegisterCloseResponse{
		ID:                r.ID,
		ClinicID:          r.ClinicID,
		CloseDate:         r.CloseDate,
		Period:            r.Period,
		TheoreticalCash:   r.TheoreticalCash,
		ActualCash:        r.ActualCash,
		CashDifference:    r.CashDifference,
		CategoryBreakdown: r.CategoryBreakdown,
		Memo:              r.Memo,
		ClosedBy:          r.ClosedBy,
		ClosedAt:          r.ClosedAt,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
