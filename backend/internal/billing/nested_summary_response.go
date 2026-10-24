package billing

// nested_summary_response.go — billing-local owner/pet summaries for estimate JSON.
// Smaller than medicalrecord's pet summary; do not unify the contracts.

import (
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// ownerSummaryResponse is the nested owner JSON for billing estimates.
type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary returns nil when o is nil. toOwnerSummary / toPetSummary は固定シグネチャの変換関数のため
// R-NIL（EMR-491）に従い MustOwnerBasic / MustPetBasic を使う。取得経路が hydrate すること（未 hydrate は panic で、空値にはしない）。
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:        o.ID,
		OwnerName: commonbasic.MustOwnerBasic(o).Name,
	}
}

type petSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toPetSummary(p *model.Pet) *petSummaryResponse {
	if p == nil {
		return nil
	}
	return &petSummaryResponse{
		ID:   p.ID,
		Name: commonbasic.MustPetBasic(p).Name,
	}
}
