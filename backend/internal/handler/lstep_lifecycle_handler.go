package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// PatchPetDeath godoc
// PATCH /pets/:id/death — ペット死亡を記録し CPM タグを再同期する（BE-017）。
func (h *Handler) PatchPetDeath(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchPetDeathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.DeceasedAt == nil {
		RespondError(c, apperrors.WrapInvalidInput("deceased_at は必須です"))
		return
	}
	if err := h.svc.LstepLifecycle.HandlePetDeath(c.Request.Context(), clinicID, petID, req.DeceasedAt.Time, req.Reason); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeletePetDeath godoc
// DELETE /pets/:id/death — ペット死亡取り消しを記録し CPM タグを再同期する（BE-017）。
func (h *Handler) DeletePetDeath(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.LstepLifecycle.HandlePetRevival(c.Request.Context(), clinicID, petID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PostOwnerLstepOptOut godoc
// POST /owners/:id/lstep-opt-out — オーナーの Lステップ配信をオプトアウトする（BE-017）。
func (h *Handler) PostOwnerLstepOptOut(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req postOwnerLstepOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.LstepLifecycle.HandleOwnerOptOut(c.Request.Context(), clinicID, ownerID, req.Reason); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOwnerLstepOptOut godoc
// DELETE /owners/:id/lstep-opt-out — オーナーを Lステップ配信にオプトインする（BE-017）。
func (h *Handler) DeleteOwnerLstepOptOut(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.LstepLifecycle.HandleOwnerOptIn(c.Request.Context(), clinicID, ownerID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchOwnerLstepOptOut godoc
// PATCH /owners/:id/lstep/opt-out — opt_out:true でオプトアウト、false でオプトインする統合エンドポイント（ISSUE-001）。
func (h *Handler) PatchOwnerLstepOptOut(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerLstepOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	var err error
	if req.OptOut {
		err = h.svc.LstepLifecycle.HandleOwnerOptOut(c.Request.Context(), clinicID, ownerID, req.Reason)
	} else {
		err = h.svc.LstepLifecycle.HandleOwnerOptIn(c.Request.Context(), clinicID, ownerID)
	}
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOwnerLine godoc
// DELETE /owners/:id/line — LINE User ID 連携を解除する（ISSUE-001）。
func (h *Handler) DeleteOwnerLine(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Owner.LinkLineUserID(c.Request.Context(), clinicID, id, nil); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
