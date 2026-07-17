package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListPets godoc
// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
// （OwnersList.tsx が useClinicScope 経由でこの一覧を消費するため owners API と同一パターンを維持する）。
func (h *Handler) ListPets(c *gin.Context) {
	clinicIDs, ok := resolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}
	filters, err := newListPetQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		RespondError(c, err)
		return
	}

	pets, total, err := h.svc.Pet.List(c.Request.Context(), clinicIDs, filters, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(pets, toPetListResponse), total, page, limit))
}

// GetPet godoc
func (h *Handler) GetPet(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := resolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	pet, err := h.svc.Pet.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetResponse(pet))
}

// GetPetFirstVisit godoc
// GET /pets/:id/first-visit (#158 飼主レポート: 初診日 = 最古の有効カルテ date)
// 値は medical_records から導出する派生値で、捏造しない（カルテ無し = null）。
// clinic 隔離・論理削除除外は service/repository が担保する。
func (h *Handler) GetPetFirstVisit(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	date, err := h.svc.Pet.GetFirstVisitDate(c.Request.Context(), clinicID, petID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetFirstVisitResponse(date))
}

// CreatePet godoc
func (h *Handler) CreatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req createPetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.svc.Pet.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/pets/%d", pet.ID))
	c.JSON(http.StatusCreated, toPetResponse(pet))
}

// UpdatePet godoc
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.svc.Pet.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetResponse(pet))
}

// DeletePet godoc
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Pet.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterPetRoutes はペット関連のルートを登録する
func (h *Handler) RegisterPetRoutes(rg *gin.RouterGroup) {
	pets := rg.Group("/pets")
	pets.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListPets)
	pets.GET("/:id", h.RequirePermission(string(model.ResourceOwners), "view"), h.GetPet)
	pets.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreatePet)
	pets.PATCH("/:id", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdatePet)
	pets.DELETE("/:id", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeletePet)
	// #158 飼主レポート: ペット単位の治療履歴（投薬/手術/治療）。
	// 治療データはカルテ内容のため medical-records:view でゲートする（owners 権限ではない）。
	pets.GET("/:id/treatment-history", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.ListPetTreatmentHistory)
	// #158 飼主レポート: ペットの初診日（最古カルテ date 由来）。カルテ内容由来のため medical-records:view でゲートする。
	pets.GET("/:id/first-visit", h.RequirePermission(string(model.ResourceMedicalRecords), "view"), h.GetPetFirstVisit)
	// BE-017: ペット死亡ライフサイクル
	pets.PATCH("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdatePetDeath)
	// P6 例外: 死亡記録の解除は PATCH 相当の状態トグル（死亡記録を PATCH で付与した状態を戻すだけで、
	// リソース自体は削除されない）のため、DELETE でも delete ではなく edit を要求する
	// （BE-refactor.md X-15、PO 決定 2026-07-12）。edit 可・delete 不可の一般ロールでも
	// 「死亡記録を解除」ボタン（PetDeceasedBanner.tsx、canEdit で表示制御）を操作可能にする現行 UX を保つ。
	pets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)
	// BE-012: 慢性疾患フラグ
	h.RegisterChronicConditionRoutes(pets)

	// ISSUE-007: FE が /clinics/:clinic_id/pets/:id/death で呼ぶエイリアス
	clinicPets := rg.Group("/clinics/:clinic_id/pets")
	clinicPets.PATCH("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.UpdatePetDeath)
	// P6 例外: 上記 pets.DELETE("/:id/death") と同じ理由（BE-refactor.md X-15、PO 決定 2026-07-12）。
	clinicPets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)
}
