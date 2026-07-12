package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *accountingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list accounting", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list accounting")
	}
	return result, total, nil
}

func (s *accountingService) ListForClinics(ctx context.Context, clinicIDs []uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAllForClinics(ctx, clinicIDs, petID, ownerID, status, startDate, endDate, page, limit)
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
	// BUG-142: 金額バリデーション
	if input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.Subtotal+input.TaxTotal != input.TotalAmount {
		return nil, apperrors.WrapInvalidInput("小計と税額の合計が請求合計と一致しません")
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
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	}
	// BE-refactor.md X-12: 会計完了(completed)での Create は billing 作成と appointment 完了化を
	// 単一 tx に統合する。従来は repo.Create のコミット後に tx 外で completeAccountingAppointments
	// を呼んでおり、後者のみ失敗すると billing が確定済みのまま残った（部分コミット）。
	// completed でない Create（waiting 等）は appointment 完了化を伴わないため従来どおり tx 不要。
	createBilling := func(cctx context.Context) error {
		if err := s.repo.Create(cctx, input.ClinicID, billing); err != nil {
			slog.ErrorContext(cctx, "failed to create accounting", "error", err)
			return apperrors.Wrap(err, "failed to create accounting")
		}
		return nil
	}
	if billing.Status == model.BillingStatusCompleted {
		if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
			if err := createBilling(txCtx); err != nil {
				return err
			}
			if err := s.completeAccountingAppointments(txCtx, input.ClinicID, billing); err != nil {
				return apperrors.Wrap(err, "failed to complete accounting appointments during create")
			}
			return nil
		}); err != nil {
			return nil, err //nolint:wrapcheck // 閉包内で文脈付き wrap 済み（"in transaction" の同義二重ラップを廃止）
		}
	} else if err := createBilling(ctx); err != nil {
		return nil, err //nolint:wrapcheck // createBilling 内で wrap 済み
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	if billing.Status == model.BillingStatusCompleted {
		s.syncCPMStageTag(ctx, input.ClinicID, billing)
	}
	return billing, nil
}

func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	// #115 / B4: レジ締め済み期間の会計編集は理由必須。service 層を権威的 enforcement 点とし、
	// handler を迂回する呼び出し元にも不変条件を強制する。
	// 注: 認可（ユーザー権限）はリクエストスコープの関心事のため handler 側に残す（service 入力に actor 権限は持たせない）。
	if input.IsPostClose && (input.PostCloseReason == nil || *input.PostCloseReason == "") {
		return nil, apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}

	_, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to find accounting")
	}
	// BUG-142: 金額バリデーション
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	// 混在会計バリデーション
	if err := validatePaymentSplits(input.PaymentSplits, input.BillingAmount); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate payment splits")
	}
	fields := buildAccountingUpdate(input)
	if len(fields) == 0 && !hasPaymentFields(input) {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}

	// #128: 書込み前に method(ENUM)→payment_methods マスタ id を解決する（tx 外・低コストな読取のみ）。
	// レジ締め・月次集計は payment_method_id をキーにし NULL を現金とみなすため、
	// 非現金 split が NULL のまま保存されると全て現金に倒れる。解決失敗時はここで会計確定を止める。
	var payment *model.Payment
	var splits []model.PaymentSplit
	if hasPaymentFields(input) {
		systemKeyToID, err := s.loadPaymentMethodSystemKeyToID(ctx, input.ClinicID)
		if err != nil {
			return nil, err // loadPaymentMethodSystemKeyToID 内で既に wrap + log 済み
		}
		payment = buildPaymentFromInput(input)
		// 代表支払方法も master id を併設（dual maintain）。method 未設定の更新（保険のみ等）は解決対象外。
		if payment.Method != "" {
			pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, systemKeyToID)
			if err != nil {
				return nil, err
			}
			payment.PaymentMethodID = pid
		}
		splits = buildPaymentSplits(input)
		for i := range splits {
			pid, err := resolvePaymentMethodMasterID(splits[i].Method, splits[i].PaymentMethodID, systemKeyToID)
			if err != nil {
				return nil, err
			}
			splits[i].PaymentMethodID = pid
		}
	}

	// BE-refactor.md R1-2 (D1): Billing 本体更新・Payment upsert・締め後編集監査を単一 tx に統合する。
	// 従来は fields 更新が tx 外、payment upsert が独立 tx、監査が tx 外 best-effort の三系統に分かれており、
	// 途中失敗時に部分コミット（例: fields 更新は成功したが payment upsert のみ失敗）が起こり得た。
	// 統合後は「本体書込（fields/payment）と締め後編集監査が原子」になる（refund/CorrectCreditPayment と同型）。
	// BE-refactor.md X-12: fields が status=completed を含む場合（buildAccountingUpdate は
	// input.Status != nil を必ず fields["status"] に反映するため、この分岐に入るなら fields は
	// 必ず非空 = s.repo.Update が必ず呼ばれる）、tx 内で得られる updatedBilling を使って
	// appointment 完了化も同一 tx に含める。従来は WithTx コミット後・tx 外の ctx で
	// completeAccountingAppointments を呼んでおり、これのみ失敗すると billing は completed で
	// 確定済みのまま部分コミットになった。
	var updatedBilling *model.Billing
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Billing 本体の更新
		if len(fields) > 0 {
			b, err := s.repo.Update(txCtx, input.ClinicID, input.ID, fields)
			if err != nil {
				slog.ErrorContext(txCtx, "failed to update accounting", "error", err)
				return apperrors.Wrap(err, "failed to update accounting")
			}
			updatedBilling = b
		}

		// Payment upsert（支払フィールドが含まれている場合）
		if hasPaymentFields(input) {
			if err := s.repo.SavePayment(txCtx, payment); err != nil {
				slog.ErrorContext(txCtx, "failed to upsert payment", "error", err)
				return apperrors.Wrap(err, "failed to upsert payment")
			}
			// payment_splits の更新（混在会計・backward compat 両対応）
			if err := s.repo.SavePaymentSplits(txCtx, splits); err != nil {
				slog.ErrorContext(txCtx, "failed to save payment splits", "error", err)
				return apperrors.Wrap(err, "failed to save payment splits")
			}
			slog.InfoContext(txCtx, "payment upserted",
				slog.Uint64("clinic_id", input.ClinicID),
				slog.Uint64("billing_id", input.ID))
		}

		// #115: 締め後編集監査ログ（fail-closed: 失敗時は tx をロールバックし更新ごと無効にする。BE-refactor.md R1-2）
		if input.IsPostClose {
			if err := s.logPostCloseEdit(txCtx, input); err != nil {
				return err
			}
		}

		// X-12: 判定は再読込後の accounting ではなく input.Status ベース（tx 内では fields 適用済みのため同値）。
		if input.Status != nil && *input.Status == model.BillingStatusCompleted {
			if err := s.completeAccountingAppointments(txCtx, input.ClinicID, updatedBilling); err != nil {
				return apperrors.Wrap(err, "failed to complete accounting appointments during update")
			}
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update accounting in transaction")
	}

	// 更新後のレコードを返す
	accounting, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload accounting after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to reload accounting after update")
	}

	slog.InfoContext(ctx, "accounting updated",
		slog.Uint64("billing_id", accounting.ID),
		slog.Uint64("clinic_id", input.ClinicID))

	// syncCPMStageTag は外部 LSTEP 同期のため従来どおり tx 外 best-effort を維持する（X-12 対応方針）。
	if input.Status != nil && *input.Status == model.BillingStatusCompleted {
		s.syncCPMStageTag(ctx, input.ClinicID, accounting)
	}
	return accounting, nil
}

// logPostCloseEdit はレジ締め済み期間の会計編集監査ログを記録する（#115 / B4）。
// BE-refactor.md R1-2: Update の ambient tx に参加する LogEntryTx を使う（fail-closed）。
// 呼び出し元は返されたエラーで tx をロールバックし、監査失敗時に編集自体も無効にする。
func (s *accountingService) logPostCloseEdit(ctx context.Context, input *UpdateAccountingInput) error {
	if s.auditTx == nil {
		return nil
	}
	billingID := input.ID
	aType := auditActorTypeFor(input.StaffID)
	meta := map[string]any{}
	if input.PostCloseReason != nil {
		meta["reason"] = *input.PostCloseReason
	}
	if err := s.auditTx.LogEntryTx(ctx, &AuditLogInput{
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
	updated, err := s.repo.CompleteAccountingAppointments(ctx, clinicID, billing.MedicalRecordID, billing.OwnerID, billing.PetID, billing.ScheduledDate)
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
