package billing

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *accountingService) List(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, filters, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list accounting", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list accounting")
	}
	return result, total, nil
}

func (s *accountingService) ListForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAllForClinics(ctx, clinicIDs, filters, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list accounting for clinics", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list accounting for clinics")
	}
	return result, total, nil
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to get accounting")
	}
	return result, nil
}

func (s *accountingService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get accounting for clinics", "error", err)
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
	if input.Status == "" {
		input.Status = model.BillingStatusWaiting
	}
	// BUG-142: 金額バリデーション
	if input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	if input.Subtotal+input.TaxTotal != input.TotalAmount {
		return nil, apperrors.WrapInvalidInput("小計と税額の合計が請求合計と一致しません")
	}
	// BUG-013: blocking unbilled warning がある pet への会計作成は fail-closed（underbilling 防止）。
	if s.unbilledGuard != nil && input.PetID != nil {
		if err := s.unbilledGuard.AssertNoBlockingUnbilled(ctx, input.ClinicID, *input.PetID); err != nil {
			return nil, err
		}
	}
	// BUG-001: 死亡ペットへの新規会計作成は BE で物理ブロック（入院登録と同型）。
	if err := s.assertAccountingPetNotDeceased(ctx, input.ClinicID, input.PetID); err != nil {
		return nil, err
	}
	billing := &model.Billing{
		ClinicID:          input.ClinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		Status:            input.Status,
		ScheduledDate:     input.ScheduledDate,
		Memo:              input.Memo,
	}
	// AUD-002: 関連 FK 所有確認は write 前に実施。確定状態は Complete command のみが書く。
	if err := s.validateAccountingRelatedFKs(
		ctx, input.ClinicID,
		input.MedicalRecordID, input.HospitalizationID, input.OwnerID, input.PetID,
	); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, input.ClinicID, billing); err != nil {
		slog.ErrorContext(ctx, "failed to create accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to create accounting")
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

// resolvePaymentWrites は書込み前に method(ENUM)→payment_methods マスタ id を解決する
// （tx 外・低コストな読取のみ。BE-refactor.md E-4）。hasPaymentFields(input) が false の場合は
// (nil, nil, nil) を返す。
func (s *accountingService) resolvePaymentWrites(ctx context.Context, input *UpdateAccountingInput) (*model.Payment, []model.PaymentSplit, error) {
	if !hasPaymentFields(input) {
		return nil, nil, nil
	}
	systemKeyToID, err := s.loadPaymentMethodSystemKeyToID(ctx, input.ClinicID)
	if err != nil {
		return nil, nil, err // loadPaymentMethodSystemKeyToID 内で既に wrap + log 済み
	}
	payment := buildPaymentFromInput(input)
	// 代表支払方法も master id を併設（dual maintain）。method 未設定の更新（保険のみ等）は解決対象外。
	if payment.Method != "" {
		pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, systemKeyToID)
		if err != nil {
			return nil, nil, err
		}
		payment.PaymentMethodID = pid
	}
	splits := buildPaymentSplits(input)
	for i := range splits {
		pid, err := resolvePaymentMethodMasterID(splits[i].Method, splits[i].PaymentMethodID, systemKeyToID)
		if err != nil {
			return nil, nil, err
		}
		splits[i].PaymentMethodID = pid
	}
	return payment, splits, nil
}

func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	// #115 / B4: レジ締め済み期間の会計編集は理由必須。service 層を権威的 enforcement 点とし、
	// handler を迂回する呼び出し元にも不変条件を強制する。
	// 注: 認可（ユーザー権限）はリクエストスコープの関心事のため handler 側に残す（service 入力に actor 権限は持たせない）。
	if input.IsPostClose && (input.PostCloseReason == nil || *input.PostCloseReason == "") {
		return nil, apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}

	existing, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to find accounting")
	}
	if err := rejectPrivilegedGenericUpdate(existing, input); err != nil {
		return nil, err
	}
	// BUG-142: 金額バリデーション
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	// 混在会計バリデーション
	if err := validatePaymentSplits(input.PaymentSplits, input.BillingAmount); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate payment splits")
	}
	fields := buildAccountingUpdate(input)
	if len(fields) == 0 && !hasPaymentFields(input) {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}

	finalMRID, finalHospID, finalOwnerID, finalPetID := resolveFinalAccountingRelatedIDs(existing, input)

	// #128: 書込み前に method(ENUM)→payment_methods マスタ id を解決する（tx 外・低コストな読取のみ）。
	// レジ締め・月次集計は payment_method_id をキーにし NULL を現金とみなすため、
	// 非現金 split が NULL のまま保存されると全て現金に倒れる。解決失敗時はここで会計確定を止める。
	payment, splits, err := s.resolvePaymentWrites(ctx, input)
	if err != nil {
		return nil, err // resolvePaymentWrites 内で既に wrap + log 済み
	}

	// BE-refactor.md R1-2 (D1): Billing 本体更新・Payment upsert・締め後編集監査を単一 tx に統合する。
	// AUD-002: 最終関連 FK の所有・相互整合検証を write 前（同一 WithTx 内）で実施する。
	var accounting *model.Billing
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		reloaded, err := s.updateAccountingInTx(
			txCtx, input, existing, fields, finalMRID, finalHospID, finalOwnerID, finalPetID, payment, splits,
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

// resolvePostCloseInTx は write 時に締め状態を再評価する（handler 候補 read の TOCTOU を閉じる）。
// closeRepo 未配線時は handler フラグのみ信頼（ユニットテスト互換）。本番 DI では closeRepo 必須経路。
func (s *accountingService) resolvePostCloseInTx(ctx context.Context, clinicID uint64, scheduledDate time.Time, handlerFlag bool) (bool, error) {
	resolved, err := s.resolvePostCloseForDatesInTx(ctx, clinicID, handlerFlag, scheduledDate)
	if err != nil {
		return false, err
	}
	return resolved.anyClosed, nil
}

type postCloseDateResolution struct {
	anyClosed    bool
	sourceClosed bool
	destClosed   bool
	adjDate      time.Time
}

func (s *accountingService) resolvePostCloseForDatesInTx(ctx context.Context, clinicID uint64, handlerFlag bool, dates ...time.Time) (postCloseDateResolution, error) {
	unique := uniqueCloseBoundaryDates(dates...)
	if len(unique) == 0 {
		return postCloseDateResolution{anyClosed: handlerFlag}, nil
	}
	res := postCloseDateResolution{adjDate: unique[len(unique)-1]}
	if s.closeRepo == nil {
		res.anyClosed = handlerFlag
		return res, nil
	}
	if err := s.lockCloseBoundaries(ctx, clinicID, unique...); err != nil {
		return postCloseDateResolution{}, err
	}
	closedByKey := make(map[string]bool, len(unique))
	for _, date := range unique {
		closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, date)
		if err != nil {
			return postCloseDateResolution{}, apperrors.Wrap(err, "failed to re-check cash register close state")
		}
		closedByKey[closeBoundaryDayKey(date)] = closed
		if closed {
			res.anyClosed = true
			res.adjDate = date
		}
	}
	if len(dates) > 0 {
		res.sourceClosed = closedByKey[closeBoundaryDayKey(dates[0])]
	}
	if len(dates) > 1 {
		res.destClosed = closedByKey[closeBoundaryDayKey(dates[len(dates)-1])]
		if res.destClosed {
			res.adjDate = dates[len(dates)-1]
		} else if res.sourceClosed {
			res.adjDate = dates[0]
		}
	} else if len(dates) == 1 {
		res.destClosed = res.sourceClosed
	}
	if handlerFlag {
		res.anyClosed = true
	}
	return res, nil
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

func rejectLegacyCreatePrivilegedFinancialState(input *CreateAccountingInput) error {
	if input.CompletedAt != nil {
		return apperrors.WrapInvalidInput("completed_at はクライアントから指定できません。確定は POST /accountings/complete を使用してください")
	}
	switch input.Status {
	case "", model.BillingStatusWaiting, model.BillingStatusPending:
		return nil
	case model.BillingStatusCompleted:
		return apperrors.WrapInvalidInput("status=completed での作成はできません。確定は POST /accountings/complete を使用してください")
	case model.BillingStatusCancelled:
		return apperrors.WrapInvalidInput("status=cancelled での作成はできません。取消は POST /accountings/:id/cancel を使用してください")
	default:
		return apperrors.WrapInvalidInput("会計の作成は waiting または pending のみ許可されます")
	}
}

func rejectPrivilegedGenericUpdate(existing *model.Billing, input *UpdateAccountingInput) error {
	if existing == nil {
		return apperrors.WrapNotFound("billing", "")
	}
	if existing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("キャンセル済みの会計は更新できません")
	}
	if input.CompletedAt != nil {
		return apperrors.WrapInvalidInput("completed_at はクライアントから指定できません")
	}
	if input.Status != nil {
		switch *input.Status {
		case model.BillingStatusCancelled:
			return apperrors.WrapForbidden("status=cancelled への変更は POST /accountings/:id/cancel を使用してください")
		case model.BillingStatusCompleted:
			if existing.Status != model.BillingStatusCompleted {
				return apperrors.WrapInvalidInput("status=completed への変更は POST /accountings/complete を使用してください")
			}
			input.Status = nil
		default:
			if *input.Status != existing.Status {
				return apperrors.WrapInvalidInput("会計ステータスは汎用 PATCH では変更できません")
			}
			input.Status = nil
		}
	}
	if hasPaymentFields(input) && existing.Status != model.BillingStatusCompleted {
		return apperrors.WrapInvalidInput("支払いの確定は POST /accountings/complete を使用してください")
	}
	return nil
}

// writePostCloseAdjustment は締め後会計編集を cash_register_close_adjustments へ append-only 追記する（W-013）。
// Update の ambient tx に参加し、失敗時は呼び出し元が rollback する（fail-closed）。
// close 自体の reverse は行わず、当該日付に存在する close のいずれかに紐付ける。
func (s *accountingService) writePostCloseAdjustment(ctx context.Context, input *UpdateAccountingInput, existing *model.Billing) error {
	if existing == nil {
		return apperrors.WrapInternalServerError("existing billing is required for post-close adjustment")
	}
	reason := ""
	if input.PostCloseReason != nil {
		reason = *input.PostCloseReason
	}
	delta := int64(0)
	if input.TotalAmount != nil {
		delta = *input.TotalAmount - existing.TotalAmount
	}
	return s.recordPostCloseAdjustment(ctx, input.ClinicID, input.ID, existing.ScheduledDate, reason, input.StaffID, delta)
}

// recordPostCloseAdjustment は締め後訂正を cash_register_close_adjustments へ fail-closed で追記する。
// Update / CorrectCreditPayment / billing-item 経路から同一 tx で呼ぶ（W-013 HIGH-2）。
func (s *accountingService) recordPostCloseAdjustment(
	ctx context.Context,
	clinicID, billingID uint64,
	scheduledDate time.Time,
	reason string,
	actorID *uint64,
	accountingDelta int64,
) error {
	return createPostCloseAdjustment(ctx, s.closeRepo, clinicID, billingID, scheduledDate, reason, actorID, accountingDelta)
}

// createPostCloseAdjustment は close repo を使った append-only 台帳追記の package 共有実装。
func createPostCloseAdjustment(
	ctx context.Context,
	closeRepo CashRegisterCloseRepository,
	clinicID, billingID uint64,
	scheduledDate time.Time,
	reason string,
	actorID *uint64,
	accountingDelta int64,
) error {
	if closeRepo == nil {
		return apperrors.WrapInternalServerError("cash register close repository is required for post-close edits")
	}
	if strings.TrimSpace(reason) == "" {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}

	closeRec, err := findCloseForBillingDate(ctx, closeRepo, clinicID, scheduledDate)
	if err != nil {
		return err
	}

	adj := &model.CashRegisterCloseAdjustment{
		ClinicID:           clinicID,
		CloseID:            closeRec.ID,
		BillingID:          billingID,
		AccountingDelta:    accountingDelta,
		CashMovementAmount: 0, // 会計のみの訂正。現金移動は別経路（現状 productize なし）
		Reason:             reason,
		ActorID:            actorID,
		ExecutedAt:         time.Now(),
	}
	if err := closeRepo.CreateAdjustment(ctx, adj); err != nil {
		slog.ErrorContext(ctx, "failed to write post-close adjustment", "error", err,
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("billing_id", billingID),
			slog.Uint64("close_id", closeRec.ID))
		return apperrors.Wrap(err, "failed to write post-close cash register adjustment")
	}
	return nil
}

// findCloseForBillingDate は会計予定日に紐づく close を period 順（am→pm→emg）で解決する。
// 締め後編集ゲートは日付単位のため、いずれかの区分が締め済みならその close に adjustment を紐付ける。
func findCloseForBillingDate(ctx context.Context, closeRepo CashRegisterCloseRepository, clinicID uint64, billingDate time.Time) (*model.CashRegisterClose, error) {
	date := time.Date(billingDate.Year(), billingDate.Month(), billingDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, period := range []string{"am", "pm", "emg"} {
		c, err := closeRepo.FindByDateAndPeriod(ctx, clinicID, date, period)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve cash register close for post-close adjustment")
		}
		if c != nil {
			return c, nil
		}
	}
	return nil, apperrors.WrapConflict("締め後編集の対象となるレジ締めレコードが見つかりません")
}

// logPostCloseEdit はレジ締め済み期間の会計編集監査ログを記録する（#115 / B4）。
// BE-refactor.md R1-2: Update の ambient tx に参加する LogEntryTx を使う（fail-closed）。
// 呼び出し元は返されたエラーで tx をロールバックし、監査失敗時に編集自体も無効にする。
func (s *accountingService) logPostCloseEdit(ctx context.Context, input *UpdateAccountingInput) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for post-close edits")
	}
	billingID := input.ID
	aType := sharedkernel.AuditActorTypeFor(input.StaffID)
	meta := map[string]any{}
	if input.PostCloseReason != nil {
		meta["reason"] = *input.PostCloseReason
	}
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &input.ClinicID,
		ActorID:    input.StaffID,
		ActorType:  aType,
		Action:     model.AuditActionBillingPostCloseEdit,
		Resource:   "billing",
		ResourceID: &billingID,
		Metadata:   meta,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write post_close_edit audit log")
	}
	return nil
}

// loadPaymentMethodSystemKeyToID は当該 clinic の payment_methods マスタを system_key→id マップとして読み込む（#197）。
// system_key が NULL の行（クリニック独自追加の非標準支払方法）はスキップする。
func (s *accountingService) loadPaymentMethodSystemKeyToID(ctx context.Context, clinicID uint64) (map[string]uint64, error) {
	methods, err := s.payMethodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load payment methods", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load payment methods")
	}
	skToID := make(map[string]uint64, len(methods))
	for i := range methods {
		if methods[i].SystemKey != nil {
			skToID[*methods[i].SystemKey] = methods[i].ID
		}
	}
	return skToID, nil
}

func (s *accountingService) completeAccountingAppointments(ctx context.Context, clinicID uint64, billing *model.Billing) error {
	if s.reservationRepo == nil {
		return apperrors.WrapInternalServerError("reservation repository is not configured")
	}
	updated, err := s.reservationRepo.CompleteForAccounting(ctx, clinicID, billing.MedicalRecordID, billing.OwnerID, billing.PetID, billing.ScheduledDate)
	if err != nil {
		slog.ErrorContext(ctx, "failed to complete accounting appointments",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("billing_id", billing.ID),
			slog.String("error", err.Error()))
		return apperrors.Wrap(err, "failed to complete accounting appointments")
	}
	if updated > 0 {
		slog.InfoContext(ctx, "accounting appointments completed",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("billing_id", billing.ID),
			slog.Int64("updated_count", updated))
	}
	return nil
}

func (s *accountingService) syncCPMStageTag(ctx context.Context, clinicID uint64, billing *model.Billing) {
	if s.tagSyncSvc == nil || billing == nil || billing.OwnerID == nil {
		return
	}
	ownerID := *billing.OwnerID
	if err := s.tagSyncSvc.SyncCPMStageTag(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to sync CPM stage tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "billing_id", billing.ID)
	}
}
