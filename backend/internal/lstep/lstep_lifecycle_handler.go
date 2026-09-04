package lstep

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// UpdatePetDeath godoc
// PATCH /pets/:id/death — ペット死亡を記録し CPM タグを再同期する（BE-017）。
func (h *Handler) UpdatePetDeath(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchPetDeathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if req.DeceasedAt == nil || req.DeceasedAt.IsZero() {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("deceased_at は必須です"))
		return
	}
	deceasedDateJST := req.DeceasedAt.Time.In(config.JST).Format(time.DateOnly)
	todayJST := time.Now().In(config.JST).Format(time.DateOnly)
	if deceasedDateJST > todayJST {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("未来の日付は指定できません"))
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	if err := h.lifecycle.HandlePetDeath(c.Request.Context(), clinicID, petID, req.DeceasedAt.Time, req.Reason, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeletePetDeath godoc
// DELETE /pets/:id/death — ペット死亡取り消しを記録し CPM タグを再同期する（BE-017）。
func (h *Handler) DeletePetDeath(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	if err := h.lifecycle.HandlePetRevival(c.Request.Context(), clinicID, petID, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateOwnerLstepOptOut godoc
// POST /owners/:id/lstep-opt-out — オーナーの Lステップ配信をオプトアウトする（BE-017）。
func (h *Handler) UpdateOwnerLstepOptOut(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req postOwnerLstepOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	if err := h.lifecycle.HandleOwnerOptOut(c.Request.Context(), clinicID, ownerID, req.Reason, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchOwnerLstepOptOut godoc
// PATCH /owners/:id/lstep/opt-out — opt_out:true でオプトアウト、false でオプトインする統合エンドポイント（ISSUE-001）。
//
// BE-refactor.md F-3 の Patch*→Update* リネーム対象外（意図的）: POST側の
// PostOwnerLstepOptOut が同フェーズで UpdateOwnerLstepOptOut にリネームされたため、
// このメソッドも同じ規則を適用すると同名になりコンパイルエラーになる
// （*Handler に同名メソッドを2つ定義できない）。よって Patch プレフィックスのまま維持する。
func (h *Handler) PatchOwnerLstepOptOut(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerLstepOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	var err error
	if req.OptOut {
		err = h.lifecycle.HandleOwnerOptOut(c.Request.Context(), clinicID, ownerID, req.Reason, actorID)
	} else {
		err = h.lifecycle.HandleOwnerOptIn(c.Request.Context(), clinicID, ownerID, actorID)
	}
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOwnerLine godoc
// DELETE /owners/:id/line — LINE User ID 連携を解除する（ISSUE-001）。
func (h *Handler) DeleteOwnerLine(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	if err := h.ownerLineLinker.LinkLineUserID(c.Request.Context(), clinicID, id, nil, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
