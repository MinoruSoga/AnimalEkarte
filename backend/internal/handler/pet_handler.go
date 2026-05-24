package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListPets godoc
func (h *Handler) ListPets(c *gin.Context) {
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

	var ownerID *uint64
	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		id, err := strconv.ParseUint(ownerIDStr, 10, 64)
		if err != nil {
			RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
			return
		}
		ownerID = &id
	}

	pets, total, err := h.svc.Pet.List(c.Request.Context(), clinicID, ownerID, page, limit, search)
	if err != nil {
		RespondError(c, err)
		return
	}

	petResponses := make([]petListResponse, 0, len(pets))
	for i := range pets {
		petResponses = append(petResponses, toPetListResponse(&pets[i]))
	}
	c.JSON(http.StatusOK, newPaginatedResponse(petResponses, total, page, limit))
}

// GetPet godoc
func (h *Handler) GetPet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetResponse(pet))
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

	input := service.CreatePetInput{
		OwnerID:         req.OwnerID,
		AnimalSpeciesID: req.AnimalSpeciesID,
		Name:            req.Name,
		PetNameKana:     req.NameKana,
		Gender:          req.Gender,
		Status:          req.Status,
		BirthDate:       jsonDatePtr(req.BirthDate),
		Breed:           req.Breed,
		Color:           req.Color,
		Weight:          req.Weight,
		NeuteredDate:    jsonDatePtr(req.NeuteredDate),
		AcquisitionType: req.AcquisitionType,
		DangerLevel:     req.DangerLevel,
		Food:            req.Food,
		Environment:     req.Environment,
		Phone:           req.Phone,
		InsuranceID:     req.InsuranceID,
		Remarks:         req.Remarks,
	}

	pet, err := h.svc.Pet.Create(c.Request.Context(), clinicID, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-005: ペット基本情報・動物分類タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncOwnerAnimalClassificationTags(c.Request.Context(), clinicID, pet.OwnerID)
	_ = h.svc.LstepTagSync.SyncPetBasicInfoTags(c.Request.Context(), clinicID, pet.OwnerID)
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

	input := service.UpdatePetInput{
		OwnerID:         req.OwnerID,
		AnimalSpeciesID: req.AnimalSpeciesID,
		PetNumber:       req.PetNumber,
		Name:            req.Name,
		PetNameKana:     req.NameKana,
		Gender:          req.Gender,
		Status:          req.Status,
		BirthDate:       jsonDatePtr(req.BirthDate),
		Breed:           req.Breed,
		Color:           req.Color,
		Weight:          req.Weight,
		NeuteredDate:    jsonDatePtr(req.NeuteredDate),
		AcquisitionType: req.AcquisitionType,
		DangerLevel:     req.DangerLevel,
		Food:            req.Food,
		Environment:     req.Environment,
		Phone:           req.Phone,
		LastVisit:       jsonDatePtr(req.LastVisit),
		InsuranceID:     req.InsuranceID,
		Remarks:         req.Remarks,
	}

	pet, err := h.svc.Pet.Update(c.Request.Context(), clinicID, id, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	// BE-005: ペット基本情報・動物分類タグ同期（best-effort）
	_ = h.svc.LstepTagSync.SyncOwnerAnimalClassificationTags(c.Request.Context(), clinicID, pet.OwnerID)
	_ = h.svc.LstepTagSync.SyncPetBasicInfoTags(c.Request.Context(), clinicID, pet.OwnerID)
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
	// BE-017: ペット死亡ライフサイクル
	pets.PATCH("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchPetDeath)
	pets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)
	// BE-012: 慢性疾患フラグ
	h.RegisterChronicConditionRoutes(pets)

	// ISSUE-007: FE が /clinics/:clinic_id/pets/:id/death で呼ぶエイリアス
	clinicPets := rg.Group("/clinics/:clinic_id/pets")
	clinicPets.PATCH("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.PatchPetDeath)
	clinicPets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)
}
