package staff_test

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

func membershipABGrantASelectedB(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == 1 && resource == string(model.ResourceMasterStaff) && action == "view"
	})
}

func membershipABGrantBSelectedB(c *gin.Context) {
	c.Set("clinic_id", "2")
	c.Set("clinic_ids", []uint64{1, 2})
	c.Set("is_system_admin", false)
	c.Set("user_id", "17")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
		return clinicID == 2 && resource == string(model.ResourceMasterStaff) && action == "view"
	})
}

func TestListStaffs_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		listFn: func(context.Context, uint64, int, int) ([]model.Staff, int64, error) {
			t.Fatal("clinic-fixed staff list must not run for selected clinic B")
			return nil, 0, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	membershipABGrantASelectedB(c)
	h.ListStaffs(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetStaff_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		verifyClinicMembershipFn: func(_ context.Context, _, clinicID uint64) error {
			assert.Equal(t, uint64(2), clinicID)
			return nil
		},
		getByIDInClinicFn: func(context.Context, uint64, uint64) (*model.Staff, error) {
			t.Fatal("clinic-fixed staff get must not run for selected clinic B")
			return nil, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	membershipABGrantASelectedB(c)
	h.GetStaff(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetStaffPermissionGroups_MembershipABGrantASelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		verifyClinicMembershipFn: func(context.Context, uint64, uint64) error { return nil },
		getPermissionGroupIDsFn: func(context.Context, uint64, uint64) ([]uint64, error) {
			t.Fatal("must not read permission groups for selected clinic B")
			return nil, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	membershipABGrantASelectedB(c)
	h.GetStaffPermissionGroups(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestListStaffs_MembershipABGrantBSelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		listFn: func(_ context.Context, clinicID uint64, _, _ int) ([]model.Staff, int64, error) {
			assert.Equal(t, uint64(2), clinicID)
			return []model.Staff{{ID: 10, Name: "医院Bスタッフ", StaffType: model.StaffTypeDoctor}}, 1, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	membershipABGrantBSelectedB(c)
	h.ListStaffs(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"医院Bスタッフ"`)
}

func TestGetStaff_MembershipABGrantBSelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		verifyClinicMembershipFn: func(_ context.Context, _, clinicID uint64) error {
			assert.Equal(t, uint64(2), clinicID)
			return nil
		},
		getByIDInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			assert.Equal(t, uint64(2), clinicID)
			assert.Equal(t, uint64(10), id)
			return &model.Staff{ID: id, Name: "医院Bスタッフ", StaffType: model.StaffTypeDoctor}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	membershipABGrantBSelectedB(c)
	h.GetStaff(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"医院Bスタッフ"`)
}

func TestGetStaffPermissionGroups_MembershipABGrantBSelectedB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithStaffSvc(&mockService{
		verifyClinicMembershipFn: func(_ context.Context, _, clinicID uint64) error {
			assert.Equal(t, uint64(2), clinicID)
			return nil
		},
		getPermissionGroupIDsFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(2), clinicID)
			assert.Equal(t, uint64(10), staffID)
			return []uint64{4}, nil
		},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	membershipABGrantBSelectedB(c)
	h.GetStaffPermissionGroups(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"group_ids":[4]`)
}
