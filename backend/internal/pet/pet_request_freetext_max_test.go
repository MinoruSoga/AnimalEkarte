package pet

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePetRequest_RejectsOverlongFreeText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"owner_id":          1,
		"animal_species_id": 1,
		"name":              "ポチ",
		"breed":             strings.Repeat("a", 101),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req createPetRequest
	err = c.ShouldBindJSON(&req)
	assert.Error(t, err, "POC-13: breed max=100 must reject 101 runes/bytes")
}

func TestCreatePetRequest_RejectsOverlongName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"owner_id":          1,
		"animal_species_id": 1,
		"name":              strings.Repeat("a", 256),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req createPetRequest
	err = c.ShouldBindJSON(&req)
	assert.Error(t, err, "BE-RC-002: name max=255 must reject 256 chars")
}

func TestCreatePetRequest_AcceptsNameAtMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"owner_id":          1,
		"animal_species_id": 1,
		"name":              strings.Repeat("a", 255),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req createPetRequest
	err = c.ShouldBindJSON(&req)
	assert.NoError(t, err, "BE-RC-002: name max=255 must accept 255 chars")
}

func TestUpdatePetRequest_RejectsOverlongName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"name": strings.Repeat("a", 256),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req updatePetRequest
	err = c.ShouldBindJSON(&req)
	assert.Error(t, err, "BE-RC-002: update name max=255 must reject 256 chars")
}

func TestUpdatePetRequest_AcceptsNameAtMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"name": strings.Repeat("a", 255),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req updatePetRequest
	err = c.ShouldBindJSON(&req)
	assert.NoError(t, err, "BE-RC-002: update name max=255 must accept 255 chars")
}
