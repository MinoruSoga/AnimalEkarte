package lstep

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// --- handlers ---

// ListTagCodeMappings godoc
// GET /api/v1/lstep-tag-code-mappings
func (h *Handler) ListTagCodeMappings(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	mappings, err := h.tagCodeMapping.ListMappings(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTagCodeMappingListResponse(mappings))
}

// ReplaceTagCodeMappingsForTag godoc
// PUT /api/v1/lstep-tag-code-mappings/:tag_name
func (h *Handler) ReplaceTagCodeMappingsForTag(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	tagName := c.Param("tag_name")
	if tagName == "" {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("tag_name is required"))
		return
	}

	var req putTagCodeMappingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	entries := make([]PutMappingEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = PutMappingEntry(e)
	}

	created, err := h.tagCodeMapping.PutMappingsForTag(c.Request.Context(), clinicID, tagName, entries)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/lstep-tag-code-mappings/%s", tagName))
	c.JSON(http.StatusOK, toTagCodeMappingListResponse(created))
}
