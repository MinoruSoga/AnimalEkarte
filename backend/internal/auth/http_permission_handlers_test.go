package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type permissionHTTPService struct {
	listFn        func(context.Context, uint64) ([]model.PermissionGroup, error)
	getFn         func(context.Context, uint64, uint64) (*model.PermissionGroup, error)
	createFn      func(context.Context, uint64, *CreatePermissionGroupInput) (*model.PermissionGroup, error)
	updateFn      func(context.Context, uint64, uint64, *UpdatePermissionGroupInput) (*model.PermissionGroup, error)
	deleteFn      func(context.Context, uint64, uint64) error
	updateRulesFn func(context.Context, uint64, uint64, []SetPermissionGroupRulesInput, uint64) error
	reorderFn     func(context.Context, uint64, []uint64) error
	audits        []PermissionMutationAudit
}

func (s *permissionHTTPService) List(
	ctx context.Context,
	clinicID uint64,
) ([]model.PermissionGroup, error) {
	if s.listFn != nil {
		return s.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (s *permissionHTTPService) GetByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if s.getFn != nil {
		return s.getFn(ctx, clinicID, id)
	}
	return &model.PermissionGroup{
		ID:       id,
		ClinicID: clinicID,
		Name:     "permission group",
		Color:    "#112233",
	}, nil
}

func (s *permissionHTTPService) Create(
	ctx context.Context,
	clinicID uint64,
	input *CreatePermissionGroupInput,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	s.audits = append(s.audits, audit)
	if s.createFn != nil {
		return s.createFn(ctx, clinicID, input)
	}
	return &model.PermissionGroup{
		ID:       7,
		ClinicID: clinicID,
		Name:     input.Name,
		Color:    input.Color,
	}, nil
}

func (s *permissionHTTPService) Update(
	ctx context.Context,
	clinicID, id uint64,
	input *UpdatePermissionGroupInput,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	s.audits = append(s.audits, audit)
	if s.updateFn != nil {
		return s.updateFn(ctx, clinicID, id, input)
	}
	name := "updated"
	if input.Name != nil {
		name = *input.Name
	}
	return &model.PermissionGroup{
		ID:       id,
		ClinicID: clinicID,
		Name:     name,
		Color:    "#112233",
	}, nil
}

func (s *permissionHTTPService) Delete(
	ctx context.Context,
	clinicID, id uint64,
	audit PermissionMutationAudit,
) error {
	s.audits = append(s.audits, audit)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (s *permissionHTTPService) UpdateRules(
	ctx context.Context,
	clinicID, groupID uint64,
	inputs []SetPermissionGroupRulesInput,
	actorStaffID uint64,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	s.audits = append(s.audits, audit)
	if s.updateRulesFn != nil {
		if err := s.updateRulesFn(
			ctx,
			clinicID,
			groupID,
			inputs,
			actorStaffID,
		); err != nil {
			return nil, err
		}
	}
	return &model.PermissionGroup{
		ID:       groupID,
		ClinicID: clinicID,
		Name:     "permission group",
		Color:    "#112233",
	}, nil
}

func (s *permissionHTTPService) Reorder(
	ctx context.Context,
	clinicID uint64,
	ids []uint64,
) error {
	if s.reorderFn != nil {
		return s.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

type permissionHTTPEffectiveService struct {
	rules []model.PermissionGroupRule
	err   error
}

func (s permissionHTTPEffectiveService) GetEffectivePermissions(
	context.Context,
	uint64,
	uint64,
) ([]model.PermissionGroupRule, error) {
	return append([]model.PermissionGroupRule(nil), s.rules...), s.err
}

type permissionHTTPAuditLogger struct {
	entries []AuthAuditEntry
	err     error
}

func (a *permissionHTTPAuditLogger) LogAuthLogin(
	context.Context,
	*uint64,
	*uint64,
	string,
	string,
	string,
) error {
	return nil
}

func (a *permissionHTTPAuditLogger) LogEntry(
	_ context.Context,
	entry AuthAuditEntry,
) error {
	a.entries = append(a.entries, entry)
	return a.err
}

func permissionHTTPHandler(
	service PermissionGroupService,
	effective EffectivePermissionService,
	audit AuthAuditLogger,
) *HTTPHandler {
	return NewHTTPHandler(HTTPDependencies{
		PermissionGroups:     service,
		EffectivePermissions: effective,
		Audit:                audit,
	}, CookieConfigForProduction(false))
}

func permissionHTTPContext(
	t *testing.T,
	method, target string,
	body any,
	configure func(*gin.Context),
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(method, target, reader)
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	if configure != nil {
		configure(c)
	}
	return c, response
}

func setPermissionHTTPIdentity(c *gin.Context) {
	c.Set("clinic_id", "23")
	c.Set("user_id", "17")
	c.Set("is_system_admin", false)
}

func setPermissionHTTPID(c *gin.Context, id string) {
	setPermissionHTTPIdentity(c)
	c.Params = gin.Params{{Key: "id", Value: id}}
}

func TestPermissionGroupHTTP_RequestAndResponseMapping(t *testing.T) {
	description := "updated description"
	name := "updated name"
	color := "#abcdef"
	active := true
	order := 4
	create := CreatePermissionGroupRequest{
		Name:        "created",
		Description: "description",
		Color:       "#123456",
		IsActive:    boolPtr(true),
		SortOrder:   3,
		Rules: []PermissionGroupRuleInput{{
			Resource: "owners",
			CanView:  true,
		}},
	}.ToInput()
	assert.Equal(t, "created", create.Name)
	assert.Equal(t, "#123456", create.Color)
	assert.True(t, create.IsActive)
	require.Len(t, create.Rules, 1)
	assert.Equal(t, "owners", create.Rules[0].Resource)

	// Presence matrix: omitted → true, false → false, true → true.
	assert.True(t, (PermissionGroupCreateRequest{
		Name:  "omitted",
		Color: "#111111",
	}).ToInput().IsActive)
	assert.False(t, (PermissionGroupCreateRequest{
		Name:     "inactive",
		Color:    "#222222",
		IsActive: boolPtr(false),
	}).ToInput().IsActive)
	assert.True(t, (PermissionGroupCreateRequest{
		Name:     "active",
		Color:    "#333333",
		IsActive: boolPtr(true),
	}).ToInput().IsActive)

	update := UpdatePermissionGroupRequest{
		Name:        &name,
		Description: &description,
		Color:       &color,
		IsActive:    &active,
		SortOrder:   &order,
		Rules: []PermissionGroupRuleInput{{
			Resource: "owners",
			CanEdit:  true,
		}},
	}.ToInput()
	assert.Equal(t, &name, update.Name)
	assert.Equal(t, &description, update.Description)
	assert.Equal(t, &color, update.Color)
	require.Len(t, update.Rules, 1)
	assert.True(t, update.Rules[0].CanEdit)
	assert.Nil(
		t,
		(UpdatePermissionGroupRequest{Name: &name}).ToInput().Rules,
		"omitted rules must preserve the legacy parent-only path",
	)

	rules := SetPermissionGroupRulesRequest{Rules: []PermissionGroupRuleInput{{
		Resource:  "owners",
		CanView:   true,
		CanCreate: true,
		CanEdit:   true,
		CanDelete: true,
	}}}.ToInput()
	require.Len(t, rules, 1)
	assert.Equal(t, "owners", rules[0].Resource)
	assert.True(t, rules[0].CanDelete)

	createdAt := time.Date(2026, 7, 24, 1, 2, 3, 0, time.UTC)
	response := ToPermissionGroupResponse(&model.PermissionGroup{
		ID:          7,
		ClinicID:    23,
		Name:        "mapped",
		Description: "mapped description",
		Color:       "#123456",
		IsActive:    true,
		SortOrder:   2,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		Rules: []model.PermissionGroupRule{{
			ID:        9,
			GroupID:   7,
			Resource:  "owners",
			CanView:   true,
			CanCreate: true,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}},
	})
	assert.Equal(t, uint64(7), response.ID)
	require.Len(t, response.Rules, 1)
	assert.Equal(t, uint64(9), response.Rules[0].ID)
	assert.True(t, response.Rules[0].CanCreate)
}

func TestHTTPHandler_HasPermissionAndMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rules := []model.PermissionGroupRule{{
		Resource:  "owners",
		CanView:   true,
		CanCreate: true,
		CanEdit:   true,
		CanDelete: true,
	}}
	handler := permissionHTTPHandler(
		&permissionHTTPService{},
		permissionHTTPEffectiveService{rules: rules},
		nil,
	)
	c, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, setPermissionHTTPIdentity)
	for _, action := range []string{"view", "create", "edit", "delete"} {
		assert.True(t, handler.HasPermission(c, "owners", action))
	}
	assert.False(t, handler.HasPermission(c, "owners", "unknown"))
	assert.False(t, handler.HasPermission(c, "pets", "view"))

	admin, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, func(c *gin.Context) {
		c.Set("is_system_admin", true)
	})
	assert.True(t, handler.HasPermission(admin, "anything", "delete"))

	missingIdentity, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, nil)
	assert.False(t, handler.HasPermission(missingIdentity, "owners", "view"))
	missingStaff, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, func(c *gin.Context) {
		c.Set("is_system_admin", false)
	})
	assert.False(t, handler.HasPermission(missingStaff, "owners", "view"))
	missingClinic, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, func(c *gin.Context) {
		c.Set("is_system_admin", false)
		c.Set("user_id", "17")
	})
	assert.False(t, handler.HasPermission(missingClinic, "owners", "view"))

	errorHandler := permissionHTTPHandler(
		&permissionHTTPService{},
		permissionHTTPEffectiveService{err: errors.New("unavailable")},
		nil,
	)
	errorContext, _ := permissionHTTPContext(
		t,
		http.MethodGet,
		"/",
		nil,
		setPermissionHTTPIdentity,
	)
	assert.False(t, errorHandler.HasPermission(errorContext, "owners", "view"))

	allowed, _ := permissionHTTPContext(
		t,
		http.MethodGet,
		"/",
		nil,
		setPermissionHTTPIdentity,
	)
	handler.RequirePermission("owners", "view")(allowed)
	assert.False(t, allowed.IsAborted())

	denied, deniedResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/",
		nil,
		setPermissionHTTPIdentity,
	)
	handler.RequirePermission("pets", "view")(denied)
	assert.True(t, denied.IsAborted())
	assert.Equal(t, http.StatusForbidden, deniedResponse.Code)

	anyAllowed, _ := permissionHTTPContext(
		t,
		http.MethodGet,
		"/",
		nil,
		setPermissionHTTPIdentity,
	)
	handler.RequirePermissionAny(
		PermissionRequirement{Resource: "pets", Action: "view"},
		PermissionRequirement{Resource: "owners", Action: "edit"},
	)(anyAllowed)
	assert.False(t, anyAllowed.IsAborted())

	anyDenied, anyDeniedResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/",
		nil,
		setPermissionHTTPIdentity,
	)
	handler.RequirePermissionAny(
		PermissionRequirement{Resource: "pets", Action: "view"},
	)(anyDenied)
	assert.True(t, anyDenied.IsAborted())
	assert.Equal(t, http.StatusForbidden, anyDeniedResponse.Code)
}

func TestHTTPHandler_DiscountPermissionFacades(t *testing.T) {
	handler := permissionHTTPHandler(
		&permissionHTTPService{},
		permissionHTTPEffectiveService{rules: []model.PermissionGroupRule{{
			Resource:  string(model.ResourceDiscount),
			CanCreate: true,
			CanEdit:   true,
		}}},
		nil,
	)
	c, _ := permissionHTTPContext(t, http.MethodGet, "/", nil, setPermissionHTTPIdentity)
	newFloat := 1.5
	newInt := int64(150)
	assert.NoError(t, handler.RequireDiscountEditFloat(c, &newFloat, 1))
	assert.NoError(t, handler.RequireDiscountEditInt(c, &newInt, 100))
	assert.NoError(t, handler.RequireDiscountCreateFloat(c, newFloat))
	assert.NoError(t, handler.RequireDiscountCreateInt(c, newInt))
	assert.True(t, FloatEquals(1, 1))
}

func TestHTTPHandler_RegisterPermissionGroupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := permissionHTTPHandler(
		&permissionHTTPService{},
		permissionHTTPEffectiveService{},
		nil,
	)
	handler.RegisterPermissionGroupRoutes(router.Group("/api/v1/masters"))

	paths := make(map[string]bool)
	for _, route := range router.Routes() {
		paths[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"GET /api/v1/masters/permission-groups",
		"GET /api/v1/masters/permission-groups/:id",
		"POST /api/v1/masters/permission-groups",
		"PATCH /api/v1/masters/permission-groups/reorder",
		"PATCH /api/v1/masters/permission-groups/:id",
		"DELETE /api/v1/masters/permission-groups/:id",
		"PUT /api/v1/masters/permission-groups/:id/rules",
	}
	for _, path := range expected {
		assert.True(t, paths[path], path)
	}
}

func TestHTTPHandler_ListAndGetPermissionGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &permissionHTTPService{
		listFn: func(_ context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
			assert.Equal(t, uint64(23), clinicID)
			return []model.PermissionGroup{{ID: 1, ClinicID: clinicID, Name: "listed"}}, nil
		},
		getFn: func(_ context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
			assert.Equal(t, uint64(23), clinicID)
			return &model.PermissionGroup{ID: id, ClinicID: clinicID, Name: "fetched"}, nil
		},
	}
	handler := permissionHTTPHandler(service, nil, nil)

	listContext, listResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups",
		nil,
		setPermissionHTTPIdentity,
	)
	handler.ListPermissionGroups(listContext)
	assert.Equal(t, http.StatusOK, listResponse.Code)
	assert.Contains(t, listResponse.Body.String(), `"name":"listed"`)

	getContext, getResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups/7",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	handler.GetPermissionGroup(getContext)
	assert.Equal(t, http.StatusOK, getResponse.Code)
	assert.Contains(t, getResponse.Body.String(), `"name":"fetched"`)

	missingContext, missingResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups",
		nil,
		nil,
	)
	handler.ListPermissionGroups(missingContext)
	assert.Equal(t, http.StatusUnauthorized, missingResponse.Code)

	invalidID, invalidIDResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups/not-an-id",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "not-an-id") },
	)
	handler.GetPermissionGroup(invalidID)
	assert.Equal(t, http.StatusBadRequest, invalidIDResponse.Code)

	errorHandler := permissionHTTPHandler(&permissionHTTPService{
		listFn: func(context.Context, uint64) ([]model.PermissionGroup, error) {
			return nil, errors.New("list failed")
		},
		getFn: func(context.Context, uint64, uint64) (*model.PermissionGroup, error) {
			return nil, apperrors.WrapNotFound("permission_group", "7")
		},
	}, nil, nil)
	listError, listErrorResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups",
		nil,
		setPermissionHTTPIdentity,
	)
	errorHandler.ListPermissionGroups(listError)
	assert.Equal(t, http.StatusInternalServerError, listErrorResponse.Code)
	getError, getErrorResponse := permissionHTTPContext(
		t,
		http.MethodGet,
		"/permission-groups/7",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	errorHandler.GetPermissionGroup(getError)
	assert.Equal(t, http.StatusNotFound, getErrorResponse.Code)
}

func TestHTTPHandler_CreateAndUpdatePermissionGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audit := &permissionHTTPAuditLogger{}
	service := &permissionHTTPService{
		createFn: func(
			_ context.Context,
			clinicID uint64,
			input *CreatePermissionGroupInput,
		) (*model.PermissionGroup, error) {
			assert.Equal(t, uint64(23), clinicID)
			assert.Equal(t, "created", input.Name)
			require.Len(t, input.Rules, 1)
			assert.Equal(t, "owners", input.Rules[0].Resource)
			return &model.PermissionGroup{
				ID:       7,
				ClinicID: clinicID,
				Name:     input.Name,
				Color:    input.Color,
			}, nil
		},
		getFn: func(
			_ context.Context,
			clinicID, id uint64,
		) (*model.PermissionGroup, error) {
			return &model.PermissionGroup{ID: id, ClinicID: clinicID, Name: "old"}, nil
		},
		updateFn: func(
			_ context.Context,
			clinicID, id uint64,
			input *UpdatePermissionGroupInput,
		) (*model.PermissionGroup, error) {
			require.NotNil(t, input.Name)
			require.Len(t, input.Rules, 1)
			assert.Equal(t, "owners", input.Rules[0].Resource)
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     *input.Name,
				Color:    "#123456",
			}, nil
		},
	}
	handler := permissionHTTPHandler(
		service,
		permissionHTTPEffectiveService{
			rules: []model.PermissionGroupRule{{
				Resource: string(model.ResourceMasterPermission),
				CanEdit:  true,
			}},
		},
		audit,
	)

	createContext, createResponse := permissionHTTPContext(
		t,
		http.MethodPost,
		"/permission-groups",
		CreatePermissionGroupRequest{
			Name:  "created",
			Color: "#123456",
			Rules: []PermissionGroupRuleInput{{
				Resource: "owners",
				CanView:  true,
			}},
		},
		setPermissionHTTPIdentity,
	)
	handler.CreatePermissionGroup(createContext)
	assert.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())
	assert.Equal(t, "/v1/masters/permission-groups/7", createResponse.Header().Get("Location"))
	require.Len(t, service.audits, 1)
	assert.Equal(t, uint64(23), service.audits[0].ClinicID)
	assert.Equal(t, uint64(17), service.audits[0].ActorStaffID)
	assert.Equal(t, model.AuditActionPermissionGroupCreate, service.audits[0].Action)
	assert.Equal(t, "permission_group", service.audits[0].Resource)
	assert.Empty(t, audit.entries, "HTTP must not write a best-effort mutation audit")

	updateContext, updateResponse := permissionHTTPContext(
		t,
		http.MethodPatch,
		"/permission-groups/7",
		map[string]any{
			"name": "updated",
			"rules": []map[string]any{{
				"resource": "owners",
				"can_edit": true,
			}},
		},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	handler.UpdatePermissionGroup(updateContext)
	assert.Equal(t, http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	require.Len(t, service.audits, 2)
	assert.Equal(t, model.AuditActionPermissionGroupUpdate, service.audits[1].Action)
	assert.Equal(t, "permission_group", service.audits[1].Resource)
	assert.Empty(t, audit.entries, "HTTP must not write a best-effort mutation audit")

	invalidCreate, invalidCreateResponse := permissionHTTPContext(
		t,
		http.MethodPost,
		"/permission-groups",
		map[string]any{"name": "missing color"},
		setPermissionHTTPIdentity,
	)
	handler.CreatePermissionGroup(invalidCreate)
	assert.Equal(t, http.StatusBadRequest, invalidCreateResponse.Code)

	createErrorHandler := permissionHTTPHandler(&permissionHTTPService{
		createFn: func(context.Context, uint64, *CreatePermissionGroupInput) (*model.PermissionGroup, error) {
			return nil, errors.New("create failed")
		},
	}, nil, nil)
	createError, createErrorResponse := permissionHTTPContext(
		t,
		http.MethodPost,
		"/permission-groups",
		CreatePermissionGroupRequest{Name: "created", Color: "#123456"},
		setPermissionHTTPIdentity,
	)
	createErrorHandler.CreatePermissionGroup(createError)
	assert.Equal(t, http.StatusInternalServerError, createErrorResponse.Code)

	updateErrorHandler := permissionHTTPHandler(&permissionHTTPService{
		getFn: func(context.Context, uint64, uint64) (*model.PermissionGroup, error) {
			return nil, errors.New("old value unavailable")
		},
		updateFn: func(context.Context, uint64, uint64, *UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
			return nil, errors.New("update failed")
		},
	}, nil, nil)
	updateError, updateErrorResponse := permissionHTTPContext(
		t,
		http.MethodPatch,
		"/permission-groups/7",
		map[string]any{"name": "updated"},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	updateErrorHandler.UpdatePermissionGroup(updateError)
	assert.Equal(t, http.StatusInternalServerError, updateErrorResponse.Code)
}

func TestHTTPHandler_CreatePermissionGroup_IsActivePresence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		body         map[string]any
		wantIsActive bool
	}{
		{
			name: "omitted defaults to true",
			body: map[string]any{
				"name":  "omitted-active",
				"color": "#112233",
			},
			wantIsActive: true,
		},
		{
			name: "explicit false is preserved",
			body: map[string]any{
				"name":      "explicit-false",
				"color":     "#112233",
				"is_active": false,
			},
			wantIsActive: false,
		},
		{
			name: "explicit true is preserved",
			body: map[string]any{
				"name":      "explicit-true",
				"color":     "#112233",
				"is_active": true,
			},
			wantIsActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *CreatePermissionGroupInput
			handler := permissionHTTPHandler(&permissionHTTPService{
				createFn: func(
					_ context.Context,
					_ uint64,
					input *CreatePermissionGroupInput,
				) (*model.PermissionGroup, error) {
					got = input
					return &model.PermissionGroup{
						ID:       9,
						ClinicID: 23,
						Name:     input.Name,
						Color:    input.Color,
						IsActive: input.IsActive,
					}, nil
				},
			}, nil, nil)

			c, response := permissionHTTPContext(
				t,
				http.MethodPost,
				"/permission-groups",
				tt.body,
				setPermissionHTTPIdentity,
			)
			handler.CreatePermissionGroup(c)

			require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
			require.NotNil(t, got)
			assert.Equal(t, tt.wantIsActive, got.IsActive)
		})
	}
}

func TestHTTPHandler_CreatePermissionGroupWithRulesRequiresEditPermission(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	called := false
	handler := permissionHTTPHandler(
		&permissionHTTPService{
			createFn: func(
				context.Context,
				uint64,
				*CreatePermissionGroupInput,
			) (*model.PermissionGroup, error) {
				called = true
				return &model.PermissionGroup{}, nil
			},
		},
		permissionHTTPEffectiveService{
			rules: []model.PermissionGroupRule{{
				Resource:  string(model.ResourceMasterPermission),
				CanCreate: true,
			}},
		},
		nil,
	)
	c, response := permissionHTTPContext(
		t,
		http.MethodPost,
		"/permission-groups",
		CreatePermissionGroupRequest{
			Name:  "create-only denied",
			Color: "#123456",
			Rules: []PermissionGroupRuleInput{{
				Resource: "owners",
				CanView:  true,
			}},
		},
		setPermissionHTTPIdentity,
	)

	handler.CreatePermissionGroup(c)

	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.False(t, called)
}

func TestHTTPHandler_DeletePermissionGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audit := &permissionHTTPAuditLogger{}
	service := &permissionHTTPService{}
	handler := permissionHTTPHandler(service, nil, audit)
	c, _ := permissionHTTPContext(
		t,
		http.MethodDelete,
		"/permission-groups/7",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	handler.DeletePermissionGroup(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	require.Len(t, service.audits, 1)
	assert.Equal(t, model.AuditActionPermissionGroupDelete, service.audits[0].Action)
	assert.Equal(t, "permission_group", service.audits[0].Resource)
	assert.Empty(t, audit.entries, "HTTP must not write a best-effort mutation audit")

	errorHandler := permissionHTTPHandler(&permissionHTTPService{
		getFn: func(context.Context, uint64, uint64) (*model.PermissionGroup, error) {
			return nil, errors.New("old value unavailable")
		},
		deleteFn: func(context.Context, uint64, uint64) error {
			return errors.New("delete failed")
		},
	}, nil, nil)
	errorContext, errorResponse := permissionHTTPContext(
		t,
		http.MethodDelete,
		"/permission-groups/7",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	errorHandler.DeletePermissionGroup(errorContext)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)

	invalidID, invalidIDResponse := permissionHTTPContext(
		t,
		http.MethodDelete,
		"/permission-groups/bad",
		nil,
		func(c *gin.Context) { setPermissionHTTPID(c, "bad") },
	)
	handler.DeletePermissionGroup(invalidID)
	assert.Equal(t, http.StatusBadRequest, invalidIDResponse.Code)
}

func TestHTTPHandler_SetRulesAndReorderPermissionGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audit := &permissionHTTPAuditLogger{err: errors.New("audit unavailable")}
	service := &permissionHTTPService{
		updateRulesFn: func(
			_ context.Context,
			clinicID, groupID uint64,
			inputs []SetPermissionGroupRulesInput,
			actorStaffID uint64,
		) error {
			assert.Equal(t, uint64(23), clinicID)
			assert.Equal(t, uint64(7), groupID)
			assert.Equal(t, uint64(17), actorStaffID)
			require.Len(t, inputs, 1)
			return nil
		},
		reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
			assert.Equal(t, uint64(23), clinicID)
			assert.Equal(t, []uint64{7, 8}, ids)
			return nil
		},
	}
	handler := permissionHTTPHandler(service, nil, audit)

	rulesContext, rulesResponse := permissionHTTPContext(
		t,
		http.MethodPut,
		"/permission-groups/7/rules",
		SetPermissionGroupRulesRequest{Rules: []PermissionGroupRuleInput{{
			Resource: "owners",
			CanView:  true,
		}}},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	handler.SetPermissionGroupRules(rulesContext)
	assert.Equal(t, http.StatusOK, rulesResponse.Code, rulesResponse.Body.String())
	require.Len(t, service.audits, 1)
	assert.Equal(t, model.AuditActionPermissionRulesUpdate, service.audits[0].Action)
	assert.Equal(t, "permission_group_rules", service.audits[0].Resource)
	assert.Empty(t, audit.entries, "HTTP must not write a best-effort mutation audit")

	reorderContext, _ := permissionHTTPContext(
		t,
		http.MethodPatch,
		"/permission-groups/reorder",
		ReorderPermissionGroupsRequest{IDs: []uint64{7, 8}},
		setPermissionHTTPIdentity,
	)
	handler.ReorderPermissionGroups(reorderContext)
	assert.Equal(t, http.StatusNoContent, reorderContext.Writer.Status())

	invalidRules, invalidRulesResponse := permissionHTTPContext(
		t,
		http.MethodPut,
		"/permission-groups/7/rules",
		map[string]any{"rules": []any{}},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	handler.SetPermissionGroupRules(invalidRules)
	assert.Equal(t, http.StatusBadRequest, invalidRulesResponse.Code)

	invalidReorder, invalidReorderResponse := permissionHTTPContext(
		t,
		http.MethodPatch,
		"/permission-groups/reorder",
		map[string]any{"ids": []uint64{}},
		setPermissionHTTPIdentity,
	)
	handler.ReorderPermissionGroups(invalidReorder)
	assert.Equal(t, http.StatusBadRequest, invalidReorderResponse.Code)

	initialLookupErrorHandler := permissionHTTPHandler(&permissionHTTPService{
		updateRulesFn: func(
			context.Context,
			uint64,
			uint64,
			[]SetPermissionGroupRulesInput,
			uint64,
		) error {
			return apperrors.WrapNotFound("permission_group", "7")
		},
	}, nil, nil)
	initialLookupError, initialLookupErrorResponse := permissionHTTPContext(
		t,
		http.MethodPut,
		"/permission-groups/7/rules",
		SetPermissionGroupRulesRequest{Rules: []PermissionGroupRuleInput{{Resource: "owners"}}},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	initialLookupErrorHandler.SetPermissionGroupRules(initialLookupError)
	assert.Equal(t, http.StatusNotFound, initialLookupErrorResponse.Code)

	updateErrorHandler := permissionHTTPHandler(&permissionHTTPService{
		updateRulesFn: func(context.Context, uint64, uint64, []SetPermissionGroupRulesInput, uint64) error {
			return errors.New("rules update failed")
		},
	}, nil, nil)
	updateError, updateErrorResponse := permissionHTTPContext(
		t,
		http.MethodPut,
		"/permission-groups/7/rules",
		SetPermissionGroupRulesRequest{Rules: []PermissionGroupRuleInput{{Resource: "owners"}}},
		func(c *gin.Context) { setPermissionHTTPID(c, "7") },
	)
	updateErrorHandler.SetPermissionGroupRules(updateError)
	assert.Equal(t, http.StatusInternalServerError, updateErrorResponse.Code)

	reorderErrorHandler := permissionHTTPHandler(&permissionHTTPService{
		reorderFn: func(context.Context, uint64, []uint64) error {
			return errors.New("reorder failed")
		},
	}, nil, nil)
	reorderError, reorderErrorResponse := permissionHTTPContext(
		t,
		http.MethodPatch,
		"/permission-groups/reorder",
		ReorderPermissionGroupsRequest{IDs: []uint64{7}},
		setPermissionHTTPIdentity,
	)
	reorderErrorHandler.ReorderPermissionGroups(reorderError)
	assert.Equal(t, http.StatusInternalServerError, reorderErrorResponse.Code)
}
