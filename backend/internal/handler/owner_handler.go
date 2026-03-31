package handler

import (
	"net/http"
	"strconv"

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
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	pets := make([]service.CreatePetForOwnerInput, 0, len(req.Pets))
	for i := range req.Pets {
		p := &req.Pets[i]
		pets = append(pets, service.CreatePetForOwnerInput{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			PetNameKana:     p.PetNameKana,
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
	c.JSON(http.StatusCreated, toOwnerResponse(owner))
}

// UpdateOwner godoc
func (h *Handler) UpdateOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
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

// DeleteOwner godoc
func (h *Handler) DeleteOwner(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Owner.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterOwnerRoutes は飼主関連のルートを登録する
func (h *Handler) RegisterOwnerRoutes(rg *gin.RouterGroup) {
	owners := rg.Group("/owners")
	owners.GET("", h.ListOwners)
	owners.POST("", h.CreateOwner)
	owners.GET("/:id", h.GetOwner)
	owners.PATCH("/:id", h.UpdateOwner)
	owners.DELETE("/:id", h.DeleteOwner)
}
