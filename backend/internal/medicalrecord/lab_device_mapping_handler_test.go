package medicalrecord

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateLabMappingDraft_HidesRawAndMeasuredValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw := []byte("measured-" + mappingSentinelValue)
	h := NewLabImportHandler(nil, nil, nil).WithMappingDrafts(&mappingDraftHTTPFake{
		create: func(_ context.Context, clinicID uint64, payload []byte, hint string) (*LabMappingDraftView, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, raw, payload)
			assert.Equal(t, "nx600", hint)
			return &LabMappingDraftView{
				Draft: model.LabDeviceMappingDraft{
					ID: 4, Status: model.LabMappingDraftStatusUnconfirmed, SourceType: "fuji_nx600",
					RawPayload: raw,
				},
				Items: []model.LabDeviceMappingDraftItem{{
					DeviceItemCode: "BUN-P", ValueRaw: mappingSentinelValue, ItemSource: model.LabMappingItemSourceBlank,
					LLMRequest: json.RawMessage(`{"code":"BUN-P"}`),
				}},
			}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"payload_base64":"` + base64.StdEncoding.EncodeToString(raw) + `","device_hint":"nx600"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device/mapping-drafts", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	h.CreateLabMappingDraft(c)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.NotContains(t, w.Body.String(), mappingSentinelValue)
	assert.NotContains(t, w.Body.String(), "llm_request")
	assert.NotContains(t, w.Body.String(), "value_raw")
	assert.Contains(t, w.Body.String(), `"source":"blank"`)
	assert.Contains(t, w.Body.String(), `"items":[`)
}

func TestCreateLabMappingDraft_RejectsBadPayloadWithoutServiceCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calls := 0
	h := NewLabImportHandler(nil, nil, nil).WithMappingDrafts(&mappingDraftHTTPFake{
		create: func(context.Context, uint64, []byte, string) (*LabMappingDraftView, error) {
			calls++
			return nil, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device/mapping-drafts", strings.NewReader(`{"payload_base64":"%%%"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	h.CreateLabMappingDraft(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Zero(t, calls)
}

func TestLabMappingDraftHTTP_AuthAndConfirm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabImportHandler(nil, nil, nil).WithMappingDrafts(&mappingDraftHTTPFake{
		confirm: func(_ context.Context, _, _, staffID uint64, input ConfirmLabMappingInput) (*LabMappingDraftView, error) {
			assert.Equal(t, uint64(1), staffID)
			assert.Empty(t, input.Items)
			return &LabMappingDraftView{Draft: model.LabDeviceMappingDraft{ID: 4, Status: model.LabMappingDraftStatusConfirmed}, Items: []model.LabDeviceMappingDraftItem{}}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device/mapping-drafts/4/confirm", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ConfirmLabMappingDraft(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device/mapping-drafts/4/confirm", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	setStaffID(c)
	h.ConfirmLabMappingDraft(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"confirmed"`)

	h = NewLabImportHandler(nil, nil, nil).WithMappingDrafts(&mappingDraftHTTPFake{
		confirm: func(context.Context, uint64, uint64, uint64, ConfirmLabMappingInput) (*LabMappingDraftView, error) {
			return nil, apperrors.WrapConflict("mapping draft is already confirmed")
		},
	})
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/lab-device/mapping-drafts/4/confirm", strings.NewReader(`{"items":[]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	setStaffID(c)
	h.ConfirmLabMappingDraft(c)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGetLabMappingDraft_SelectedClinicLacksViewGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &mappingDraftHTTPFake{}
	h := NewLabImportHandler(nil, nil, nil).WithMappingDrafts(fake)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/lab-device/mapping-drafts/4", http.NoBody)
	setSelectedClinicWithoutLabImportGrant(c, "view")
	h.GetLabMappingDraft(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Zero(t, fake.gets)
}

func TestPutLabMappingGeneration_ReturnsStopped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewLabImportHandler(nil, nil, nil).WithMappingDrafts(&mappingDraftHTTPFake{
		stop: func(_ context.Context, clinicID uint64, stopped bool) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.True(t, stopped)
			return nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/lab-device/mapping-generation", strings.NewReader(`{"stopped":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	h.PutLabMappingGeneration(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"stopped":true}`, w.Body.String())
}

type mappingDraftHTTPFake struct {
	create  func(context.Context, uint64, []byte, string) (*LabMappingDraftView, error)
	confirm func(context.Context, uint64, uint64, uint64, ConfirmLabMappingInput) (*LabMappingDraftView, error)
	stop    func(context.Context, uint64, bool) error
	gets    int
}

func (f *mappingDraftHTTPFake) Create(ctx context.Context, clinicID uint64, payload []byte, hint string) (*LabMappingDraftView, error) {
	return f.create(ctx, clinicID, payload, hint)
}
func (f *mappingDraftHTTPFake) Get(context.Context, uint64, uint64) (*LabMappingDraftView, error) {
	f.gets++
	return nil, apperrors.WrapNotFound("lab_device_mapping_draft", "missing")
}
func (f *mappingDraftHTTPFake) Confirm(ctx context.Context, clinicID, id, staffID uint64, input ConfirmLabMappingInput) (*LabMappingDraftView, error) {
	return f.confirm(ctx, clinicID, id, staffID, input)
}
func (f *mappingDraftHTTPFake) SetGenerationStopped(ctx context.Context, clinicID uint64, stopped bool) error {
	return f.stop(ctx, clinicID, stopped)
}
