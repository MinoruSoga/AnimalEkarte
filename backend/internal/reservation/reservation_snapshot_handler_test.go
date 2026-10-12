package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type snapshotHandlerService struct {
	*mockReservationService
	called     bool
	snapshotFn func(context.Context, []uint64, listReservationFilters) ([]model.Reservation, error)
}

func (s *snapshotHandlerService) listSnapshot(ctx context.Context, clinicIDs []uint64, filters listReservationFilters) ([]model.Reservation, error) {
	s.called = true
	if s.snapshotFn != nil {
		return s.snapshotFn(ctx, clinicIDs, filters)
	}
	if len(clinicIDs) != 1 || clinicIDs[0] != 1 || filters.Date == nil {
		panic("snapshot scope or date lost")
	}
	return []model.Reservation{{ID: 1}}, nil
}
func TestListReservationsSnapshotHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("snapshot opts into coherent capability", func(t *testing.T) {
		svc := &snapshotHandlerService{mockReservationService: &mockReservationService{
			listFn: func(context.Context, []uint64, int, int, *time.Time, *time.Time, *time.Time, *string, *string, *uint64, *uint64) ([]model.Reservation, int64, error) {
				t.Error("snapshot must not use legacy pagination")
				return nil, 0, nil
			},
		}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/reservations?snapshot=true&date=2026-10-11", nil)
		setClinicID(c)
		newHandlerWithReservationSvc(svc).ListReservations(c)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, svc.called)
		require.Contains(t, rec.Body.String(), `"snapshot":true`)
		require.Contains(t, rec.Body.String(), `"total":1`)
	})
}

func TestListReservationsSnapshotBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{
		"snapshot=yes&date=2026-10-11", "snapshot=&date=2026-10-11",
		"snapshot=true&snapshot=false&date=2026-10-11", "snapshot=true",
		"snapshot=true&status=tentative", "snapshot=true&start_date=2026-10-11",
		"snapshot=true&end_date=2026-10-11&pet_id=2", "snapshot=true&date=bad",
		"snapshot=true&start_date=2026-10-12&end_date=2026-10-11",
	} {
		t.Run("invalid "+query, func(t *testing.T) {
			svc := &snapshotHandlerService{mockReservationService: &mockReservationService{}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/reservations?"+query, nil)
			setClinicID(c)
			newHandlerWithReservationSvc(svc).ListReservations(c)
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.False(t, svc.called)
		})
	}
	t.Run("snapshot ignores pagination and preserves dates and clinic scope", func(t *testing.T) {
		svc := &snapshotHandlerService{mockReservationService: &mockReservationService{}, snapshotFn: func(_ context.Context, ids []uint64, f listReservationFilters) ([]model.Reservation, error) {
			require.Equal(t, []uint64{1}, ids)
			require.Equal(t, "2026-10-11", f.StartDate.Format(time.DateOnly))
			require.Equal(t, "2026-10-13", f.EndDate.Format(time.DateOnly))
			require.Equal(t, uint64(2), *f.PetID)
			require.Equal(t, "tentative", *f.Status)
			return nil, nil
		}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/reservations?snapshot=true&page=bad&limit=10000&start_date=2026-10-11&end_date=2026-10-12&pet_id=2&status=tentative", nil)
		setClinicID(c)
		newHandlerWithReservationSvc(svc).ListReservations(c)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, svc.called)
		var body struct {
			Data     []reservationResponse `json:"data"`
			Total    int                   `json:"total"`
			Snapshot bool                  `json:"snapshot"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.NotNil(t, body.Data)
		require.Empty(t, body.Data)
		require.True(t, body.Snapshot)
	})
	for _, query := range []string{"date=2026-10-11&page=2&limit=20", "snapshot=false&date=2026-10-11&page=2&limit=20"} {
		t.Run("legacy "+query, func(t *testing.T) {
			legacyCalled := false
			svc := &snapshotHandlerService{mockReservationService: &mockReservationService{listFn: func(_ context.Context, _ []uint64, page, limit int, _ *time.Time, _ *time.Time, _ *time.Time, _ *string, _ *string, _ *uint64, _ *uint64) ([]model.Reservation, int64, error) {
				legacyCalled = true
				require.Equal(t, 2, page)
				require.Equal(t, 20, limit)
				return []model.Reservation{{ID: 1}}, 25, nil
			}}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/reservations?"+query, nil)
			setClinicID(c)
			newHandlerWithReservationSvc(svc).ListReservations(c)
			require.Equal(t, http.StatusOK, rec.Code)
			require.True(t, legacyCalled)
			require.False(t, svc.called)
			require.NotContains(t, rec.Body.String(), `"snapshot"`)
			require.Contains(t, rec.Body.String(), `"total":25`)
		})
	}
	t.Run("denied clinic prevents snapshot read", func(t *testing.T) {
		svc := &snapshotHandlerService{mockReservationService: &mockReservationService{}}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/reservations?snapshot=true&date=2026-10-11", nil)
		setClinicID(c)
		httpapi.SetClinicPermissionChecker(c, func(*gin.Context, uint64, string, string) bool { return false })
		newHandlerWithReservationSvc(svc).ListReservations(c)
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.False(t, svc.called)
	})
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{{"overflow", apperrors.WrapInvalidInput("期間を絞り込んでください"), http.StatusBadRequest}, {"read failed", errors.New("private SQL fixture failure"), http.StatusInternalServerError}} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &snapshotHandlerService{mockReservationService: &mockReservationService{}, snapshotFn: func(context.Context, []uint64, listReservationFilters) ([]model.Reservation, error) {
				return nil, tc.err
			}}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/reservations?snapshot=true&date=2026-10-11", nil)
			setClinicID(c)
			newHandlerWithReservationSvc(svc).ListReservations(c)
			require.Equal(t, tc.status, rec.Code)
			require.NotContains(t, rec.Body.String(), "private SQL")
		})
	}
}
