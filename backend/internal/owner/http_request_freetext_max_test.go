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

func TestCreateOwnerRequest_OwnerNameMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "255 chars accepted", length: 255, wantErr: false},
		{name: "256 chars rejected", length: 256, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body, err := json.Marshal(map[string]any{
				"owner_name": strings.Repeat("a", tt.length),
			})
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			var req createOwnerRequest
			err = c.ShouldBindJSON(&req)
			if tt.wantErr {
				assert.Error(t, err, "BE-RC-002: owner_name max=255 must reject %d chars", tt.length)
				return
			}
			assert.NoError(t, err, "BE-RC-002: owner_name max=255 must accept %d chars", tt.length)
		})
	}
}

func TestCreatePetForOwnerRequest_NameMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "255 chars accepted", length: 255, wantErr: false},
		{name: "256 chars rejected", length: 256, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body, err := json.Marshal(map[string]any{
				"name":              strings.Repeat("a", tt.length),
				"animal_species_id": 1,
			})
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			var req createPetForOwnerRequest
			err = c.ShouldBindJSON(&req)
			if tt.wantErr {
				assert.Error(t, err, "BE-RC-002: name max=255 must reject %d chars", tt.length)
				return
			}
			assert.NoError(t, err, "BE-RC-002: name max=255 must accept %d chars", tt.length)
		})
	}
}

func TestUpdateOwnerRequest_OwnerNameMax(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "255 chars accepted", length: 255, wantErr: false},
		{name: "256 chars rejected", length: 256, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body, err := json.Marshal(map[string]any{
				"owner_name": strings.Repeat("a", tt.length),
			})
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			var req updateOwnerRequest
			err = c.ShouldBindJSON(&req)
			if tt.wantErr {
				assert.Error(t, err, "BE-RC-002: owner_name max=255 must reject %d chars", tt.length)
				return
			}
			assert.NoError(t, err, "BE-RC-002: owner_name max=255 must accept %d chars", tt.length)
		})
	}
}
