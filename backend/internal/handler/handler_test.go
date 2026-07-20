package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// mockFileUploader is a no-op infra.FileUploader for wiring/identity tests. The
// medical-record-image handler's own upload-path mock moved to internal/medicalrecord with that
// handler (BE9-2D sub-batch④a); this residual copy exists only to satisfy New()'s uploader arg.
type mockFileUploader struct{}

func (m *mockFileUploader) Upload(_ context.Context, _ string, _ io.Reader, _ string) (string, error) {
	return "", nil
}

func (m *mockFileUploader) Delete(_ context.Context, _ string) error { return nil }

var _ infra.FileUploader = (*mockFileUploader)(nil)

// ---- New ----

func TestNew(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	svc := &service.Services{}
	// nil *gorm.DB は問題ない — このテストは identity 配線のみ検証し、メソッドは呼び出さない。
	customerLookup := repository.NewLineCustomerRepository(nil)
	settingLookup := repository.NewLineReservationSettingRepository(nil)
	repos := &repository.Repositories{
		LineCustomerMgr:        customerLookup,
		LineReservationSetting: settingLookup,
	}
	uploader := &mockFileUploader{}

	h := New(cfg, svc, repos, uploader)

	require.NotNil(t, h)
	assert.Same(t, cfg, h.cfg)
	assert.Same(t, svc, h.svc)
	assert.Same(t, customerLookup, h.liffCustomerLookup)
	assert.Same(t, settingLookup, h.liffSettingLookup)
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
