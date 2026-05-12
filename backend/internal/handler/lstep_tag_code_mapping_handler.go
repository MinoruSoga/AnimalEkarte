package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// --- request / response types ---

type putTagCodeMappingEntryRequest struct {
	CodeType     string   `json:"code_type"     binding:"required"`
	Codes        []string `json:"codes"         binding:"required"`
	SpeciesScope string   `json:"species_scope"`
	AgeMin       *int     `json:"age_min"`
}

type putTagCodeMappingsRequest struct {
	Entries []putTagCodeMappingEntryRequest `json:"entries" binding:"required"`
}

type tagCodeMappingResponse struct {
	ID           uint64   `json:"id"`
	ClinicID     uint64   `json:"clinic_id"`
	TagName      string   `json:"tag_name"`
	CodeType     string   `json:"code_type"`
	Codes        []string `json:"codes"`
	SpeciesScope *string  `json:"species_scope,omitempty"`
	AgeMin       *int     `json:"age_min,omitempty"`
}

func toTagCodeMappingResponse(m *model.LstepTagCodeMapping) tagCodeMappingResponse {
	codes := []string(m.Codes)
	if codes == nil {
		codes = []string{}
	}
	return tagCodeMappingResponse{
		ID:           m.ID,
		ClinicID:     m.ClinicID,
		TagName:      m.TagName,
		CodeType:     m.CodeType,
		Codes:        codes,
		SpeciesScope: m.SpeciesScope,
		AgeMin:       m.AgeMin,
	}
}

func toTagCodeMappingListResponse(ms []*model.LstepTagCodeMapping) []tagCodeMappingResponse {
	out := make([]tagCodeMappingResponse, len(ms))
	for i, m := range ms {
		out[i] = toTagCodeMappingResponse(m)
	}
	return out
}

// --- handlers ---

// ListTagCodeMappings godoc
// GET /api/v1/lstep-tag-code-mappings
func (h *Handler) ListTagCodeMappings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	mappings, err := h.svc.LstepTagCodeMapping.ListMappings(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagCodeMappingListResponse(mappings))
}

// PutTagCodeMappingsForTag godoc
// PUT /api/v1/lstep-tag-code-mappings/:tag_name
func (h *Handler) PutTagCodeMappingsForTag(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	tagName := c.Param("tag_name")
	if tagName == "" {
		RespondError(c, apperrors.WrapInvalidInput("tag_name is required"))
		return
	}

	var req putTagCodeMappingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	entries := make([]service.PutMappingEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = service.PutMappingEntry{
			CodeType:     e.CodeType,
			Codes:        e.Codes,
			SpeciesScope: e.SpeciesScope,
			AgeMin:       e.AgeMin,
		}
	}

	created, err := h.svc.LstepTagCodeMapping.PutMappingsForTag(c.Request.Context(), clinicID, tagName, entries)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-code-mappings/%s", tagName))
	c.JSON(http.StatusOK, toTagCodeMappingListResponse(created))
}

// --- routes ---

// RegisterLstepTagCodeMappingRoutes はタグコードマッピング設定のルートを登録する。
func (h *Handler) RegisterLstepTagCodeMappingRoutes(r *gin.RouterGroup) {
	r.GET("/lstep-tag-code-mappings", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.ListTagCodeMappings)
	r.PUT("/lstep-tag-code-mappings/:tag_name", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.PutTagCodeMappingsForTag)

	// FE が /clinics/:clinic_id/lstep-tag-code-mappings で呼ぶエイリアス
	alias := r.Group("/clinics/:clinic_id/lstep-tag-code-mappings")
	alias.GET("", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.ListTagCodeMappings)
	alias.PUT("/:tag_name", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.PutTagCodeMappingsForTag)
}
