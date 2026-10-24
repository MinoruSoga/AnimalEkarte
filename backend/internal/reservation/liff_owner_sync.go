package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// findCustomerWithOwnerSync は LINE 顧客を取得する。
//
// EMR-504: 旧実装は owners.line_user_id を頼りに line_customers.owner_id を読み取り時に自己修復していたが、
// 両列とも DROP 済みで、顧客と飼主の紐付けは owner_line_binding_states が唯一の正本である。
// 書き手は EMR-469 の binding writer だけで、読み取り経路からの紐付け書込み（heal）は廃止した。
// 飼主は LineCustomerRepository.FindByID が state='linked' の binding から解決して customer.Owner に載せる
// （未連携・unlink_pending・legacy_unknown は Owner=nil）。name/phone 照合は行わない（SEC-CS2-F02）。
func (s *liffService) findCustomerWithOwnerSync(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
	return s.customerRepo.FindByID(ctx, clinicID, customerID)
}
