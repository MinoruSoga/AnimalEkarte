package billing

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *accountingService) List(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, filters, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list accounting")
	}
	return result, total, nil
}

func (s *accountingService) ListForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAllForClinics(ctx, clinicIDs, filters, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list accounting for clinics")
	}
	return result, total, nil
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get accounting")
	}
	return result, nil
}

func (s *accountingService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get accounting for clinics")
	}
	return result, nil
}

func (s *accountingService) Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error) {
	if input.ScheduledDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("scheduled_date is required")
	}
	if err := rejectLegacyCreatePrivilegedFinancialState(input); err != nil {
		return nil, err
	}
	// BUG-013: blocking unbilled warning がある pet への会計作成は fail-closed（underbilling 防止）。
	// held for EMR-481: 未請求候補の projection（billing_source_bindings）が未接続の間、AssertNoBlockingUnbilled は
	// 固定エラー（422）を返すため、pet に紐付く請求ヘッダは作成できない（C1。成功を装って通さない）。
	if s.unbilledGuard != nil && input.PetID != nil {
		if err := s.unbilledGuard.AssertNoBlockingUnbilled(ctx, input.ClinicID, *input.PetID); err != nil {
			return nil, err
		}
	}
	// BUG-001: 死亡ペットへの新規会計作成は BE で物理ブロック（入院登録と同型）。
	if err := s.assertAccountingPetNotDeceased(ctx, input.ClinicID, input.PetID); err != nil {
		return nil, err
	}
	// 作成は draft のヘッダだけを書く。請求状態は claim_state（NULL は状態不明として明細ガードが拒否する）で、
	// 合計・確定日時・保険有無・支払いはヘッダに持たず、算定固定（snapshot）の writer が接続されるまで保存しない。
	claimState := billingClaimStateDraft
	billing := &model.Billing{
		ClinicID:          input.ClinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		ScheduledDate:     input.ScheduledDate,
		Memo:              input.Memo,
		ClaimState:        &claimState,
	}
	// AUD-002: 関連 FK 所有確認は write 前に実施。確定状態は Complete command のみが書く。
	if err := s.validateAccountingRelatedFKs(
		ctx, input.ClinicID,
		input.MedicalRecordID, input.HospitalizationID, input.OwnerID, input.PetID,
	); err != nil {
		return nil, err
	}
	createFn := func(txCtx context.Context) error {
		if err := s.lockCloseBoundaries(txCtx, input.ClinicID, input.ScheduledDate); err != nil {
			return err
		}
		return s.repo.Create(txCtx, input.ClinicID, billing)
	}
	var err error
	if s.transactor != nil {
		err = s.transactor.WithTx(ctx, createFn)
	} else {
		err = createFn(ctx)
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create accounting")
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

// Update は会計ヘッダ（関連 ID・予定日・メモ）の部分更新。
// 金額・状態・支払いは PATCH で書かない: 金額・状態・確定日時・保険有無は 400、支払い・保険・割引は固定エラー（422）で、
// 読取より先に拒否する。held for EMR-506（受領と配賦の writer）/ EMR-509（締め後の変更記録）。
func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	if err := rejectUnsupportedUpdateFields(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find accounting")
	}
	if err := rejectHeaderUpdateUnlessDraft(existing); err != nil {
		return nil, err
	}
	cmd := accountingUpdateFromInput(input)
	if len(cmd.toFields()) == 0 {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}

	finalMRID, finalHospID, finalOwnerID, finalPetID := resolveFinalAccountingRelatedIDs(existing, input)

	// BE-refactor.md R1-2 (D1): ヘッダ更新を単一 tx に統合する。
	// AUD-002: 最終関連 FK の所有・相互整合検証を write 前（同一 WithTx 内）で実施する。
	var accounting *model.Billing
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		reloaded, err := s.updateAccountingInTx(
			txCtx, input, existing, cmd, finalMRID, finalHospID, finalOwnerID, finalPetID,
		)
		if err != nil {
			return err
		}
		accounting = reloaded
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update accounting in transaction")
	}

	slog.InfoContext(ctx, "accounting updated",
		slog.Uint64("billing_id", accounting.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return accounting, nil
}

// anyDateClosedInTx は write 時に各日付の締め状態を再評価する（handler 候補 read の TOCTOU を閉じる）。
// closeRepo 未配線時は handler フラグのみ信頼（ユニットテスト互換）。本番 DI では closeRepo 必須経路。
func (s *accountingService) anyDateClosedInTx(ctx context.Context, clinicID uint64, handlerFlag bool, dates ...time.Time) (bool, error) {
	unique := uniqueCloseBoundaryDates(dates...)
	if len(unique) == 0 || s.closeRepo == nil {
		return handlerFlag, nil
	}
	if err := s.lockCloseBoundaries(ctx, clinicID, unique...); err != nil {
		return false, err
	}
	anyClosed := handlerFlag
	for _, date := range unique {
		closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, date)
		if err != nil {
			return false, apperrors.Wrap(err, "failed to re-check cash register close state")
		}
		if closed {
			anyClosed = true
		}
	}
	return anyClosed, nil
}

func (s *accountingService) lockCloseBoundaries(ctx context.Context, clinicID uint64, dates ...time.Time) error {
	if s.closeRepo == nil {
		return nil
	}
	for _, date := range uniqueCloseBoundaryDates(dates...) {
		if err := s.closeRepo.LockCloseBoundary(ctx, clinicID, date); err != nil {
			return err
		}
	}
	return nil
}

func closeBoundaryDayKey(date time.Time) string {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
}

func uniqueCloseBoundaryDates(dates ...time.Time) []time.Time {
	seen := make(map[string]time.Time, len(dates))
	for _, date := range dates {
		if date.IsZero() {
			continue
		}
		key := closeBoundaryDayKey(date)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]time.Time, 0, len(keys))
	for _, key := range keys {
		out = append(out, seen[key])
	}
	return out
}

// rejectLegacyCreatePrivilegedFinancialState は作成リクエストに含まれる金額・状態の項目を拒否する（Q4(c)）。
// 作成は draft のヘッダ（関連 ID・予定日・メモ）だけを受け付ける。合計・確定日時・保険有無・請求状態は
// server が算定固定と確定 command で決めるため、クライアント供給値は無視せず 400 で返す。held for EMR-506。
func rejectLegacyCreatePrivilegedFinancialState(input *CreateAccountingInput) error {
	if input.Subtotal != 0 || input.TaxTotal != 0 || input.TotalAmount != 0 {
		return apperrors.WrapInvalidInput("小計・税額・請求合計はクライアントから指定できません。合計は server が算定します")
	}
	if input.HasInsurance || input.CompletedAt != nil {
		return apperrors.WrapInvalidInput("保険有無・確定日時はクライアントから指定できません。確定は POST /accountings/complete を使用してください")
	}
	if input.Status != "" {
		return apperrors.WrapInvalidInput("会計の状態はクライアントから指定できません。作成される会計は draft です")
	}
	return nil
}

// rejectUnsupportedUpdateFields は PATCH に含まれる金額・状態・支払いの項目を拒否する。
// 金額・状態・確定日時・保険有無は server が算定固定と確定 command で決めるため 400。
// 支払い・保険・割引は受領と配賦の writer が未接続のため固定エラー（422）。held for EMR-506。
func rejectUnsupportedUpdateFields(input *UpdateAccountingInput) error {
	if input.Subtotal != nil || input.TaxTotal != nil || input.TotalAmount != nil || input.HasInsurance != nil || input.CompletedAt != nil {
		return apperrors.WrapInvalidInput("小計・税額・請求合計・保険有無・確定日時はクライアントから指定できません")
	}
	if input.Status != nil {
		return apperrors.WrapInvalidInput("会計ステータスは汎用 PATCH では変更できません")
	}
	if hasPaymentFields(input) {
		return apperrors.WrapFinanceConsumerPending("会計の支払い・保険・割引の更新は、受領と配賦の保存先が未接続のため利用できません")
	}
	return nil
}

// rejectHeaderUpdateUnlessDraft はヘッダ更新を draft の会計に限る。
// 確定済み（finalized）の変更は訂正の変更イベントと新しい算定版（correct）が必要で未接続のため固定エラー（422）。
// cancelled と状態不明（claim_state が NULL・未知）は Conflict。
func rejectHeaderUpdateUnlessDraft(existing *model.Billing) error {
	if existing == nil {
		return apperrors.WrapNotFound("billing", "")
	}
	switch billingClaimState(existing) {
	case billingClaimStateDraft:
		return nil
	case billingClaimStateFinalized:
		return apperrors.WrapFinanceConsumerPending("確定済み会計のヘッダ変更は、訂正の変更イベントと新しい算定版（correct）が未接続のため利用できません")
	case billingClaimStateCancelled:
		return apperrors.WrapConflict("キャンセル済みの会計は更新できません")
	default:
		return apperrors.WrapConflict("会計の請求状態（claim_state）を確認できないため、更新できません")
	}
}
