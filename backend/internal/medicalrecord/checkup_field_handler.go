package medicalrecord

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListCheckupTypeFields は GET /v1/masters/checkup-types/:id/fields —
// 健診パッケージのフィールド定義を返す（FE 動的フォーム構築用）。
func (h *CheckupHandler) ListCheckupTypeFields(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceCheckups), "view") {
		return
	}
	checkupTypeID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	fields, err := h.fieldResultService.ListFields(c.Request.Context(), clinicID, checkupTypeID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(fields, toCheckupTypeFieldResponse))
}

// ListCheckupFieldResults は GET /v1/medical-records/:id/checkups/:checkupId/field-results。
func (h *CheckupHandler) ListCheckupFieldResults(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMedicalRecords), "view") {
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
	results, err := h.fieldResultService.ListByCheckup(c.Request.Context(), clinicID, medicalRecordID, checkupID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(results, toCheckupFieldResultResponse))
}

// ReplaceCheckupFieldResults は PUT /v1/medical-records/:id/checkups/:checkupId/field-results。
// 既存全削除→一括登録の PUT セマンティクス。
func (h *CheckupHandler) ReplaceCheckupFieldResults(c *gin.Context) {
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
	var req replaceCheckupFieldResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	saved, err := h.fieldResultService.ReplaceForCheckup(c.Request.Context(), clinicID, medicalRecordID, checkupID, httpapi.OptionalStaffID(c), req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(saved, toCheckupFieldResultResponse))
}

// ListPetCheckupResults は GET /v1/checkups/field-results?pet_id=X —
// pet 単位の健診結果（飼い主レポート用）。
func (h *CheckupHandler) ListPetCheckupResults(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceCheckups), "view") {
		return
	}
	petID, ok := httpapi.ParseOptionalUint64Query(c, "pet_id")
	if !ok {
		return
	}
	if petID == nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("pet_id is required"))
		return
	}
	results, err := h.fieldResultService.ListByPet(c.Request.Context(), clinicID, *petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(results, toPetCheckupResultResponse))
}
