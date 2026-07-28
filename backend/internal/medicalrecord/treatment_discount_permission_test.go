package medicalrecord

// treatment_discount_permission_test.go — BE9-2D ④b: BUG-372 discount ガード4メソッドの
// medicalrecord 側直接テスト。原本 internal/handler/discount_permission.go のコピー
// （treatment_handler.go）が原本側のテストから外れて独立フォークになったため、
// nil/zero/equal スキップ・権限分岐を本 package で固定し、原本との silent 乖離を検出する
// （Phase 2 敵対レビュー MEDIUM 指摘の是正）。

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func denyAllPermission(_ *gin.Context, _, _ string) bool  { return false }
func allowAllPermission(_ *gin.Context, _, _ string) bool { return true }

func newDiscountTestHandler(check PermissionChecker) *TreatmentHandler {
	return NewTreatmentHandler(&mockTreatmentService{}, check)
}

func discountTestCtx(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func TestTreatmentDiscountGuards_CreateFloat(t *testing.T) {
	c := discountTestCtx(t)

	t.Run("zero value requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountCreateFloat(c, 0))
	})
	t.Run("non-zero without permission is forbidden", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		err := h.requireDiscountCreateFloat(c, 10)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	})
	t.Run("non-zero with permission passes", func(t *testing.T) {
		h := newDiscountTestHandler(allowAllPermission)
		assert.NoError(t, h.requireDiscountCreateFloat(c, 10))
	})
}

func TestTreatmentDiscountGuards_CreateInt(t *testing.T) {
	c := discountTestCtx(t)

	t.Run("zero value requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountCreateInt(c, 0))
	})
	t.Run("non-zero without permission is forbidden", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		err := h.requireDiscountCreateInt(c, 100)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	})
	t.Run("non-zero with permission passes", func(t *testing.T) {
		h := newDiscountTestHandler(allowAllPermission)
		assert.NoError(t, h.requireDiscountCreateInt(c, 100))
	})
}

func TestTreatmentDiscountGuards_EditFloat(t *testing.T) {
	c := discountTestCtx(t)
	fv := func(v float64) *float64 { return &v }

	t.Run("nil (unspecified) requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountEditFloat(c, nil, 5))
	})
	t.Run("unchanged value (epsilon-equal) requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountEditFloat(c, fv(5), 5))
		assert.NoError(t, h.requireDiscountEditFloat(c, fv(5+discountFloatEpsilon/2), 5))
	})
	t.Run("changed value without permission is forbidden", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		err := h.requireDiscountEditFloat(c, fv(10), 5)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	})
	t.Run("changed value with permission passes", func(t *testing.T) {
		h := newDiscountTestHandler(allowAllPermission)
		assert.NoError(t, h.requireDiscountEditFloat(c, fv(10), 5))
	})
}

func TestTreatmentDiscountGuards_EditInt(t *testing.T) {
	c := discountTestCtx(t)
	iv := func(v int64) *int64 { return &v }

	t.Run("nil (unspecified) requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountEditInt(c, nil, 100))
	})
	t.Run("unchanged value requires no permission", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		assert.NoError(t, h.requireDiscountEditInt(c, iv(100), 100))
	})
	t.Run("changed value without permission is forbidden", func(t *testing.T) {
		h := newDiscountTestHandler(denyAllPermission)
		err := h.requireDiscountEditInt(c, iv(200), 100)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	})
	t.Run("changed value with permission passes", func(t *testing.T) {
		h := newDiscountTestHandler(allowAllPermission)
		assert.NoError(t, h.requireDiscountEditInt(c, iv(200), 100))
	})
}

// TestCreateTreatment_DiscountForbidden は CreateTreatment 経路にガードが配線されていること
// （権限なし + 非ゼロ discount_rate → 403・service 未到達）を固定する。
func TestCreateTreatment_DiscountForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// createFn nil の mock: ガードをすり抜けて service に到達すると nil fn 呼び出しで
	// panic → test fail するため、「未到達」自体が構造的に検証される。
	h := NewTreatmentHandler(&mockTreatmentService{}, denyAllPermission)

	body, err := json.Marshal(map[string]any{
		"item_type":     "procedure",
		"unit_price":    1000,
		"quantity":      1.0,
		"discount_rate": 10,
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	h.CreateTreatment(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
