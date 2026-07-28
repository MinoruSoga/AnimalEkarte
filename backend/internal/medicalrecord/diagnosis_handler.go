package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// DiagnosisHandler serves the DiagnosisType/DiagnosisName HTTP boundary. The two entities
// share one handler struct (mirroring their pre-move file grouping in
// internal/handler/diagnosis_handler.go) because DiagnosisName has a genuine FK coupling to
// DiagnosisType (#020 FK validation) — they are not independent resources.
type DiagnosisHandler struct {
	typeService DiagnosisTypeService
	nameService DiagnosisNameService
}

// NewDiagnosisHandler initializes a DiagnosisHandler.
func NewDiagnosisHandler(typeService DiagnosisTypeService, nameService DiagnosisNameService) *DiagnosisHandler {
	return &DiagnosisHandler{typeService: typeService, nameService: nameService}
}

// ---- DiagnosisType ----

// ListDiagnosisTypes godoc
func (h *DiagnosisHandler) ListDiagnosisTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	categories, total, err := h.typeService.List(c.Request.Context(), clinicID, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(categories, ToDiagnosisTypeResponse), total, page, limit))
}

// GetDiagnosisType godoc
func (h *DiagnosisHandler) GetDiagnosisType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	category, err := h.typeService.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToDiagnosisTypeResponse(category))
}

// CreateDiagnosisType godoc
func (h *DiagnosisHandler) CreateDiagnosisType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createDiagnosisTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	category, err := h.typeService.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/diagnosis-types/%d", category.ID))
	c.JSON(http.StatusCreated, ToDiagnosisTypeResponse(category))
}

// UpdateDiagnosisType godoc
func (h *DiagnosisHandler) UpdateDiagnosisType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateDiagnosisTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	category, err := h.typeService.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToDiagnosisTypeResponse(category))
}

// DeleteDiagnosisType godoc
func (h *DiagnosisHandler) DeleteDiagnosisType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.typeService.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderDiagnosisTypes godoc (#019)
func (h *DiagnosisHandler) ReorderDiagnosisTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.typeService.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- DiagnosisName ----

// GetDiagnosisName godoc
func (h *DiagnosisHandler) GetDiagnosisName(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	diagnosisName, err := h.nameService.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToDiagnosisNameResponse(diagnosisName))
}

// ListDiagnosisNames godoc
func (h *DiagnosisHandler) ListDiagnosisNames(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, paginationErr := httpapi.ParsePagination(c)
	if paginationErr != nil {
		httpapi.RespondError(c, paginationErr)
		return
	}

	typeID, ok2 := httpapi.ParseOptionalUint64Query(c, "type_id")
	if !ok2 {
		return
	}

	names, total, err := h.nameService.List(c.Request.Context(), clinicID, typeID, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(names, ToDiagnosisNameResponse), total, page, limit))
}

// ListDiagnosisNamesAll godoc
// ページネーションなしで有効な診断名の一覧を返す。
// type_id クエリパラメータが指定された場合は該当カテゴリのみ、未指定の場合は全件を返す。
func (h *DiagnosisHandler) ListDiagnosisNamesAll(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	typeID, ok := httpapi.ParseOptionalUint64Query(c, "type_id")
	if !ok {
		return
	}

	names, err := h.nameService.ListNames(c.Request.Context(), clinicID, typeID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(names, ToDiagnosisNameResponse))
}

// CreateDiagnosisName godoc
func (h *DiagnosisHandler) CreateDiagnosisName(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	diagnosisName, err := h.nameService.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/diagnosis-names/%d", diagnosisName.ID))
	c.JSON(http.StatusCreated, ToDiagnosisNameResponse(diagnosisName))
}

// UpdateDiagnosisName godoc
func (h *DiagnosisHandler) UpdateDiagnosisName(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateDiagnosisNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	diagnosisName, err := h.nameService.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToDiagnosisNameResponse(diagnosisName))
}

// DeleteDiagnosisName godoc
func (h *DiagnosisHandler) DeleteDiagnosisName(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.nameService.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderDiagnosisNames godoc (#019)
func (h *DiagnosisHandler) ReorderDiagnosisNames(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.nameService.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
