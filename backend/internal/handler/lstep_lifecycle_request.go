package handler

// patchPetDeathRequest はペット死亡記録リクエスト（BE-017）。
type patchPetDeathRequest struct {
	DeceasedAt *jsonDate `json:"deceased_at" binding:"required"`
	Reason     string    `json:"reason"`
}

// postOwnerLstepOptOutRequest はオーナーオプトアウトリクエスト（BE-017）。
type postOwnerLstepOptOutRequest struct {
	Reason string `json:"reason"`
}

// patchOwnerLstepOptOutRequest は PATCH /owners/:id/lstep/opt-out の統合リクエスト（ISSUE-001）。
// opt_out: true でオプトアウト、false でオプトイン。
type patchOwnerLstepOptOutRequest struct {
	OptOut bool   `json:"opt_out"`
	Reason string `json:"reason"`
}
