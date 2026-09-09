package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func membershipABGrantASelectedBAccounting(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	setAccountingPermissionOnlyClinic(c, 1, "view")
}

func TestListAccountings_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithAccountingSvc(&mockAccountingService{
			listFn: func(context.Context, uint64, AccountingListFilters, int, int) ([]model.Billing, int64, error) {
				t.Fatal("must not list selected clinic B without accounting:view")
				return nil, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantASelectedBAccounting(c)
		h.ListAccountings(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), `"total":0`)
	})

	t.Run("filters mixed clinic_ids to A and returns an empty A list as 200", func(t *testing.T) {
		h := newHandlerWithAccountingSvc(&mockAccountingService{
			listFn: func(_ context.Context, clinicID uint64, _ AccountingListFilters, _, _ int) ([]model.Billing, int64, error) {
				assert.Equal(t, uint64(1), clinicID)
				return []model.Billing{}, 0, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedBAccounting(c)
		h.ListAccountings(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":0`)
	})
}

func TestGetAccounting_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithAccountingSvc(&mockAccountingService{
		getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			return &model.Billing{ID: id, ClinicID: 1}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	membershipABGrantASelectedBAccounting(c)
	h.GetAccounting(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}

func TestGetDailySummary_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithAccountingSvc(&mockAccountingService{
			getDailySummaryFn: func(context.Context, uint64, string) (*DailySummaryResult, error) {
				t.Fatal("must not summarize selected clinic B without accounting:view")
				return nil, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?date=2026-01-01", http.NoBody)
		membershipABGrantASelectedBAccounting(c)
		h.GetDailySummary(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("filters mixed clinic_ids to A", func(t *testing.T) {
		h := newHandlerWithAccountingSvc(&mockAccountingService{
			getDailySummaryFn: func(_ context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, "2026-01-01", dateStr)
				return &DailySummaryResult{BillingCount: 0, GrandTotal: 0}, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?date=2026-01-01&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedBAccounting(c)
		h.GetDailySummary(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"billing_count":0`)
	})
}
