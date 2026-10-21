package medicalrecord

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// TreatmentPackageHandler serves the treatment package master HTTP boundary
// (EMR-518/EMR-521 master 6 operations). Route registration/DI is the single
// composition writer's job (routes_masters.go / main.go are out of this stage).
type TreatmentPackageHandler struct {
	service TreatmentPackageService
}

func NewTreatmentPackageHandler(service TreatmentPackageService) *TreatmentPackageHandler {
	return &TreatmentPackageHandler{service: service}
}

// Package routes must answer every early failure with the TreatmentPackageError
// envelope (code required). The shared Extract*/Require* helpers write plain
// {error} bodies, so the package boundary uses peek variants and maps failures
// through respondTreatmentPackageError.
func packageClinicID(c *gin.Context) (uint64, bool) {
	if _, exists := c.Get("clinic_id"); !exists {
		respondTreatmentPackageError(c, apperrors.WrapUnauthorized("missing clinic context"))
		return 0, false
	}
	id, ok := httpapi.PeekClinicID(c)
	if !ok || id == 0 {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput("invalid clinic context"))
		return 0, false
	}
	return id, true
}

func packageStaffID(c *gin.Context) (uint64, bool) {
	if _, exists := c.Get("user_id"); !exists {
		respondTreatmentPackageError(c, apperrors.WrapUnauthorized("missing user context"))
		return 0, false
	}
	id, ok := httpapi.PeekStaffID(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput("invalid user context"))
		return 0, false
	}
	return id, true
}

func packageRequireGrant(c *gin.Context, resource, action string) bool {
	if !httpapi.PeekSelectedClinicGrant(c, resource, action) {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return false
	}
	return true
}

func packageIDParam(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput("id は正の整数で指定してください"))
		return 0, false
	}
	return id, true
}

// ListTreatmentPackages handles GET /masters/treatment-packages.
func (h *TreatmentPackageHandler) ListTreatmentPackages(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	views, revision, err := h.service.List(ctx, clinicID)
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	items := make([]treatmentPackageResponse, 0, len(views))
	for i := range views {
		items = append(items, toTreatmentPackageResponse(&views[i]))
	}
	c.JSON(http.StatusOK, treatmentPackageListResponse{
		Items:        items,
		OrderVersion: formatInt64String(revision),
	})
}

// CreateTreatmentPackage handles POST /masters/treatment-packages.
// No C06 master-create receipt exists; an ambiguous response must not be
// rescued by same-name search or blind re-POST.
func (h *TreatmentPackageHandler) CreateTreatmentPackage(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "create") {
		return
	}
	var req createTreatmentPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := packageStaffID(c)
	if !ok {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	view, err := h.service.Create(ctx, clinicID, &CreateTreatmentPackageInput{
		ActorID:     &actorID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Components:  req.Components,
	})
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/masters/treatment-packages/%d", view.Package.ID))
	c.JSON(http.StatusCreated, toTreatmentPackageResponse(view))
}

// GetTreatmentPackage handles GET /masters/treatment-packages/:id.
func (h *TreatmentPackageHandler) GetTreatmentPackage(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	id, ok := packageIDParam(c, "id")
	if !ok {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	view, err := h.service.GetByID(ctx, clinicID, id)
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPackageResponse(view))
}

// UpdateTreatmentPackage handles PATCH /masters/treatment-packages/:id.
func (h *TreatmentPackageHandler) UpdateTreatmentPackage(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "edit") {
		return
	}
	id, ok := packageIDParam(c, "id")
	if !ok {
		return
	}
	var req updateTreatmentPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := packageStaffID(c)
	if !ok {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	view, err := h.service.Update(ctx, clinicID, id, &UpdateTreatmentPackageInput{
		Version:     req.Version,
		ActorID:     &actorID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Components:  req.Components,
	})
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPackageResponse(view))
}

// DeleteTreatmentPackage handles DELETE /masters/treatment-packages/:id?version=.
// version query is mandatory; logical deletion preserves history.
func (h *TreatmentPackageHandler) DeleteTreatmentPackage(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "delete") {
		return
	}
	id, ok := packageIDParam(c, "id")
	if !ok {
		return
	}
	versionText := c.Query("version")
	if versionText == "" {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput("version query は必須です"))
		return
	}
	version, err := parsePackageCanonicalInt64(versionText, "version")
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	actorID, ok := packageStaffID(c)
	if !ok {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	if err := h.service.Delete(ctx, clinicID, id, version, actorID); err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTreatmentPackages handles PATCH /masters/treatment-packages/reorder.
// Route registration must place this static segment before /:id.
func (h *TreatmentPackageHandler) ReorderTreatmentPackages(c *gin.Context) {
	clinicID, ok := packageClinicID(c)
	if !ok {
		return
	}
	if !packageRequireGrant(c, string(model.ResourceMasterMedical), "edit") {
		return
	}
	var req reorderTreatmentPackagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := packageStaffID(c)
	if !ok {
		return
	}
	ctx, ok := treatmentPackageViewCheckContext(c)
	if !ok {
		respondTreatmentPackageError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	if err := h.service.Reorder(ctx, clinicID, req.OrderVersion, req.IDs, actorID); err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
