package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- New ----

func TestNew(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	svc := &service.Services{}
	repos := &repository.Repositories{}
	uploader := &mockMedicalRecordImageUploader{}

	h := New(cfg, svc, repos, uploader)

	require.NotNil(t, h)
	assert.Same(t, cfg, h.cfg)
	assert.Same(t, svc, h.svc)
	assert.Same(t, repos, h.repos)
	assert.Same(t, uploader, h.uploader)
}

// ---- Health ----

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", http.NoBody)

	h.Health(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
