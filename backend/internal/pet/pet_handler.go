package pet

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// ListPets godoc
// #86: 拠点横断一覧 — clinic_ids クエリ指定時は所属検証済みの複数医院、未指定は現在の医院のみ
// （OwnersList.tsx が useClinicScope 経由でこの一覧を消費するため owners API と同一パターンを維持する）。
func (h *Handler) ListPets(c *gin.Context) {
	clinicIDs, ok := httpapi.ResolveListClinicIDs(c)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	filters, err := newListPetQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	pets, total, err := h.pets.List(c.Request.Context(), clinicIDs, filters, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(pets, toPetListResponse), total, page, limit))
}

// GetPet godoc
func (h *Handler) GetPet(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属医院全体をスコープにしてレコードを取得する
	clinicIDs, ok := httpapi.ResolveAllClinicIDs(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	pet, err := h.pets.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetResponse(pet))
}

// GetPetFirstVisit godoc
// GET /pets/:id/first-visit (#158 飼主レポート: 初診日 = 最古の有効カルテ date)
// 値は medical_records から導出する派生値で、捏造しない（カルテ無し = null）。
// clinic 隔離・論理削除除外は service/repository が担保する。
func (h *Handler) GetPetFirstVisit(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	date, err := h.pets.GetFirstVisitDate(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetFirstVisitResponse(date))
}

// CreatePet godoc
func (h *Handler) CreatePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createPetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.pets.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/pets/%d", pet.ID))
	c.JSON(http.StatusCreated, toPetResponse(pet))
}

// UpdatePet godoc
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.pets.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetResponse(pet))
}

// DeletePet godoc
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.pets.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
