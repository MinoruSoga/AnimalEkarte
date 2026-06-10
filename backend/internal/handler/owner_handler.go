package handler

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListOwners godoc
func (h *Handler) ListOwners(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	query := newListOwnersQuery(c.Request.URL.Query())

	owners, total, err := h.svc.Owner.List(c.Request.Context(), clinicID, page, limit, query.Search)
	if err != nil {
		RespondError(c, err)
		return
	}

	ownerResponses := make([]ownerResponse, 0, len(owners))
	for i := range owners {
		ownerResponses = append(ownerResponses, toOwnerResponse(&owners[i]))
	}
	c.JSON(http.StatusOK, newPaginatedResponse(ownerResponses, total, page, limit))
}

// GetOwner godoc
func (h *Handler) GetOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	owner, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// CreateOwner godoc
func (h *Handler) CreateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// #84: 登録時の医院指定 (Q11=A 所属医院のみ / Q12=A 登録フォームのみ)。
	// req.ClinicID はユーザー入力のためここでの所属検証が唯一の防壁。
	// 検証なしで service へ渡すとクロステナント書き込みになる。
	// system_admin は X-Clinic-ID 検証 (middleware/auth.go) と同様に全医院を許可する。
	if req.ClinicID != nil && *req.ClinicID != clinicID {
		isAdmin, ok := extractIsSystemAdmin(c)
		if !ok {
			return
		}
		if !isAdmin {
			clinicIDs, ok := extractClinicIDs(c)
			if !ok {
				return
			}
			if !slices.Contains(clinicIDs, *req.ClinicID) {
				RespondError(c, apperrors.WrapForbidden("not assigned to this clinic"))
				return
			}
		}
		clinicID = *req.ClinicID
	}

	// BUG-372: discount_rate にゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateFloat(c, req.DiscountRate); err != nil {
		RespondError(c, err)
		return
	}

	input := req.toServiceInput()
	owner, err := h.svc.Owner.CreateWithPets(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/owners/%d", owner.ID))
	c.JSON(http.StatusCreated, toOwnerResponse(owner))
}

// UpdateOwner godoc
func (h *Handler) UpdateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// BUG-372: discount_rate を変更する場合は既存値と比較し権限チェック
	if req.DiscountRate != nil {
		existing, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
		if err != nil {
			RespondError(c, err)
			return
		}
		if err := h.requireDiscountEditFloat(c, req.DiscountRate, existing.DiscountRate); err != nil {
			RespondError(c, err)
			return
		}
	}

	input := req.toServiceInput()
	owner, err := h.svc.Owner.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// PatchOwnerLineUserID godoc
// PATCH /owners/:id/line-user-id — LINE User ID を連携または解除する（BE-005）。
func (h *Handler) PatchOwnerLineUserID(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerLineUserIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	actorID := optionalStaffID(c)
	if err := h.svc.Owner.LinkLineUserID(c.Request.Context(), clinicID, id, req.LineUserID, actorID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PatchOwnerDeliveryExclusion godoc
// PATCH /owners/:id/delivery-exclusion — 配信除外フラグを更新する（FEAT-381）。
func (h *Handler) PatchOwnerDeliveryExclusion(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerDeliveryExclusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.svc.Owner.UpdateDeliveryExclusion(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// PatchOwnerDeliveryCaution godoc
// PATCH /owners/:id/delivery-caution — 配信注意フラグを更新する（FEAT-381-2）。
func (h *Handler) PatchOwnerDeliveryCaution(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerDeliveryCautionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.svc.Owner.UpdateDeliveryCaution(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// PatchOwnerTransferStatus godoc
// PATCH /owners/:id/transfer-status — 転院フラグを更新する（FEAT-381）。
func (h *Handler) PatchOwnerTransferStatus(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerTransferStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.svc.Owner.UpdateTransferStatus(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// PatchOwnerLineIDConfirm godoc
// PATCH /owners/:id/line-id-confirm — LINE ID 紐付け確認日時を現在時刻に設定する（FEAT-381）。
func (h *Handler) PatchOwnerLineIDConfirm(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := optionalStaffID(c)
	owner, err := h.svc.Owner.ConfirmLineID(c.Request.Context(), clinicID, id, actorID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// DeleteOwner godoc
func (h *Handler) DeleteOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	// BE-017: 削除前に Lステップタグを全解除（best-effort、失敗しても削除は続行）
	_ = h.svc.LstepLifecycle.HandleOwnerDeletion(c.Request.Context(), clinicID, id)
	if err := h.svc.Owner.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
