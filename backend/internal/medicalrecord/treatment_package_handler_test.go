package medicalrecord

// EMR-521 handler tests (NOT_RUN — mock service, execution deferred).

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockTreatmentPackageService struct {
	listFn    func(ctx context.Context, clinicID uint64) ([]TreatmentPackageView, int64, error)
	getFn     func(ctx context.Context, clinicID, id uint64) (*TreatmentPackageView, error)
	createFn  func(ctx context.Context, clinicID uint64, in *CreateTreatmentPackageInput) (*TreatmentPackageView, error)
	updateFn  func(ctx context.Context, clinicID, id uint64, in *UpdateTreatmentPackageInput) (*TreatmentPackageView, error)
	deleteFn  func(ctx context.Context, clinicID, id uint64, version int64, actorID uint64) error
	reorderFn func(ctx context.Context, clinicID uint64, ov int64, ids []uint64, actorID uint64) error
}

func (m *mockTreatmentPackageService) List(ctx context.Context, c uint64) ([]TreatmentPackageView, int64, error) {
	return m.listFn(ctx, c)
}
func (m *mockTreatmentPackageService) GetByID(ctx context.Context, c, id uint64) (*TreatmentPackageView, error) {
	return m.getFn(ctx, c, id)
}
func (m *mockTreatmentPackageService) Create(ctx context.Context, c uint64, in *CreateTreatmentPackageInput) (*TreatmentPackageView, error) {
	return m.createFn(ctx, c, in)
}
func (m *mockTreatmentPackageService) Update(ctx context.Context, c, id uint64, in *UpdateTreatmentPackageInput) (*TreatmentPackageView, error) {
	return m.updateFn(ctx, c, id, in)
}
func (m *mockTreatmentPackageService) Delete(ctx context.Context, c, id uint64, v int64, a uint64) error {
	return m.deleteFn(ctx, c, id, v, a)
}
func (m *mockTreatmentPackageService) Reorder(ctx context.Context, c uint64, ov int64, ids []uint64, a uint64) error {
	return m.reorderFn(ctx, c, ov, ids, a)
}

func newPackageHandlerContext(w *httptest.ResponseRecorder, method, url, body string) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	var rdr *strings.Reader
	rdr = strings.NewReader(body)
	c.Request = httptest.NewRequest(method, url, rdr)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("clinic_id", "7")
	c.Set("user_id", "9")
	c.Set("is_system_admin", true) // bypass per-clinic checker for boundary tests
	return c
}

func TestTreatmentPackageHandler_List(t *testing.T) {
	svc := &mockTreatmentPackageService{
		listFn: func(context.Context, uint64) ([]TreatmentPackageView, int64, error) {
			return []TreatmentPackageView{{Package: mustPkg()}}, 3, nil
		},
	}
	h := NewTreatmentPackageHandler(svc)
	w := httptest.NewRecorder()
	h.ListTreatmentPackages(newPackageHandlerContext(w, "GET", "/", ""))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"order_version":"3"`)
	assert.Contains(t, w.Body.String(), `"id":"201"`) // canonical string wire.
}

func TestTreatmentPackageHandler_Create_InvalidAndCreated(t *testing.T) {
	svc := &mockTreatmentPackageService{
		createFn: func(_ context.Context, _ uint64, in *CreateTreatmentPackageInput) (*TreatmentPackageView, error) {
			return &TreatmentPackageView{Package: mustPkg()}, nil
		},
	}
	h := NewTreatmentPackageHandler(svc)

	w := httptest.NewRecorder()
	h.CreateTreatmentPackage(newPackageHandlerContext(w, "POST", "/", `{"name":"x"}`)) // missing fields
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"invalid-input"`)

	w = httptest.NewRecorder()
	h.CreateTreatmentPackage(newPackageHandlerContext(w, "POST", "/", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`))
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "/api/v1/masters/treatment-packages/201", w.Header().Get("Location"))
}

func TestTreatmentPackageHandler_Update_VersionConflict(t *testing.T) {
	svc := &mockTreatmentPackageService{
		updateFn: func(context.Context, uint64, uint64, *UpdateTreatmentPackageInput) (*TreatmentPackageView, error) {
			return nil, apperrors.WrapConflict("stale")
		},
	}
	h := NewTreatmentPackageHandler(svc)
	w := httptest.NewRecorder()
	c := newPackageHandlerContext(w, "PATCH", "/", `{"version":"3","name":"y"}`)
	c.Params = gin.Params{{Key: "id", Value: "201"}}
	h.UpdateTreatmentPackage(c)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"version-conflict"`)
}

func TestTreatmentPackageHandler_Delete_VersionRequired(t *testing.T) {
	svc := &mockTreatmentPackageService{
		deleteFn: func(context.Context, uint64, uint64, int64, uint64) error { return nil },
	}
	h := NewTreatmentPackageHandler(svc)
	w := httptest.NewRecorder()
	c := newPackageHandlerContext(w, "DELETE", "/", "")
	c.Params = gin.Params{{Key: "id", Value: "201"}}
	h.DeleteTreatmentPackage(c)
	assert.Equal(t, http.StatusBadRequest, w.Code) // missing version query

	w = httptest.NewRecorder()
	c = newPackageHandlerContext(w, "DELETE", "/?version=3", "")
	c.Params = gin.Params{{Key: "id", Value: "201"}}
	h.DeleteTreatmentPackage(c)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTreatmentPackageHandler_Reorder(t *testing.T) {
	svc := &mockTreatmentPackageService{
		reorderFn: func(_ context.Context, _ uint64, ov int64, ids []uint64, _ uint64) error {
			assert.EqualValues(t, 7, ov)
			assert.Equal(t, []uint64{201}, ids)
			return nil
		},
	}
	h := NewTreatmentPackageHandler(svc)
	w := httptest.NewRecorder()
	h.ReorderTreatmentPackages(newPackageHandlerContext(w, "PATCH", "/", `{"order_version":"7","ids":["201"]}`))
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func mustPkg() model.TreatmentPackage {
	return model.TreatmentPackage{ID: 201, ClinicID: 7, Version: 3, Name: "pkg"}
}
