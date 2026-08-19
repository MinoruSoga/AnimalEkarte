package medicalrecord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockLabDeviceItemMasterService struct {
	listFn   func(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error)
	ensureFn func(ctx context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error)
	updateFn func(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error)
}

func (m *mockLabDeviceItemMasterService) List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error) {
	return m.listFn(ctx, clinicID, sourceType)
}

func (m *mockLabDeviceItemMasterService) EnsureDefaults(ctx context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error) {
	return m.ensureFn(ctx, clinicID)
}

func (m *mockLabDeviceItemMasterService) Update(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockLabDeviceItemMasterService) ResolveItems(context.Context, uint64, string, []string) (*LabDeviceMasterResolution, error) {
	return nil, apperrors.WrapInvalidInput("not used")
}

func newDeviceMasterHandler(svc LabDeviceItemMasterService) *LabImportHandler {
	return NewLabImportHandler(nil, nil, nil).WithDeviceMasters(svc)
}

func TestListLabDeviceItemMasters_MissingClinicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device-item-masters", http.NoBody)
	h.ListLabDeviceItemMasters(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListLabDeviceItemMasters_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		listFn: func(_ context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "fuji_nx600", sourceType)
			return []model.LabDeviceItemMaster{{
				ID: 9, ClinicID: 1, SourceType: "fuji_nx600", DeviceItemCode: "Na-P",
				DisplayName: "Na", ValueShape: model.LabDeviceValueShapeNumeric, IsActive: true,
			}}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device-item-masters?source_type=fuji_nx600", http.NoBody)
	setClinicID(c)
	h.ListLabDeviceItemMasters(c)
	require.Equal(t, http.StatusOK, w.Code)
	var body []labDeviceItemMasterResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, "Na-P", body[0].DeviceItemCode)
	assert.NotContains(t, w.Body.String(), "legacy")
}

func TestEnsureLabDeviceItemMasters_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		ensureFn: func(_ context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error) {
			assert.Equal(t, uint64(1), clinicID)
			return 25, []model.LabDeviceItemMaster{{ID: 1, DeviceItemCode: "Na-P", DisplayName: "Na"}}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device-item-masters/ensure", http.NoBody)
	setClinicID(c)
	h.EnsureLabDeviceItemMasters(c)
	require.Equal(t, http.StatusOK, w.Code)
	var body labDeviceItemMasterEnsureResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, int64(25), body.InsertedCount)
	require.Len(t, body.Items, 1)
}

func TestUpdateLabDeviceItemMaster_InvalidField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		updateFn: func(_ context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(4), id)
			return nil, apperrors.WrapInvalidInput("exam_type_field_id is not in this clinic")
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/lab-device-item-masters/4",
		jsonBody(map[string]any{"display_name": "BUN", "unit": "mg/dl", "is_active": true, "exam_type_field_id": 99}))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	setClinicID(c)
	h.UpdateLabDeviceItemMaster(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
