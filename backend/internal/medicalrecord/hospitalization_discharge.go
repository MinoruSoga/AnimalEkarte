package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// DischargeWithBillingInput は退院+会計作成の入力DTO
type DischargeWithBillingInput struct {
	DischargeDate    time.Time
	CreateAccounting bool
	ActorID          *uint64
}

// DischargeWithBillingResult は退院+会計作成のレスポンスDTO
type DischargeWithBillingResult struct {
	HospitalizationID uint64
	AccountingID      *uint64
	Status            string
}

// DischargeWithBilling is the discharge entry point (EMR-499 U1 movement side).
//
// create_accounting=true is refused first with stay_discharge_billing_unavailable (422): the billing builder
// is owned by EMR-505/506/508 and nothing may be written before it exists. The movement side reads the stay's
// effective state under the hospitalization row lock; see dischargeWithBillingInTx for what it refuses.
// There is no read of, and no write to, the removed hospitalizations.status / end_date.
func (s *hospitalizationService) DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
	if input.CreateAccounting {
		return nil, apperrors.WrapUnprocessableWithCode(codeStayDischargeBillingUnavailable,
			"退院と同時の会計作成は、会計明細の再設計（EMR-505/506/508）まで利用できません")
	}
	result := &DischargeWithBillingResult{HospitalizationID: id}
	// MRB-02: same nil-transactor guard as Create/Update (panic vs explicit 500).
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return s.dischargeWithBillingInTx(txCtx, clinicID, id, input, result)
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to discharge hospitalization")
	}
	slog.InfoContext(ctx, "hospitalization discharged",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}
