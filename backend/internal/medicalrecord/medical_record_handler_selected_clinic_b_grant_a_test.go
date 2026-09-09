package medicalrecord

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

func membershipABGrantASelectedBMedicalRecords(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	setResourcePermissionOnlyClinic(c, 1, string(model.ResourceMedicalRecords), "view")
}

func TestListMedicalRecords_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to selected clinic and rejects B without grant", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
			listFn: func(context.Context, []uint64, MedicalRecordListFilters, int, int) ([]model.MedicalRecord, int64, error) {
				t.Fatal("must not list selected clinic B without medical-records:view")
				return nil, 0, nil
			},
		}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", http.NoBody)
		membershipABGrantASelectedBMedicalRecords(c)
		h.ListMedicalRecords(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), `"total":0`)
	})

	t.Run("filters mixed clinic_ids to A and returns an empty A list as 200", func(t *testing.T) {
		h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
			listFn: func(_ context.Context, clinicIDs []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				assert.Equal(t, []uint64{1}, clinicIDs)
				return []model.MedicalRecord{}, 0, nil
			},
		}, &mockClinicalPlanService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&clinic_ids=1,2", http.NoBody)
		membershipABGrantASelectedBMedicalRecords(c)
		h.ListMedicalRecords(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":0`)
	})
}

func TestGetMedicalRecord_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithMedicalRecordSvc(&mockMedicalRecordService{
		getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
			require.Equal(t, []uint64{1}, clinicIDs)
			return &model.MedicalRecord{ID: id, ClinicID: 1, RecordNo: "A-1"}, nil
		},
	}, &mockClinicalPlanService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	membershipABGrantASelectedBMedicalRecords(c)
	h.GetMedicalRecord(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), `"clinic_id":2`)
}
