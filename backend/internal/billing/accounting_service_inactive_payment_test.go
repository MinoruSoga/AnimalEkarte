package billing

import (
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-465: 停止中の支払方法による新受領の拒否。既存受領と同じ method の再保存・
// idempotent replay は拒否しない（EMR-449 調査の引渡し契約どおり）。
func TestEnsureNoInactiveNewReceiptMethods(t *testing.T) {
	tests := []struct {
		name        string
		methods     []model.PaymentMethod
		saved       map[model.PaymentMethod]bool
		activeKeys  map[string]bool
		wantReject  bool
		wantContain string
	}{
		{
			name:       "有効な method の新受領は拒否されない",
			methods:    []model.PaymentMethod{model.PaymentMethodCash},
			saved:      nil,
			activeKeys: map[string]bool{"cash": true},
			wantReject: false,
		},
		{
			name:        "停止中の method の新受領は拒否される",
			methods:     []model.PaymentMethod{model.PaymentMethodCreditCard},
			saved:       nil,
			activeKeys:  map[string]bool{"cash": true},
			wantReject:  true,
			wantContain: "停止中の支払方法",
		},
		{
			name:       "既存受領と同じ method の再保存は拒否されない（停止後に停止された場合を含む）",
			methods:    []model.PaymentMethod{model.PaymentMethodCreditCard},
			saved:      map[model.PaymentMethod]bool{model.PaymentMethodCreditCard: true},
			activeKeys: map[string]bool{"cash": true},
			wantReject: false,
		},
		{
			name: "複合決済で1つでも新規採用の停止 method があれば拒否",
			methods: []model.PaymentMethod{
				model.PaymentMethodCash,
				model.PaymentMethodElectronicMoney,
			},
			saved:      map[model.PaymentMethod]bool{model.PaymentMethodCash: true},
			activeKeys: map[string]bool{"cash": true},
			wantReject: true,
			wantContain: "electronic_money",
		},
		{
			name:       "空 method は検査対象外（解決経路で扱う）",
			methods:    []model.PaymentMethod{""},
			saved:      nil,
			activeKeys: map[string]bool{},
			wantReject: false,
		},
		{
			name:       "未知の method は skip（resolvePaymentMethodMasterID で拒否される）",
			methods:    []model.PaymentMethod{model.PaymentMethod("unknown_method")},
			saved:      nil,
			activeKeys: map[string]bool{"cash": true},
			wantReject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureNoInactiveNewReceiptMethods(tt.methods, tt.saved, tt.activeKeys)
			if tt.wantReject {
				if err == nil {
					t.Fatalf("expected rejection, got nil")
				}
				if tt.wantContain != "" && !strings.Contains(err.Error(), tt.wantContain) {
					t.Fatalf("error %q should contain %q", err.Error(), tt.wantContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}


