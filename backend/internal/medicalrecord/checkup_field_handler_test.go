package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock CheckupFieldResultService ----

type mockCheckupFieldResultService struct {
	listFieldsFn        func(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
	listByCheckupFn     func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error)
	listByPetFn         func(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	replaceForCheckupFn func(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error)
}

func (m *mockCheckupFieldResultService) ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	return m.listFieldsFn(ctx, clinicID, checkupTypeID)
}

func (m *mockCheckupFieldResultService) ListByCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	return m.listByCheckupFn(ctx, clinicID, medicalRecordID, checkupID)
}

func (m *mockCheckupFieldResultService) ListByPet(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	return m.listByPetFn(ctx, clinicID, petID)
}

func (m *mockCheckupFieldResultService) ReplaceForCheckup(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
	return m.replaceForCheckupFn(ctx, clinicID, medicalRecordID, checkupID, actorID, inputs)
}

func newHandlerWithCheckupFieldResultSvc(svc CheckupFieldResultService) *CheckupHandler {
	return NewCheckupHandler(nil, svc)
}

// ---- ListCheckupTypeFields ----

func TestListCheckupTypeFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of checkup type fields",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listFieldsFn: func(_ context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), checkupTypeID)
					return []model.CheckupTypeField{{ID: 1, CheckupTypeID: 1, Name: "体重", FieldType: model.CheckupFieldTypeNumber}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid id param",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listFieldsFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListCheckupTypeFields(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListCheckupFieldResults ----

func TestListCheckupFieldResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		recordID   string
		checkupID  string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:      "returns list of checkup field results",
			recordID:  "1",
			checkupID: "2",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByCheckupFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					return []model.CheckupFieldResult{{ID: 1, CheckupID: 2, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			recordID:   "1",
			checkupID:  "2",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid medical record id",
			recordID:   "abc",
			checkupID:  "2",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			recordID:   "1",
			checkupID:  "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			recordID:  "1",
			checkupID: "2",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByCheckupFn: func(_ context.Context, _, _, _ uint64) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.recordID}, {Key: "checkupId", Value: tt.checkupID}}
			tt.setupCtx(c)

			h.ListCheckupFieldResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ReplaceCheckupFieldResults ----

func TestReplaceCheckupFieldResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fieldID := uint64(10)

	tests := []struct {
		name       string
		recordID   string
		checkupID  string
		body       any
		bodyRaw    string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:      "replaces field results successfully",
			recordID:  "1",
			checkupID: "2",
			body: map[string]any{
				"results": []map[string]any{
					{"checkup_type_field_id": 10, "value_number": 12.5},
				},
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				replaceForCheckupFn: func(_ context.Context, clinicID, medicalRecordID, checkupID uint64, actorID *uint64, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), medicalRecordID)
					assert.Equal(t, uint64(2), checkupID)
					require.Len(t, inputs, 1)
					require.NotNil(t, inputs[0].CheckupTypeFieldID)
					assert.Equal(t, fieldID, *inputs[0].CheckupTypeFieldID)
					return []model.CheckupFieldResult{{ID: 1, CheckupID: 2, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			recordID:   "1",
			checkupID:  "2",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid medical record id",
			recordID:   "abc",
			checkupID:  "2",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid checkup id",
			recordID:   "1",
			checkupID:  "abc",
			body:       map[string]any{"results": []map[string]any{}},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON body",
			recordID:   "1",
			checkupID:  "2",
			bodyRaw:    `{"results":`,
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			recordID:  "1",
			checkupID: "2",
			body:      map[string]any{"results": []map[string]any{}},
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				replaceForCheckupFn: func(_ context.Context, _, _, _ uint64, _ *uint64, _ []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)

			var bodyBytes []byte
			if tt.bodyRaw != "" {
				bodyBytes = []byte(tt.bodyRaw)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.recordID}, {Key: "checkupId", Value: tt.checkupID}}
			tt.setupCtx(c)

			h.ReplaceCheckupFieldResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- ListPetCheckupResults ----

func TestListPetCheckupResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockCheckupFieldResultService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns pet checkup results",
			query:    "pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByPetFn: func(_ context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), petID)
					return []model.CheckupFieldResult{{ID: 1, FieldName: "体重"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"field_name":"体重"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "pet_id=5",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pet_id",
			query:      "pet_id=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when pet_id is missing",
			query:      "",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockCheckupFieldResultService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			query:    "pet_id=5",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockCheckupFieldResultService{
				listByPetFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithCheckupFieldResultSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListPetCheckupResults(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
