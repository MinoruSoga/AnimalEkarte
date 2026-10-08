package reservation

// reservation_intent_handler_test.go — EMR-308 next-visit-intent HTTP boundary:
// PUT/DELETE /reservations/pets/:petId/next-visit-intent と
// POST /reservations/:id/confirm-tentative の handler 層を pin する。
// 認証 context・path param・JSON binding・JST 暦日パース・service エラー伝播・
// confirm 成功時のカルテ自動作成フックをモック service で検証する。

import (
	"bytes"
	"context"
	"encoding/json"
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

// setClinicAndStaffID は intent handler が要求する clinic_id + user_id 双方を
// 認証済み context として設定するミドルウェア。
func setClinicAndStaffID(c *gin.Context) {
	setClinicID(c)
	c.Set("user_id", "7")
}

func newSetNextVisitIntentRouter(svc ReservationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationSvc(svc)
	r.PUT("/reservations/pets/:petId/next-visit-intent", setClinicAndStaffID, h.SetNextVisitIntent)
	return r
}

func newClearNextVisitIntentRouter(svc ReservationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationSvc(svc)
	r.DELETE("/reservations/pets/:petId/next-visit-intent", setClinicAndStaffID, h.ClearNextVisitIntent)
	return r
}

func newConfirmTentativeRouter(svc ReservationService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithReservationSvc(svc)
	r.POST("/reservations/:id/confirm-tentative", setClinicAndStaffID, h.ConfirmTentativeAppointment)
	return r
}

// ---- SetNextVisitIntent ----

func TestSetNextVisitIntent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withJSTLocal(t)

	jstDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.Local)
	sourceID := uint64(42)

	tests := []struct {
		name       string
		petID      string
		body       any
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name:  "200 success parses JST date, provenance, and staff provenance",
			petID: "5",
			body:  map[string]any{"intended_date": "2026-11-01", "source_medical_record_id": 42},
			svc: &mockReservationService{
				setNextVisitIntentFn: func(_ context.Context, clinicID, petID uint64, input SetNextVisitIntentInput) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), petID)
					assert.True(t, input.IntendedDate.Equal(jstDate), "intended_date must parse as JST calendar date, got %v", input.IntendedDate)
					require.NotNil(t, input.SourceMedicalRecordID)
					assert.Equal(t, sourceID, *input.SourceMedicalRecordID)
					require.NotNil(t, input.CreatedBy)
					assert.Equal(t, uint64(7), *input.CreatedBy)
					return &model.Reservation{ID: 9, IntendedDate: &jstDate, SourceMedicalRecordID: &sourceID}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "200 success without source_medical_record_id",
			petID: "5",
			body:  map[string]any{"intended_date": "2026-11-01"},
			svc: &mockReservationService{
				setNextVisitIntentFn: func(_ context.Context, _, _ uint64, input SetNextVisitIntentInput) (*model.Reservation, error) {
					assert.Nil(t, input.SourceMedicalRecordID)
					return &model.Reservation{ID: 9, IntendedDate: &jstDate}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 malformed JSON",
			petID:      "5",
			body:       "not-json",
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 missing intended_date",
			petID:      "5",
			body:       map[string]any{},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 invalid intended_date format",
			petID:      "5",
			body:       map[string]any{"intended_date": "not-a-date"},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 non-numeric petId",
			petID:      "abc",
			body:       map[string]any{"intended_date": "2026-11-01"},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "404 pet not found",
			petID: "999",
			body:  map[string]any{"intended_date": "2026-11-01"},
			svc: &mockReservationService{
				setNextVisitIntentFn: func(_ context.Context, _, _ uint64, _ SetNextVisitIntentInput) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("pet", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newSetNextVisitIntentRouter(tt.svc)
			w := httptest.NewRecorder()
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPut, "/reservations/pets/"+tt.petID+"/next-visit-intent", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
		})
	}

	t.Run("200 response carries intended_date and provenance fields", func(t *testing.T) {
		svc := &mockReservationService{
			setNextVisitIntentFn: func(_ context.Context, _, _ uint64, _ SetNextVisitIntentInput) (*model.Reservation, error) {
				return &model.Reservation{ID: 9, IntendedDate: &jstDate, SourceMedicalRecordID: &sourceID}, nil
			},
		}
		router := newSetNextVisitIntentRouter(svc)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/reservations/pets/5/next-visit-intent",
			bytes.NewReader([]byte(`{"intended_date":"2026-11-01","source_medical_record_id":42}`)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"intended_date":"2026-11-01"`)
		assert.Contains(t, w.Body.String(), `"source_medical_record_id":42`)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		r := gin.New()
		h := newHandlerWithReservationSvc(&mockReservationService{})
		r.PUT("/reservations/pets/:petId/next-visit-intent", h.SetNextVisitIntent)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/reservations/pets/5/next-visit-intent",
			bytes.NewReader([]byte(`{"intended_date":"2026-11-01"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("401 when user_id is missing", func(t *testing.T) {
		r := gin.New()
		h := newHandlerWithReservationSvc(&mockReservationService{})
		r.PUT("/reservations/pets/:petId/next-visit-intent", setClinicID, h.SetNextVisitIntent)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/reservations/pets/5/next-visit-intent",
			bytes.NewReader([]byte(`{"intended_date":"2026-11-01"}`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ClearNextVisitIntent ----

func TestClearNextVisitIntent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		petID      string
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name:  "204 success",
			petID: "5",
			svc: &mockReservationService{
				clearNextVisitIntentFn: func(_ context.Context, clinicID, petID uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(5), petID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:  "204 idempotent when no active tentative exists",
			petID: "5",
			svc: &mockReservationService{
				clearNextVisitIntentFn: func(_ context.Context, _, _ uint64) error { return nil },
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "400 non-numeric petId",
			petID:      "abc",
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "500 service internal error",
			petID: "5",
			svc: &mockReservationService{
				clearNextVisitIntentFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapInternalServerError("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newClearNextVisitIntentRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/reservations/pets/"+tt.petID+"/next-visit-intent", http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
		})
	}

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		r := gin.New()
		h := newHandlerWithReservationSvc(&mockReservationService{})
		r.DELETE("/reservations/pets/:petId/next-visit-intent", h.ClearNextVisitIntent)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/reservations/pets/5/next-visit-intent", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- ConfirmTentativeAppointment ----

func TestConfirmTentativeAppointment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withJSTLocal(t)

	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.Local)
	end := start.Add(30 * time.Minute)

	tests := []struct {
		name       string
		id         string
		body       any
		svc        *mockReservationService
		wantStatus int
	}{
		{
			name: "200 success maps optional fields",
			id:   "9",
			body: map[string]any{
				"start_time":          start.Format(time.RFC3339),
				"end_time":            end.Format(time.RFC3339),
				"reservation_type_id": 3,
				"doctor_id":           11,
				"is_designated":       true,
				"status":              "pending",
				"notes":               "再検のため",
			},
			svc: &mockReservationService{
				confirmTentativeFn: func(_ context.Context, clinicID, id uint64, input *ConfirmTentativeInput) (*model.Reservation, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(9), id)
					assert.True(t, input.StartTime.Equal(start))
					assert.True(t, input.EndTime.Equal(end))
					require.NotNil(t, input.ReservationTypeID)
					assert.Equal(t, uint64(3), *input.ReservationTypeID)
					require.NotNil(t, input.DoctorID)
					assert.Equal(t, uint64(11), *input.DoctorID)
					require.NotNil(t, input.IsDesignated)
					assert.True(t, *input.IsDesignated)
					require.NotNil(t, input.Status)
					assert.Equal(t, model.ReservationStatusPending, *input.Status)
					require.NotNil(t, input.Notes)
					assert.Equal(t, "再検のため", *input.Notes)
					return &model.Reservation{ID: 9, Status: model.ReservationStatusPending}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "200 success with minimal body",
			id:   "9",
			body: map[string]any{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
			},
			svc: &mockReservationService{
				confirmTentativeFn: func(_ context.Context, _, _ uint64, input *ConfirmTentativeInput) (*model.Reservation, error) {
					assert.Nil(t, input.ReservationTypeID)
					assert.Nil(t, input.DoctorID)
					assert.Nil(t, input.Status)
					return &model.Reservation{ID: 9, Status: model.ReservationStatusPending}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 malformed JSON",
			id:         "9",
			body:       "not-json",
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 missing start_time/end_time",
			id:         "9",
			body:       map[string]any{},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "400 status outside pending/confirmed",
			id:   "9",
			body: map[string]any{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
				"status":     "cancelled",
			},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "400 non-numeric id",
			id:   "abc",
			body: map[string]any{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
			},
			svc:        &mockReservationService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "404 non-tentative reservation",
			id:   "9",
			body: map[string]any{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
			},
			svc: &mockReservationService{
				confirmTentativeFn: func(_ context.Context, _, _ uint64, _ *ConfirmTentativeInput) (*model.Reservation, error) {
					return nil, apperrors.WrapNotFound("reservation", "9")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "409 slot conflict",
			id:   "9",
			body: map[string]any{
				"start_time": start.Format(time.RFC3339),
				"end_time":   end.Format(time.RFC3339),
			},
			svc: &mockReservationService{
				confirmTentativeFn: func(_ context.Context, _, _ uint64, _ *ConfirmTentativeInput) (*model.Reservation, error) {
					return nil, apperrors.WrapConflict("slot occupied")
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newConfirmTentativeRouter(tt.svc)
			w := httptest.NewRecorder()
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/reservations/"+tt.id+"/confirm-tentative", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
		})
	}

	t.Run("confirmed result triggers medical-record auto-create", func(t *testing.T) {
		var autoCreated *model.Reservation
		medicalSvc := &mockMedicalRecordService{
			autoCreateFromReservationFn: func(_ context.Context, clinicID uint64, reservation *model.Reservation) {
				assert.Equal(t, uint64(1), clinicID)
				autoCreated = reservation
			},
		}
		svc := &mockReservationService{
			confirmTentativeFn: func(_ context.Context, _, _ uint64, _ *ConfirmTentativeInput) (*model.Reservation, error) {
				return &model.Reservation{ID: 9, Status: model.ReservationStatusConfirmed}, nil
			},
		}
		r := gin.New()
		h := newHandlerWithReservationAndMedicalRecordSvc(svc, medicalSvc)
		r.POST("/reservations/:id/confirm-tentative", setClinicAndStaffID, h.ConfirmTentativeAppointment)

		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
		})
		req := httptest.NewRequest(http.MethodPost, "/reservations/9/confirm-tentative", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NotNil(t, autoCreated, "confirmed tentative must reach AutoCreateFromReservation")
		assert.Equal(t, uint64(9), autoCreated.ID)
	})

	t.Run("pending result does not trigger medical-record auto-create", func(t *testing.T) {
		called := false
		medicalSvc := &mockMedicalRecordService{
			autoCreateFromReservationFn: func(_ context.Context, _ uint64, _ *model.Reservation) {
				called = true
			},
		}
		svc := &mockReservationService{
			confirmTentativeFn: func(_ context.Context, _, _ uint64, _ *ConfirmTentativeInput) (*model.Reservation, error) {
				return &model.Reservation{ID: 9, Status: model.ReservationStatusPending}, nil
			},
		}
		r := gin.New()
		h := newHandlerWithReservationAndMedicalRecordSvc(svc, medicalSvc)
		r.POST("/reservations/:id/confirm-tentative", setClinicAndStaffID, h.ConfirmTentativeAppointment)

		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
		})
		req := httptest.NewRequest(http.MethodPost, "/reservations/9/confirm-tentative", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.False(t, called)
	})

	t.Run("401 when clinic_id is missing", func(t *testing.T) {
		r := gin.New()
		h := newHandlerWithReservationSvc(&mockReservationService{})
		r.POST("/reservations/:id/confirm-tentative", h.ConfirmTentativeAppointment)
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
		})
		req := httptest.NewRequest(http.MethodPost, "/reservations/9/confirm-tentative", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
