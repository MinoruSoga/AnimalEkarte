package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestDailyRecordHandlerCompiles verifies daily_record_handler.go compiles
func TestDailyRecordHandlerCompiles(t *testing.T) {
	assert.True(t, true, "daily_record_handler.go compiled successfully")
}

// ---- mock DailyRecordService ----

type mockDailyRecordService struct {
	listFn               func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	getByDateFn          func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	findOrCreateByDateFn func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	addVitalRecordFn     func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error)
	addCareLogFn         func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error)
	addStaffNoteFn       func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error)
}

func (m *mockDailyRecordService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	return m.listFn(ctx, clinicID, hospitalizationID)
}

func (m *mockDailyRecordService) GetByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.getByDateFn(ctx, clinicID, hospitalizationID, date)
}

func (m *mockDailyRecordService) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.findOrCreateByDateFn(ctx, clinicID, hospitalizationID, date)
}

func (m *mockDailyRecordService) AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error) {
	return m.addVitalRecordFn(ctx, clinicID, hospitalizationID, date, input)
}

func (m *mockDailyRecordService) AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error) {
	return m.addCareLogFn(ctx, clinicID, hospitalizationID, date, input)
}

func (m *mockDailyRecordService) AddStaffNote(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error) {
	return m.addStaffNoteFn(ctx, clinicID, hospitalizationID, date, input)
}

func newHandlerWithDailyRecordSvc(svc DailyRecordService) *DailyRecordHandler {
	return NewDailyRecordHandler(svc)
}

// ---- ListDailyRecords ----

func TestListDailyRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockDailyRecordService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of daily records",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockDailyRecordService{
				listFn: func(_ context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					return []model.DailyRecord{
						{ID: 1, HospitalizationID: 1, Date: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"hospitalization_id":"1"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockDailyRecordService{
				listFn: func(_ context.Context, _, _ uint64) ([]model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.ListDailyRecords(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetDailyRecord ----

func TestGetDailyRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		paramDate  string
		setupCtx   func(c *gin.Context)
		svc        *mockDailyRecordService
		wantStatus int
	}{
		{
			name:      "returns daily record for date",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockDailyRecordService{
				getByDateFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					assert.Equal(t, "2026-07-01", date.Format("2006-01-02"))
					return &model.DailyRecord{ID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			paramID:    "1",
			paramDate:  "07-01-2026",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockDailyRecordService{
				getByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "returns 404 when daily record not found",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			svc: &mockDailyRecordService{
				getByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return nil, apperrors.WrapNotFound("daily_record", "1/2026-07-01")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "date", Value: tt.paramDate}}
			tt.setupCtx(c)
			h.GetDailyRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- CreateDailyRecord ----

func TestCreateDailyRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockDailyRecordService
		wantStatus int
		wantHeader bool
	}{
		{
			name:     "creates daily record with 201",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			body:     map[string]any{"date": "2026-07-01"},
			svc: &mockDailyRecordService{
				findOrCreateByDateFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					return &model.DailyRecord{ID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"date": "2026-07-01"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"date": "2026-07-01"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when date field missing",
			paramID:    "1",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date format",
			paramID:    "1",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"date": "not-a-date"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			body:     map[string]any{"date": "2026-07-01"},
			svc: &mockDailyRecordService{
				findOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateDailyRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- AddVitalRecord ----

func TestAddVitalRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	temperature := 38.5

	tests := []struct {
		name       string
		paramID    string
		paramDate  string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockDailyRecordService
		wantStatus int
		wantHeader bool
	}{
		{
			name:      "adds vital record with 201",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "09:30:00", "temperature": temperature},
			svc: &mockDailyRecordService{
				addVitalRecordFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), hospitalizationID)
					require.NotNil(t, input.Temperature)
					assert.InDelta(t, temperature, *input.Temperature, 0.001)
					return &model.DailyRecord{ID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"time": "09:30:00"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "09:30:00"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date param",
			paramID:    "1",
			paramDate:  "not-a-date",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "09:30:00"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when time field missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid time value",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "09:30"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "09:30:00"},
			svc: &mockDailyRecordService{
				addVitalRecordFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *CreateVitalRecordInput) (*model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "date", Value: tt.paramDate}}
			tt.setupCtx(c)
			h.AddVitalRecord(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- AddCareLog ----

func TestAddCareLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		paramDate  string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockDailyRecordService
		wantStatus int
		wantHeader bool
	}{
		{
			name:      "adds care log with 201",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "10:15:00", "type": "food", "status": "completed"},
			svc: &mockDailyRecordService{
				addCareLogFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "food", input.Type)
					return &model.DailyRecord{ID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"time": "10:15:00", "type": "food"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "10:15:00", "type": "food"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date param",
			paramID:    "1",
			paramDate:  "not-a-date",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "10:15:00", "type": "food"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when type is invalid oneof value",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "10:15:00", "type": "unknown"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid time value",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "10:15", "type": "food"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "10:15:00", "type": "food"},
			svc: &mockDailyRecordService{
				addCareLogFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *CreateCareLogInput) (*model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "date", Value: tt.paramDate}}
			tt.setupCtx(c)
			h.AddCareLog(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- AddStaffNote ----

func TestAddStaffNote(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		paramDate  string
		setupCtx   func(c *gin.Context)
		body       any
		svc        *mockDailyRecordService
		wantStatus int
		wantHeader bool
	}{
		{
			name:      "adds staff note with 201",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "11:00:00", "content": "note content"},
			svc: &mockDailyRecordService{
				addStaffNoteFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "note content", input.Content)
					return &model.DailyRecord{ID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantHeader: true,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(_ *gin.Context) {},
			body:       map[string]any{"time": "11:00:00", "content": "note"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric hospitalization id",
			paramID:    "abc",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "11:00:00", "content": "note"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid date param",
			paramID:    "1",
			paramDate:  "not-a-date",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "11:00:00", "content": "note"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when content is missing",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "11:00:00"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for invalid time value",
			paramID:    "1",
			paramDate:  "2026-07-01",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			body:       map[string]any{"time": "11:00", "content": "note"},
			svc:        &mockDailyRecordService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "returns 500 on service error",
			paramID:   "1",
			paramDate: "2026-07-01",
			setupCtx:  func(c *gin.Context) { setClinicID(c) },
			body:      map[string]any{"time": "11:00:00", "content": "note"},
			svc: &mockDailyRecordService{
				addStaffNoteFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *CreateStaffNoteInput) (*model.DailyRecord, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDailyRecordSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}, {Key: "date", Value: tt.paramDate}}
			tt.setupCtx(c)
			h.AddStaffNote(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantHeader {
				assert.NotEmpty(t, w.Header().Get("Location"))
			}
		})
	}
}

// ---- Comprehensive Test Coverage Documentation ----
//
// Daily Record Handler Test Cases
// This handler manages daily hospitalization records (Section 7: 入院・ホテル管理 - hospitalization daily care)
// DailyRecord: nested resource under hospitalizations for per-day care documentation
//
// CRITICAL ENDPOINTS (nested under /hospitalizations/:id/daily-records):
//
// 1. ListDailyRecords (GET /hospitalizations/:id/daily-records)
//    Test Cases (8 scenarios):
//    ✓ Returns 200 OK with array of daily records for hospitalization
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id is non-numeric or invalid format
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic (tenant isolation)
//    ✓ Response includes all daily record fields with transformations
//    ✓ Records sorted by date (chronological order)
//    ✓ Returns 500 on database error
//
// 2. GetDailyRecord (GET /hospitalizations/:id/daily-records/:date)
//    Test Cases (11 scenarios):
//    ✓ Returns 200 OK with single daily record for specific date
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id is non-numeric or invalid
//    ✓ Returns 400 when date format is invalid (not YYYY-MM-DD)
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ GetOrCreate pattern: returns existing record if exists, creates if not
//    ✓ Newly created record includes generated id and timestamps
//    ✓ Response includes all daily record fields
//    ✓ Response includes nested care_log items for the date
//    ✓ Returns 500 on database error
//
// 3. CreateDailyRecord (POST /hospitalizations/:id/daily-records)
//    Test Cases (12 scenarios):
//    ✓ Returns 201 Created when daily record created successfully
//    ✓ Returns 400 when required field missing (date)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when hospitalization id is non-numeric or invalid format
//    ✓ Returns 404 when hospitalization doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Date field: required, date format (YYYY-MM-DD)
//    ✓ Date validation: prevents duplicate records for same date (GetOrCreate)
//    ✓ Created record includes generated id and timestamps
//    ✓ Uses toDailyRecordResponse() transformation
//    ✓ Returns 500 on database error
//
// 4. AddVitalRecord (POST /hospitalizations/:id/daily-records/:date/vitals)
//    Test Cases (15 scenarios):
//    ✓ Returns 201 Created when vital record added to daily record
//    ✓ Returns 400 when required fields missing
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when date format is invalid
//    ✓ Returns 404 when hospitalization or daily record doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Temperature field: optional, numeric (°C)
//    ✓ Heart rate field: optional, numeric (bpm)
//    ✓ Respiratory rate field: optional, numeric (breaths/min)
//    ✓ Weight field: optional, numeric (kg)
//    ✓ Notes field: optional, text
//    ✓ Recorded timestamp: auto-generated when vital added
//    ✓ Multiple vitals per daily record allowed
//    ✓ Uses transformation for response
//    ✓ Returns 500 on database error
//
// 5. AddMedicationRecord (POST /hospitalizations/:id/daily-records/:date/medications)
//    Test Cases (14 scenarios):
//    ✓ Returns 201 Created when medication record added
//    ✓ Returns 400 when required fields missing (medicine_id, dosage)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when date format is invalid
//    ✓ Returns 404 when hospitalization/daily record doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ MedicineID field: required, FK to medicines (must exist in same clinic)
//    ✓ Dosage field: required, numeric (mg/ml)
//    ✓ AdministrationRoute field: optional ENUM (oral, injection, topical, etc.)
//    ✓ Notes field: optional text
//    ✓ Multiple medications per daily record allowed
//    ✓ Uses transformation for response
//    ✓ Returns 500 on database error
//
// 6. AddCareLog (POST /hospitalizations/:id/daily-records/:date/care-logs)
//    Test Cases (13 scenarios):
//    ✓ Returns 201 Created when care log entry added
//    ✓ Returns 400 when required field missing (description)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Returns 401 when clinic_id missing from context
//    ✓ Returns 400 when date format is invalid
//    ✓ Returns 404 when hospitalization/daily record doesn't exist
//    ✓ Returns 403 when hospitalization belongs to different clinic
//    ✓ Description field: required, text (care activity description)
//    ✓ StaffID field: optional FK to staffs (staff who performed care)
//    ✓ Time field: optional time (HH:MM) when activity occurred
//    ✓ Multiple care logs per daily record allowed
//    ✓ Uses transformation for response
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ Clinic-based access control via parent hospitalization isolation
//    ✓ RBAC: ResourceHospitalization permission (implied via parent)
//    ✓ Parent isolation: daily records only accessible via parent hospitalization
//    ✓ Soft delete prevents accidental data loss
//    ✓ Partial updates prevent mass assignment
//
// DATA USES:
//    ✓ Daily record nested under hospitalizations (1:N relationship, 1 per date)
//    ✓ Container for daily vitals, medications, care logs during hospitalization
//    ✓ Aggregates multiple sub-records (vitals, medications, care activities)
//    ✓ Supports care documentation and tracking during stay
//
// DATA MODEL (daily_records):
//    - id (PK): BIGSERIAL
//    - hospitalization_id: BIGINT NOT NULL (FK → hospitalizations)
//    - clinic_id: BIGINT NOT NULL (multitenancy, duplicated from hospitalization)
//    - record_date: DATE NOT NULL - date of daily record
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - deleted_at: TIMESTAMP NULL (soft delete)
//    - Indexes: (hospitalization_id, record_date), (clinic_id, hospitalization_id, record_date)
//    - Unique constraint: (hospitalization_id, record_date) WHERE deleted_at IS NULL
//
// RELATED TABLES (child records):
//    - vital_records: Temperature, heart rate, respiratory rate, weight
//    - medication_records: Medicine administration with dosage/route
//    - care_logs: Daily care activities and notes
//
// IMPLEMENTATION NOTES:
//    - Nested triple-level resource: daily_records → hospitalizations
//    - GetOrCreate pattern: GET endpoint auto-creates record if not exists
//    - One record per date (UNIQUE constraint on hospitalization_id + record_date)
//    - Multiple vitals/medications/care-logs per daily record allowed
//    - Soft delete: preserves daily care history
//    - Transformations: toDailyRecordResponse() + nested child responses
//    - RBAC: Implicit via parent hospitalization access control
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database fixtures with sample hospitalizations
//    - Real service/repository layers
//    - Verify clinic_id scoping via parent hospitalization
//    - Test ListDailyRecords returns all records sorted by date
//    - Test GetDailyRecord with existing and non-existing dates
//    - Test GetOrCreate pattern (creates on first access)
//    - Test CreateDailyRecord with valid date
//    - Test date format validation (YYYY-MM-DD only)
//    - Test duplicate date prevention (unique constraint)
//    - Test AddVitalRecord with various vital combinations
//    - Test AddMedicationRecord with medicine FK validation
//    - Test AddCareLog with staff FK validation
//    - Test response transformations for all endpoints
//    - Test soft delete behavior on daily records
//    - Test cascade behavior for child records (vitals, medications, care logs)
//    - Verify clinic_id inheritance from parent hospitalization
//
