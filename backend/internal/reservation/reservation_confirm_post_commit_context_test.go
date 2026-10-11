package reservation

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type confirmPostCommitTraceKey struct{}

func TestConfirmTentative_PostCommitContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name          string
		status        model.ReservationStatus
		category      model.ReservationTypeCategory
		confirmErr    error
		withoutRecord bool
		cancelBefore  bool
	}{
		{name: "committed general confirm survives cancellation", status: model.ReservationStatusConfirmed},
		{name: "pending remains ineligible", status: model.ReservationStatusPending},
		{name: "trimming remains ineligible", status: model.ReservationStatusConfirmed, category: model.ReservationTypeCategoryTrimming},
		{name: "failed confirm does not auto-create", confirmErr: errors.New("confirm failed")},
		{name: "missing automation dependency remains optional", status: model.ReservationStatusConfirmed, withoutRecord: true},
		{name: "business confirm still follows request cancellation", cancelBefore: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), confirmPostCommitTraceKey{}, "confirm"))
			defer cancel()
			ambient := &gorm.DB{}
			ctx = persistence.WithTxValue(ctx, ambient)
			if tt.cancelBefore {
				cancel()
			}
			committed, called := false, false
			svc := &mockReservationService{confirmTentativeFn: func(writeCtx context.Context, _, _ uint64, _ *ConfirmTentativeInput) (*model.Reservation, error) {
				assert.Same(t, ambient, persistence.TxFromContext(writeCtx))
				if tt.cancelBefore {
					return nil, writeCtx.Err()
				}
				require.NoError(t, writeCtx.Err())
				if tt.confirmErr != nil {
					return nil, tt.confirmErr
				}
				committed = true
				cancel()
				return &model.Reservation{ID: 9, Status: tt.status, ReservationType: &model.ReservationType{Category: tt.category}}, nil
			}}
			var done <-chan struct{}
			medical := &mockMedicalRecordService{autoCreateFromReservationFn: func(autoCtx context.Context, clinicID uint64, r *model.Reservation) {
				require.True(t, committed)
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(9), r.ID)
				require.NoError(t, autoCtx.Err())
				assert.Nil(t, persistence.TxFromContext(autoCtx))
				assert.Equal(t, "confirm", autoCtx.Value(confirmPostCommitTraceKey{}))
				deadline, ok := autoCtx.Deadline()
				require.True(t, ok)
				assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, time.Second)
				done = autoCtx.Done()
				called = true
			}}
			if tt.withoutRecord {
				medical = nil
			}
			h := newHandlerWithReservationAndMedicalRecordSvc(svc, medical)
			// Avoid a typed nil inside the interface in the optional-dependency case.
			if tt.withoutRecord {
				h.medicalRecord = nil
			}
			router := gin.New()
			router.POST("/reservations/:id/confirm-tentative", setClinicAndStaffID, h.ConfirmTentativeAppointment)
			req := httptest.NewRequest("POST", "/reservations/9/confirm-tentative", strings.NewReader(`{"start_time":"2030-01-01T01:00:00Z","end_time":"2030-01-01T02:00:00Z"}`)).WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if tt.confirmErr != nil || tt.cancelBefore {
				assert.GreaterOrEqual(t, w.Code, 400)
				assert.False(t, called)
				return
			}
			assert.Equal(t, 200, w.Code)
			eligible := tt.status == model.ReservationStatusConfirmed && tt.category != model.ReservationTypeCategoryTrimming && !tt.withoutRecord
			assert.Equal(t, eligible, called)
			if called {
				select {
				case <-done:
				default:
					t.Fatal("automation context must be canceled before handler returns")
				}
			}
		})
	}
}
