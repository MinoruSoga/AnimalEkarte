package identitylink

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware gates routes on (resource, action).
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler is the identity-link HTTP boundary.
type Handler struct {
	service           Service
	requirePermission PermissionMiddleware
}

// NewHandler constructs the HTTP boundary.
func NewHandler(service Service, requirePermission PermissionMiddleware) *Handler {
	return &Handler{service: service, requirePermission: requirePermission}
}

func (h *Handler) actorFromContext(c *gin.Context) (ActorContext, bool) {
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return ActorContext{}, false
	}
	homeClinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return ActorContext{}, false
	}
	// Full verified assignment set (not just current X-Clinic-ID).
	clinicIDs, ok := httpapi.ExtractClinicIDs(c)
	if !ok {
		return ActorContext{}, false
	}
	return ActorContext{
		StaffID:         staffID,
		HomeClinicID:    homeClinicID,
		VerifiedClinics: clinicIDs,
		IPAddress:       c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
	}, true
}

// SearchOwners GET /identity-links/owners/search?q=
func (h *Handler) SearchOwners(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	q := c.Query("q")
	if len(q) > 255 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("q must be at most 255 characters"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	owners, err := h.service.SearchOwners(c.Request.Context(), actor, q, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toOwnerSearchItems(owners)})
}

// SearchPets GET /identity-links/pets/search?q=
func (h *Handler) SearchPets(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	q := c.Query("q")
	if len(q) > 255 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("q must be at most 255 characters"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	pets, err := h.service.SearchPets(c.Request.Context(), actor, q, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toPetSearchItems(pets)})
}

// GetOwnerGroup GET /identity-links/owner-groups/:id
func (h *Handler) GetOwnerGroup(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	group, members, err := h.service.GetOwnerGroup(c.Request.Context(), actor, groupID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerGroupResponse(group, members))
}

// FindOwnerGroupByMember GET /identity-links/owners/:clinic_id/:owner_id/group
func (h *Handler) FindOwnerGroupByMember(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	clinicID, err := strconv.ParseUint(c.Param("clinic_id"), 10, 64)
	if err != nil || clinicID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid clinic_id"))
		return
	}
	ownerID, err := strconv.ParseUint(c.Param("owner_id"), 10, 64)
	if err != nil || ownerID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid owner_id"))
		return
	}
	group, members, err := h.service.FindOwnerGroupByMember(c.Request.Context(), actor, clinicID, ownerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerGroupResponse(group, members))
}

// CreateOwnerGroup POST /identity-links/owner-groups
func (h *Handler) CreateOwnerGroup(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	var req CreateOwnerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	group, members, err := h.service.CreateOwnerGroup(c.Request.Context(), actor, req.Members)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toOwnerGroupResponse(group, members))
}

// AddOwnerMembers POST /identity-links/owner-groups/:id/members
func (h *Handler) AddOwnerMembers(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	var req AddOwnerMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	group, members, err := h.service.AddOwnerMembers(c.Request.Context(), actor, groupID, req.Members)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOwnerGroupResponse(group, members))
}

// UnlinkOwnerMember DELETE /identity-links/owner-groups/:id/members
func (h *Handler) UnlinkOwnerMember(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	var req UnlinkOwnerMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.UnlinkOwnerMember(c.Request.Context(), actor, groupID, OwnerMemberRef(req)); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetPetGroup GET /identity-links/pet-groups/:id
func (h *Handler) GetPetGroup(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	group, members, err := h.service.GetPetGroup(c.Request.Context(), actor, groupID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetGroupResponse(group, members))
}

// FindPetGroupByMember GET /identity-links/pets/:clinic_id/:pet_id/group
func (h *Handler) FindPetGroupByMember(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	clinicID, err := strconv.ParseUint(c.Param("clinic_id"), 10, 64)
	if err != nil || clinicID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid clinic_id"))
		return
	}
	petID, err := strconv.ParseUint(c.Param("pet_id"), 10, 64)
	if err != nil || petID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
		return
	}
	group, members, err := h.service.FindPetGroupByMember(c.Request.Context(), actor, clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetGroupResponse(group, members))
}

// CreatePetGroup POST /identity-links/pet-groups
func (h *Handler) CreatePetGroup(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	var req CreatePetGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	group, members, err := h.service.CreatePetGroup(c.Request.Context(), actor, req.OwnerGroupID, req.Members)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toPetGroupResponse(group, members))
}

// AddPetMembers POST /identity-links/pet-groups/:id/members
func (h *Handler) AddPetMembers(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	var req AddPetMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	group, members, err := h.service.AddPetMembers(c.Request.Context(), actor, groupID, req.Members)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetGroupResponse(group, members))
}

// UnlinkPetMember DELETE /identity-links/pet-groups/:id/members
func (h *Handler) UnlinkPetMember(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid group id"))
		return
	}
	var req UnlinkPetMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.UnlinkPetMember(c.Request.Context(), actor, groupID, PetMemberRef(req)); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListLinkedTreatmentHistory GET /identity-links/pets/:clinic_id/:pet_id/treatment-history
func (h *Handler) ListLinkedTreatmentHistory(c *gin.Context) {
	actor, ok := h.actorFromContext(c)
	if !ok {
		return
	}
	clinicID, err := strconv.ParseUint(c.Param("clinic_id"), 10, 64)
	if err != nil || clinicID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid clinic_id"))
		return
	}
	petID, err := strconv.ParseUint(c.Param("pet_id"), 10, 64)
	if err != nil || petID == 0 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid pet_id"))
		return
	}
	includeLinked := c.Query("include_linked") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, total, err := h.service.ListLinkedTreatmentHistory(
		c.Request.Context(), actor, clinicID, petID, includeLinked, page, limit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, LinkedTreatmentHistoryResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// resourceAction helpers keep permission strings centralized.
func identityLinksView() (string, string) {
	return string(model.ResourceIdentityLinks), "view"
}

func identityLinksEdit() (string, string) {
	return string(model.ResourceIdentityLinks), "edit"
}
