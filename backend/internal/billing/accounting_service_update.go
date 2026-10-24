package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *accountingService) updateAccountingInTx(
	txCtx context.Context,
	input *UpdateAccountingInput,
	existing *model.Billing,
	cmd AccountingUpdate,
	finalMRID, finalHospID, finalOwnerID, finalPetID *uint64,
) (*model.Billing, error) {
	destDate := existing.ScheduledDate
	if input.ScheduledDate != nil {
		destDate = *input.ScheduledDate
	}
	anyClosed, err := s.anyDateClosedInTx(txCtx, input.ClinicID, input.IsPostClose, existing.ScheduledDate, destDate)
	if err != nil {
		return nil, err
	}
	if anyClosed {
		// held for EMR-509: 締め後の台帳追記（cash_register_close_adjustments）は退役し、後継の遅延イベント・実査訂正は未接続のため、
		// 締め済み期間の会計ヘッダ変更は書き込み前に固定エラー（422）で拒否する。
		return nil, apperrors.WrapFinanceConsumerPending("レジ締め済み期間の会計の変更は、締め後の変更記録（遅延イベント・実査訂正）が未接続のため利用できません")
	}

	if err := s.validateAccountingRelatedFKs(txCtx, input.ClinicID, finalMRID, finalHospID, finalOwnerID, finalPetID); err != nil {
		return nil, err
	}
	if len(cmd.toFields()) > 0 {
		if _, err := s.repo.Update(txCtx, input.ClinicID, input.ID, cmd); err != nil {
			return nil, apperrors.Wrap(err, "failed to update accounting")
		}
	}

	accounting, err := s.repo.FindByID(txCtx, input.ClinicID, input.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload accounting after update")
	}
	return accounting, nil
}
