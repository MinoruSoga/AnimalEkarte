package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestRespondErrorPreferringConflictCode_EmitsCodeAndParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := apperrors.WrapNameConflict(apperrors.CodePermissionGroupNameConflict, "執行")
	RespondErrorPreferringConflictCode(c, err)

	assert.Equal(t, http.StatusConflict, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, apperrors.CodePermissionGroupNameConflict, body["code"])
	params, ok := body["params"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "執行", params["name"])
	// Must not leak constraint/table/SQL in body.
	raw := w.Body.String()
	assert.NotContains(t, raw, "uk_permission")
	assert.NotContains(t, raw, "permission_groups")
	assert.NotContains(t, raw, "23505")
}

func TestRespondErrorPreferringConflictCode_AnimalSpecies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := apperrors.WrapNameConflict(apperrors.CodeAnimalSpeciesNameConflict, "V04動物種類")
	RespondErrorPreferringConflictCode(c, err)

	assert.Equal(t, http.StatusConflict, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, apperrors.CodeAnimalSpeciesNameConflict, body["code"])
	params := body["params"].(map[string]any)
	assert.Equal(t, "V04動物種類", params["name"])
}

func TestRespondErrorPreferringConflictCode_ShiftTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := apperrors.WrapNameConflict(apperrors.CodeShiftTemplateNameConflict, "早番")
	RespondErrorPreferringConflictCode(c, err)

	assert.Equal(t, http.StatusConflict, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, apperrors.CodeShiftTemplateNameConflict, body["code"])
	params := body["params"].(map[string]any)
	assert.Equal(t, "早番", params["name"])
	assert.NotContains(t, w.Body.String(), "uk_shift_templates")
}

func TestRespondErrorPreferringConflictCode_LstepAutoManagedPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	err := apperrors.WrapNameConflict(apperrors.CodeLstepAutoManagedPrefixConflict, "checkup_")
	RespondErrorPreferringConflictCode(c, err)

	assert.Equal(t, http.StatusConflict, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, apperrors.CodeLstepAutoManagedPrefixConflict, body["code"])
	params := body["params"].(map[string]any)
	assert.Equal(t, "checkup_", params["name"])
	assert.NotContains(t, w.Body.String(), "lstep_auto_managed_prefixes_prefix_key")
}

func TestRespondErrorPreferringConflictCode_FallsBackToRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	// Ordinary AlreadyExists keeps legacy {"error": ...} without forcing code.
	RespondErrorPreferringConflictCode(c, apperrors.WrapAlreadyExists("permission_group", ""))
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "permission_group '' already exists")
	// code key may be absent when using RespondError
	assert.NotContains(t, w.Body.String(), `"code"`)
}

func TestRespondError_DatabaseErrorDoesNotLeakConstraint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newTestContext()

	// Unknown DB path (FromGORM default Wrap → "database error") maps to 500 generic.
	// Response must not expose constraint/table/SQL fragments.
	dbErr := apperrors.Wrap(
		errors.New("ERROR: duplicate key value violates unique constraint \"uk_secret\""),
		"database error",
	)
	RespondError(c, dbErr)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	raw := w.Body.String()
	assert.Contains(t, raw, "internal server error")
	assert.NotContains(t, raw, "database error")
	assert.NotContains(t, raw, "uk_secret")
	assert.NotContains(t, raw, "duplicate key")
	assert.NotContains(t, raw, "constraint")
}
