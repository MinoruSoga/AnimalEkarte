package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// medicine_dose_param_handler.go — #201 B-2c: 薬剤 × 種の投与量パラメータ authoring API。
// ルート: GET /masters/medicines/:id/dose-params,
//        PUT /masters/medicines/:id/dose-params/:species,
//        DELETE /masters/medicines/:id/dose-params/:species

// parseDoseSpeciesParam は :species path param を検証付きで取得する（dog/cat のみ）。
func parseDoseSpeciesParam(c *gin.Context) (model.MedicineDoseSpecies, bool) {
	species := model.MedicineDoseSpecies(c.Param("species"))
	if !model.ValidMedicineDoseSpecies(species) {
		RespondError(c, apperrors.WrapInvalidInput("species は dog または cat である必要があります"))
		return "", false
	}
	return species, true
}

// ListMedicineDoseParams は1薬剤の dose パラメータ集合を返す。
func (h *Handler) ListMedicineDoseParams(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicineID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	params, err := h.svc.MedicineDoseParam.List(c.Request.Context(), clinicID, medicineID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineDoseParamListResponse(params))
}

// UpsertMedicineDoseParam は (薬剤, 種) の dose パラメータを作成または全列置換する（idempotent upsert）。
func (h *Handler) UpsertMedicineDoseParam(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicineID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	species, ok := parseDoseSpeciesParam(c)
	if !ok {
		return
	}
	var req upsertMedicineDoseParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	param, err := h.svc.MedicineDoseParam.Upsert(c.Request.Context(), clinicID, medicineID, req.toServiceInput(species), optionalStaffID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineDoseParamResponse(param))
}

// DeleteMedicineDoseParam は (薬剤, 種) の dose パラメータを論理削除する。
func (h *Handler) DeleteMedicineDoseParam(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	medicineID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	species, ok := parseDoseSpeciesParam(c)
	if !ok {
		return
	}
	if err := h.svc.MedicineDoseParam.Delete(c.Request.Context(), clinicID, medicineID, species, optionalStaffID(c)); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
