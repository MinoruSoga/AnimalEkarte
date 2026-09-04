package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (s *accountingService) completeInTx(
	txCtx context.Context,
	input *CompleteAccountingInput,
	digest string,
	systemKeyToID map[string]uint64,
) (*CompleteAccountingResult, error) {
	replay, err := s.replayCompleteIfExisting(txCtx, input.ClinicID, input.IdempotencyKey, digest)
	if err != nil {
		return nil, err
	}
	if replay != nil {
		return replay, nil
	}

	// BUG-004: 締め後理由は FK 解決より先。締め済みなのに参照組み合わせエラーだけ出ると確定導線が消える。
	postClose, err := s.resolvePostCloseInTx(txCtx, input.ClinicID, input.ScheduledDate, input.IsPostClose)
	if err != nil {
		return nil, err
	}
	if postClose {
		if input.PostCloseReason == nil || strings.TrimSpace(*input.PostCloseReason) == "" {
			return nil, apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
		}
		input.IsPostClose = true
	}

	// BUG-011: treatment 付き明細があるとき billing.medical_record_id が必須。
	// FE が未送信でも treatment から一意に解決する（明示値は優先・不一致は拒否）。
	medicalRecordID, err := resolveCompleteMedicalRecordID(txCtx, input.ClinicID, input.MedicalRecordID, input.Items)
	if err != nil {
		return nil, err
	}

	if err := s.validateAccountingRelatedFKs(
		txCtx, input.ClinicID,
		medicalRecordID, input.HospitalizationID, input.OwnerID, input.PetID,
	); err != nil {
		return nil, err
	}

	// BUG-013: blocking unbilled を同一 tx 内で再検証（TOCTOU 解消）。
	if s.unbilledGuard != nil && input.PetID != nil {
		if err := s.unbilledGuard.AssertNoBlockingUnbilled(txCtx, input.ClinicID, *input.PetID); err != nil {
			return nil, err
		}
	}

	// BUG-001: 死亡ペットへの complete 確定を同一 tx 内で拒否（URL 直叩き経路の物理ブロック）。
	if err := s.assertAccountingPetNotDeceased(txCtx, input.ClinicID, input.PetID); err != nil {
		return nil, err
	}

	billing, replay, err := s.createCompleteBillingHeader(txCtx, input, digest, medicalRecordID)
	if err != nil {
		return nil, err
	}
	if replay != nil {
		return replay, nil
	}

	// Items: ambient tx 参加。N 番目失敗で全 rollback。
	if err := s.createCompleteItems(txCtx, input, billing.ID); err != nil {
		return nil, err
	}

	subtotal, taxTotal, totalAmount, err := s.totalsWriter.RecalculateTotalsForComplete(txCtx, input.ClinicID, billing.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to recalculate complete totals")
	}

	if err := s.persistCompletePayments(txCtx, input, billing, systemKeyToID, subtotal, taxTotal, totalAmount); err != nil {
		return nil, err
	}

	if err := s.writeCompletePostCloseIfNeeded(txCtx, input, billing, totalAmount); err != nil {
		return nil, err
	}

	// Reload before commit so response failure rolls back the whole complete.
	reloaded, err := s.repo.FindByID(txCtx, input.ClinicID, billing.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload accounting after complete")
	}
	return &CompleteAccountingResult{Accounting: reloaded, Created: true}, nil
}

func (s *accountingService) replayCompleteIfExisting(
	txCtx context.Context,
	clinicID uint64,
	idempotencyKey, digest string,
) (*CompleteAccountingResult, error) {
	existing, err := s.repo.FindByCompletionRequestID(txCtx, clinicID, idempotencyKey)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to re-lookup completion request in tx")
	}
	if existing == nil {
		return nil, nil
	}
	return s.resolveIdempotentReplay(txCtx, clinicID, existing, digest)
}

func (s *accountingService) createCompleteBillingHeader(
	txCtx context.Context,
	input *CompleteAccountingInput,
	digest string,
	medicalRecordID *uint64,
) (*model.Billing, *CompleteAccountingResult, error) {
	reqID := input.IdempotencyKey
	hash := digest
	billing := &model.Billing{
		ClinicID:              input.ClinicID,
		MedicalRecordID:       medicalRecordID,
		HospitalizationID:     input.HospitalizationID,
		OwnerID:               input.OwnerID,
		PetID:                 input.PetID,
		Subtotal:              0,
		TaxTotal:              0,
		TotalAmount:           0,
		HasInsurance:          input.HasInsurance,
		Status:                model.BillingStatusWaiting,
		ScheduledDate:         input.ScheduledDate,
		Memo:                  input.Memo,
		CompletionRequestID:   &reqID,
		CompletionRequestHash: &hash,
	}
	if err := s.repo.Create(txCtx, input.ClinicID, billing); err != nil {
		// Create は UNIQUE を AlreadyExists に変換する（pg 23505 は chain されない）。
		// completion_request_id 衝突時のみ replay。他 UNIQUE は existing==nil のまま元エラー。
		if !apperrors.IsAlreadyExists(err) && !persistence.IsUniqueConstraintErr(err) {
			return nil, nil, apperrors.Wrap(err, "failed to create accounting header for complete")
		}
		existing, lookupErr := s.repo.FindByCompletionRequestID(txCtx, input.ClinicID, input.IdempotencyKey)
		if lookupErr != nil {
			return nil, nil, apperrors.Wrap(lookupErr, "failed to resolve completion unique conflict")
		}
		if existing == nil {
			return nil, nil, apperrors.Wrap(err, "failed to create accounting header for complete")
		}
		replay, replayErr := s.resolveIdempotentReplay(txCtx, input.ClinicID, existing, digest)
		if replayErr != nil {
			return nil, nil, replayErr
		}
		return nil, replay, nil
	}
	return billing, nil, nil
}

func (s *accountingService) createCompleteItems(txCtx context.Context, input *CompleteAccountingInput, billingID uint64) error {
	for i := range input.Items {
		it := input.Items[i]
		itemInput := &CreateBillingItemInput{
			ClinicID:              input.ClinicID,
			BillingID:             billingID,
			Category:              it.Category,
			Name:                  it.Name,
			UnitPrice:             it.UnitPrice,
			Quantity:              it.Quantity,
			DiscountRate:          it.DiscountRate,
			DiscountAmount:        it.DiscountAmount,
			TaxType:               it.TaxType,
			TaxRate:               it.TaxRate,
			IsInsuranceApplicable: it.IsInsuranceApplicable,
			Source:                it.Source,
			OtherReason:           it.OtherReason,
			MerchandiseItemID:     it.MerchandiseItemID,
			TreatmentID:           it.TreatmentID,
			VaccinationID:         it.VaccinationID,
			ExamID:                it.ExamID,
			AppointmentID:         it.AppointmentID,
			TrimmingCourseID:      it.TrimmingCourseID,
			TrimmingOptionID:      it.TrimmingOptionID,
			SortOrder:             it.SortOrder,
			StaffID:               input.StaffID,
			CreatedBy:             input.StaffID,
		}
		if _, err := s.itemWriter.CreateItemForComplete(txCtx, itemInput); err != nil {
			return apperrors.Wrap(err, fmt.Sprintf("failed to create complete item index=%d", i))
		}
	}
	return nil
}

func (s *accountingService) persistCompletePayments(
	txCtx context.Context,
	input *CompleteAccountingInput,
	billing *model.Billing,
	systemKeyToID map[string]uint64,
	subtotal, taxTotal, totalAmount int64,
) error {
	insuranceAmount := int64(0)
	if input.InsuranceAmount != nil {
		insuranceAmount = *input.InsuranceAmount
	}
	discountAmount := int64(0)
	if input.DiscountAmount != nil {
		discountAmount = *input.DiscountAmount
	}
	billingAmount := totalAmount - insuranceAmount - discountAmount
	if billingAmount < 0 {
		return apperrors.WrapInvalidInput("請求金額が負になります（保険・割引の指定を確認してください）")
	}
	// BUG-006: 請求額が正なのに内訳未指定だと buildPaymentSplits が全額1行を合成し、
	// 部分入金が 201 で黙って上書きされる。UI は remaining!==0 で到達不可だが API 契約は 400。
	if billingAmount > 0 && len(input.PaymentSplits) == 0 {
		return apperrors.WrapInvalidInput("支払い内訳は必須です")
	}
	if err := validatePaymentSplits(input.PaymentSplits, &billingAmount); err != nil {
		return err
	}

	updateInput := &UpdateAccountingInput{
		ID:              billing.ID,
		ClinicID:        input.ClinicID,
		StaffID:         input.StaffID,
		Subtotal:        &subtotal,
		TaxTotal:        &taxTotal,
		TotalAmount:     &totalAmount,
		InsuranceRatio:  input.InsuranceRatio,
		InsuranceName:   input.InsuranceName,
		InsuranceAmount: &insuranceAmount,
		DiscountAmount:  &discountAmount,
		BillingAmount:   &billingAmount,
		PaymentSplits:   input.PaymentSplits,
		PostCloseReason: input.PostCloseReason,
		IsPostClose:     input.IsPostClose,
	}
	if len(input.PaymentSplits) > 0 {
		method := representativeMethod(input.PaymentSplits)
		updateInput.PaymentMethod = &method
	}
	payment := buildPaymentFromInput(updateInput)
	if payment.Method != "" {
		pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, systemKeyToID)
		if err != nil {
			return err
		}
		payment.PaymentMethodID = pid
	}
	splits := buildPaymentSplits(updateInput)
	for i := range splits {
		pid, err := resolvePaymentMethodMasterID(splits[i].Method, splits[i].PaymentMethodID, systemKeyToID)
		if err != nil {
			return err
		}
		splits[i].PaymentMethodID = pid
	}
	if err := s.repo.SavePayment(txCtx, payment); err != nil {
		return apperrors.Wrap(err, "failed to save payment for complete")
	}
	if err := s.repo.SavePaymentSplits(txCtx, splits); err != nil {
		return apperrors.Wrap(err, "failed to save payment splits for complete")
	}

	now := time.Now()
	completedStatus := model.BillingStatusCompleted
	updated, err := s.repo.Update(txCtx, input.ClinicID, billing.ID, map[string]any{
		"status":       completedStatus,
		"completed_at": now,
		"subtotal":     subtotal,
		"tax_total":    taxTotal,
		"total_amount": totalAmount,
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to mark accounting completed")
	}
	if err := s.completeAccountingAppointments(txCtx, input.ClinicID, updated); err != nil {
		return apperrors.Wrap(err, "failed to complete accounting appointments during complete")
	}
	return nil
}

func (s *accountingService) writeCompletePostCloseIfNeeded(
	txCtx context.Context,
	input *CompleteAccountingInput,
	billing *model.Billing,
	totalAmount int64,
) error {
	if !input.IsPostClose {
		return nil
	}
	adjInput := &UpdateAccountingInput{
		ID:              billing.ID,
		ClinicID:        input.ClinicID,
		StaffID:         input.StaffID,
		TotalAmount:     &totalAmount,
		PostCloseReason: input.PostCloseReason,
		IsPostClose:     true,
	}
	existingForAdj := &model.Billing{ID: billing.ID, TotalAmount: 0, ScheduledDate: input.ScheduledDate}
	if err := s.writePostCloseAdjustment(txCtx, adjInput, existingForAdj); err != nil {
		return err
	}
	return s.logPostCloseEdit(txCtx, adjInput)
}
