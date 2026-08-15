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
