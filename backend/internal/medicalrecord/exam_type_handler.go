package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExamTypeHandler serves the ExaminationType HTTP boundary.
type ExamTypeHandler struct {
	service ExaminationTypeService
}

// NewExamTypeHandler initializes an ExamTypeHandler.
func NewExamTypeHandler(service ExaminationTypeService) *ExamTypeHandler {
	return &ExamTypeHandler{service: service}
}

// ListExaminationTypes godoc
func (h *ExamTypeHandler) ListExaminationTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	exTypes, err := h.service.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	fieldIDs := examTypeFieldIDs(exTypes)
	ranges, err := h.service.ListReferenceRanges(c.Request.Context(), clinicID, fieldIDs)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(exTypes, func(examType *model.ExaminationType) examTypeResponse {
		return toExaminationTypeResponseWithRanges(examType, ranges)
	}))
}

// GetExaminationType godoc
func (h *ExamTypeHandler) GetExaminationType(c *gin.Context) {
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
	et, err := h.service.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	ranges, err := h.service.ListReferenceRanges(c.Request.Context(), clinicID, examTypeFieldIDs([]model.ExaminationType{*et}))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationTypeResponseWithRanges(et, ranges))
}

func (h *ExamTypeHandler) CreateExaminationTypeField(c *gin.Context) {
	clinicID, examTypeID, ok := parseExamTypeFieldParent(c)
	if !ok {
		return
	}
	var req createExamTypeFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	field, err := h.service.CreateField(c.Request.Context(), clinicID, req.toServiceCommand(examTypeID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/examination-types/%d/fields/%d", examTypeID, field.ID))
	c.JSON(http.StatusCreated, toExamTypeItemResponse(field))
}

func (h *ExamTypeHandler) UpdateExaminationTypeField(c *gin.Context) {
	clinicID, examTypeID, fieldID, ok := parseExamTypeFieldIDs(c)
	if !ok {
		return
	}
	var req updateExamTypeFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	result, err := h.service.UpdateField(c.Request.Context(), clinicID, examTypeID, fieldID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeItemResponseWithRanges(&result.Field, result.ReferenceRanges))
}

func (h *ExamTypeHandler) DeleteExaminationTypeField(c *gin.Context) {
	clinicID, examTypeID, fieldID, ok := parseExamTypeFieldIDs(c)
	if !ok {
		return
	}
	if err := h.service.DeleteField(c.Request.Context(), clinicID, examTypeID, fieldID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExamTypeHandler) ReorderExaminationTypeFields(c *gin.Context) {
	clinicID, examTypeID, ok := parseExamTypeFieldParent(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.ReorderFields(c.Request.Context(), clinicID, examTypeID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExamTypeHandler) ReplaceExaminationTypeFieldReferenceRanges(c *gin.Context) {
	clinicID, examTypeID, fieldID, ok := parseExamTypeFieldIDs(c)
	if !ok {
		return
	}
	var req replaceExamReferenceRangesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if req.Ranges == nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("ranges must be a non-null array"))
		return
	}
	result, err := h.service.ReplaceReferenceRanges(
		c.Request.Context(),
		clinicID,
		examTypeID,
		req.toServiceCommand(fieldID),
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExamTypeItemResponseWithRanges(&result.Field, result.ReferenceRanges))
}

func parseExamTypeFieldParent(c *gin.Context) (uint64, uint64, bool) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return 0, 0, false
	}
	examTypeID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return 0, 0, false
	}
	return clinicID, examTypeID, true
}

func parseExamTypeFieldIDs(c *gin.Context) (uint64, uint64, uint64, bool) {
	clinicID, examTypeID, ok := parseExamTypeFieldParent(c)
	if !ok {
		return 0, 0, 0, false
	}
	fieldID, ok := httpapi.ParseIDParam(c, "fieldId")
	if !ok {
		return 0, 0, 0, false
	}
	return clinicID, examTypeID, fieldID, true
}

func examTypeFieldIDs(examTypes []model.ExaminationType) []uint64 {
	fieldIDs := make([]uint64, 0)
	for _, examType := range examTypes {
		for _, item := range examType.Items {
			fieldIDs = append(fieldIDs, item.ID)
		}
	}
	return fieldIDs
}

// CreateExaminationType godoc
func (h *ExamTypeHandler) CreateExaminationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	var req createExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	examType, err := h.service.Create(c.Request.Context(), clinicID, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/examination-types/%d", examType.ID))
	c.JSON(http.StatusCreated, toExaminationTypeResponse(examType))
}

// UpdateExaminationType godoc
func (h *ExamTypeHandler) UpdateExaminationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateExaminationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	exType, err := h.service.Update(c.Request.Context(), clinicID, id, req.toServiceInput())
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.JSON(http.StatusOK, toExaminationTypeResponse(exType))
}

// ReorderExaminationTypes godoc
func (h *ExamTypeHandler) ReorderExaminationTypes(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.service.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteExaminationType godoc
func (h *ExamTypeHandler) DeleteExaminationType(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
