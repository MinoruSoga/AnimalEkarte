package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockLabDeviceItemMasterService struct {
	listFn         func(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error)
	ensureFn       func(ctx context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error)
	updateFn       func(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error)
	listDevicesFn  func(ctx context.Context, clinicID uint64) ([]model.LabDevice, error)
	createDeviceFn func(ctx context.Context, clinicID uint64, input CreateLabDeviceInput) (*model.LabDevice, error)
	updateDeviceFn func(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceInput) (*model.LabDevice, error)
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

func (m *mockLabDeviceItemMasterService) ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error) {
	if m.listDevicesFn == nil {
		return nil, apperrors.WrapInvalidInput("not used")
	}
	return m.listDevicesFn(ctx, clinicID)
}

func (m *mockLabDeviceItemMasterService) CreateDevice(ctx context.Context, clinicID uint64, input CreateLabDeviceInput) (*model.LabDevice, error) {
	if m.createDeviceFn == nil {
		return nil, apperrors.WrapInvalidInput("not used")
	}
	return m.createDeviceFn(ctx, clinicID, input)
}

func (m *mockLabDeviceItemMasterService) UpdateDevice(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceInput) (*model.LabDevice, error) {
	if m.updateDeviceFn == nil {
		return nil, apperrors.WrapInvalidInput("not used")
	}
	return m.updateDeviceFn(ctx, clinicID, id, input)
}
func (m *mockLabDeviceItemMasterService) SaveConfiguration(context.Context, uint64, uint64, SaveLabDeviceConfigurationInput) (*SaveLabDeviceConfigurationResult, error) {
	return nil, errors.New("not implemented")
}

func newDeviceMasterHandler(svc LabDeviceItemMasterService) *LabImportHandler {
	return NewLabImportHandler(nil, nil, nil).WithDeviceMasters(svc)
}

type mockLabDeviceReceiveService struct {
	boardFn      func(ctx context.Context, clinicID uint64) (*LabDeviceBoard, error)
	unlinkedFn   func(ctx context.Context, clinicID uint64) ([]LabDeviceJobCard, error)
	getStationFn func(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error)
}

func (m *mockLabDeviceReceiveService) ReceiveFrames(context.Context, uint64, []byte, string) (*LabDeviceReceiveResult, error) {
	return nil, errors.New("not used")
}

func (m *mockLabDeviceReceiveService) PutWait(context.Context, uint64, uint64, uint64) (*LabDeviceWaitView, error) {
	return nil, errors.New("not used")
}

func (m *mockLabDeviceReceiveService) ClearWait(context.Context, uint64) error {
	return errors.New("not used")
}

func (m *mockLabDeviceReceiveService) Board(ctx context.Context, clinicID uint64) (*LabDeviceBoard, error) {
	if m.boardFn == nil {
		return nil, errors.New("not used")
	}
	return m.boardFn(ctx, clinicID)
}

func (m *mockLabDeviceReceiveService) Unlinked(ctx context.Context, clinicID uint64) ([]LabDeviceJobCard, error) {
	if m.unlinkedFn == nil {
		return nil, errors.New("not used")
	}
	return m.unlinkedFn(ctx, clinicID)
}

func (m *mockLabDeviceReceiveService) Attach(context.Context, uint64, uuid.UUID, uint64) (*LabDeviceJobCard, error) {
	return nil, errors.New("not used")
}

func (m *mockLabDeviceReceiveService) Detach(context.Context, uint64, uuid.UUID) (*LabDeviceJobCard, error) {
	return nil, errors.New("not used")
}

func (m *mockLabDeviceReceiveService) GetStation(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error) {
	if m.getStationFn == nil {
		return nil, errors.New("not used")
	}
	return m.getStationFn(ctx, clinicID)
}

func (m *mockLabDeviceReceiveService) PutStation(context.Context, uint64, string) (*LabDeviceStationView, error) {
	return nil, errors.New("not used")
}

func setSelectedClinicWithoutLabImportGrant(c *gin.Context, action string) {
	setClinicID(c)
	c.Set("clinic_id", "2")
	c.Set("is_system_admin", false)
	c.Set("clinic_ids", []uint64{1, 2})
	setResourcePermissionOnlyClinic(c, 1, string(model.ResourceLabImport), action)
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
				ValueShape: model.LabDeviceValueShapeNumeric, IsActive: true,
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
	assert.NotContains(t, w.Body.String(), "display_name")
}

func TestEnsureLabDeviceItemMasters_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		ensureFn: func(_ context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error) {
			assert.Equal(t, uint64(1), clinicID)
			return 25, []model.LabDeviceItemMaster{{ID: 1, DeviceItemCode: "Na-P"}}, nil
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
		jsonBody(map[string]any{"unit": "mg/dl", "is_active": true, "exam_type_field_id": 99}))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	setClinicID(c)
	h.UpdateLabDeviceItemMaster(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateLabDeviceItemMaster_RequiresIsActive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/lab-device-item-masters/4",
		jsonBody(map[string]any{"unit": "mg/dl"}))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	setClinicID(c)
	h.UpdateLabDeviceItemMaster(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLabDevices_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	examTypeID := uint64(10)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		listDevicesFn: func(_ context.Context, clinicID uint64) ([]model.LabDevice, error) {
			assert.Equal(t, uint64(1), clinicID)
			return []model.LabDevice{{
				ID: 3, ClinicID: 1, SourceType: "fuji_nx600", Name: "NX600",
				ExamTypeID: &examTypeID, IsActive: true, SortOrder: 10,
			}}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-devices", http.NoBody)
	setClinicID(c)
	h.ListLabDevices(c)
	require.Equal(t, http.StatusOK, w.Code)
	var body []labDeviceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, "NX600", body[0].Name)
	assert.Equal(t, "fuji_nx600", body[0].SourceType)
	require.NotNil(t, body[0].ExamTypeID)
	assert.Equal(t, uint64(10), *body[0].ExamTypeID)
}

func TestCreateLabDevice_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		createDeviceFn: func(_ context.Context, clinicID uint64, input CreateLabDeviceInput) (*model.LabDevice, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "院内NX", input.Name)
			assert.Equal(t, "fuji_nx600", input.SourceType)
			return &model.LabDevice{
				ID: 8, ClinicID: clinicID, Name: input.Name, SourceType: input.SourceType, IsActive: true,
			}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-devices",
		jsonBody(map[string]any{"name": "院内NX", "source_type": "fuji_nx600", "is_active": true}))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	h.CreateLabDevice(c)
	require.Equal(t, http.StatusCreated, w.Code)
	var body labDeviceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, uint64(8), body.ID)
	assert.Equal(t, "院内NX", body.Name)
}

func TestUpdateLabDevice_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	examTypeID := uint64(21)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		updateDeviceFn: func(_ context.Context, clinicID, id uint64, input UpdateLabDeviceInput) (*model.LabDevice, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(8), id)
			assert.Equal(t, "院内NX", input.Name)
			require.NotNil(t, input.ExamTypeID)
			assert.Equal(t, uint64(21), *input.ExamTypeID)
			return &model.LabDevice{
				ID: id, ClinicID: clinicID, Name: input.Name, SourceType: "fuji_nx600",
				ExamTypeID: input.ExamTypeID, IsActive: input.IsActive,
			}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/lab-devices/8",
		jsonBody(map[string]any{"name": "院内NX", "exam_type_id": 21, "is_active": true, "sort_order": 10}))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	setClinicID(c)
	h.UpdateLabDevice(c)
	require.Equal(t, http.StatusOK, w.Code)
	var body labDeviceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "院内NX", body.Name)
	require.NotNil(t, body.ExamTypeID)
	assert.Equal(t, examTypeID, *body.ExamTypeID)
}

func TestListLabDeviceItemMasters_SelectedClinicLacksViewGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		listFn: func(context.Context, uint64, string) ([]model.LabDeviceItemMaster, error) {
			t.Fatal("lab device item master List must not be reached")
			return nil, errors.New("not used")
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device-item-masters", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "view")
	h.ListLabDeviceItemMasters(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListLabDevices_SelectedClinicLacksViewGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDeviceMasterHandler(&mockLabDeviceItemMasterService{
		listDevicesFn: func(context.Context, uint64) ([]model.LabDevice, error) {
			t.Fatal("lab device ListDevices must not be reached")
			return nil, errors.New("not used")
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-devices", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "view")
	h.ListLabDevices(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetLabDeviceBoard_SelectedClinicLacksCreateGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabImportHandler(nil, nil, nil).
		WithDeviceMasters(&mockLabDeviceItemMasterService{}).
		WithDeviceReceive(&mockLabDeviceReceiveService{
			boardFn: func(context.Context, uint64) (*LabDeviceBoard, error) {
				t.Fatal("lab device receive Board must not be reached")
				return nil, errors.New("not used")
			},
		})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device/board", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "create")
	h.GetLabDeviceBoard(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetLabDeviceUnlinked_SelectedClinicLacksViewGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabImportHandler(nil, nil, nil).
		WithDeviceMasters(&mockLabDeviceItemMasterService{}).
		WithDeviceReceive(&mockLabDeviceReceiveService{
			unlinkedFn: func(context.Context, uint64) ([]LabDeviceJobCard, error) {
				t.Fatal("lab device receive Unlinked must not be reached")
				return nil, errors.New("not used")
			},
		})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device/unlinked", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "view")
	h.GetLabDeviceUnlinked(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetLabDeviceStation_SelectedClinicLacksViewGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabImportHandler(nil, nil, nil).
		WithDeviceMasters(&mockLabDeviceItemMasterService{}).
		WithDeviceReceive(&mockLabDeviceReceiveService{
			getStationFn: func(context.Context, uint64) (*LabDeviceStationView, error) {
				t.Fatal("lab device receive GetStation must not be reached")
				return nil, errors.New("not used")
			},
		})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device/station", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "view")
	h.GetLabDeviceStation(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
