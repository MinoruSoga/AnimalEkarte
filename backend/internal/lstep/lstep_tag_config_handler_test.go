package lstep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// --- mock ---

type mockLstepTagConfigService struct {
	listAutoManagedPrefixesFn    func(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	createAutoManagedPrefixFn    func(ctx context.Context, input CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error)
	deleteAutoManagedPrefixFn    func(ctx context.Context, id uint64) error
	listConditionTagMappingsFn   func(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	createConditionTagMappingFn  func(ctx context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error)
	deleteConditionTagMappingFn  func(ctx context.Context, id uint64) error
	listSendPurposeTagPrefixesFn func(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
	createSendPurposeTagPrefixFn func(ctx context.Context, input CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error)
	deleteSendPurposeTagPrefixFn func(ctx context.Context, id uint64) error
}

func (m *mockLstepTagConfigService) ListAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.listAutoManagedPrefixesFn != nil {
		return m.listAutoManagedPrefixesFn(ctx)
	}
	return []*model.LstepAutoManagedPrefix{}, nil
}

func (m *mockLstepTagConfigService) CreateAutoManagedPrefix(ctx context.Context, input CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error) {
	if m.createAutoManagedPrefixFn != nil {
		return m.createAutoManagedPrefixFn(ctx, input)
	}
	return &model.LstepAutoManagedPrefix{}, nil
}

func (m *mockLstepTagConfigService) DeleteAutoManagedPrefix(ctx context.Context, id uint64) error {
	if m.deleteAutoManagedPrefixFn != nil {
		return m.deleteAutoManagedPrefixFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigService) ListConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	if m.listConditionTagMappingsFn != nil {
		return m.listConditionTagMappingsFn(ctx)
	}
	return []*model.LstepConditionTagMapping{}, nil
}

func (m *mockLstepTagConfigService) CreateConditionTagMapping(ctx context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
	if m.createConditionTagMappingFn != nil {
		return m.createConditionTagMappingFn(ctx, input)
	}
	return &model.LstepConditionTagMapping{}, nil
}

func (m *mockLstepTagConfigService) DeleteConditionTagMapping(ctx context.Context, id uint64) error {
	if m.deleteConditionTagMappingFn != nil {
		return m.deleteConditionTagMappingFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigService) ListSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	if m.listSendPurposeTagPrefixesFn != nil {
		return m.listSendPurposeTagPrefixesFn(ctx)
	}
	return []*model.LstepSendPurposeTagPrefix{}, nil
}

func (m *mockLstepTagConfigService) CreateSendPurposeTagPrefix(ctx context.Context, input CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error) {
	if m.createSendPurposeTagPrefixFn != nil {
		return m.createSendPurposeTagPrefixFn(ctx, input)
	}
	return &model.LstepSendPurposeTagPrefix{}, nil
}

func (m *mockLstepTagConfigService) DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error {
	if m.deleteSendPurposeTagPrefixFn != nil {
		return m.deleteSendPurposeTagPrefixFn(ctx, id)
	}
	return nil
}

type denyTagConfigPermission struct{}

func newLstepTagConfigRouter(svc LstepTagConfigService, deny *denyTagConfigPermission, setupCtx gin.HandlerFunc) *gin.Engine {
	permission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {
			if deny != nil {
				c.AbortWithStatus(http.StatusForbidden)
			}
		}
	}
	h := &Handler{tagConfig: svc, requirePermission: permission}
	r := gin.New()
	perm := func(action string) gin.HandlerFunc {
		return h.requirePermission(string(model.ResourceHospitalSettings), action)
	}
	r.GET("/lstep-tag-config/auto-managed-prefixes", setupCtx, perm("view"), h.ListAutoManagedPrefixes)
	r.POST("/lstep-tag-config/auto-managed-prefixes", setupCtx, perm("create"), h.CreateAutoManagedPrefix)
	r.DELETE("/lstep-tag-config/auto-managed-prefixes/:id", setupCtx, perm("delete"), h.DeleteAutoManagedPrefix)
	r.GET("/lstep-tag-config/condition-tag-mappings", setupCtx, perm("view"), h.ListConditionTagMappings)
	r.POST("/lstep-tag-config/condition-tag-mappings", setupCtx, perm("create"), h.CreateConditionTagMapping)
	r.DELETE("/lstep-tag-config/condition-tag-mappings/:id", setupCtx, perm("delete"), h.DeleteConditionTagMapping)
	r.GET("/lstep-tag-config/send-purpose-tag-prefixes", setupCtx, perm("view"), h.ListSendPurposeTagPrefixes)
	r.POST("/lstep-tag-config/send-purpose-tag-prefixes", setupCtx, perm("create"), h.CreateSendPurposeTagPrefix)
	r.DELETE("/lstep-tag-config/send-purpose-tag-prefixes/:id", setupCtx, perm("delete"), h.DeleteSendPurposeTagPrefix)
	return r
}

// --- auto_managed_prefixes tests ---

func TestListAutoManagedPrefixes_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("403 no permission", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, &denyTagConfigPermission{},
			func(c *gin.Context) { setNonSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/auto-managed-prefixes", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("200 returns list", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
					return []*model.LstepAutoManagedPrefix{
						{ID: 1, Prefix: "vaccine_", Category: "C2"},
					}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/auto-managed-prefixes", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var body []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		if assert.Len(t, body, 1) {
			assert.Equal(t, "vaccine_", body[0]["prefix"])
			assert.Equal(t, "C2", body[0]["category"])
		}
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/auto-managed-prefixes", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCreateAutoManagedPrefix_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("400 missing prefix", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		body := `{"category":"C2"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/auto-managed-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("201 created with location", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createAutoManagedPrefixFn: func(_ context.Context, input CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error) {
					return &model.LstepAutoManagedPrefix{ID: 42, Prefix: input.Prefix, Category: input.Category}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"prefix":"vaccine_","category":"C2"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/auto-managed-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "42")
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "vaccine_", resp["prefix"])
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createAutoManagedPrefixFn: func(_ context.Context, _ CreateAutoManagedPrefixInput) (*model.LstepAutoManagedPrefix, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"prefix":"vaccine_","category":"C2"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/auto-managed-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteAutoManagedPrefix_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("404 not found", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				deleteAutoManagedPrefixFn: func(_ context.Context, _ uint64) error {
					return apperrors.WrapNotFound("lstep_auto_managed_prefix", "99")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/auto-managed-prefixes/99", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("204 deleted", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/auto-managed-prefixes/1", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400 invalid id", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/auto-managed-prefixes/abc", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- condition_tag_mappings tests ---

func TestListConditionTagMappings_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns list", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return []*model.LstepConditionTagMapping{
						{ID: 1, ConditionCode: "DM", TagName: "CHRON_DM"},
					}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/condition-tag-mappings", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var body []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body, 1)
		assert.Equal(t, "DM", body[0]["condition_code"])
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/condition-tag-mappings", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCreateConditionTagMapping_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("400 missing field", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		body := `{"condition_code":"DM"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/condition-tag-mappings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// BUG-025: VARCHAR(50) overflow must be 400 with explicit bind message, not 500.
	t.Run("400 condition_code longer than 50", func(t *testing.T) {
		called := false
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createConditionTagMappingFn: func(_ context.Context, _ CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
					called = true
					return &model.LstepConditionTagMapping{ID: 1}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		code51 := strings.Repeat("A", 51)
		body := fmt.Sprintf(`{"condition_code":%q,"tag_name":"CHRON_DM"}`, code51)
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/condition-tag-mappings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "condition_code は 50 以下で入力してください")
		assert.False(t, called, "service must not run when condition_code exceeds max")
	})

	t.Run("201 created", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createConditionTagMappingFn: func(_ context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
					return &model.LstepConditionTagMapping{ID: 5, ConditionCode: input.ConditionCode, TagName: input.TagName}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"condition_code":"DM","tag_name":"CHRON_DM"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/condition-tag-mappings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "5")
	})

	// BUG-025 regression: boundary length 50 is accepted (201).
	t.Run("201 condition_code exactly 50", func(t *testing.T) {
		var gotCode string
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createConditionTagMappingFn: func(_ context.Context, input CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
					gotCode = input.ConditionCode
					return &model.LstepConditionTagMapping{ID: 6, ConditionCode: input.ConditionCode, TagName: input.TagName}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		code50 := strings.Repeat("B", 50)
		body := fmt.Sprintf(`{"condition_code":%q,"tag_name":"CHRON_DM"}`, code50)
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/condition-tag-mappings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, code50, gotCode)
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createConditionTagMappingFn: func(_ context.Context, _ CreateConditionTagMappingInput) (*model.LstepConditionTagMapping, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"condition_code":"DM","tag_name":"CHRON_DM"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/condition-tag-mappings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteConditionTagMapping_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("404 not found", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				deleteConditionTagMappingFn: func(_ context.Context, _ uint64) error {
					return apperrors.WrapNotFound("lstep_condition_tag_mapping", "99")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/condition-tag-mappings/99", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("204 deleted", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/condition-tag-mappings/1", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400 invalid id", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/condition-tag-mappings/abc", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// --- send_purpose_tag_prefixes tests ---

func TestListSendPurposeTagPrefixes_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 returns list", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
					return []*model.LstepSendPurposeTagPrefix{
						{ID: 1, Purpose: "vaccine_reminder", TagPrefix: "PREV_"},
					}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/send-purpose-tag-prefixes", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var body []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body, 1)
		assert.Equal(t, "PREV_", body[0]["tag_prefix"])
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				listSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodGet, "/lstep-tag-config/send-purpose-tag-prefixes", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCreateSendPurposeTagPrefix_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("400 missing field", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		body := `{"purpose":"vaccine_reminder"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/send-purpose-tag-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("201 created", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createSendPurposeTagPrefixFn: func(_ context.Context, input CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error) {
					return &model.LstepSendPurposeTagPrefix{ID: 9, Purpose: input.Purpose, TagPrefix: input.TagPrefix}, nil
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"purpose":"vaccine_reminder","tag_prefix":"PREV_"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/send-purpose-tag-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "9")
	})

	t.Run("500 service error", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				createSendPurposeTagPrefixFn: func(_ context.Context, _ CreateSendPurposeTagPrefixInput) (*model.LstepSendPurposeTagPrefix, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		body := `{"purpose":"vaccine_reminder","tag_prefix":"PREV_"}`
		req := httptest.NewRequest(http.MethodPost, "/lstep-tag-config/send-purpose-tag-prefixes", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteSendPurposeTagPrefix_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("404 not found", func(t *testing.T) {
		r := newLstepTagConfigRouter(
			&mockLstepTagConfigService{
				deleteSendPurposeTagPrefixFn: func(_ context.Context, _ uint64) error {
					return apperrors.WrapNotFound("lstep_send_purpose_tag_prefix", "99")
				},
			},
			nil,
			func(c *gin.Context) { setSystemAdmin(c) },
		)
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/send-purpose-tag-prefixes/99", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("204 deleted", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/send-purpose-tag-prefixes/1", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400 invalid id", func(t *testing.T) {
		r := newLstepTagConfigRouter(&mockLstepTagConfigService{}, nil,
			func(c *gin.Context) { setSystemAdmin(c) })
		req := httptest.NewRequest(http.MethodDelete, "/lstep-tag-config/send-purpose-tag-prefixes/abc", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
