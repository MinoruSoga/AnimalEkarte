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
