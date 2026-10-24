package httpapi

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// discountExactRecorder は PermissionChecker の呼出しを記録し、固定の判定を返す。
type discountExactRecorder struct {
	allow bool
	calls []string
}

func (r *discountExactRecorder) check(_ *gin.Context, resource, action string) bool {
	r.calls = append(r.calls, resource+":"+action)
	return r.allow
}

// discountExactToken はテスト用に10進 token から ExactNumeric を作る（不正な literal はテストの誤り）。
func discountExactToken(t *testing.T, token string) model.ExactNumeric {
	t.Helper()
	value, err := model.ParseExactNumeric(token)
	require.NoError(t, err)
	return value
}

func discountExactPtr(value model.ExactNumeric) *model.ExactNumeric {
	return &value
}

func TestRequireDiscountEditExact(t *testing.T) {
	nan := model.ExactNumeric{Numeric: pgtype.Numeric{NaN: true, Valid: true}}
	cases := []struct {
		name      string
		newVal    *model.ExactNumeric
		oldVal    model.ExactNumeric
		allow     bool
		wantErr   bool
		wantCalls []string
	}{
		{name: "新値なしは権限不要", oldVal: discountExactToken(t, "10")},
		{name: "scaleだけ違う同値は権限不要", newVal: discountExactPtr(discountExactToken(t, "10.00")), oldVal: discountExactToken(t, "10")},
		{name: "変更かつ権限なしは拒否", newVal: discountExactPtr(discountExactToken(t, "12")), oldVal: discountExactToken(t, "10"), wantErr: true, wantCalls: []string{"discount:edit"}},
		{name: "変更かつ権限ありは許可", newVal: discountExactPtr(discountExactToken(t, "12")), oldVal: discountExactToken(t, "10"), allow: true, wantCalls: []string{"discount:edit"}},
		{name: "epsilon内の差も変更として扱う", newVal: discountExactPtr(discountExactToken(t, "0.10001")), oldVal: discountExactToken(t, "0.1"), wantErr: true, wantCalls: []string{"discount:edit"}},
		{name: "NaNは変更として扱う", newVal: discountExactPtr(nan), oldVal: discountExactToken(t, "0"), wantErr: true, wantCalls: []string{"discount:edit"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &discountExactRecorder{allow: tc.allow}
			err := RequireDiscountEditExact(nil, rec.check, tc.newVal, tc.oldVal)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantCalls, rec.calls)
		})
	}
}

func TestRequireDiscountCreateExact(t *testing.T) {
	cases := []struct {
		name      string
		val       model.ExactNumeric
		allow     bool
		wantErr   bool
		wantCalls []string
	}{
		{name: "scaleだけ違うゼロは権限不要", val: discountExactToken(t, "0.00")},
		{name: "非ゼロかつ権限なしは拒否", val: discountExactToken(t, "5"), wantErr: true, wantCalls: []string{"discount:create"}},
		{name: "非ゼロかつ権限ありは許可", val: discountExactToken(t, "5"), allow: true, wantCalls: []string{"discount:create"}},
		{name: "NULLはゼロとみなさず権限を要求する", val: model.ExactNumeric{}, wantErr: true, wantCalls: []string{"discount:create"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &discountExactRecorder{allow: tc.allow}
			err := RequireDiscountCreateExact(nil, rec.check, tc.val)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantCalls, rec.calls)
		})
	}
}
