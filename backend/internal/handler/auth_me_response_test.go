package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- toMeResponse additional branch coverage ----

// TestToMeResponse_OccupationSet は staff.Occupation が設定されている場合に
// occupation ポインタが埋められることを検証する。
func TestToMeResponse_OccupationSet(t *testing.T) {
	account := &model.Account{ID: 1, Email: "vet@test.com", IsSystemAdmin: false}
	staff := &model.Staff{
		ID:         30,
		Name:       "獣医師A",
		Occupation: &model.Occupation{ID: 1, Name: "獣医師"},
		ClinicAssignments: []model.StaffClinicAssignment{
			{ClinicID: 1, IsMain: true},
		},
	}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA"}}

	resp := toMeResponse(staff, account, "1", nil, allClinics, nil)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Occupation)
	assert.Equal(t, "獣医師", *resp.Occupation)
}

// TestToMeResponse_OccupationNil は staff.Occupation が nil の場合、
// occupation フィールドが nil のままであることを検証する。
func TestToMeResponse_OccupationNil(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 31, Name: "受付スタッフ", Occupation: nil}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA"}}

	resp := toMeResponse(staff, account, "1", nil, allClinics, nil)

	require.NotNil(t, resp)
	assert.Nil(t, resp.Occupation)
}

// TestToMeResponse_LogoURLSet は clinic.LogoURL が非空の場合、LogoURL ポインタが
// 埋められることを検証する（空文字なら nil、というポインタ変換分岐）。
func TestToMeResponse_LogoURLSet(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 32, Name: "スタッフ"}
	allClinics := []model.Clinic{
		{ID: 1, Name: "クリニックA", LogoURL: "https://example.com/logo.png"},
	}

	resp := toMeResponse(staff, account, "1", nil, allClinics, nil)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Clinic)
	require.NotNil(t, resp.Clinic.LogoURL)
	assert.Equal(t, "https://example.com/logo.png", *resp.Clinic.LogoURL)
}

// TestToMeResponse_LogoURLEmpty は clinic.LogoURL が空文字の場合、LogoURL ポインタが
// nil のままであることを検証する。
func TestToMeResponse_LogoURLEmpty(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 33, Name: "スタッフ"}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA", LogoURL: ""}}

	resp := toMeResponse(staff, account, "1", nil, allClinics, nil)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Clinic)
	assert.Nil(t, resp.Clinic.LogoURL)
}

// TestToMeResponse_MainClinicNoMatch は mainClinicID がどの allClinics とも一致しない場合、
// Clinic フィールドが nil のままであることを検証する（回帰防止）。
func TestToMeResponse_MainClinicNoMatch(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 34, Name: "スタッフ"}
	allClinics := []model.Clinic{
		{ID: 1, Name: "クリニックA"},
		{ID: 2, Name: "クリニックB"},
	}

	resp := toMeResponse(staff, account, "999", nil, allClinics, nil)

	require.NotNil(t, resp)
	assert.Nil(t, resp.Clinic)
	assert.Equal(t, "999", resp.MainClinicID)
}

// TestToMeResponse_NilEffectivePermsDefaultsToEmptyMap は effectivePerms=nil の場合、
// Permissions が nil ではなく空の EffectivePermissions マップになることを検証する。
func TestToMeResponse_NilEffectivePermsDefaultsToEmptyMap(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 35, Name: "スタッフ"}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA"}}

	resp := toMeResponse(staff, account, "1", nil, allClinics, nil)

	require.NotNil(t, resp)
	require.NotNil(t, resp.Permissions)
	assert.Empty(t, resp.Permissions)
}

// TestToMeResponse_EffectivePermsPassthrough は effectivePerms が非 nil の場合、
// そのままレスポンスに設定されることを検証する。
func TestToMeResponse_EffectivePermsPassthrough(t *testing.T) {
	account := &model.Account{ID: 1, Email: "staff@test.com"}
	staff := &model.Staff{ID: 36, Name: "スタッフ"}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA"}}
	perms := EffectivePermissions{
		"owner": ResourcePermission{View: true, Create: false, Edit: false, Delete: false},
	}

	resp := toMeResponse(staff, account, "1", nil, allClinics, perms)

	require.NotNil(t, resp)
	assert.Equal(t, perms, resp.Permissions)
}

// TestToMeResponse_NilAccountDoesNotPanic は account が nil（staff に紐付く
// account 未作成）の場合でも panic せず、Email が空文字になることを検証する。
func TestToMeResponse_NilAccountDoesNotPanic(t *testing.T) {
	staff := &model.Staff{ID: 37, Name: "スタッフ"}
	allClinics := []model.Clinic{{ID: 1, Name: "クリニックA"}}

	require.NotPanics(t, func() {
		resp := toMeResponse(staff, nil, "1", nil, allClinics, nil)

		require.NotNil(t, resp)
		assert.Equal(t, "", resp.Email)
	})
}

// ---- buildAllPermissions ----

// TestBuildAllPermissions は全リソースに対して全CRUD trueのマップが返ることを検証する。
func TestBuildAllPermissions(t *testing.T) {
	perms := buildAllPermissions()

	require.Len(t, perms, len(model.AllResources))
	for _, res := range model.AllResources {
		perm, ok := perms[string(res)]
		require.True(t, ok, "resource %s should be present", res)
		assert.True(t, perm.View, "resource %s View should be true", res)
		assert.True(t, perm.Create, "resource %s Create should be true", res)
		assert.True(t, perm.Edit, "resource %s Edit should be true", res)
		assert.True(t, perm.Delete, "resource %s Delete should be true", res)
	}
}

// ---- calculateEffectivePermissions ----

// mockEffectivePermissionService は permission_group_handler_test.go で定義済みのものを再利用する。

func newHandlerWithMeEffectivePermSvc(svc service.EffectivePermissionService) *Handler {
	return &Handler{svc: &service.Services{EffectivePermission: svc}}
}

func newMeTestContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/auth/me", http.NoBody)
	return c
}

func TestCalculateEffectivePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("system_admin bypasses and returns all permissions", func(t *testing.T) {
		h := newHandlerWithMeEffectivePermSvc(&mockEffectivePermissionService{})
		c := newMeTestContext()

		perms := h.calculateEffectivePermissions(c, true, 1)

		assert.Equal(t, buildAllPermissions(), perms)
	})

	t.Run("non-admin with missing clinic context returns empty map", func(t *testing.T) {
		h := newHandlerWithMeEffectivePermSvc(&mockEffectivePermissionService{})
		c := newMeTestContext()
		// clinic_id を設定しない → extractClinicID が失敗する

		perms := h.calculateEffectivePermissions(c, false, 1)

		assert.NotNil(t, perms)
		assert.Empty(t, perms)
	})

	t.Run("non-admin with service error returns empty map", func(t *testing.T) {
		svc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
				assert.Equal(t, uint64(7), staffID)
				assert.Equal(t, uint64(1), clinicID)
				return nil, errors.New("db failure")
			},
		}
		h := newHandlerWithMeEffectivePermSvc(svc)
		c := newMeTestContext()
		c.Set("clinic_id", "1")

		perms := h.calculateEffectivePermissions(c, false, 7)

		assert.NotNil(t, perms)
		assert.Empty(t, perms)
	})

	t.Run("non-admin with rules builds effective permission map", func(t *testing.T) {
		svc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
				return []model.PermissionGroupRule{
					{Resource: "owner", CanView: true, CanCreate: true, CanEdit: false, CanDelete: false},
					{Resource: "pet", CanView: true, CanCreate: false, CanEdit: false, CanDelete: false},
				}, nil
			},
		}
		h := newHandlerWithMeEffectivePermSvc(svc)
		c := newMeTestContext()
		c.Set("clinic_id", "1")

		perms := h.calculateEffectivePermissions(c, false, 7)

		require.Len(t, perms, 2)
		assert.Equal(t, ResourcePermission{View: true, Create: true, Edit: false, Delete: false}, perms["owner"])
		assert.Equal(t, ResourcePermission{View: true, Create: false, Edit: false, Delete: false}, perms["pet"])
	})
}
