package medicalrecord

import (
	"fmt"
	"net/http"

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

// ListTreatmentPackages handles GET /masters/treatment-packages.
func (h *TreatmentPackageHandler) ListTreatmentPackages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	views, revision, err := h.service.List(c.Request.Context(), clinicID)
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
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "create") {
		return
	}
	var req createTreatmentPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	view, err := h.service.Create(c.Request.Context(), clinicID, &CreateTreatmentPackageInput{
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
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	view, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTreatmentPackageResponse(view))
}

// UpdateTreatmentPackage handles PATCH /masters/treatment-packages/:id.
func (h *TreatmentPackageHandler) UpdateTreatmentPackage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "edit") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateTreatmentPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	view, err := h.service.Update(c.Request.Context(), clinicID, id, &UpdateTreatmentPackageInput{
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
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "delete") {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
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
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id, uint64(version), actorID); err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderTreatmentPackages handles PATCH /masters/treatment-packages/reorder.
// Route registration must place this static segment before /:id.
func (h *TreatmentPackageHandler) ReorderTreatmentPackages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "edit") {
		return
	}
	var req reorderTreatmentPackagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondTreatmentPackageError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	actorID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.OrderVersion, req.IDs, actorID); err != nil {
		respondTreatmentPackageError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
