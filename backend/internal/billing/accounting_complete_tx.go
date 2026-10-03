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

	// EMR-253: takeover 対象を解決する（billing_id 明示 or hospitalization スロット占有者）。
	// 行を FOR UPDATE でロックした上で status を判定する。waiting 行は新規 INSERT せず
	// in-place で確定し、cancelled 占有者は部分 UNIQUE スロット解放のため同一 tx で soft-delete する。
	// この段階では行ロックと cancelled の解放だけで billing items / 未請求集約には触れない。
	bound, err := s.resolveCompleteTakeoverTarget(txCtx, input, medicalRecordID)
	if err != nil {
		return nil, err
	}

	// BUG-013 + EMR-196②: blocking unbilled と表示済み集約の版を同一 tx 内で再検証（TOCTOU 解消）。
	// 自分の billing_items 挿入後に再集計すると自分明細の分だけ集約が縮退するため、
	// billing header/items の作成前であるこの位置で照合する。
	// EMR-253: takeover では対象行が既存 billing で「表示した未請求集約」が前提にならないため、
	// revision 未送信時は blocking warning のみ検証する（送られた場合は従来どおり版照合も行う）。
	if s.unbilledGuard != nil && input.PetID != nil {
		if bound != nil && input.ExpectedUnbilledRevision == "" {
			if err := s.unbilledGuard.AssertNoBlockingUnbilled(txCtx, input.ClinicID, *input.PetID); err != nil {
				return nil, err
			}
		} else if err := s.unbilledGuard.AssertUnbilledForComplete(txCtx, input.ClinicID, *input.PetID, input.ExpectedUnbilledRevision); err != nil {
			return nil, err
		}
	}

	// BUG-001: 死亡ペットへの complete 確定を同一 tx 内で拒否（URL 直叩き経路の物理ブロック）。
	// takeover では request が pet_id を持たなくてもバインド済み行の pet で判定する。
	effectivePetID := input.PetID
	if effectivePetID == nil && bound != nil {
		effectivePetID = bound.PetID
	}
	if err := s.assertAccountingPetNotDeceased(txCtx, input.ClinicID, effectivePetID); err != nil {
		return nil, err
	}

	var billing *model.Billing
	var takeoverPrevTotal int64
	if bound == nil {
		// EMR-66: medical_record / hospitalization スロットが既存会計で占有済みなら INSERT の
		// UNIQUE 違反を待たずにここで 409 + 既存会計を返す。INSERT 失敗後に同 tx で再検索すると
		// PostgreSQL 25P02（aborted transaction）になり 500 へ落ちるため、必ず INSERT 前に検出する。
		if err := s.assertCompleteSlotAvailable(txCtx, input, medicalRecordID); err != nil {
			return nil, err
		}
		if billing, err = s.createCompleteBillingHeader(txCtx, input, digest, medicalRecordID); err != nil {
			return nil, err
		}
	} else {
		// takeover: 既存 waiting 行へ complete command を適用する（重複行は作らない）。
		// 締め後訂正の delta 計算用に takeover 前の total を退避する。
		takeoverPrevTotal = bound.TotalAmount
		if err := s.applyCompleteTakeoverHeader(txCtx, input, digest, bound); err != nil {
			return nil, err
		}
		// request items が確定内容の権威のため、既存明細（退院ケアプラン由来）は一括削除で置換する。
		if err := s.itemWriter.DeleteItemsForComplete(txCtx, input.ClinicID, bound.ID); err != nil {
			return nil, apperrors.Wrap(err, "failed to replace billing items for complete takeover")
		}
		billing = bound
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

	if err := s.writeCompletePostCloseIfNeeded(txCtx, input, billing, takeoverPrevTotal, totalAmount); err != nil {
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

// assertCompleteSlotAvailable は EMR-66: medical_record_id / hospitalization_id の
// 確定スロットが既存会計で占有済みかを INSERT 前に検出する。
// 占有時は ACCOUNTING_ALREADY_COMPLETED（409）と既存会計を返す。
func (s *accountingService) assertCompleteSlotAvailable(txCtx context.Context, input *CompleteAccountingInput, medicalRecordID *uint64) error {
	if medicalRecordID == nil && input.HospitalizationID == nil {
		return nil
	}
	blocking, err := s.repo.FindCompleteConflict(txCtx, input.ClinicID, medicalRecordID, input.HospitalizationID)
	if err != nil {
		return apperrors.Wrap(err, "failed to check accounting slot conflict")
	}
	if blocking == nil {
		return nil
	}
	return s.alreadyCompletedConflict(txCtx, input.ClinicID, blocking)
}

func (s *accountingService) createCompleteBillingHeader(
	txCtx context.Context,
	input *CompleteAccountingInput,
	digest string,
	medicalRecordID *uint64,
) (*model.Billing, error) {
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
		if !apperrors.IsAlreadyExists(err) && !persistence.IsUniqueConstraintErr(err) {
			return nil, apperrors.Wrap(err, "failed to create accounting header for complete")
		}
		// EMR-66: UNIQUE 衝突後に aborted tx 上で再検索すると 25P02 → 500 になる。
		// ここでは sentinel だけ返し、rollback 後の resolveCompleteUniqueConflict で
		// completion_request_id replay / slot 衝突 409 を再解決する。
		return nil, &completeUniqueConflictError{
			medicalRecordID:   medicalRecordID,
			hospitalizationID: input.HospitalizationID,
			cause:             err,
		}
	}
	return billing, nil
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
	// EMR-62: insurance_amount は正の magnitude（円）が正規契約（billing = total − insurance − discount）。
	// 旧クライアントの負値送信も絶対値として正規化して保存し、符号解釈ずれによる確定失敗を防ぐ。
	if insuranceAmount < 0 {
		insuranceAmount = -insuranceAmount
	}
	// 保険なし会計に保険負担額が非ゼロは矛盾レコードのため拒否する。
	if !input.HasInsurance && insuranceAmount != 0 {
		return apperrors.WrapInvalidInput("保険なし会計に保険負担額は設定できません")
	}
	discountAmount := int64(0)
	if input.DiscountAmount != nil {
		discountAmount = *input.DiscountAmount
	}
	if discountAmount < 0 {
		return apperrors.WrapInvalidInput("割引額は0円以上で指定してください")
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
	payment := buildPaymentFromInput(updateInput, nil)
	if payment.Method != "" {
		pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, systemKeyToID)
		if err != nil {
			return err
		}
		payment.PaymentMethodID = pid
	}
	splits := buildPaymentSplits(updateInput, billingAmount)
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
	updated, err := s.repo.Update(txCtx, input.ClinicID, billing.ID, AccountingUpdate{
		Status:      &completedStatus,
		CompletedAt: &now,
		Subtotal:    &subtotal,
		TaxTotal:    &taxTotal,
		TotalAmount: &totalAmount,
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
	prevTotalAmount int64,
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
	// prevTotalAmount: 新規 INSERT は 0、EMR-253 takeover は takeover 前の行 total。
	// close の紐付けは最終 scheduled_date（= input の値に takeover でも更新される）で解決する。
	existingForAdj := &model.Billing{ID: billing.ID, TotalAmount: prevTotalAmount, ScheduledDate: input.ScheduledDate}
	if err := s.writePostCloseAdjustment(txCtx, adjInput, existingForAdj); err != nil {
		return err
	}
	return s.logPostCloseEdit(txCtx, adjInput)
}

// resolveCompleteTakeoverTarget は EMR-253: complete command が新規行ではなく既存 waiting 行を
// 確定対象にする takeover を解決する。billing_id 明示時は当該行のみを対象にし、未指定で
// hospitalization_id 指定時は入院スロット占有者を対象にする。
// 返り値 nil は「takeover なし = 従来どおり新規 INSERT 経路」。
func (s *accountingService) resolveCompleteTakeoverTarget(
	txCtx context.Context,
	input *CompleteAccountingInput,
	medicalRecordID *uint64,
) (*model.Billing, error) {
	if input.BillingID != nil {
		return s.resolveExplicitCompleteTakeover(txCtx, input, medicalRecordID)
	}
	if input.HospitalizationID == nil {
		return nil, nil
	}
	return s.resolveHospitalizationCompleteTakeover(txCtx, input)
}

// resolveExplicitCompleteTakeover は billing_id で明示された行を FOR UPDATE でロックして
// takeover 可否を判定する。completed は親参照の一致/不一致に関わらず 409 を返し、
// waiting 行は親参照の不一致を InvalidInput で拒否する。
func (s *accountingService) resolveExplicitCompleteTakeover(
	txCtx context.Context,
	input *CompleteAccountingInput,
	medicalRecordID *uint64,
) (*model.Billing, error) {
	billing, err := s.repo.LockAndFindByID(txCtx, input.ClinicID, *input.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock billing for complete takeover")
	}
	switch billing.Status {
	case model.BillingStatusCompleted:
		return nil, s.alreadyCompletedConflict(txCtx, input.ClinicID, billing)
	case model.BillingStatusCancelled:
		return nil, apperrors.WrapConflict("キャンセル済みの会計は complete の takeover 対象にできません")
	}
	if billing.Status != model.BillingStatusWaiting {
		return nil, apperrors.WrapConflict(fmt.Sprintf("会計ステータス %s は complete の takeover 対象にできません", billing.Status))
	}
	if err := assertCompleteTakeoverInputConsistent(input, medicalRecordID, billing); err != nil {
		return nil, err
	}
	return billing, nil
}

// resolveHospitalizationCompleteTakeover は hospitalization スロット占有者を takeover 対象に解決する。
//   - 占有者なし → nil（新規 INSERT 経路へ）
//   - cancelled → SoftDeleteCancelled で部分 UNIQUE スロットを解放し nil（新規 INSERT 経路へ）
//   - waiting → FOR UPDATE でロックして status 再判定のうえ takeover 対象として返す
//   - その他（completed 等）→ ACCOUNTING_ALREADY_COMPLETED（従来の slot 409 と同型）
func (s *accountingService) resolveHospitalizationCompleteTakeover(
	txCtx context.Context,
	input *CompleteAccountingInput,
) (*model.Billing, error) {
	occupant, err := s.repo.FindByHospitalizationID(txCtx, input.ClinicID, *input.HospitalizationID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lookup hospitalization billing")
	}
	if occupant == nil {
		return nil, nil
	}
	if occupant.Status == model.BillingStatusCancelled {
		if err := s.repo.SoftDeleteCancelled(txCtx, input.ClinicID, occupant.ID); err != nil {
			return nil, apperrors.Wrap(err, "failed to release cancelled hospitalization billing")
		}
		return nil, nil
	}
	if occupant.Status != model.BillingStatusWaiting {
		return nil, s.alreadyCompletedConflict(txCtx, input.ClinicID, occupant)
	}
	// waiting: 検索→ロック間の status 遷移を閉じるため FOR UPDATE で取り直して再判定する。
	billing, err := s.repo.LockAndFindByID(txCtx, input.ClinicID, occupant.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock hospitalization billing for complete takeover")
	}
	switch billing.Status {
	case model.BillingStatusWaiting:
		return billing, nil
	case model.BillingStatusCancelled:
		if err := s.repo.SoftDeleteCancelled(txCtx, input.ClinicID, billing.ID); err != nil {
			return nil, apperrors.Wrap(err, "failed to release cancelled hospitalization billing")
		}
		return nil, nil
	default:
		return nil, s.alreadyCompletedConflict(txCtx, input.ClinicID, billing)
	}
}

// assertCompleteTakeoverInputConsistent は takeover 対象行と request の親参照が一致するかを検証する。
// 別カルテ/入院/owner/pet の waiting 行を誤って確定しないための fail-closed ガード。
func assertCompleteTakeoverInputConsistent(input *CompleteAccountingInput, medicalRecordID *uint64, billing *model.Billing) error {
	if !sameOptionalBillingReference(medicalRecordID, billing.MedicalRecordID) {
		return apperrors.WrapInvalidInput("billing_id の会計と medical_record_id が一致しません")
	}
	if !sameOptionalBillingReference(input.HospitalizationID, billing.HospitalizationID) {
		return apperrors.WrapInvalidInput("billing_id の会計と hospitalization_id が一致しません")
	}
	if !sameOptionalBillingReference(input.OwnerID, billing.OwnerID) {
		return apperrors.WrapInvalidInput("billing_id の会計と owner_id が一致しません")
	}
	if !sameOptionalBillingReference(input.PetID, billing.PetID) {
		return apperrors.WrapInvalidInput("billing_id の会計と pet_id が一致しません")
	}
	return nil
}

// applyCompleteTakeoverHeader は takeover 対象の既存 waiting 行へ complete command の
// ヘッダ値（予定日・メモ・保険フラグ・冪等キー）を in-place で適用する。
// 親参照列は assertCompleteTakeoverInputConsistent で一致保証済みのため更新しない。
// status / totals / completed_at は後続の persistCompletePayments が確定する。
func (s *accountingService) applyCompleteTakeoverHeader(
	txCtx context.Context,
	input *CompleteAccountingInput,
	digest string,
	billing *model.Billing,
) error {
	reqID := input.IdempotencyKey
	hash := digest
	update := AccountingUpdate{
		ScheduledDate:         &input.ScheduledDate,
		Memo:                  &input.Memo,
		HasInsurance:          &input.HasInsurance,
		CompletionRequestID:   &reqID,
		CompletionRequestHash: &hash,
	}
	if _, err := s.repo.Update(txCtx, input.ClinicID, billing.ID, update); err != nil {
		return apperrors.Wrap(err, "failed to apply complete takeover header")
	}
	return nil
}
