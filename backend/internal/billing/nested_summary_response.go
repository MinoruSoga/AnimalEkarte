package billing

// nested_summary_response.go — internal/handler/owner_response.go の ownerSummaryResponse の
// documented local copy（estimate 一覧の Owner 埋め込み用・BE9-2C B②。JSON tag 同一で
// byte-identical 出力。billing で実使用があるのは owner summary のみ——staff/pet 等は
// medicalrecord 先例ファイルから複製せず YAGNI で持たない）。

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// ownerSummaryResponse mirrors internal/handler.ownerSummaryResponse (owner_response.go).
type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary mirrors internal/handler.toOwnerSummary. nil の場合は nil を返す。
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}
