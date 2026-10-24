package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Estimate successor draft creation (TASK-012).
// Split from estimate_service.go (ARCH-A4-billing S3).

// CreateSuccessor は承認済み/却下済み見積の後継ドラフト作成を、固定の 422 で拒否する（EMR-508 / EMR-549 U4）。
// 後継改訂の置換先は estimate_successor_links で、提示原本（estimate_presentations。header_snapshot の DDL を含む）の
// 保存を前提にするが、その writer は未接続（選定済み wire は未実装）。旧実装は estimates のヘッダー合計列と
// supersedes_estimate_id 列を複製して後継ドラフトを作っていたが、どちらも廃止されたため同じ複製では成立しない。
// 成功を装う部分実装は置かず、DB には一切触れない。
// 解除条件: 提示原本の writer と estimate_successor_links の writer が接続されたとき、この実装を置き換える。
func (s *estimateService) CreateSuccessor(
	_ context.Context,
	_, _ uint64,
	_ *CreateSuccessorInput,
) (*model.Estimate, error) {
	return nil, apperrors.WrapFinanceConsumerPending("確定済み見積書の後継ドラフトは、提示原本の保存が未接続のため作成できません")
}
