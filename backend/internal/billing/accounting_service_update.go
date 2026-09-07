package billing

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *accountingService) updateAccountingInTx(
	txCtx context.Context,
	input *UpdateAccountingInput,
	existing *model.Billing,
	cmd AccountingUpdate,
	finalMRID, finalHospID, finalOwnerID, finalPetID *uint64,
	payment *model.Payment,
	splits []model.PaymentSplit,
) (*model.Billing, error) {
	destDate := existing.ScheduledDate
	if input.ScheduledDate != nil {
		destDate = *input.ScheduledDate
	}
	postCloseRes, err := s.resolvePostCloseForDatesInTx(txCtx, input.ClinicID, input.IsPostClose, existing.ScheduledDate, destDate)
	if err != nil {
		return nil, err
	}
	if postCloseRes.anyClosed {
		if input.PostCloseReason == nil || *input.PostCloseReason == "" {
			return nil, apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
		}
		input.IsPostClose = true
	}

	if err := s.validateAccountingRelatedFKs(txCtx, input.ClinicID, finalMRID, finalHospID, finalOwnerID, finalPetID); err != nil {
		return nil, err
	}
	if len(cmd.toFields()) > 0 {
		if _, err := s.repo.Update(txCtx, input.ClinicID, input.ID, cmd); err != nil {
			return nil, apperrors.Wrap(err, "failed to update accounting")
		}
	}

	if hasPaymentFields(input) {
		if err := s.repo.SavePayment(txCtx, payment); err != nil {
			return nil, apperrors.Wrap(err, "failed to upsert payment")
		}
		if err := s.repo.SavePaymentSplits(txCtx, splits); err != nil {
			return nil, apperrors.Wrap(err, "failed to save payment splits")
		}
		slog.InfoContext(txCtx, "payment upserted",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.ID))
	}

	if postCloseRes.anyClosed {
		adjExisting := *existing
		adjExisting.ScheduledDate = postCloseRes.adjDate
		if err := s.writePostCloseAdjustment(txCtx, input, &adjExisting); err != nil {
			return nil, err
		}
		if err := s.logPostCloseEdit(txCtx, input); err != nil {
			return nil, err
		}
	}

	accounting, err := s.repo.FindByID(txCtx, input.ClinicID, input.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload accounting after update")
	}
	return accounting, nil
}
