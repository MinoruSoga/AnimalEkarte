package medicalrecord

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// CheckupHandler serves the Checkup (定期健診) HTTP boundary, including the typed
// checkup-field-result sub-resource. The two services share one handler struct (mirroring
// their pre-move file grouping in internal/handler/checkup_{handler,field_handler}.go): the
// field results are checkup children (checkup_type field definitions + per-checkup result
// values), not an independent resource. The checkup-field methods live in
// checkup_field_handler.go on this same receiver.
type CheckupHandler struct {
	service            CheckupService
	fieldResultService CheckupFieldResultService
}

// NewCheckupHandler initializes a CheckupHandler.
func NewCheckupHandler(service CheckupService, fieldResultService CheckupFieldResultService) *CheckupHandler {
	return &CheckupHandler{service: service, fieldResultService: fieldResultService}
}

// ListCheckups は指定カルテに紐づく健診記録の一覧を返す
func (h *CheckupHandler) ListCheckups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	checkups, err := h.service.List(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(checkups, toCheckupResponse))
}

// CreateCheckup は指定カルテに健診記録を作成する
func (h *CheckupHandler) CreateCheckup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req createCheckupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput(clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	checkup, err := h.service.Create(c.Request.Context(), id, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/checkups/%d", id, checkup.ID))
	c.JSON(http.StatusCreated, toCheckupResponse(checkup))
}

// UpdateCheckup は健診記録を部分更新する
func (h *CheckupHandler) UpdateCheckup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	checkupID, ok := httpapi.ParseIDParam(c, "checkupId")
	if !ok {
		return
	}

	var req updateCheckupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	svcInput, err := req.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	checkup, err := h.service.Update(c.Request.Context(), clinicID, medicalRecordID, checkupID, svcInput)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupResponse(checkup))
}

// DeleteCheckup は健診記録を soft delete する
func (h *CheckupHandler) DeleteCheckup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	checkupID, ok := httpapi.ParseIDParam(c, "checkupId")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, checkupID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListGlobalCheckups は GET /v1/checkups — クリニック横断の健診記録一覧を返す
func (h *CheckupHandler) ListGlobalCheckups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	query := newListGlobalCheckupsQuery(clinicID, c.Request.URL.Query())
	input, err := query.toServiceInput()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	input.Page = page
	input.Limit = limit

	checkups, total, err := h.service.ListByClinic(c.Request.Context(), input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	responses := httpapi.MapSlice(checkups, toCheckupGlobalResponse)
	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(responses, total, page, limit))
}
