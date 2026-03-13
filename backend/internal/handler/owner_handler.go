package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// CreateOwnerInput はクライアントから受け付けるフィールドのみを定義する DTO
type CreateOwnerInput struct {
	OwnerName      string               `json:"owner_name" binding:"required"`
	OwnerNameKana  string               `json:"owner_name_kana"`
	BirthDate      *time.Time           `json:"birth_date"`
	Company        string               `json:"company"`
	PostalCode     string               `json:"postal_code"`
	Address1       string               `json:"address1"`
	Address2       string               `json:"address2"`
	HomePostalCode string               `json:"home_postal_code"`
	HomeAddress1   string               `json:"home_address1"`
	HomeAddress2   string               `json:"home_address2"`
	Phone          string               `json:"phone"`
	CompanyPhone   string               `json:"company_phone"`
	Email          string               `json:"email"`
	Remarks        string               `json:"remarks"`
	IsDangerous    bool                 `json:"is_dangerous"`
	DiscountRate   float64              `json:"discount_rate"`
	MembershipType model.MembershipType `json:"membership_type"`
	Pets           []CreatePetForOwnerInput `json:"pets"`
}

// CreatePetForOwnerInput は飼主登録時に同時作成するペットの DTO
type CreatePetForOwnerInput struct {
	Name            string     `json:"name" binding:"required"`
	AnimalSpeciesID uint64     `json:"animal_species_id" binding:"required"`
	PetNumber       string     `json:"pet_number"`
	PetNameKana     string     `json:"pet_name_kana"`
	Breed           string     `json:"breed"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Remarks         string     `json:"remarks"`
}

// UpdateOwnerInput は更新可能なフィールドのみを定義する DTO
type UpdateOwnerInput struct {
	OwnerName      string               `json:"owner_name"`
	OwnerNameKana  string               `json:"owner_name_kana"`
	BirthDate      *time.Time           `json:"birth_date"`
	Company        string               `json:"company"`
	PostalCode     string               `json:"postal_code"`
	Address1       string               `json:"address1"`
	Address2       string               `json:"address2"`
	HomePostalCode string               `json:"home_postal_code"`
	HomeAddress1   string               `json:"home_address1"`
	HomeAddress2   string               `json:"home_address2"`
	Phone          string               `json:"phone"`
	CompanyPhone   string               `json:"company_phone"`
	Email          string               `json:"email"`
	Remarks        string               `json:"remarks"`
	IsDangerous    bool                 `json:"is_dangerous"`
	DiscountRate   float64              `json:"discount_rate"`
	MembershipType model.MembershipType `json:"membership_type"`
}

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
	c.JSON(http.StatusOK, PaginatedResponse{Data: owners, Total: total, Page: page, Limit: limit})
}

// GetOwner godoc
func (h *Handler) GetOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	owner, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, owner)
}

// CreateOwner godoc
func (h *Handler) CreateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var input CreateOwnerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner := model.Owner{
		ClinicID:       clinicID,
		OwnerName:      input.OwnerName,
		OwnerNameKana:  input.OwnerNameKana,
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
	}

	// DTO のペットを model.Pet スライスに変換
	pets := make([]model.Pet, 0, len(input.Pets))
	for _, p := range input.Pets {
		pet := model.Pet{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			PetNumber:       p.PetNumber,
			PetNameKana:     p.PetNameKana,
			Breed:           p.Breed,
			BirthDate:       p.BirthDate,
			Remarks:         p.Remarks,
		}
		if p.Gender != "" {
			pet.Gender = model.PetGender(p.Gender)
		}
		pets = append(pets, pet)
	}

	ctx := c.Request.Context()
	if err := h.svc.Owner.CreateWithPets(ctx, &owner, pets); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "owner created",
		slog.Uint64("owner_id", owner.ID),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)),
		slog.Int("pets_count", len(pets)))
	c.JSON(http.StatusCreated, owner)
}

// UpdateOwner godoc
func (h *Handler) UpdateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input UpdateOwnerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner := model.Owner{
		ID:             id,
		ClinicID:       clinicID,
		OwnerName:      input.OwnerName,
		OwnerNameKana:  input.OwnerNameKana,
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
	}

	ctx := c.Request.Context()
	if err := h.svc.Owner.Update(ctx, &owner); err != nil {
		RespondError(c, err)
		return
	}
	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.String("clinic_id", strconv.FormatUint(clinicID, 10)))
	c.JSON(http.StatusOK, owner)
}

// DeleteOwner godoc
func (h *Handler) DeleteOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Owner.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
