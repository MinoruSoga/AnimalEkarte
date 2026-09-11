package reservation

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

func TestGetLineReservationSetting_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects path clinic B without hospital-settings:view", func(t *testing.T) {
		h := newHandlerWithLineReservationSettingSvc(&mockLineReservationSettingService{
			getFn: func(context.Context, uint64) (*model.LineReservationSetting, error) {
				t.Fatal("must not read LINE reservation settings for clinic B")
				return nil, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "clinic_id", Value: "2"}}
		c.Set("clinic_id", "2")
		c.Set("clinic_ids", []uint64{1, 2})
		c.Set("is_system_admin", false)
		c.Set("user_id", "17")
		httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
			return clinicID == 1 && resource == string(model.ResourceHospitalSettings) && action == "view"
		})
		h.GetLineReservationSetting(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("allows path clinic A", func(t *testing.T) {
		h := newHandlerWithLineReservationSettingSvc(&mockLineReservationSettingService{
			getFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				assert.Equal(t, uint64(1), clinicID)
				return &model.LineReservationSetting{ID: 1, ClinicID: 1, Status: "running"}, nil
			},
		})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		c.Params = gin.Params{{Key: "clinic_id", Value: "1"}}
		c.Set("clinic_id", "2")
		c.Set("clinic_ids", []uint64{1, 2})
		c.Set("is_system_admin", false)
		c.Set("user_id", "17")
		httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
			return clinicID == 1 && resource == string(model.ResourceHospitalSettings) && action == "view"
		})
		h.GetLineReservationSetting(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"running"`)
	})
}
