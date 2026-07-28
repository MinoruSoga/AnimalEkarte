package reservation_test

// liff_customer_key_contract_test.go — R⑤レビューMEDIUM対応:
// reservation 内部の extractLiffCustomerID 複製が使う key リテラル "liff_customer_id" と
// middleware.LiffCustomerIDKey() の一致を実行時に固定する契約テスト。
// package reservation_test（外部テストパッケージ）なので middleware import は循環しない。

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/middleware"
)

func TestLiffCustomerIDKey_ContractWithReservationCopy(t *testing.T) {
	assert.Equal(t, "liff_customer_id", middleware.LiffCustomerIDKey(),
		"middleware の key を変更する場合は reservation/liff_handler.go の extractLiffCustomerID 複製も追随すること")
}
