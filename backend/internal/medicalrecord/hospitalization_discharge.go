package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

// DischargeWithBilling は退院処理を行い、オプションで会計レコードを自動生成する。
// care_plan_items を billing_items に変換してトランザクション内で原子的に実行する。
func (s *hospitalizationService) DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
	// 入院レコード取得
	hosp, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	if hosp.Status == model.HospitalizationStatusDischarged {
		return nil, apperrors.WrapInvalidInput("hospitalization is already discharged")
	}

	result := &DischargeWithBillingResult{
		HospitalizationID: id,
		Status:            string(model.HospitalizationStatusDischarged),
	}

	// MRB-02: same nil-transactor guard as Create/Update (panic vs explicit 500).
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	// BE9-2D ⑤ Phase1: repos.Transaction（tx-bound clone）→ Transactor.WithTx（ctx-txKey）へ変換。
	// 閉包内の read/write は各 repo の dbOrTx が txCtx の ambient tx へ参加する（挙動は旧機構と等価）。
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 0. Q2-C: FOR UPDATE で直列化し、locked 行の OwnerID/PetID で Q2-A 再検証する。
		// LockByIDForUpdate の行スナップショットにスカラーが含まれるため、検証用 ID が空にならない。
		locked, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization for discharge")
		}
		if locked.Status == model.HospitalizationStatusDischarged {
			return apperrors.WrapInvalidInput("hospitalization is already discharged")
		}
		// MRB-06: discharge date must not precede admission start.
		if err := validateHospitalizationDateRange(locked.StartDate, input.DischargeDate); err != nil {
			return err
		}

		// 1. 汚染行対策: CreateAccounting 有無に関わらず、Update 前に Owner/Pet の clinic 所有を再検証する（AUD-004 Q2-A）。
		if err := s.reservationRepo.AssertOwnerInClinic(txCtx, clinicID, locked.OwnerID); err != nil {
			return apperrors.Wrap(err, "failed to verify hospitalization owner ownership")
		}
		if _, err := s.reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, locked.PetID); err != nil {
			return apperrors.Wrap(err, "failed to verify hospitalization pet ownership")
		}
		if input.CreateAccounting {
			if input.ActorID == nil || *input.ActorID == 0 {
				return apperrors.WrapInvalidInput("staff actor is required to create discharge billing")
			}
			if s.auditTx == nil {
				return apperrors.WrapInternalServerError("hospitalization discharge billing audit dependency is required")
			}
		}

		// 2. 退院ステータスに更新
		dischargedStatus := model.HospitalizationStatusDischarged
		dischargeFields := map[string]any{
			"status":   dischargedStatus,
			"end_date": input.DischargeDate,
		}
		if _, err := s.hospRepo.UpdateIfNotDischarged(txCtx, clinicID, id, dischargeFields); err != nil {
			return apperrors.Wrap(err, "failed to discharge hospitalization")
		}

		if !input.CreateAccounting {
			return nil
		}

		// 2. ケアプラン取得
		carePlanItems, err := s.carePlanItemRepo.FindByHospitalizationID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get care plan items")
		}

		// 3. 会計レコード作成
		billing := &model.Billing{
			ClinicID:          clinicID,
			HospitalizationID: &id,
			PetID:             &locked.PetID,
			OwnerID:           &locked.OwnerID,
			Status:            model.BillingStatusWaiting,
			ScheduledDate:     input.DischargeDate,
		}
		if err := s.accountingRepo.Create(txCtx, clinicID, billing); err != nil {
			return apperrors.Wrap(err, "failed to create billing")
		}

		// 4. ケアプラン → 会計明細に変換
		var subtotalAmount int64
		for i := range carePlanItems {
			item := &carePlanItems[i]
			billingItem := &model.BillingItem{
				BillingID: billing.ID,
				Category: sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
					Source:              model.ItemSourceHospitalization,
					CarePlanType:        item.Type,
					IsSurgery:           item.Procedure != nil && item.Procedure.IsSurgery,
					HospitalizationType: locked.HospitalizationType,
				}),
				Name:      item.Name,
				UnitPrice: item.UnitPrice,
				Quantity:  1.0,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   sharedkernel.DefaultTaxRate,
				Source:    model.ItemSourceHospitalization,
				SortOrder: i,
			}
			if err := s.billingItemRepo.Create(txCtx, billingItem); err != nil {
				return apperrors.Wrap(err, "failed to create billing item")
			}
			subtotalAmount += item.UnitPrice
		}

		// 5. 合計金額更新
		taxAmount := int64(float64(subtotalAmount) * sharedkernel.DefaultTaxRate)
		totalAmount := subtotalAmount + taxAmount
		if len(carePlanItems) > 0 {
			if err := s.billingItemRepo.UpdateBillingTotals(txCtx, clinicID, billing.ID, subtotalAmount, taxAmount, totalAmount); err != nil {
				return apperrors.Wrap(err, "failed to update billing totals")
			}
		}

		// 6. 会計を伴う退院を同じ tx 内で監査する。失敗時は status/会計/明細/合計を一括 rollback する。
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    input.ActorID,
			ActorType:  model.AuditActorTypeStaff,
			Action:     model.AuditActionHospitalizationDischargeWithBilling,
			Resource:   model.AuditResourceHospitalization,
			ResourceID: &resourceID,
			NewValue: map[string]any{
				"billing_id":      billing.ID,
				"subtotal_amount": subtotalAmount,
				"tax_amount":      taxAmount,
				"total_amount":    totalAmount,
			},
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit hospitalization discharge billing")
		}

		result.AccountingID = &billing.ID
		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to discharge hospitalization with billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to discharge hospitalization with billing")
	}

	slog.InfoContext(ctx, "hospitalization discharged",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("create_accounting", input.CreateAccounting))

	return result, nil
}
