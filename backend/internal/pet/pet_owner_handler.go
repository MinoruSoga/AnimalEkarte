package pet

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ListOwnerSharedPets returns pets for which the owner is a secondary owner.
func (h *Handler) ListOwnerSharedPets(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if h.petOwners == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("pet owner handler dependencies are unavailable"))
		return
	}

	sharedPets, err := h.petOwners.GetSharedPetsByOwnerID(
		c.Request.Context(),
		clinicID,
		ownerID,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newOwnerSharedPetsResponse(sharedPets))
}

// ListPetOwners returns the clinic-scoped secondary owners for a pet.
func (h *Handler) ListPetOwners(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if h.petOwners == nil || h.petOwnerDetails == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("pet owner handler dependencies are unavailable"))
		return
	}

	links, err := h.petOwners.GetByPetID(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	owners, err := h.petOwnerDetails.FindByIDs(
		c.Request.Context(),
		clinicID,
		petOwnerIDs(links),
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	response, err := newPetOwnersResponse(links, owners)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

// ReplacePetOwners completely replaces the clinic-scoped secondary owners for a pet.
func (h *Handler) ReplacePetOwners(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var request replacePetOwnersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if h.petOwners == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("pet owner handler dependencies are unavailable"))
		return
	}

	input := request.toServiceInput()
	input.ActorID = &staffID
	input.ActorType = sharedkernel.AuditActorTypeFor(input.ActorID)
	input.IPAddress = c.ClientIP()
	input.UserAgent = c.GetHeader("User-Agent")
	if err := h.petOwners.ReplaceForPet(
		c.Request.Context(),
		clinicID,
		petID,
		input,
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func petOwnerIDs(links []model.PetOwner) []uint64 {
	ids := make([]uint64, 0, len(links))
	seen := make(map[uint64]struct{}, len(links))
	for _, link := range links {
		if _, exists := seen[link.OwnerID]; exists {
			continue
		}
		seen[link.OwnerID] = struct{}{}
		ids = append(ids, link.OwnerID)
	}
	return ids
}
