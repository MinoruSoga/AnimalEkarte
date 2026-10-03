package medicalrecord

import (
	"fmt"
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

// ── フィールド定義 write 側（EMR-225: POST/PATCH/DELETE /v1/masters/checkup-types/:id/fields*）──
// 権限はルート登録側の perm(model.ResourceCheckups, ...) で強制する（examination-types/:id/fields と同型）。
// リクエスト struct は本ファイルに置く — checkup_field_request.go は結果値（field-results）
// 専用で、定義マスタ用の別拡張子ファイルを増やさない。

// checkupFieldOptionRequest は選択肢1件分のバインド struct（{value,label} 形は
// パッケージ import のマニフェストと同一。manifest 変更は本ユニットの非目標）。
type checkupFieldOptionRequest struct {
	Value string `json:"value" binding:"required"`
	Label string `json:"label" binding:"required"`
}

func (r checkupFieldOptionRequest) toServiceInput() CheckupFieldOptionInput {
	return CheckupFieldOptionInput(r)
}

// createCheckupTypeFieldRequest はフィールド定義作成 POST のバインド struct。
type createCheckupTypeFieldRequest struct {
	Name      string                      `json:"name" binding:"required"`
	FieldType string                      `json:"field_type" binding:"required"`
	Unit      string                      `json:"unit"`
	MinValue  *float64                    `json:"min_value"`
	MaxValue  *float64                    `json:"max_value"`
	Options   []checkupFieldOptionRequest `json:"options" binding:"max=200,dive"`
	SortOrder int                         `json:"sort_order"`
}

func (r createCheckupTypeFieldRequest) toServiceInput() *CreateCheckupTypeFieldInput {
	options := make([]CheckupFieldOptionInput, 0, len(r.Options))
	for _, opt := range r.Options {
		options = append(options, opt.toServiceInput())
	}
	return &CreateCheckupTypeFieldInput{
		Name:      r.Name,
		FieldType: r.FieldType,
		Unit:      r.Unit,
		MinValue:  r.MinValue,
		MaxValue:  r.MaxValue,
		Options:   options,
		SortOrder: r.SortOrder,
	}
}

// updateCheckupTypeFieldRequest はフィールド定義部分更新 PATCH のバインド struct。
// options は省略（キーなし）= 維持、明示配列（空配列含む）= 上書き。min/max は
// clear_min_value / clear_max_value で NULL へ戻せる（clear_parent_id 先例）。
type updateCheckupTypeFieldRequest struct {
	Name          *string                      `json:"name"`
	FieldType     *string                      `json:"field_type"`
	Unit          *string                      `json:"unit"`
	MinValue      *float64                     `json:"min_value"`
	MaxValue      *float64                     `json:"max_value"`
	Options       *[]checkupFieldOptionRequest `json:"options" binding:"omitempty,max=200,dive"`
	SortOrder     *int                         `json:"sort_order"`
	ClearMinValue bool                         `json:"clear_min_value"`
	ClearMaxValue bool                         `json:"clear_max_value"`
}

func (r updateCheckupTypeFieldRequest) toServiceInput() *UpdateCheckupTypeFieldInput {
	var options *[]CheckupFieldOptionInput
	if r.Options != nil {
		inputs := make([]CheckupFieldOptionInput, 0, len(*r.Options))
		for _, opt := range *r.Options {
			inputs = append(inputs, opt.toServiceInput())
		}
		options = &inputs
	}
	return &UpdateCheckupTypeFieldInput{
		Name:          r.Name,
		FieldType:     r.FieldType,
		Unit:          r.Unit,
		MinValue:      r.MinValue,
		MaxValue:      r.MaxValue,
		Options:       options,
		SortOrder:     r.SortOrder,
		ClearMinValue: r.ClearMinValue,
		ClearMaxValue: r.ClearMaxValue,
	}
}

func parseCheckupTypeFieldParent(c *gin.Context) (uint64, uint64, bool) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return 0, 0, false
	}
	checkupTypeID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return 0, 0, false
	}
	return clinicID, checkupTypeID, true
}

func parseCheckupTypeFieldIDs(c *gin.Context) (uint64, uint64, uint64, bool) {
	clinicID, checkupTypeID, ok := parseCheckupTypeFieldParent(c)
	if !ok {
		return 0, 0, 0, false
	}
	fieldID, ok := httpapi.ParseIDParam(c, "fieldId")
	if !ok {
		return 0, 0, 0, false
	}
	return clinicID, checkupTypeID, fieldID, true
}

// checkupFieldServiceAvailable はフィールド定義 write 依存の配線ガード（fail-closed）。
// ルートは fieldService 非nil 時のみ登録されるため通常到達しないが、直接呼出しでも
// nil 参照で panic せず 500 に落とす。
func (h *CheckupHandler) checkupFieldServiceAvailable(c *gin.Context) bool {
	if h.fieldService == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("checkup field service is not configured"))
		return false
	}
	return true
}

// CreateCheckupTypeField は POST /v1/masters/checkup-types/:id/fields。
func (h *CheckupHandler) CreateCheckupTypeField(c *gin.Context) {
	if !h.checkupFieldServiceAvailable(c) {
		return
	}
	clinicID, checkupTypeID, ok := parseCheckupTypeFieldParent(c)
	if !ok {
		return
	}
	var req createCheckupTypeFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	field, err := h.fieldService.CreateField(c.Request.Context(), clinicID, checkupTypeID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/checkup-types/%d/fields/%d", checkupTypeID, field.ID))
	c.JSON(http.StatusCreated, toCheckupTypeFieldResponse(field))
}

// UpdateCheckupTypeField は PATCH /v1/masters/checkup-types/:id/fields/:fieldId。
// field_type 変更も許可する（checkup_field_results はスナップショット保持のため履歴不変）。
func (h *CheckupHandler) UpdateCheckupTypeField(c *gin.Context) {
	if !h.checkupFieldServiceAvailable(c) {
		return
	}
	clinicID, checkupTypeID, fieldID, ok := parseCheckupTypeFieldIDs(c)
	if !ok {
		return
	}
	var req updateCheckupTypeFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	field, err := h.fieldService.UpdateField(c.Request.Context(), clinicID, checkupTypeID, fieldID, req.toServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCheckupTypeFieldResponse(field))
}

// DeleteCheckupTypeField は DELETE /v1/masters/checkup-types/:id/fields/:fieldId（ソフトデリート）。
func (h *CheckupHandler) DeleteCheckupTypeField(c *gin.Context) {
	if !h.checkupFieldServiceAvailable(c) {
		return
	}
	clinicID, checkupTypeID, fieldID, ok := parseCheckupTypeFieldIDs(c)
	if !ok {
		return
	}
	if err := h.fieldService.DeleteField(c.Request.Context(), clinicID, checkupTypeID, fieldID); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderCheckupTypeFields は PATCH /v1/masters/checkup-types/:id/fields/reorder（{ids:[...]}）。
func (h *CheckupHandler) ReorderCheckupTypeFields(c *gin.Context) {
	if !h.checkupFieldServiceAvailable(c) {
		return
	}
	clinicID, checkupTypeID, ok := parseCheckupTypeFieldParent(c)
	if !ok {
		return
	}
	var req httpapi.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	if err := h.fieldService.ReorderFields(c.Request.Context(), clinicID, checkupTypeID, req.IDs); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
