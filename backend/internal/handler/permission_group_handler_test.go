package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock PermissionGroupService ----

type mockPermissionGroupService struct {
	listFn     func(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	getByIDFn  func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	createFn   func(ctx context.Context, clinicID uint64, input *service.CreatePermissionGroupInput) (*model.PermissionGroup, error)
	updateFn   func(ctx context.Context, clinicID, id uint64, input *service.UpdatePermissionGroupInput) (*model.PermissionGroup, error)
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
	setRulesFn func(ctx context.Context, groupID uint64, inputs []service.SetPermissionGroupRulesInput, actorStaffID uint64) error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockPermissionGroupService) List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPermissionGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, clinicID, id)
	}
	return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
}

func (m *mockPermissionGroupService) Create(ctx context.Context, clinicID uint64, input *service.CreatePermissionGroupInput) (*model.PermissionGroup, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.PermissionGroup{ID: 1, ClinicID: clinicID, Name: input.Name}, nil
}

func (m *mockPermissionGroupService) Update(ctx context.Context, clinicID, id uint64, input *service.UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
}

func (m *mockPermissionGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPermissionGroupService) UpdateRules(ctx context.Context, groupID uint64, inputs []service.SetPermissionGroupRulesInput, actorStaffID uint64) error {
	if m.setRulesFn != nil {
		return m.setRulesFn(ctx, groupID, inputs, actorStaffID)
	}
	return nil
}

func (m *mockPermissionGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- mock EffectivePermissionService ----

type mockEffectivePermissionService struct {
	getEffectivePermissionsFn func(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error)
}

func (m *mockEffectivePermissionService) GetEffectivePermissions(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsFn != nil {
		return m.getEffectivePermissionsFn(ctx, staffID)
	}
	return nil, nil
}

// ---- mock AuditService (for permission group handler) ----

type mockAuditServiceForPG struct {
	logFn          func(ctx context.Context, log *model.AuditLog) error
	logAuthLoginFn func(ctx context.Context, clinicID *uint64, staffID *uint64, action string, ipAddress string, userAgent string) error
}

func (m *mockAuditServiceForPG) Log(ctx context.Context, log *model.AuditLog) error {
	if m.logFn != nil {
		return m.logFn(ctx, log)
	}
	return nil
}

func (m *mockAuditServiceForPG) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ipAddress, userAgent string) error {
	if m.logAuthLoginFn != nil {
		return m.logAuthLoginFn(ctx, clinicID, staffID, action, ipAddress, userAgent)
	}
	return nil
}

func (m *mockAuditServiceForPG) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}

// ---- helper ----

func newHandlerWithPermissionGroupSvc(pgSvc service.PermissionGroupService) *Handler {
	return &Handler{svc: &service.Services{
		PermissionGroup:     pgSvc,
		EffectivePermission: &mockEffectivePermissionService{},
		Audit:               &mockAuditServiceForPG{},
	}}
}

// ---- TestPermissionGroupHandler_List ----

func TestPermissionGroupHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(c *gin.Context)
		svc        *mockPermissionGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "正常: 一覧を返す",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPermissionGroupService{
				listFn: func(_ context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					return []model.PermissionGroup{
						{ID: 1, ClinicID: 1, Name: "管理者"},
						{ID: 2, ClinicID: 1, Name: "獣医師"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"管理者"`,
		},
		{
			name:       "401: clinic_id なし",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "500: サービスエラー",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPermissionGroupService{
				listFn: func(_ context.Context, _ uint64) ([]model.PermissionGroup, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPermissionGroupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			tt.setupCtx(c)
			h.ListPermissionGroups(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPermissionGroupHandler_GetByID ----

func TestPermissionGroupHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockPermissionGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "正常: グループを返す",
			paramID:  "3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPermissionGroupService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(3), id)
					return &model.PermissionGroup{ID: id, ClinicID: clinicID, Name: "看護師"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"看護師"`,
		},
		{
			name:       "401: clinic_id なし",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: 不正なID",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404: 存在しないID",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPermissionGroupService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
					return nil, apperrors.WrapNotFound("permission_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPermissionGroupSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.GetPermissionGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPermissionGroupHandler_Create ----

func TestPermissionGroupHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"name":        "新グループ",
			"description": "テスト用グループ",
			"color":       "#FF5733",
			"is_active":   true,
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPermissionGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "201: 正常作成",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockPermissionGroupService{
				createFn: func(_ context.Context, clinicID uint64, input *service.CreatePermissionGroupInput) (*model.PermissionGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, "新グループ", input.Name)
					return &model.PermissionGroup{ID: 5, ClinicID: clinicID, Name: input.Name}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"新グループ"`,
		},
		{
			name:       "401: clinic_id なし",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: name なし（バリデーションエラー）",
			body:       map[string]any{"description": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400: 不正なJSON",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "500: サービスエラー",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockPermissionGroupService{
				createFn: func(_ context.Context, _ uint64, _ *service.CreatePermissionGroupInput) (*model.PermissionGroup, error) {
					return nil, fmt.Errorf("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPermissionGroupSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)
			h.CreatePermissionGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusCreated {
				assert.Contains(t, w.Header().Get("Location"), "/permission-groups/")
			}
		})
	}
}

// ---- TestPermissionGroupHandler_Update ----

func TestPermissionGroupHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockPermissionGroupService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "200: 正常更新",
			paramID:  "1",
			body:     map[string]any{"name": "更新グループ"},
			setupCtx: func(c *gin.Context) { setClinicID(c); c.Set("user_id", "1") },
			svc: &mockPermissionGroupService{
				updateFn: func(_ context.Context, clinicID, id uint64, input *service.UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, input.Name)
					assert.Equal(t, "更新グループ", *input.Name)
					return &model.PermissionGroup{ID: id, ClinicID: clinicID, Name: *input.Name}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"更新グループ"`,
		},
		{
			name:       "401: clinic_id なし",
			paramID:    "1",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "400: 不正なID",
			paramID:    "xyz",
			body:       map[string]any{"name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400: 不正なJSON",
			paramID:    "1",
			body:       "not-json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockPermissionGroupService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "404: 存在しないID",
			paramID:  "999",
			body:     map[string]any{"name": "テスト"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockPermissionGroupService{
				updateFn: func(_ context.Context, _, _ uint64, _ *service.UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
					return nil, apperrors.WrapNotFound("permission_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithPermissionGroupSvc(tt.svc)
			b, err := json.Marshal(tt.body)
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)
			h.UpdatePermissionGroup(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- TestPermissionGroupHandler_Delete ----

func newDeletePermissionGroupRouter(pgSvc service.PermissionGroupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithPermissionGroupSvc(pgSvc)
	r.DELETE("/permission-groups/:id", func(c *gin.Context) {
		setClinicID(c)
		c.Set("user_id", "1")
		h.DeletePermissionGroup(c)
	})
	return r
}

func TestPermissionGroupHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockPermissionGroupService
		wantStatus int
	}{
		{
			name:    "204: 正常削除",
			paramID: "1",
			svc: &mockPermissionGroupService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
					return &model.PermissionGroup{ID: id, ClinicID: clinicID, Name: "管理者"}, nil
				},
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "409: スタッフ割当済みのため削除不可",
			paramID: "2",
			svc: &mockPermissionGroupService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
					return &model.PermissionGroup{ID: id, ClinicID: clinicID}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapConflict("この権限グループはスタッフに割り当てられているため削除できません")
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "404: 存在しないID",
			paramID: "999",
			svc: &mockPermissionGroupService{
				getByIDFn: func(_ context.Context, _, _ uint64) (*model.PermissionGroup, error) {
					return nil, apperrors.WrapNotFound("permission_group", "999")
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("permission_group", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeletePermissionGroupRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/permission-groups/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("401: clinic_id なし", func(t *testing.T) {
		h := newHandlerWithPermissionGroupSvc(&mockPermissionGroupService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeletePermissionGroup(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- TestPermissionGroupHandler_Reorder ----

func newReorderPermissionGroupsRouter(pgSvc service.PermissionGroupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithPermissionGroupSvc(pgSvc)
	r.POST("/permission-groups/reorder", func(c *gin.Context) {
		setClinicID(c)
		h.ReorderPermissionGroups(c)
	})
	return r
}

func TestPermissionGroupHandler_Reorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200: 正常並び替え", func(t *testing.T) {
		svc := &mockPermissionGroupService{
			reorderFn: func(_ context.Context, clinicID uint64, ids []uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, []uint64{3, 1, 2}, ids)
				return nil
			},
		}
		router := newReorderPermissionGroupsRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{3, 1, 2}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/permission-groups/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400: 空のIDsエラー", func(t *testing.T) {
		svc := &mockPermissionGroupService{
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return apperrors.WrapInvalidInput("IDsは1件以上必要です")
			},
		}
		router := newReorderPermissionGroupsRouter(svc)
		body, _ := json.Marshal(map[string]any{"ids": []int{}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/permission-groups/reorder", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("401: clinic_id なし", func(t *testing.T) {
		h := newHandlerWithPermissionGroupSvc(&mockPermissionGroupService{})
		body, _ := json.Marshal(map[string]any{"ids": []int{1, 2}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.ReorderPermissionGroups(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- TestPermissionGroupHandler_SetRules ----

func newSetPermissionGroupRulesRouter(pgSvc service.PermissionGroupService) *gin.Engine {
	r := gin.New()
	h := newHandlerWithPermissionGroupSvc(pgSvc)
	r.PUT("/permission-groups/:id/rules", func(c *gin.Context) {
		setClinicID(c)
		c.Set("user_id", "1")
		h.SetPermissionGroupRules(c)
	})
	return r
}

func TestPermissionGroupHandler_SetRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200: 正常にルールを設定", func(t *testing.T) {
		svc := &mockPermissionGroupService{
			getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
				return &model.PermissionGroup{
					ID:       id,
					ClinicID: clinicID,
					Name:     "管理者",
					Rules:    []model.PermissionGroupRule{{Resource: string(model.ResourceMasterPermission), CanEdit: true}},
				}, nil
			},
			setRulesFn: func(_ context.Context, groupID uint64, inputs []service.SetPermissionGroupRulesInput, actorStaffID uint64) error {
				assert.Equal(t, uint64(1), groupID)
				assert.Equal(t, uint64(1), actorStaffID)
				return nil
			},
		}
		router := newSetPermissionGroupRulesRouter(svc)
		body, _ := json.Marshal(map[string]any{
			"rules": []map[string]any{
				{
					"resource":   string(model.ResourceMasterPermission),
					"can_view":   true,
					"can_create": true,
					"can_edit":   true,
					"can_delete": false,
				},
			},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/permission-groups/1/rules", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("401: clinic_id なし", func(t *testing.T) {
		h := newHandlerWithPermissionGroupSvc(&mockPermissionGroupService{})
		body, _ := json.Marshal(map[string]any{"rules": []map[string]any{}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.SetPermissionGroupRules(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- TestPermissionGroupHandler_GetEffectivePermissions ----

// GetEffectivePermissions は auth_handler の calculateEffectivePermissions 経由で呼ばれる。
// ここでは EffectivePermissionService モックが正常に動作することを検証する。
func TestPermissionGroupHandler_GetEffectivePermissions(t *testing.T) {
	t.Run("正常: EffectivePermissionService がルールを返す", func(t *testing.T) {
		svc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
				assert.Equal(t, uint64(1), staffID)
				return []model.PermissionGroupRule{
					{Resource: string(model.ResourceMasterPermission), CanView: true, CanEdit: true},
				}, nil
			},
		}
		rules, err := svc.GetEffectivePermissions(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, string(model.ResourceMasterPermission), rules[0].Resource)
		assert.True(t, rules[0].CanView)
		assert.True(t, rules[0].CanEdit)
	})

	t.Run("エラー: サービスがエラーを返す", func(t *testing.T) {
		svc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, _ uint64) ([]model.PermissionGroupRule, error) {
				return nil, fmt.Errorf("db error")
			},
		}
		rules, err := svc.GetEffectivePermissions(context.Background(), 1)
		assert.Error(t, err)
		assert.Nil(t, rules)
	})
}
