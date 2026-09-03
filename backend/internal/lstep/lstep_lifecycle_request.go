package lstep

import (
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type jsonDate struct{ time.Time }

func (d *jsonDate) UnmarshalJSON(data []byte) error {
	value := strings.Trim(string(data), `"`)
	if value == "null" || value == "" {
		return nil
	}
	parsed, err := httpapi.ParseFlexibleDate(value)
	if err != nil {
		return apperrors.WrapInvalidInput(httpapi.FlexibleDateInvalidInputMsg)
	}
	d.Time = parsed
	return nil
}

// patchPetDeathRequest はペット死亡記録リクエスト（BE-017）。
type patchPetDeathRequest struct {
	DeceasedAt *jsonDate `json:"deceased_at" binding:"required"`
	Reason     string    `json:"reason" binding:"omitempty,max=500"`
}

// postOwnerLstepOptOutRequest はオーナーオプトアウトリクエスト（BE-017）。
type postOwnerLstepOptOutRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// patchOwnerLstepOptOutRequest は PATCH /owners/:id/lstep/opt-out の統合リクエスト（ISSUE-001）。
// opt_out: true でオプトアウト、false でオプトイン。
type patchOwnerLstepOptOutRequest struct {
	OptOut bool   `json:"opt_out"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}
