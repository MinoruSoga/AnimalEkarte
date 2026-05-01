package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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
	search := c.Query("search")

	owners, total, err := h.svc.Owner.List(c.Request.Context(), clinicID, page, limit, search)
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

	// BUG-372: discount_rate にゼロ以外を指定する場合は権限要
	if err := h.requireDiscountCreateFloat(c, req.DiscountRate); err != nil {
		RespondError(c, err)
		return
	}

	pets := make([]service.CreatePetForOwnerInput, 0, len(req.Pets))
	for i := range req.Pets {
		p := &req.Pets[i]
		pets = append(pets, service.CreatePetForOwnerInput{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			PetNameKana:     p.NameKana,
			Breed:           p.Breed,
			Color:           p.Color,
			Gender:          p.Gender,
			Status:          p.Status,
			BirthDate:       jsonDatePtr(p.BirthDate),
			Weight:          p.Weight,
			NeuteredDate:    jsonDatePtr(p.NeuteredDate),
			AcquisitionType: p.AcquisitionType,
			DangerLevel:     p.DangerLevel,
			Food:            p.Food,
			Environment:     p.Environment,
			InsuranceID:     p.InsuranceID,
			Remarks:         p.Remarks,
		})
	}
	input := service.CreateOwnerInput{
		OwnerName:      req.OwnerName,
		OwnerNameKana:  req.OwnerNameKana,
		BirthDate:      jsonDatePtr(req.BirthDate),
		Company:        req.Company,
		PostalCode:     req.PostalCode,
		Address1:       req.Address1,
		Address2:       req.Address2,
		HomePostalCode: req.HomePostalCode,
		HomeAddress1:   req.HomeAddress1,
		HomeAddress2:   req.HomeAddress2,
		Phone:          req.Phone,
		CompanyPhone:   req.CompanyPhone,
		Email:          req.Email,
		Remarks:        req.Remarks,
		IsDangerous:    req.IsDangerous,
		DiscountRate:   req.DiscountRate,
		MembershipType: model.MembershipType(req.MembershipType),
		Pets:           pets,
	}

	owner, err := h.svc.Owner.CreateWithPets(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-005: 動物分類タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncOwnerAnimalClassificationTags(c.Request.Context(), clinicID, owner.ID)
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

	var membershipType *model.MembershipType
	if req.MembershipType != nil {
		mt := model.MembershipType(*req.MembershipType)
		membershipType = &mt
	}
	input := service.UpdateOwnerInput{
		OwnerName:      req.OwnerName,
		OwnerNameKana:  req.OwnerNameKana,
		BirthDate:      jsonDatePtr(req.BirthDate),
		Company:        req.Company,
		PostalCode:     req.PostalCode,
		Address1:       req.Address1,
		Address2:       req.Address2,
		HomePostalCode: req.HomePostalCode,
		HomeAddress1:   req.HomeAddress1,
		HomeAddress2:   req.HomeAddress2,
		Phone:          req.Phone,
		CompanyPhone:   req.CompanyPhone,
		Email:          req.Email,
		Remarks:        req.Remarks,
		IsDangerous:    req.IsDangerous,
		DiscountRate:   req.DiscountRate,
		MembershipType: membershipType,
	}

	owner, err := h.svc.Owner.Update(c.Request.Context(), clinicID, id, &input)
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
	if err := h.svc.Owner.LinkLineUserID(c.Request.Context(), clinicID, id, req.LineUserID); err != nil {
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
	input := service.UpdateDeliveryExclusionInput{
		Excluded: req.Excluded,
		Reason:   req.Reason,
	}
	owner, err := h.svc.Owner.UpdateDeliveryExclusion(c.Request.Context(), clinicID, id, input)
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
	input := service.UpdateTransferStatusInput{
		IsTransferred: req.IsTransferred,
	}
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
	owner, err := h.svc.Owner.ConfirmLineID(c.Request.Context(), clinicID, id)
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
