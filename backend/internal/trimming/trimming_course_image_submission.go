package trimming

import (
	"bytes"
	"encoding/json"
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"strings"
)

// Probe the cached JSON before required/type/business validation rejects fields.
// Only a canonical, nonempty image key participates. Malformed JSON is rejected.
func protectCourseImageBody(c *gin.Context, clinic uint64, drafts CourseImageDrafts, fields ...string) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
	var body json.RawMessage
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid image-bearing body"))
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("object body required"))
		return false
	}
	// Preserve duplicate/case aliases too: struct JSON binding accepts case-insensitive tags.
	var candidates []json.RawMessage
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		name, ok := token.(string)
		if !ok {
			return false
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return false
		}
		for _, field := range fields {
			if strings.EqualFold(name, field) {
				candidates = append(candidates, value)
				break
			}
		}
	}
	var firstErr error
	for _, value := range candidates {
		var key string
		if err := json.Unmarshal(value, &key); err != nil || key == "" {
			continue
		}
		owned, err := ownedImageKey(clinic, key)
		if err != nil {
			continue
		}
		actor, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return false
		}
		if drafts == nil {
			httpapi.RespondError(c, apperrors.WrapInternalServerError("course image provenance unavailable"))
			return false
		}
		if err := drafts.ProtectSubmission(c.Request.Context(), clinic, actor, owned.value); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		httpapi.RespondError(c, firstErr)
		return false
	}
	return true
}
