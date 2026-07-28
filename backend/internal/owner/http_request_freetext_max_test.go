package owner

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

func TestCreateOwnerRequest_RejectsOverlongFreeText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, err := json.Marshal(map[string]any{
		"owner_name": "山田太郎",
		"company":    strings.Repeat("a", 201),
	})
	require.NoError(t, err)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req createOwnerRequest
	err = c.ShouldBindJSON(&req)
	assert.Error(t, err, "POC-13: company max=200 must reject 201 chars")
}
