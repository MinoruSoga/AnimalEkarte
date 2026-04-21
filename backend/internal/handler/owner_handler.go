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
	if err := h.svc.Owner.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
