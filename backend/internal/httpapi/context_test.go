package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractClinicID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		setupContext    func(c *gin.Context)
		wantClinicID    uint64
		wantOK          bool
		wantStatus      int
		wantBodyContain string
	}{
		{
			name: "extracts valid numeric clinic_id from context",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "42")
			},
			wantClinicID: 42,
			wantOK:       true,
		},
		{
			name:            "returns false when clinic_id key is missing",
			setupContext:    func(_ *gin.Context) {},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusUnauthorized,
			wantBodyContain: "missing clinic context",
		},
		{
			name: "returns false when clinic_id is not a string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", 42) // int instead of string
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false when clinic_id is non-numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "not-a-number")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for negative numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "-1")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for empty string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusBadRequest,
			wantBodyContain: "invalid clinic context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupContext(c)

			clinicID, ok := ExtractClinicID(c)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantClinicID, clinicID)
			} else {
				assert.Equal(t, uint64(0), clinicID)
				assert.Equal(t, tt.wantStatus, w.Code)
				assert.Contains(t, w.Body.String(), tt.wantBodyContain)
			}
		})
	}
}

func TestPeekContextHelpersDoNotWriteResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	_, ok := PeekStaffID(c)
	assert.False(t, ok)
	_, ok = PeekClinicID(c)
	assert.False(t, ok)
	_, ok = PeekIsSystemAdmin(c)
	assert.False(t, ok)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	assert.False(t, c.Writer.Written())

	c.Set("user_id", "9")
	c.Set("clinic_id", "11")
	c.Set("is_system_admin", true)
	staffID, ok := PeekStaffID(c)
	assert.True(t, ok)
	assert.Equal(t, uint64(9), staffID)
	clinicID, ok := PeekClinicID(c)
	assert.True(t, ok)
	assert.Equal(t, uint64(11), clinicID)
	isAdmin, ok := PeekIsSystemAdmin(c)
	assert.True(t, ok)
	assert.True(t, isAdmin)
	assert.False(t, c.Writer.Written())
}

func TestAuthorizeClinicIDs_SystemAdminUsesTrustedActiveClinicScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		setupContext    func(c *gin.Context)
		requested       []uint64
		wantAuthorized  bool
		wantStatus      int
		wantBodyContain string
	}{
		{
			name: "allows an active clinic from the trusted context",
			setupContext: func(c *gin.Context) {
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			requested:      []uint64{2},
			wantAuthorized: true,
		},
		{
			name: "rejects a clinic outside the trusted active set",
			setupContext: func(c *gin.Context) {
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			requested:       []uint64{99},
			wantAuthorized:  false,
			wantStatus:      http.StatusForbidden,
			wantBodyContain: "not assigned to this clinic",
		},
		{
			name: "fails closed when the trusted clinic set is missing",
			setupContext: func(c *gin.Context) {
				c.Set("is_system_admin", true)
			},
			requested:       []uint64{1},
			wantAuthorized:  false,
			wantStatus:      http.StatusUnauthorized,
			wantBodyContain: "missing clinic context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupContext(c)

			authorized := AuthorizeClinicIDs(c, tt.requested)

			assert.Equal(t, tt.wantAuthorized, authorized)
			if !tt.wantAuthorized {
				assert.Equal(t, tt.wantStatus, w.Code)
				assert.Contains(t, w.Body.String(), tt.wantBodyContain)
			}
		})
	}
}
