package billing

// billing_item_request_move_test.go — BE9-2C B③: handler/accounting_request_test.go から
// 明細/返金 request のテストを実装と同 package へ分離移動。

import (
	"github.com/gin-gonic/gin/binding"

	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestPaymentMethodPointer_Binding_Oneof_Refund は createRefundRequest の oneof binding を検証する
// （旧 handler 側 TestPaymentMethodPointer_Binding_Oneof から返金分を分離・B③）。
func TestPaymentMethodPointer_Binding_Oneof_Refund(t *testing.T) {
	bankTransfer := string(model.PaymentMethodBankTransfer)
	cash := string(model.PaymentMethodCash)
	t.Run("createRefundRequest accepts bank_transfer", func(t *testing.T) {
		req := createRefundRequest{Amount: 1, PaymentMethod: &bankTransfer}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Fatalf("ValidateStruct = %v, want nil for bank_transfer refund", err)
		}
	})
	t.Run("createRefundRequest keeps existing cash valid", func(t *testing.T) {
		req := createRefundRequest{Amount: 1, PaymentMethod: &cash}
		if err := binding.Validator.ValidateStruct(&req); err != nil {
			t.Fatalf("ValidateStruct = %v, want nil for cash refund", err)
		}
	})
}
