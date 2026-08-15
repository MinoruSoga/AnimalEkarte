package owner

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- DeleteOwner ----
//
// c.Status(http.StatusNoContent) は Gin の ResponseWriter にステータスをバッファするだけで
// httptest.ResponseRecorder には即時書き込まれない。
// そのため NoContent 系レスポンスは gin.Engine 経由でリクエストを送る。

func newDeleteOwnerRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.DELETE("/owners/:id", func(c *gin.Context) {
		setClinicID(c)
	}, h.DeleteOwner)
	return r
}

func TestDeleteOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "deletes owner successfully",
			paramID: "1",
			svc: &mockOwnerService{
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			svc: &mockOwnerService{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newDeleteOwnerRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/owners/"+tt.paramID, http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	// clinic_id なしのケースは CreateTestContext で直接テスト
	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/", http.NoBody)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.DeleteOwner(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func newPatchDeliveryExclusionRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.PATCH("/owners/:id/delivery-exclusion", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateOwnerDeliveryExclusion)
	return r
}

func TestPatchOwnerDeliveryExclusion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "sets delivery_excluded=true",
			paramID: "1",
			body:    `{"excluded":true,"reason":"配信不要"}`,
			svc: &mockOwnerService{
				updateDeliveryExclusionFn: func(_ context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: clinicID, DeliveryExcluded: input.Excluded, DeliveryExcludedReason: input.Reason}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "sets delivery_excluded=false without reason",
			paramID: "1",
			body:    `{"excluded":false}`,
			svc: &mockOwnerService{
				updateDeliveryExclusionFn: func(_ context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: clinicID, DeliveryExcluded: input.Excluded}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       `{invalid}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       `{"excluded":true}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			body:    `{"excluded":true}`,
			svc: &mockOwnerService{
				updateDeliveryExclusionFn: func(_ context.Context, _, _ uint64, _ UpdateDeliveryExclusionInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchDeliveryExclusionRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/delivery-exclusion", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func newPatchDeliveryCautionRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.PATCH("/owners/:id/delivery-caution", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateOwnerDeliveryCaution)
	return r
}

func TestPatchOwnerDeliveryCaution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "sets delivery_caution=true",
			paramID: "1",
			body:    `{"caution":true,"reason":"注意が必要"}`,
			svc: &mockOwnerService{
				updateDeliveryCautionFn: func(_ context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error) {
					reason := input.Reason
					return &model.Owner{ID: id, ClinicID: clinicID, DeliveryCaution: input.Caution, DeliveryCautionReason: &reason}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "sets delivery_caution=false without reason",
			paramID: "1",
			body:    `{"caution":false}`,
			svc: &mockOwnerService{
				updateDeliveryCautionFn: func(_ context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: clinicID, DeliveryCaution: input.Caution}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       `{invalid}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       `{"caution":true}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			body:    `{"caution":true}`,
			svc: &mockOwnerService{
				updateDeliveryCautionFn: func(_ context.Context, _, _ uint64, _ UpdateDeliveryCautionInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchDeliveryCautionRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/delivery-caution", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func newPatchTransferStatusRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.PATCH("/owners/:id/transfer-status", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateOwnerTransferStatus)
	return r
}

func TestPatchOwnerTransferStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "sets is_transferred=true",
			paramID: "1",
			body:    `{"is_transferred":true}`,
			svc: &mockOwnerService{
				updateTransferStatusFn: func(_ context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: clinicID, IsTransferred: input.IsTransferred}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "sets is_transferred=false",
			paramID: "1",
			body:    `{"is_transferred":false}`,
			svc: &mockOwnerService{
				updateTransferStatusFn: func(_ context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: clinicID, IsTransferred: input.IsTransferred}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       `{invalid}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       `{"is_transferred":true}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			body:    `{"is_transferred":true}`,
			svc: &mockOwnerService{
				updateTransferStatusFn: func(_ context.Context, _, _ uint64, _ UpdateTransferStatusInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchTransferStatusRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/transfer-status", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func newPatchLineIDConfirmRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.PATCH("/owners/:id/line-id-confirm", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateOwnerLineIDConfirm)
	return r
}

func TestPatchOwnerLineIDConfirm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "confirms line id successfully",
			paramID: "1",
			svc: &mockOwnerService{
				confirmLineIDFn: func(_ context.Context, clinicID, id uint64, _ *uint64) (*model.Owner, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					now := time.Now()
					return &model.Owner{ID: id, ClinicID: clinicID, LineIDConfirmedAt: &now}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			svc: &mockOwnerService{
				confirmLineIDFn: func(_ context.Context, _, _ uint64, _ *uint64) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchLineIDConfirmRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/line-id-confirm", http.NoBody)
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- UpdateOwner: discount_rate authorization branch ----

// newHandlerWithOwnerAndPermSvc は discount_rate 権限チェック (BUG-372) の検証用に
// EffectivePermission も注入したハンドラを構築する。
func newHandlerWithOwnerAndPermSvc(
	svc Service,
	permSvc *mockEffectivePermissionService,
) *Handler {
	hasPermission := func(c *gin.Context, resource, action string) bool {
		isAdmin, ok := httpapi.ExtractIsSystemAdmin(c)
		if !ok {
			return false
		}
		if isAdmin {
			return true
		}
		staffID, ok := httpapi.ExtractStaffID(c)
		if !ok {
			return false
		}
		clinicID, ok := httpapi.ExtractClinicID(c)
		if !ok {
			return false
		}
		rules, err := permSvc.GetEffectivePermissions(c.Request.Context(), staffID, clinicID)
		if err != nil {
			return false
		}
		for i := range rules {
			rule := rules[i]
			if rule.Resource != resource {
				continue
			}
			switch action {
			case "view":
				return rule.CanView
			case "create":
				return rule.CanCreate
			case "edit":
				return rule.CanEdit
			case "delete":
				return rule.CanDelete
			}
		}
		return false
	}
	return NewHandler(svc, &mockOwnerDeletionLifecycle{}, allowPermission, hasPermission)
}

func TestUpdateOwner_DiscountRateAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newDiscountRate := 15.0

	t.Run("returns error from GetByID when discount_rate lookup fails", func(t *testing.T) {
		svc := &mockOwnerService{
			getByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "1")
			},
		}
		h := newHandlerWithOwnerAndPermSvc(svc, &mockEffectivePermissionService{})

		bodyBytes, err := json.Marshal(map[string]any{"discount_rate": newDiscountRate})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)

		h.UpdateOwner(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 403 when discount_rate changed without discount:edit permission", func(t *testing.T) {
		svc := &mockOwnerService{
			getByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, DiscountRate: 0}, nil
			},
		}
		permSvc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
				return []model.PermissionGroupRule{}, nil
			},
		}
		h := newHandlerWithOwnerAndPermSvc(svc, permSvc)

		bodyBytes, err := json.Marshal(map[string]any{"discount_rate": newDiscountRate})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setNonSystemAdmin(c)

		h.UpdateOwner(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("succeeds without permission lookup when discount_rate unchanged", func(t *testing.T) {
		svc := &mockOwnerService{
			getByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, DiscountRate: newDiscountRate}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, input *UpdateOwnerInput) (*model.Owner, error) {
				require.NotNil(t, input.DiscountRate)
				return &model.Owner{ID: 1, DiscountRate: *input.DiscountRate}, nil
			},
		}
		// EffectivePermission が nil でも floatEquals early-return のため呼ばれない
		h := newHandlerWithOwnerSvc(svc)

		bodyBytes, err := json.Marshal(map[string]any{"discount_rate": newDiscountRate})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)

		h.UpdateOwner(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("succeeds when discount_rate changed with discount:edit permission", func(t *testing.T) {
		svc := &mockOwnerService{
			getByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 1, DiscountRate: 0}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, input *UpdateOwnerInput) (*model.Owner, error) {
				require.NotNil(t, input.DiscountRate)
				return &model.Owner{ID: 1, DiscountRate: *input.DiscountRate}, nil
			},
		}
		permSvc := &mockEffectivePermissionService{
			getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
				return []model.PermissionGroupRule{
					{Resource: string(model.ResourceDiscount), CanEdit: true},
				}, nil
			},
		}
		h := newHandlerWithOwnerAndPermSvc(svc, permSvc)

		bodyBytes, err := json.Marshal(map[string]any{"discount_rate": newDiscountRate})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setNonSystemAdmin(c)

		h.UpdateOwner(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ---- UpdateOwnerLineUserID ----
//
// c.Status(http.StatusNoContent) は httptest.ResponseRecorder に即時反映されないため
// gin.Engine 経由でリクエストを送る（DeleteOwner と同様のパターン）。

func newPatchLineUserIDRouter(svc Service) *gin.Engine {
	r := gin.New()
	h := newHandlerWithOwnerSvc(svc)
	r.PATCH("/owners/:id/line-user-id", func(c *gin.Context) {
		setClinicID(c)
	}, h.UpdateOwnerLineUserID)
	return r
}

func TestPatchOwnerLineUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		body       string
		svc        *mockOwnerService
		wantStatus int
	}{
		{
			name:    "links line user id successfully",
			paramID: "1",
			body:    `{"line_user_id":"U1234567890"}`,
			svc: &mockOwnerService{
				linkLineUserIDFn: func(_ context.Context, clinicID, id uint64, lineUserID *string, _ *uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(1), id)
					require.NotNil(t, lineUserID)
					assert.Equal(t, "U1234567890", *lineUserID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "unlinks line user id when null",
			paramID: "1",
			body:    `{"line_user_id":null}`,
			svc: &mockOwnerService{
				linkLineUserIDFn: func(_ context.Context, _, _ uint64, lineUserID *string, _ *uint64) error {
					assert.Nil(t, lineUserID)
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "returns 400 for invalid JSON",
			paramID:    "1",
			body:       `{invalid}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			body:       `{"line_user_id":"U1"}`,
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "returns 404 when owner not found",
			paramID: "999",
			body:    `{"line_user_id":"U1"}`,
			svc: &mockOwnerService{
				linkLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string, _ *uint64) error {
					return apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newPatchLineUserIDRouter(tt.svc)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/owners/"+tt.paramID+"/line-user-id", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	t.Run("returns 401 when clinic_id is missing", func(t *testing.T) {
		h := newHandlerWithOwnerSvc(&mockOwnerService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"line_user_id":"U1"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.UpdateOwnerLineUserID(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// ---- view permission gate tests ----

// TestListOwners_ViewPermissionDenied は view 権限を持たない非 system_admin が 403 を受けることを確認する。
func TestListOwners_ViewPermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	denyPermission := func(string, string) gin.HandlerFunc {
		return func(c *gin.Context) {
			httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
			c.Abort()
		}
	}
	h := NewHandler(
		&mockOwnerService{},
		&mockOwnerDeletionLifecycle{},
		denyPermission,
		func(*gin.Context, string, string) bool { return false },
	)
	r := gin.New()
	protected := r.Group("")
	protected.Use(func(c *gin.Context) { setNonSystemAdmin(c); setClinicID(c) })
	h.RegisterRoutes(protected)

	req := httptest.NewRequest(http.MethodGet, "/owners", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
