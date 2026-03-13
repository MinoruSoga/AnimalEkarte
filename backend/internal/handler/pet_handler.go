package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// CreatePetInput はペット作成時のクライアント DTO
type CreatePetInput struct {
	OwnerID         uint64     `json:"owner_id" binding:"required"`
	AnimalSpeciesID uint64     `json:"animal_species_id" binding:"required"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name" binding:"required"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           string     `json:"breed"`
	Weight          *float64   `json:"weight"`
	Environment     string     `json:"environment"`
	Status          string     `json:"status"`
	InsuranceID     *uint64    `json:"insurance_id"`
	Remarks         string     `json:"remarks"`
}

// UpdatePetInput はペット更新時のクライアント DTO
type UpdatePetInput struct {
	OwnerID         *uint64    `json:"owner_id"`
	AnimalSpeciesID *uint64    `json:"animal_species_id"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           string     `json:"breed"`
	Weight          *float64   `json:"weight"`
	Environment     string     `json:"environment"`
	Status          string     `json:"status"`
	InsuranceID     *uint64    `json:"insurance_id"`
	Remarks         string     `json:"remarks"`
}

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
	c.JSON(http.StatusOK, PaginatedResponse{Data: pets, Total: total, Page: page, Limit: limit})
}

// GetPet godoc
func (h *Handler) GetPet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	pet, err := h.svc.Pet.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pet)
}

// CreatePet godoc
func (h *Handler) CreatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input CreatePetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pet := model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       input.PetNumber,
		Name:            input.Name,
		PetNameKana:     input.PetNameKana,
		BirthDate:       input.BirthDate,
		Breed:           input.Breed,
		Weight:          input.Weight,
		Environment:     input.Environment,
		InsuranceID:     input.InsuranceID,
		Remarks:         input.Remarks,
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}

	ctx := c.Request.Context()
	if err := h.svc.Pet.Create(ctx, &pet); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "pet created",
		slog.Uint64("pet_id", pet.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
	c.JSON(http.StatusCreated, pet)
}

// UpdatePet godoc
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input UpdatePetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pet := model.Pet{
		ID:          id,
		ClinicID:    clinicID,
		PetNumber:   input.PetNumber,
		Name:        input.Name,
		PetNameKana: input.PetNameKana,
		BirthDate:   input.BirthDate,
		Breed:       input.Breed,
		Weight:      input.Weight,
		Environment: input.Environment,
		InsuranceID: input.InsuranceID,
		Remarks:     input.Remarks,
	}
	if input.OwnerID != nil {
		pet.OwnerID = *input.OwnerID
	}
	if input.AnimalSpeciesID != nil {
		pet.AnimalSpeciesID = *input.AnimalSpeciesID
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}

	ctx := c.Request.Context()
	if err := h.svc.Pet.Update(ctx, &pet); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "pet updated",
		slog.Uint64("pet_id", id),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
	c.JSON(http.StatusOK, pet)
}

// DeletePet godoc
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Pet.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
