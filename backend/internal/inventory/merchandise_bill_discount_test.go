package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EMR-449 A01 fixture F（wire 部分）: 省略 / true / false / null / 非 boolean。
func TestBillDiscountExcludedField_Decode(t *testing.T) {
	type body struct {
		Field billDiscountExcludedField `json:"is_bill_discount_excluded"`
	}
	tests := []struct {
		name      string
		payload   string
		wantErr   bool
		wantInput *bool
	}{
		{name: "omitted keeps nil", payload: `{}`, wantInput: nil},
		{name: "explicit true", payload: `{"is_bill_discount_excluded":true}`, wantInput: boolPtrForBillDiscountTest(true)},
		{name: "explicit false is not dropped", payload: `{"is_bill_discount_excluded":false}`, wantInput: boolPtrForBillDiscountTest(false)},
		{name: "null is rejected", payload: `{"is_bill_discount_excluded":null}`, wantErr: true},
		{name: "string is rejected", payload: `{"is_bill_discount_excluded":"true"}`, wantErr: true},
		{name: "number is rejected", payload: `{"is_bill_discount_excluded":1}`, wantErr: true},
		{name: "object is rejected", payload: `{"is_bill_discount_excluded":{}}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got body
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantInput, got.Field.toUpdateInput())
			assert.Equal(t, tt.wantInput, got.Field.toCreateInput())
		})
	}
}

func TestCreateBillDiscountExcluded_DefaultsToExplicitFalse(t *testing.T) {
	tests := []struct {
		name  string
		input *bool
		want  bool
	}{
		{name: "omitted saves explicit false", input: nil, want: false},
		{name: "explicit true", input: boolPtrForBillDiscountTest(true), want: true},
		{name: "explicit false", input: boolPtrForBillDiscountTest(false), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createBillDiscountExcluded(tt.input)
			require.NotNil(t, got, "create must never leave the column to the DB default")
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestSetBillDiscountExcludedField(t *testing.T) {
	tests := []struct {
		name    string
		value   *bool
		wantSet bool
		want    bool
	}{
		{name: "omitted keeps the stored NULL/false/true", value: nil, wantSet: false},
		{name: "explicit false is written", value: boolPtrForBillDiscountTest(false), wantSet: true, want: false},
		{name: "explicit true is written", value: boolPtrForBillDiscountTest(true), wantSet: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := map[string]any{}
			setBillDiscountExcludedField(fields, tt.value)
			got, ok := fields[colIsBillDiscountExcluded]
			assert.Equal(t, tt.wantSet, ok)
			if tt.wantSet {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLogBillDiscountAuditTx(t *testing.T) {
	actorID := uint64(7)
	tests := []struct {
		name       string
		change     billDiscountChange
		wantBefore any
		wantAfter  map[string]any
	}{
		{
			name:       "create has no before row",
			change:     billDiscountChange{ResourceID: 3, BeforeExists: false, After: boolPtrForBillDiscountTest(false)},
			wantBefore: nil,
			wantAfter:  map[string]any{colIsBillDiscountExcluded: false},
		},
		{
			name:       "stored raw NULL is kept distinct from create",
			change:     billDiscountChange{ResourceID: 3, BeforeExists: true, Before: nil, After: boolPtrForBillDiscountTest(true)},
			wantBefore: map[string]any{colIsBillDiscountExcluded: nil},
			wantAfter:  map[string]any{colIsBillDiscountExcluded: true},
		},
		{
			name:       "true to false",
			change:     billDiscountChange{ResourceID: 3, BeforeExists: true, Before: boolPtrForBillDiscountTest(true), After: boolPtrForBillDiscountTest(false)},
			wantBefore: map[string]any{colIsBillDiscountExcluded: true},
			wantAfter:  map[string]any{colIsBillDiscountExcluded: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *AuditEntry
			audit := auditTxLoggerFuncForBillDiscountTest(func(_ context.Context, entry *AuditEntry) error {
				got = entry
				return nil
			})
			err := logBillDiscountAuditTx(context.Background(), audit, 1, &actorID, "merchandise_item", "update", tt.change)
			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.wantBefore == nil {
				assert.Nil(t, got.OldValue)
			} else {
				assert.Equal(t, tt.wantBefore, got.OldValue)
			}
			assert.Equal(t, tt.wantAfter, got.NewValue)
			assert.Equal(t, map[string]any{"field": colIsBillDiscountExcluded}, got.Metadata)
			require.NotNil(t, got.ResourceID)
			assert.Equal(t, uint64(3), *got.ResourceID)
		})
	}
}

func TestLogBillDiscountAuditTx_FailsClosed(t *testing.T) {
	actorID := uint64(7)
	change := billDiscountChange{ResourceID: 3, BeforeExists: true, After: boolPtrForBillDiscountTest(true)}

	t.Run("missing audit dependency", func(t *testing.T) {
		err := logBillDiscountAuditTx(context.Background(), nil, 1, &actorID, "merchandise_item", "update", change)
		require.Error(t, err)
	})
	t.Run("missing actor", func(t *testing.T) {
		audit := auditTxLoggerFuncForBillDiscountTest(func(context.Context, *AuditEntry) error { return nil })
		err := logBillDiscountAuditTx(context.Background(), audit, 1, nil, "merchandise_item", "update", change)
		require.Error(t, err)
	})
	t.Run("audit write failure is returned for rollback", func(t *testing.T) {
		sentinel := errors.New("audit sink down")
		audit := auditTxLoggerFuncForBillDiscountTest(func(context.Context, *AuditEntry) error { return sentinel })
		err := logBillDiscountAuditTx(context.Background(), audit, 1, &actorID, "merchandise_item", "update", change)
		require.ErrorIs(t, err, sentinel)
	})
}

func boolPtrForBillDiscountTest(v bool) *bool { return &v }

// auditTxLoggerFuncForBillDiscountTest は手書き fn-field モック（golang-testing 規約）。
type auditTxLoggerFuncForBillDiscountTest func(ctx context.Context, entry *AuditEntry) error

func (f auditTxLoggerFuncForBillDiscountTest) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	return f(ctx, entry)
}
