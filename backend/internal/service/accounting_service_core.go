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
	if err := s.repo.Create(ctx, input.ClinicID, billing); err != nil {
		slog.ErrorContext(ctx, "failed to create accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to create accounting")
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	if billing.Status == model.BillingStatusCompleted {
		if err := s.completeAccountingAppointments(ctx, input.ClinicID, billing); err != nil {
			return nil, apperrors.Wrap(err, "failed to complete accounting appointments during create")
		}
		s.syncCPMStageTag(ctx, input.ClinicID, billing)
	}
	return billing, nil
}

func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
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

	// Billing 本体の更新
	if len(fields) > 0 {
		if _, err := s.repo.Update(ctx, input.ClinicID, input.ID, fields); err != nil {
			slog.ErrorContext(ctx, "failed to update accounting", "error", err)
			return nil, apperrors.Wrap(err, "failed to update accounting")
		}
	}

	// Payment upsert（支払フィールドが含まれている場合）— トランザクション内で SavePayment + SavePaymentSplits を実行
	if hasPaymentFields(input) {
		// #128: 書込み前に method(ENUM)→payment_methods マスタ id を解決する。
		// レジ締め・月次集計は payment_method_id をキーにし NULL を現金とみなすため、
		// 非現金 split が NULL のまま保存されると全て現金に倒れる。解決失敗時はここで会計確定を止める。
		nameToID, err := s.loadPaymentMethodNameToID(ctx, input.ClinicID)
		if err != nil {
			return nil, err // loadPaymentMethodNameToID 内で既に wrap + log 済み
		}
		payment := buildPaymentFromInput(input)
		// 代表支払方法も master id を併設（dual maintain）。method 未設定の更新（保険のみ等）は解決対象外。
		if payment.Method != "" {
			pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, nameToID)
			if err != nil {
				return nil, err
			}
			payment.PaymentMethodID = pid
		}
		splits := buildPaymentSplits(input)
		for i := range splits {
			pid, err := resolvePaymentMethodMasterID(splits[i].Method, splits[i].PaymentMethodID, nameToID)
			if err != nil {
				return nil, err
			}
			splits[i].PaymentMethodID = pid
		}

		if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
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
			return nil
		}); err != nil {
			return nil, apperrors.Wrap(err, "failed to upsert payment")
		}
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

	// #115: 締め後編集監査ログ（ベストエフォート）
	if input.IsPostClose && s.auditSvc != nil {
		billingID := input.ID
		aType := "system"
		if input.StaffID != nil {
			aType = "staff"
		}
		meta := map[string]any{}
		if input.PostCloseReason != nil {
			meta["reason"] = *input.PostCloseReason
		}
		if logErr := s.auditSvc.LogEntry(ctx, &AuditLogInput{
			ClinicID:   &input.ClinicID,
			ActorID:    input.StaffID,
			ActorType:  aType,
			Action:     model.AuditActionBillingPostCloseEdit,
			Resource:   "billing",
			ResourceID: &billingID,
			Metadata:   meta,
		}); logErr != nil {
			slog.WarnContext(ctx, "audit log failed for post_close_edit", "error", logErr, "billing_id", input.ID)
		}
	}

	if input.Status != nil && *input.Status == model.BillingStatusCompleted {
		if err := s.completeAccountingAppointments(ctx, input.ClinicID, accounting); err != nil {
			return nil, apperrors.Wrap(err, "failed to complete accounting appointments during update")
		}
		s.syncCPMStageTag(ctx, input.ClinicID, accounting)
	}
	return accounting, nil
}

// loadPaymentMethodNameToID は当該 clinic の payment_methods マスタを name→id マップとして読み込む（#128）。
func (s *accountingService) loadPaymentMethodNameToID(ctx context.Context, clinicID uint64) (map[string]uint64, error) {
	methods, err := s.payMethodRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load payment methods", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load payment methods")
	}
	nameToID := make(map[string]uint64, len(methods))
	for i := range methods {
		nameToID[methods[i].Name] = methods[i].ID
	}
	return nameToID, nil
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
