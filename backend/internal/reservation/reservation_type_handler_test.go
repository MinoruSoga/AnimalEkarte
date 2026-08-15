package reservation

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

// ---- mock ReservationTypeService ----

type mockReservationTypeService struct {
	listFn              func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	getByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	createFn            func(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error)
	updateFn            func(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	reorderFn           func(ctx context.Context, clinicID uint64, ids []uint64) error
	listUnavailableFn   func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	createUnavailableFn func(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error)
	deleteUnavailableFn func(ctx context.Context, clinicID, reservationTypeID, id uint64) error
	listAvailableFn     func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	createAvailableFn   func(ctx context.Context, clinicID, reservationTypeID uint64, input CreateAvailableSlotInput) (*model.ReservationTypeAvailableSlot, error)
	deleteAvailableFn   func(ctx context.Context, clinicID, reservationTypeID, id uint64) error
	listOccupationsFn   func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	linkOccupationFn    func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	unlinkOccupationFn  func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
}

func (m *mockReservationTypeService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockReservationTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationType{ID: id}, nil
}

func (m *mockReservationTypeService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.ReservationType{ID: 1, Name: input.Name}, nil
}

func (m *mockReservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.ReservationType{ID: id}, nil
}

func (m *mockReservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockReservationTypeService) ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	if m.listUnavailableFn != nil {
		return m.listUnavailableFn(ctx, clinicID, reservationTypeID)
	}
	return nil, nil
}

func (m *mockReservationTypeService) CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error) {
	if m.createUnavailableFn != nil {
		return m.createUnavailableFn(ctx, clinicID, reservationTypeID, input)
	}
	return &model.ReservationTypeUnavailableTime{ID: 1}, nil
}

func (m *mockReservationTypeService) DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
	if m.deleteUnavailableFn != nil {
		return m.deleteUnavailableFn(ctx, clinicID, reservationTypeID, id)
	}
	return nil
}

func (m *mockReservationTypeService) ListAvailableSlots(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.listAvailableFn != nil {
		return m.listAvailableFn(ctx, clinicID, reservationTypeID)
	}
	return nil, nil
}

func (m *mockReservationTypeService) CreateAvailableSlot(ctx context.Context, clinicID, reservationTypeID uint64, input CreateAvailableSlotInput) (*model.ReservationTypeAvailableSlot, error) {
	if m.createAvailableFn != nil {
		return m.createAvailableFn(ctx, clinicID, reservationTypeID, input)
	}
	return &model.ReservationTypeAvailableSlot{ID: 1}, nil
}

func (m *mockReservationTypeService) DeleteAvailableSlot(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
	if m.deleteAvailableFn != nil {
		return m.deleteAvailableFn(ctx, clinicID, reservationTypeID, id)
	}
	return nil
}

func (m *mockReservationTypeService) ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if m.listOccupationsFn != nil {
		return m.listOccupationsFn(ctx, clinicID, reservationTypeID)
	}
	return nil, nil
}

func (m *mockReservationTypeService) LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if m.linkOccupationFn != nil {
		return m.linkOccupationFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return &model.ReservationTypeOccupation{ID: 1, OccupationID: occupationID}, nil
}

func (m *mockReservationTypeService) UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if m.unlinkOccupationFn != nil {
		return m.unlinkOccupationFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return nil
}

// ---- helper ----

func newHandlerWithReservationTypeSvc(svc ReservationTypeService) *ReservationTypeHandler {
	return NewReservationTypeHandler(svc, svc, svc, svc)
}

// ---- ListReservationTypes ----

func TestListReservationTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns list of reservation types",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.ReservationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.ReservationType{{ID: 1, Name: "初診"}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"初診"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns 500 on service error",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				listFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListReservationTypes(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetReservationType ----

func TestGetReservationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns reservation type for valid id",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				getByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationType, error) {
					return &model.ReservationType{ID: id, Name: "定期健診"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"定期健診"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
					return nil, apperrors.WrapNotFound("reservation_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 500 on service error",
			paramID:  "1",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetReservationType(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateReservationType ----

func TestCreateReservationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{"name": "ワクチン接種"}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates reservation type successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				createFn: func(_ context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "ワクチン接種", input.Name)
					return &model.ReservationType{ID: 1, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"ワクチン接種"`,
		},
		{
			name:       "returns 401 when clinic_id missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when name is missing",
			body:       map[string]any{"color": "#FF0000"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on service error",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				createFn: func(_ context.Context, _ uint64, _ *CreateReservationTypeInput) (*model.ReservationType, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreateReservationType(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateReservationType ----

func TestUpdateReservationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	name := "更新済み種別"

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeService
		wantStatus int
	}{
		{
			name:     "updates reservation type successfully",
			paramID:  "1",
			body:     map[string]any{"name": "更新済み種別"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				updateFn: func(_ context.Context, _, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新済み種別", *input.Name)
					return &model.ReservationType{ID: id, Name: name}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"name": "test"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when not found",
			paramID:  "999",
			body:     map[string]any{"name": "test"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeInput) (*model.ReservationType, error) {
					return nil, apperrors.WrapNotFound("reservation_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 when empty body (no fields)",
			paramID:  "1",
			body:     map[string]any{},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateReservationTypeInput) (*model.ReservationType, error) {
					return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdateReservationType(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- DeleteReservationType ----

func newDeleteReservationTypeRouter(svc ReservationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeSvc(svc)
	r.DELETE("/reservation-types/:id", func(c *gin.Context) { setClinicID(c) }, h.DeleteReservationType)
	return r
}

func TestDeleteReservationType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockReservationTypeService
		wantStatus int
	}{
		{
			name:    "deletes successfully",
			paramID: "1",
			svc: &mockReservationTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "returns 409 when type is in use",
			paramID: "2",
			svc: &mockReservationTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "returns 404 when not found",
			paramID: "999",
			svc: &mockReservationTypeService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("reservation_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteReservationTypeRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservation-types/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteReservationType(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ReorderReservationTypes ----

func newReorderReservationTypesRouter(svc ReservationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeSvc(svc)
	r.POST("/reservation-types/reorder", func(c *gin.Context) { setClinicID(c) }, h.ReorderReservationTypes)
	return r
}

func TestReorderReservationTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reorders successfully", func(t *testing.T) {
		svc := &mockReservationTypeService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderReservationTypesRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/reservation-types/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderReservationTypes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- CreateUnavailableTime (sub-resource) ----

func TestCreateUnavailableTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dow := int8(1)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockReservationTypeService
		wantStatus int
	}{
		{
			name:    "creates unavailable time successfully",
			paramID: "1",
			body: map[string]any{
				"unavailable_type": "weekly",
				"day_of_week":      1,
				"start_time":       "09:00",
				"end_time":         "10:00",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				createUnavailableFn: func(_ context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), reservationTypeID)
					assert.Equal(t, "weekly", input.UnavailableType)
					assert.Equal(t, &dow, input.DayOfWeek)
					return &model.ReservationTypeUnavailableTime{
						ID:                1,
						ReservationTypeID: reservationTypeID,
						UnavailableType:   model.UnavailableTypeWeekly,
						DayOfWeek:         &dow,
						StartTime:         "09:00",
						EndTime:           "10:00",
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 401 when clinic_id missing",
			paramID:    "1",
			body:       map[string]any{"unavailable_type": "weekly", "day_of_week": 1, "start_time": "09:00", "end_time": "10:00"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 when required fields missing",
			paramID:    "1",
			body:       map[string]any{"day_of_week": 1},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockReservationTypeService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when reservation type not found",
			paramID: "999",
			body: map[string]any{
				"unavailable_type": "weekly",
				"day_of_week":      1,
				"start_time":       "09:00",
				"end_time":         "10:00",
			},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockReservationTypeService{
				createUnavailableFn: func(_ context.Context, _, _ uint64, _ CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error) {
					return nil, apperrors.WrapNotFound("reservation_type", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithReservationTypeSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.CreateUnavailableTime(c)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ListUnavailableTimes ----

func TestListUnavailableTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns array of unavailable times", func(t *testing.T) {
		dow := int8(2)
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{
			listUnavailableFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeUnavailableTime, error) {
				return []model.ReservationTypeUnavailableTime{
					{ID: 1, DayOfWeek: &dow, StartTime: "10:00", EndTime: "11:00"},
				}, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.ListUnavailableTimes(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"start_time":"10:00"`)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ListUnavailableTimes(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- DeleteUnavailableTime ----

func newDeleteUnavailableTimeRouter(svc ReservationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeSvc(svc)
	r.DELETE("/reservation-types/:id/unavailable-times/:unavailable_time_id",
		func(c *gin.Context) { setClinicID(c) },
		h.DeleteUnavailableTime)
	return r
}

func TestDeleteUnavailableTime(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("deletes unavailable time successfully", func(t *testing.T) {
		svc := &mockReservationTypeService{
			deleteUnavailableFn: func(_ context.Context, _, reservationTypeID, id uint64) error {
				assert.Equal(t, uint64(1), reservationTypeID)
				assert.Equal(t, uint64(5), id)
				return nil
			},
		}
		router := newDeleteUnavailableTimeRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/reservation-types/1/unavailable-times/5", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc := &mockReservationTypeService{
			deleteUnavailableFn: func(_ context.Context, _, _, _ uint64) error {
				return apperrors.WrapNotFound("unavailable_time", "999")
			},
		}
		router := newDeleteUnavailableTimeRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/reservation-types/1/unavailable-times/999", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ---- AvailableSlots ----

func TestCreateAvailableSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dow := int8(1)
	h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{
		createAvailableFn: func(_ context.Context, clinicID, reservationTypeID uint64, input CreateAvailableSlotInput) (*model.ReservationTypeAvailableSlot, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), reservationTypeID)
			assert.Equal(t, "weekly", input.AvailableType)
			assert.Equal(t, &dow, input.DayOfWeek)
			assert.Equal(t, "09:45", input.StartTime)
			return &model.ReservationTypeAvailableSlot{
				ID:                1,
				ReservationTypeID: reservationTypeID,
				AvailableType:     model.AvailableSlotTypeWeekly,
				DayOfWeek:         &dow,
				StartTime:         "09:45",
				IsActive:          true,
			}, nil
		},
	})
	body, err := json.Marshal(map[string]any{
		"available_type": "weekly",
		"day_of_week":    1,
		"start_time":     "09:45",
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)

	h.CreateAvailableSlot(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"start_time":"09:45"`)
}

func TestListAvailableSlots(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dow := int8(1)
	h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{
		listAvailableFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeAvailableSlot, error) {
			return []model.ReservationTypeAvailableSlot{
				{ID: 1, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dow, StartTime: "12:30", IsActive: true},
			}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)

	h.ListAvailableSlots(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"start_time":"12:30"`)
}

func TestDeleteAvailableSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockReservationTypeService{
		deleteAvailableFn: func(_ context.Context, _, reservationTypeID, id uint64) error {
			assert.Equal(t, uint64(1), reservationTypeID)
			assert.Equal(t, uint64(5), id)
			return nil
		},
	}
	h := newHandlerWithReservationTypeSvc(svc)
	router := gin.New()
	router.DELETE("/reservation-types/:id/available-slots/:available_slot_id",
		func(c *gin.Context) { setClinicID(c) },
		h.DeleteAvailableSlot)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/reservation-types/1/available-slots/5", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// ---- LinkReservationTypeOccupation ----

func TestLinkReservationTypeOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("links occupation and returns 201 with Location header", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{
			linkOccupationFn: func(_ context.Context, _, _, occupationID uint64) (*model.ReservationTypeOccupation, error) {
				return &model.ReservationTypeOccupation{ID: 10, OccupationID: occupationID, CreatedAt: time.Time{}}, nil
			},
		})
		body, _ := json.Marshal(map[string]any{"occupation_id": 7})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.LinkReservationTypeOccupation(c)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/occupations/")
	})

	t.Run("returns 400 when occupation_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{})
		body, _ := json.Marshal(map[string]any{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.LinkReservationTypeOccupation(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---- ListReservationTypeOccupations ----

func TestListReservationTypeOccupations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns array of occupations", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{
			listOccupationsFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeOccupation, error) {
				return []model.ReservationTypeOccupation{{ID: 1, OccupationID: 3}}, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		h.ListReservationTypeOccupations(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 401 when clinic_id missing", func(t *testing.T) {
		h := newHandlerWithReservationTypeSvc(&mockReservationTypeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.ListReservationTypeOccupations(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- UnlinkReservationTypeOccupation ----

func newUnlinkOccupationRouter(svc ReservationTypeService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationTypeSvc(svc)
	r.DELETE("/reservation-types/:id/occupations/:occupation_id",
		func(c *gin.Context) { setClinicID(c) },
		h.UnlinkReservationTypeOccupation)
	return r
}

func TestUnlinkReservationTypeOccupation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("unlinks occupation successfully", func(t *testing.T) {
		svc := &mockReservationTypeService{
			unlinkOccupationFn: func(_ context.Context, _, rtID, occID uint64) error {
				assert.Equal(t, uint64(1), rtID)
				assert.Equal(t, uint64(7), occID)
				return nil
			},
		}
		router := newUnlinkOccupationRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/reservation-types/1/occupations/7", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("returns 404 when occupation link not found", func(t *testing.T) {
		svc := &mockReservationTypeService{
			unlinkOccupationFn: func(_ context.Context, _, _, _ uint64) error {
				return apperrors.WrapNotFound("occupation", "7")
			},
		}
		router := newUnlinkOccupationRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/reservation-types/1/occupations/7", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
