package medicalrecord

import (
	"net/http"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicineDoseParamHandler serves the medicine-dose-param master HTTP boundary. Moved from internal/handler (BE9-2D ⑥ Batch C).
type MedicineDoseParamHandler struct {
	service MedicineDoseParamService
}

// NewMedicineDoseParamHandler initializes a MedicineDoseParamHandler.
func NewMedicineDoseParamHandler(service MedicineDoseParamService) *MedicineDoseParamHandler {
	return &MedicineDoseParamHandler{service: service}
}

// medicine_dose_param_handler.go — #201 B-2c: 薬剤 × 種の投与量パラメータ authoring API。
// ルート: GET /masters/medicines/:id/dose-params,
//        PUT /masters/medicines/:id/dose-params/:species,
//        DELETE /masters/medicines/:id/dose-params/:species

// parseDoseSpeciesParam は :species path param を検証付きで取得する（dog/cat のみ）。
func parseDoseSpeciesParam(c *gin.Context) (model.MedicineDoseSpecies, bool) {
	species := model.MedicineDoseSpecies(c.Param("species"))
	if !model.ValidMedicineDoseSpecies(species) {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("species は dog または cat である必要があります"))
		return "", false
	}
	return species, true
}

// ListMedicineDoseParams は1薬剤の dose パラメータ集合を返す。
func (h *MedicineDoseParamHandler) ListMedicineDoseParams(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMasterMedical), "view") {
		return
	}
	medicineID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	params, err := h.service.List(c.Request.Context(), clinicID, medicineID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineDoseParamListResponse(params))
}

// UpsertMedicineDoseParam は (薬剤, 種) の dose パラメータを作成または全列置換する（idempotent upsert）。
func (h *MedicineDoseParamHandler) UpsertMedicineDoseParam(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicineID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	species, ok := parseDoseSpeciesParam(c)
	if !ok {
		return
	}
	var req upsertMedicineDoseParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	param, err := h.service.Upsert(c.Request.Context(), clinicID, medicineID, req.toServiceInput(species), httpapi.OptionalStaffID(c))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMedicineDoseParamResponse(param))
}

// DeleteMedicineDoseParam は (薬剤, 種) の dose パラメータを論理削除する。
func (h *MedicineDoseParamHandler) DeleteMedicineDoseParam(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicineID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	species, ok := parseDoseSpeciesParam(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), clinicID, medicineID, species, httpapi.OptionalStaffID(c)); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
