package clinic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestListClinics_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithClinicSvc(&mockService{
		listByStaffIDFn: func(_ context.Context, staffID uint64) ([]model.Clinic, error) {
			assert.Equal(t, uint64(17), staffID)
			return []model.Clinic{
				{ID: 1, Name: "医院A", IsActive: true},
				{ID: 2, Name: "医院B", IsActive: true},
			}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/clinics", http.NoBody)
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == 1 && resource == string(model.ResourceHospitalSettings) && action == "view"
	})
	h.ListClinics(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":1`)
	assert.Contains(t, w.Body.String(), `"id":2`)
}
