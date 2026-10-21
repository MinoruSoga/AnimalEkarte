package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-463: technical-spec#owners-list-scope-contract-emr463 の BE fixture 行。
// 値はすべて合成 ID（実患者・実飼主の値は含めない）。実行は NOT_RUN。

type ownersScopeFixture struct {
	staffID    uint64
	accountID  uint64
	clinics    []model.Clinic
	assignment []model.StaffClinicAssignment
	isAdmin    bool
}

func ownersScopeHandler(
	fixture ownersScopeFixture,
	effective EffectivePermissionService,
) *HTTPHandler {
	accountID := fixture.accountID
	return NewHTTPHandler(HTTPDependencies{
		Staff: sessionStaffReader{
			getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{
					ID:        id,
					Name:      "Owners Scope Staff",
					AccountID: &accountID,
					IsActive:  true,
				}, nil
			},
		},
		Accounts: sessionAccountService{
			getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
				return &model.Account{
					ID:            id,
					Email:         "scope-staff@example.test",
					IsSystemAdmin: fixture.isAdmin,
				}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{
			assignments: fixture.assignment,
		},
		Clinics: sessionClinicLister{
			clinics: fixture.clinics,
		},
		EffectivePermissions: effective,
	}, CookieConfigForProduction(false))
}

func ownersScopeContext(
	response *httptest.ResponseRecorder,
	clinicIDs any,
	isAdmin bool,
) *gin.Context {
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", http.NoBody)
	c.Set("user_id", "17")
	c.Set("clinic_id", "1")
	c.Set("is_system_admin", isAdmin)
	if clinicIDs != nil {
		c.Set("clinic_ids", clinicIDs)
	}
	return c
}

func TestHTTPHandler_GetMe_OwnersViewClinicIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("交差だけを順次評価し true のみ採用する", func(t *testing.T) {
		var evaluated []uint64
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				clinics: []model.Clinic{
					{ID: 1, Name: "Clinic One", IsActive: true},
					{ID: 2, Name: "Clinic Two", IsActive: true},
					{ID: 3, Name: "Clinic Three", IsActive: true},
					{ID: 4, Name: "Clinic Four", IsActive: true},
				},
				assignment: []model.StaffClinicAssignment{
					{StaffID: 17, ClinicID: 1, IsMain: true},
					{StaffID: 17, ClinicID: 2},
					{StaffID: 17, ClinicID: 4},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(_ context.Context, _, clinicID uint64) ([]model.PermissionGroupRule, error) {
					evaluated = append(evaluated, clinicID)
					if clinicID == 1 {
						return []model.PermissionGroupRule{{
							Resource: string(model.ResourceOwners),
							CanView:  true,
						}}, nil
					}
					return nil, nil
				},
			},
		)

		response := httptest.NewRecorder()
		// context=[1,2,3]、profile=[1,2,4] → 交差 [1,2]。1=true・2=false → ["1"]。
		handler.GetMe(ownersScopeContext(response, []uint64{1, 2, 3}, false))

		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.Contains(t, response.Body.String(), `"owners_view_clinic_ids":["1"]`)
		// context 外の 4（profile のみ）と profile 外の 3（context のみ）は評価しない。
		assert.NotContains(t, evaluated, uint64(3))
		assert.NotContains(t, evaluated, uint64(4))
	})

	t.Run("system_admin も現在contextとの交差内に限定する", func(t *testing.T) {
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				isAdmin:   true,
				clinics: []model.Clinic{
					{ID: 1, Name: "Clinic One", IsActive: true},
					{ID: 2, Name: "Clinic Two", IsActive: true},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(context.Context, uint64, uint64) ([]model.PermissionGroupRule, error) {
					return nil, nil
				},
			},
		)

		response := httptest.NewRecorder()
		// 現在context [1]、active profile [1,2] → admin でも 1 だけ。
		handler.GetMe(ownersScopeContext(response, []uint64{1}, true))

		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.Contains(t, response.Body.String(), `"owners_view_clinic_ids":["1"]`)
	})

	t.Run("全候補 false は 200 と非nullの空配列", func(t *testing.T) {
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				clinics: []model.Clinic{
					{ID: 5, Name: "Clinic Five", IsActive: true},
					{ID: 6, Name: "Clinic Six", IsActive: true},
				},
				assignment: []model.StaffClinicAssignment{
					{StaffID: 17, ClinicID: 5, IsMain: true},
					{StaffID: 17, ClinicID: 6},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(context.Context, uint64, uint64) ([]model.PermissionGroupRule, error) {
					return nil, nil
				},
			},
		)

		response := httptest.NewRecorder()
		handler.GetMe(ownersScopeContext(response, []uint64{5, 6}, false))

		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.Contains(t, response.Body.String(), `"owners_view_clinic_ids":[]`)
		assert.NotContains(t, response.Body.String(), `"owners_view_clinic_ids":null`)
	})

	t.Run("評価 error は 5xx で部分集合や空の 200 を返さない", func(t *testing.T) {
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				clinics: []model.Clinic{
					{ID: 1, Name: "Clinic One", IsActive: true},
					{ID: 2, Name: "Clinic Two", IsActive: true},
				},
				assignment: []model.StaffClinicAssignment{
					{StaffID: 17, ClinicID: 1, IsMain: true},
					{StaffID: 17, ClinicID: 2},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(_ context.Context, _, clinicID uint64) ([]model.PermissionGroupRule, error) {
					if clinicID == 2 {
						return nil, errors.New("permission repository unavailable")
					}
					return []model.PermissionGroupRule{{
						Resource: string(model.ResourceOwners),
						CanView:  true,
					}}, nil
				},
			},
		)

		response := httptest.NewRecorder()
		// 1=true の後に 2 が評価 error → 5xx（["1"] や [] の 200 を返さない）。
		handler.GetMe(ownersScopeContext(response, []uint64{1, 2}, false))

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.NotContains(t, response.Body.String(), "owners_view_clinic_ids")
	})

	t.Run("context 欠落は 401 で二重応答なし", func(t *testing.T) {
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				clinics:   []model.Clinic{{ID: 1, Name: "Clinic One", IsActive: true}},
				assignment: []model.StaffClinicAssignment{
					{StaffID: 17, ClinicID: 1, IsMain: true},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(context.Context, uint64, uint64) ([]model.PermissionGroupRule, error) {
					return nil, nil
				},
			},
		)

		response := httptest.NewRecorder()
		handler.GetMe(ownersScopeContext(response, nil, false))

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.NotContains(t, response.Body.String(), "owners_view_clinic_ids")
	})

	t.Run("context 不正型は 400 で二重応答なし", func(t *testing.T) {
		handler := ownersScopeHandler(
			ownersScopeFixture{
				staffID:   17,
				accountID: 41,
				clinics:   []model.Clinic{{ID: 1, Name: "Clinic One", IsActive: true}},
				assignment: []model.StaffClinicAssignment{
					{StaffID: 17, ClinicID: 1, IsMain: true},
				},
			},
			authServiceEffectivePermissionStub{
				getFn: func(context.Context, uint64, uint64) ([]model.PermissionGroupRule, error) {
					return nil, nil
				},
			},
		)

		response := httptest.NewRecorder()
		handler.GetMe(ownersScopeContext(response, "1,2", false))

		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.NotContains(t, response.Body.String(), "owners_view_clinic_ids")
	})
}

func TestGetMeResponse_FieldIsolation(t *testing.T) {
	// GET /me 専用 field は login/refresh・共通 MeResponse・既存 builder に現れない。
	loginJSON, err := json.Marshal(LoginResponse{
		IsSystemAdmin: false,
		User:          &MeResponse{ID: "17"},
	})
	require.NoError(t, err)
	assert.NotContains(t, string(loginJSON), "owners_view_clinic_ids")

	meJSON, err := json.Marshal(&MeResponse{ID: "17"})
	require.NoError(t, err)
	assert.NotContains(t, string(meJSON), "owners_view_clinic_ids")

	getMeJSON, err := json.Marshal(&GetMeResponse{
		MeResponse:          MeResponse{ID: "17"},
		OwnersViewClinicIDs: []string{},
	})
	require.NoError(t, err)
	assert.Contains(t, string(getMeJSON), `"owners_view_clinic_ids":[]`)
}

func TestHTTPHandler_Login_ResponseHasNoOwnersViewClinicIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: sessionLoginService{
			account: &model.Account{
				ID:        accountID,
				Email:     "staff@example.test",
				IsActive:  true,
				UpdatedAt: time.Unix(1_721_000_000, 0),
			},
			staff: &model.Staff{
				ID:        17,
				AccountID: &accountID,
				Name:      "Login Staff",
				IsActive:  true,
			},
		},
		Tokens: NewTokenService("owners-scope-login-secret", nil),
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 1, IsMain: true},
			},
		},
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{{ID: 1, Name: "Clinic One", IsActive: true}},
		},
		Audit: &sessionAuditLogger{},
	}, CookieConfigForProduction(false))

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		strings.NewReader(`{"email":"staff@example.test","password":"password1"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.NotContains(t, response.Body.String(), "owners_view_clinic_ids")
}
