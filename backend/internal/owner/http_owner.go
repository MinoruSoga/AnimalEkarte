package owner

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListOwners godoc
func (h *Handler) ListOwners(c *gin.Context) {
	// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	query := newListOwnersQuery(c.Request.URL.Query())

	owners, total, err := h.service.List(c.Request.Context(), clinicIDs, page, limit, query.Search)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(owners, toOwnerResponse), total, page, limit))
}

// GetOwner godoc
func (h *Handler) GetOwner(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := httpapi.ResolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	owner, err := h.service.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// CreateOwner godoc
func (h *Handler) CreateOwner(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// #84: 登録時の医院指定 (Q11=A 所属医院のみ / Q12=A 登録フォームのみ)。
	// req.ClinicID はユーザー入力のためここでの所属検証が唯一の防壁。
	// 検証なしで service へ渡すとクロステナント書き込みになる。
	// system_admin も middleware が確定した active clinic 集合の範囲だけを許可する。
	if req.ClinicID != nil && *req.ClinicID != clinicID {
		if !httpapi.AuthorizeClinicIDs(c, []uint64{*req.ClinicID}) {
			return
		}
		clinicID = *req.ClinicID
	}

	// BUG-372: discount_rate にゼロ以外を指定する場合は権限要
	if err := httpapi.RequireDiscountCreateFloat(c, httpapi.PermissionChecker(h.hasPermission), req.DiscountRate); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	input := req.toServiceInput()
	owner, err := h.service.CreateWithPets(c.Request.Context(), clinicID, &input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/owners/%d", owner.ID))
	c.JSON(http.StatusCreated, toOwnerResponse(owner))
}

// UpdateOwner godoc
func (h *Handler) UpdateOwner(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	// BUG-372: discount_rate を変更する場合は既存値と比較し権限チェック
	if req.DiscountRate != nil {
		existing, err := h.service.GetByID(c.Request.Context(), clinicID, id)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		if err := httpapi.RequireDiscountEditFloat(c, httpapi.PermissionChecker(h.hasPermission), req.DiscountRate, existing.DiscountRate); err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	input := req.toServiceInput()
	// SEC-CS-F15: pass discount:edit into the write TX for locked-row recheck.
	if h.hasPermission != nil {
		input.DiscountEditAllowed = h.hasPermission(c, string(model.ResourceDiscount), "edit")
	}
	owner, err := h.service.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// UpdateOwnerLineUserID godoc
// PATCH /owners/:id/line-user-id — LINE User ID を連携または解除する（BE-005）。
func (h *Handler) UpdateOwnerLineUserID(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerLineUserIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	if err := h.service.LinkLineUserID(c.Request.Context(), clinicID, id, req.LineUserID, actorID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateOwnerDeliveryExclusion godoc
// PATCH /owners/:id/delivery-exclusion — 配信除外フラグを更新する（FEAT-381）。
func (h *Handler) UpdateOwnerDeliveryExclusion(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerDeliveryExclusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.service.UpdateDeliveryExclusion(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// UpdateOwnerDeliveryCaution godoc
// PATCH /owners/:id/delivery-caution — 配信注意フラグを更新する（FEAT-381-2）。
func (h *Handler) UpdateOwnerDeliveryCaution(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerDeliveryCautionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.service.UpdateDeliveryCaution(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// UpdateOwnerTransferStatus godoc
// PATCH /owners/:id/transfer-status — 転院フラグを更新する（FEAT-381）。
func (h *Handler) UpdateOwnerTransferStatus(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req patchOwnerTransferStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	input := req.toServiceInput()
	owner, err := h.service.UpdateTransferStatus(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// UpdateOwnerLineIDConfirm godoc
// PATCH /owners/:id/line-id-confirm — LINE ID 紐付け確認日時を現在時刻に設定する（FEAT-381）。
func (h *Handler) UpdateOwnerLineIDConfirm(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	actorID := httpapi.OptionalStaffID(c)
	owner, err := h.service.ConfirmLineID(c.Request.Context(), clinicID, id, actorID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerResponse(owner))
}

// DeleteOwner godoc
func (h *Handler) DeleteOwner(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	// BE-017: 削除前に Lステップタグを全解除（best-effort、失敗しても削除は続行）
	if h.deletionLifecycle != nil {
		_ = h.deletionLifecycle.HandleOwnerDeletion(c.Request.Context(), clinicID, id)
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
