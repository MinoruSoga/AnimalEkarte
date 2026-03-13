package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := service.CreatePetInput{
		OwnerID:         req.OwnerID,
		AnimalSpeciesID: req.AnimalSpeciesID,
		Name:            req.Name,
		PetNameKana:     req.PetNameKana,
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
	c.Header("Location", fmt.Sprintf("/v1/pets/%d", pet.ID))
	c.JSON(http.StatusCreated, toPetResponse(pet))
}

// UpdatePet godoc
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	input := service.UpdatePetInput{
		OwnerID:         req.OwnerID,
		AnimalSpeciesID: req.AnimalSpeciesID,
		PetNumber:       req.PetNumber,
		Name:            req.Name,
		PetNameKana:     req.PetNameKana,
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
	c.JSON(http.StatusOK, toPetResponse(pet))
}

// DeletePet godoc
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
	pets.GET("", h.ListPets)
	pets.POST("", h.CreatePet)
	pets.GET("/:id", h.GetPet)
	pets.PATCH("/:id", h.UpdatePet)
	pets.DELETE("/:id", h.DeletePet)
}
