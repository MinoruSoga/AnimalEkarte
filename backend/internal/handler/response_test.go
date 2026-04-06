package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	return c, w
}

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
			wantStatus:      http.StatusInternalServerError,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false when clinic_id is non-numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "not-a-number")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusInternalServerError,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for negative numeric string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "-1")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusInternalServerError,
			wantBodyContain: "invalid clinic context",
		},
		{
			name: "returns false for empty string",
			setupContext: func(c *gin.Context) {
				c.Set("clinic_id", "")
			},
			wantClinicID:    0,
			wantOK:          false,
			wantStatus:      http.StatusInternalServerError,
			wantBodyContain: "invalid clinic context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupContext(c)

			clinicID, ok := extractClinicID(c)

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

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
		notInBody  string
	}{
		{
			name:       "not found error maps to 404",
			err:        apperrors.WrapNotFound("owner", "123"),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "already exists error maps to 409",
			err:        apperrors.WrapAlreadyExists("owner", "test@example.com"),
			wantStatus: http.StatusConflict,
			wantBody:   "owner 'test@example.com' already exists",
		},
		{
			name:       "invalid input error maps to 400",
			err:        apperrors.WrapInvalidInput("name is required"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "name is required",
		},
		{
			name:       "unauthorized sentinel maps to 401",
			err:        apperrors.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
		{
			name:       "forbidden sentinel maps to 403",
			err:        apperrors.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   "forbidden",
		},
		{
			name:       "unknown error maps to 500 without sensitive info",
			err:        fmt.Errorf("some internal db error with sensitive info"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
			notInBody:  "sensitive info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext()

			RespondError(c, tt.err)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.notInBody != "" {
				assert.NotContains(t, w.Body.String(), tt.notInBody)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
		wantErr   bool
	}{
		{
			name:      "defaults return page 1 and limit 20",
			query:     "",
			wantPage:  1,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "custom valid values are accepted",
			query:     "page=2&limit=50",
			wantPage:  2,
			wantLimit: 50,
			wantErr:   false,
		},
		{
			name:      "limit of 100 is accepted",
			query:     "page=1&limit=100",
			wantPage:  1,
			wantLimit: 100,
			wantErr:   false,
		},
		{
			name:    "page zero is invalid",
			query:   "page=0",
			wantErr: true,
		},
		{
			name:    "negative page is invalid",
			query:   "page=-1",
			wantErr: true,
		},
		{
			name:    "limit over 100 is invalid",
			query:   "limit=101",
			wantErr: true,
		},
		{
			name:    "limit zero is invalid",
			query:   "limit=0",
			wantErr: true,
		},
		{
			name:    "non-numeric page is invalid",
			query:   "page=abc",
			wantErr: true,
		},
		{
			name:    "non-numeric limit is invalid",
			query:   "limit=xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)

			page, limit, err := parsePagination(c)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, page)
				assert.Equal(t, 0, limit)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPage, page)
				assert.Equal(t, tt.wantLimit, limit)
			}
		})
	}
}
