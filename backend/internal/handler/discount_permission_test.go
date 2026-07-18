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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// TestFloatEquals は浮動小数等価判定の挙動を検証する（BUG-372）。
func TestFloatEquals(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want bool
	}{
		{"完全一致 - ゼロ", 0, 0, true},
		{"完全一致 - 整数", 10, 10, true},
		{"完全一致 - 小数", 10.5, 10.5, true},
		{"epsilon 内の差は等価とみなす", 10.0, 10.00001, true},
		{"epsilon 超える差は不一致", 10.0, 10.001, false},
		{"負値 - 等価", -5.0, -5.0, true},
		{"負値 - 不一致", -5.0, -5.5, false},
		{"異符号", 0.5, -0.5, false},
		{"極小差（浮動小数誤差想定）", 0.1 + 0.2, 0.3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := floatEquals(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("floatEquals(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ---- discount permission helper test infrastructure ----

func discountFloatPtr(v float64) *float64 { return &v }
func discountInt64Ptr(v int64) *int64     { return &v }

// newHandlerWithDiscountPermSvc は EffectivePermission サービスのみを注入した Handler を返す。
func newHandlerWithDiscountPermSvc(svc service.EffectivePermissionService) *Handler {
	return &Handler{svc: &service.Services{EffectivePermission: svc}}
}

// setNonAdminWithClinic は is_system_admin=false + user_id + clinic_id をコンテキストに設定する。
func setNonAdminWithClinic(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
}

func newDiscountTestContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", http.NoBody)
	return c
}

// ---- requireDiscountEditFloat ----

func TestRequireDiscountEditFloat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		newVal    *float64
		oldVal    float64
		setupCtx  func(c *gin.Context)
		svc       *mockEffectivePermissionService
		wantError bool
	}{
		{
			name:      "nil newVal requires no permission",
			newVal:    nil,
			oldVal:    10.0,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "equal values (within epsilon) requires no permission",
			newVal:    discountFloatPtr(10.00001),
			oldVal:    10.0,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "different values with system_admin bypass",
			newVal:    discountFloatPtr(20.0),
			oldVal:    10.0,
			setupCtx:  func(c *gin.Context) { c.Set("is_system_admin", true) },
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:     "different values with edit rule granted",
			newVal:   discountFloatPtr(20.0),
			oldVal:   10.0,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanEdit: true}}, nil
				},
			},
			wantError: false,
		},
		{
			name:     "different values without edit rule returns forbidden",
			newVal:   discountFloatPtr(20.0),
			oldVal:   10.0,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanEdit: false}}, nil
				},
			},
			wantError: true,
		},
		{
			name:      "different values with missing user context returns forbidden",
			newVal:    discountFloatPtr(20.0),
			oldVal:    10.0,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiscountPermSvc(tt.svc)
			c := newDiscountTestContext()
			tt.setupCtx(c)
			err := h.requireDiscountEditFloat(c, tt.newVal, tt.oldVal)
			if tt.wantError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrForbidden))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- requireDiscountEditInt ----

func TestRequireDiscountEditInt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		newVal    *int64
		oldVal    int64
		setupCtx  func(c *gin.Context)
		svc       *mockEffectivePermissionService
		wantError bool
	}{
		{
			name:      "nil newVal requires no permission",
			newVal:    nil,
			oldVal:    100,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "equal values requires no permission",
			newVal:    discountInt64Ptr(100),
			oldVal:    100,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "different values with system_admin bypass",
			newVal:    discountInt64Ptr(200),
			oldVal:    100,
			setupCtx:  func(c *gin.Context) { c.Set("is_system_admin", true) },
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:     "different values with edit rule granted",
			newVal:   discountInt64Ptr(200),
			oldVal:   100,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanEdit: true}}, nil
				},
			},
			wantError: false,
		},
		{
			name:     "different values without edit rule returns forbidden",
			newVal:   discountInt64Ptr(200),
			oldVal:   100,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return nil, nil
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiscountPermSvc(tt.svc)
			c := newDiscountTestContext()
			tt.setupCtx(c)
			err := h.requireDiscountEditInt(c, tt.newVal, tt.oldVal)
			if tt.wantError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrForbidden))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- requireDiscountCreateFloat ----

func TestRequireDiscountCreateFloat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		val       float64
		setupCtx  func(c *gin.Context)
		svc       *mockEffectivePermissionService
		wantError bool
	}{
		{
			name:      "zero value requires no permission",
			val:       0,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "epsilon-zero value requires no permission",
			val:       0.00001,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "nonzero value with system_admin bypass",
			val:       5.0,
			setupCtx:  func(c *gin.Context) { c.Set("is_system_admin", true) },
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:     "nonzero value with create rule granted",
			val:      5.0,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanCreate: true}}, nil
				},
			},
			wantError: false,
		},
		{
			name:     "nonzero value without create rule returns forbidden",
			val:      5.0,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanCreate: false}}, nil
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiscountPermSvc(tt.svc)
			c := newDiscountTestContext()
			tt.setupCtx(c)
			err := h.requireDiscountCreateFloat(c, tt.val)
			if tt.wantError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrForbidden))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- requireDiscountCreateInt ----

func TestRequireDiscountCreateInt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		val       int64
		setupCtx  func(c *gin.Context)
		svc       *mockEffectivePermissionService
		wantError bool
	}{
		{
			name:      "zero value requires no permission",
			val:       0,
			setupCtx:  func(_ *gin.Context) {},
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:      "nonzero value with system_admin bypass",
			val:       500,
			setupCtx:  func(c *gin.Context) { c.Set("is_system_admin", true) },
			svc:       &mockEffectivePermissionService{},
			wantError: false,
		},
		{
			name:     "nonzero value with create rule granted",
			val:      500,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return []model.PermissionGroupRule{{Resource: string(model.ResourceDiscount), CanCreate: true}}, nil
				},
			},
			wantError: false,
		},
		{
			name:     "nonzero value without create rule returns forbidden",
			val:      500,
			setupCtx: setNonAdminWithClinic,
			svc: &mockEffectivePermissionService{
				getEffectivePermissionsFn: func(_ context.Context, _, _ uint64) ([]model.PermissionGroupRule, error) {
					return nil, errors.New("db failure")
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithDiscountPermSvc(tt.svc)
			c := newDiscountTestContext()
			tt.setupCtx(c)
			err := h.requireDiscountCreateInt(c, tt.val)
			if tt.wantError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrForbidden))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
