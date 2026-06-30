package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ListCheckupTypeFields は GET /v1/masters/checkup-types/:id/fields —
// 健診パッケージのフィールド定義を返す（FE 動的フォーム構築用）。
func (h *Handler) ListCheckupTypeFields(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	checkupTypeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	fields, err := h.svc.CheckupFieldResult.ListFields(c.Request.Context(), clinicID, checkupTypeID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(fields, toCheckupTypeFieldResponse))
}

// ListCheckupFieldResults は GET /v1/medical-records/:id/checkups/:checkupId/field-results。
func (h *Handler) ListCheckupFieldResults(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	checkupID, ok := parseIDParam(c, "checkupId")
	if !ok {
		return
	}
	results, err := h.svc.CheckupFieldResult.ListByCheckup(c.Request.Context(), clinicID, medicalRecordID, checkupID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(results, toCheckupFieldResultResponse))
}

// ReplaceCheckupFieldResults は PUT /v1/medical-records/:id/checkups/:checkupId/field-results。
// 既存全削除→一括登録の PUT セマンティクス。
func (h *Handler) ReplaceCheckupFieldResults(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	checkupID, ok := parseIDParam(c, "checkupId")
	if !ok {
		return
	}
	var req replaceCheckupFieldResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	saved, err := h.svc.CheckupFieldResult.ReplaceForCheckup(c.Request.Context(), clinicID, medicalRecordID, checkupID, optionalStaffID(c), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(saved, toCheckupFieldResultResponse))
}

// ListPetCheckupResults は GET /v1/checkups/field-results?pet_id=X —
// pet 単位の健診結果（飼い主レポート用）。
func (h *Handler) ListPetCheckupResults(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	petID, err := parseOptionalUintQueryFilter(c.Query("pet_id"), "pet_id")
	if err != nil {
		RespondError(c, err)
		return
	}
	if petID == nil {
		RespondError(c, apperrors.WrapInvalidInput("pet_id is required"))
		return
	}
	results, err := h.svc.CheckupFieldResult.ListByPet(c.Request.Context(), clinicID, *petID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapSlice(results, toPetCheckupResultResponse))
}
