package reservation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func membershipABGrantASelectedB(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	setReservationsViewOnlyClinic(c, 1)
}

func TestListReservations_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithReservationSvc(&mockReservationService{
			listFn: func(context.Context, []uint64, int, int, *time.Time, *time.Time, *time.Time, *string, *string, *uint64, *uint64) ([]model.Reservation, int64, error) {
				t.Fatal("must not list selected clinic B without reservations:view")
				return nil, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListReservations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), `"total":0`)
	})

	t.Run("filters mixed clinic_ids to A and returns an empty A list as 200", func(t *testing.T) {
		h := newHandlerWithReservationSvc(&mockReservationService{
			listFn: func(_ context.Context, clinicIDs []uint64, _, _ int, _, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
				assert.Equal(t, []uint64{1}, clinicIDs)
				return []model.Reservation{}, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedB(c)
		h.ListReservations(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":0`)
	})
}

func TestGetReservation_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithReservationSvc(&mockReservationService{
		getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			return &model.Reservation{ID: id, ClinicID: 1, Notes: "A予約"}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	membershipABGrantASelectedB(c)
	h.GetReservation(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"notes":"A予約"`)
}
