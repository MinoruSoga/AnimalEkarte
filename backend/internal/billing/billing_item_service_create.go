// Package service provides business logic implementations for BillingItem entity.
package billing

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// validateBillingItemOwnership は billing および trimming マスタFKのテナント所有権を検証する。
func (s *billingItemService) validateBillingItemOwnership(ctx context.Context, input *CreateBillingItemInput) error {
	// テナント所有権確認: billing が同一クリニックに属することを確認
	if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
		slog.ErrorContext(ctx, "billing not found or belongs to different clinic", "error", err)
		return apperrors.Wrap(err, "billing not found or belongs to different clinic")
	}

	// X-4: クロステナント write 防止: trimming_course_id/trimming_option_id が
	// caller の clinic に属することを検証する(#124/#125 と同型の master FK 所有権チェック)。
	if input.TrimmingCourseID != nil {
		if _, err := s.trimmingCourseRepo.FindByID(ctx, input.ClinicID, *input.TrimmingCourseID); err != nil {
			slog.ErrorContext(ctx, "trimming course not found or belongs to different clinic", "error", err)
			return apperrors.Wrap(err, "failed to verify trimming course ownership")
		}
	}
	if input.TrimmingOptionID != nil {
		if _, err := s.trimmingOptionRepo.FindByID(ctx, input.ClinicID, *input.TrimmingOptionID); err != nil {
			slog.ErrorContext(ctx, "trimming option not found or belongs to different clinic", "error", err)
			return apperrors.Wrap(err, "failed to verify trimming option ownership")
		}
	}
	return nil
}

// resolveOwnerDiscountRate は会計に紐付く飼主の割引率を返す。ownerRepo 未配線または取得失敗は 0。
func (s *billingItemService) resolveOwnerDiscountRate(ctx context.Context, clinicID uint64, ownerID *uint64) float64 {
	if ownerID == nil || s.ownerRepo == nil {
		return 0
	}
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID)
	if err != nil || owner == nil {
		return 0
	}
	return owner.DiscountRate
}

// resolveAutoDiscount は #81 段階2b: 明細に適用するキャンペーン/飼主割引額を算出する(best-effort)。
// campaignRepo 未配線時は 0。会計日(billing.ScheduledDate)・明細カテゴリ・個別商品IDで該当キャンペーンを検索し、
// 飼主割引と高い方を採用する(CalculateItemCampaignDiscount)。
func (s *billingItemService) resolveAutoDiscount(
	ctx context.Context,
	input *CreateBillingItemInput,
	category model.ItemCategory,
	unitPrice int64,
	quantity float64,
) int64 {
	if s.campaignRepo == nil {
		return 0
	}
	billing, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil || billing == nil {
		return 0
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, input.ClinicID, billing.OwnerID)
	campaign, cerr := s.campaignRepo.FindApplicableForItem(ctx, input.ClinicID, billing.ScheduledDate, category, input.MerchandiseItemID)
	if cerr != nil {
		// A-4: best-effort 継続自体は妥当（自動割引はあくまで補助機能）だが、クエリ障害で
		// 自動割引が静かに止まると運用から不可視になるため Warn ログを追加する。
		slog.WarnContext(ctx, "campaign lookup failed; skipping auto discount", "error", cerr, "clinic_id", input.ClinicID, "billing_id", input.BillingID)
		campaign = nil // best-effort: キャンペーン検索失敗は割引なしで続行
	}
	itemSubtotal := int64(float64(unitPrice) * quantity)
	return CalculateItemCampaignDiscount(itemSubtotal, campaign, ownerRate)
}

func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if err := validateCreateBillingItemInput(input); err != nil {
		return nil, err
	}
	// #115 / BUG-463: 締め後編集は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return nil, err
	}
	if err := s.validateBillingItemOwnership(ctx, input); err != nil {
		return nil, err
	}
	// 未請求予防接種の blocking warning は会計確定と会計作成だけを止める（BUG-015）。
	// 明細追加に流用すると物販・その他が常時 409 になり会計業務が停止する。

	var item *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.createItemInAmbientTx(txCtx, input, createItemAmbientOpts{
			recalculate:     true,
			recordPostClose: true,
		})
		if err != nil {
			return err
		}
		item = created
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create billing item in transaction")
	}

	slog.InfoContext(ctx, "billing item created",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("billing_id", input.BillingID),
		slog.Uint64("item_id", item.ID),
	)
	return item, nil
}

// createItemInAmbientTx は ambient tx 内で明細1件を作成する（WithTx を開始しない）。
// BUG-018 Complete 経路から呼び出され、独立 tx の部分 commit を防ぐ。
func (s *billingItemService) createItemInAmbientTx(ctx context.Context, input *CreateBillingItemInput, opts createItemAmbientOpts) (*model.BillingItem, error) {
	taxType, taxRate, source, err := resolveBillingItemDefaults(input)
	if err != nil {
		return nil, err
	}

	item := newBillingItemFromCreateInput(input, taxType, taxRate, source)

	billing, err := s.billingRepo.LockAndFindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock billing before creating item")
	}
	if err := rejectIfBillingFinalized(billing, "登録"); err != nil {
		return nil, err
	}

	needsResolve := input.AppointmentID == nil && (input.TrimmingCourseID != nil || input.TrimmingOptionID != nil)
	if tx := persistence.TxFromContext(ctx); needsResolve && tx != nil {
		resolved, resolveErr := resolveUniqueTrimmingAppointmentID(
			tx.WithContext(ctx), input.ClinicID, billing.PetID, input.TrimmingCourseID, input.TrimmingOptionID,
		)
		if resolveErr != nil {
			return nil, resolveErr
		}
		input.AppointmentID = resolved
		item.AppointmentID = resolved
	}

	category, err := s.repo.ValidateCreateReferences(
		ctx,
		input.ClinicID,
		input.BillingID,
		input.MerchandiseItemID,
		input.TreatmentID,
		input.AppointmentID,
		input.TrimmingCourseID,
		input.TrimmingOptionID,
	)
	if err != nil {
		slog.WarnContext(ctx, "billing item references rejected", "error", err)
		return nil, apperrors.Wrap(err, "failed to validate billing item references")
	}
	if input.MerchandiseItemID != nil {
		item.Category = category
	}
	if input.ExamID != nil && input.VaccinationID != nil {
		return nil, invalidBillingItemReferenceCombination()
	}
	if err := applyExamBillingProvenance(ctx, s.repo, input, item); err != nil {
		return nil, err
	}
	if err := s.applyVaccinationBillingItemCreate(ctx, input, item); err != nil {
		return nil, err
	}

	if err := applyBillingItemOtherMetadata(input, item); err != nil {
		return nil, err
	}
	if item.CreatedBy != nil {
		if err := s.repo.LockActiveStaffAssignment(ctx, input.ClinicID, *item.CreatedBy); err != nil {
			return nil, err
		}
	}

	if item.DiscountAmount == 0 && input.VaccinationID == nil && input.ExamID == nil {
		item.DiscountAmount = s.resolveAutoDiscount(
			ctx,
			input,
			item.Category,
			item.UnitPrice,
			item.Quantity,
		)
	}

	if err := s.repo.Create(ctx, item); err != nil {
		slog.ErrorContext(ctx, "failed to create billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to create billing item")
	}

	if opts.recalculate {
		if err := s.recalculateTotals(ctx, input.ClinicID, input.BillingID, false); err != nil {
			slog.ErrorContext(ctx, "failed to recalculate billing totals after create",
				slog.Uint64("billing_id", input.BillingID),
				slog.String("error", err.Error()),
			)
			return nil, apperrors.Wrap(err, "failed to recalculate billing totals")
		}
	}

	if opts.recordPostClose {
		if err := s.recordBillingItemPostClose(ctx, input.IsPostClose, input.ClinicID, input.BillingID, billing, input.StaffID, input.PostCloseReason, "create", &item.ID); err != nil {
			return nil, err
		}
	}

	return item, nil
}

func newBillingItemFromCreateInput(
	input *CreateBillingItemInput,
	taxType model.TaxType,
	taxRate float64,
	source model.ItemSource,
) *model.BillingItem {
	return &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              model.ItemCategory(input.Category),
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		DiscountRate:          input.DiscountRate,
		DiscountAmount:        input.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: input.IsInsuranceApplicable,
		Source:                source,
		MerchandiseItemID:     input.MerchandiseItemID,
		TreatmentID:           input.TreatmentID,
		VaccinationID:         input.VaccinationID,
		ExamID:                input.ExamID,
		AppointmentID:         input.AppointmentID,
		TrimmingCourseID:      input.TrimmingCourseID,
		TrimmingOptionID:      input.TrimmingOptionID,
		SortOrder:             input.SortOrder,
	}
}

func (s *billingItemService) applyVaccinationBillingItemCreate(
	ctx context.Context,
	input *CreateBillingItemInput,
	item *model.BillingItem,
) error {
	if input.VaccinationID == nil {
		return nil
	}
	if input.MerchandiseItemID != nil ||
		input.TreatmentID != nil ||
		input.AppointmentID != nil ||
		input.TrimmingCourseID != nil ||
		input.TrimmingOptionID != nil ||
		input.ExamID != nil {
		return invalidBillingItemReferenceCombination()
	}
	_, err := s.repo.ValidateVaccinationCreateReference(
		ctx,
		input.ClinicID,
		input.BillingID,
		*input.VaccinationID,
	)
	if err != nil {
		return err
	}
	item.Category = model.ItemCategoryVaccine
	item.Source = model.ItemSourceMedicalRecord
	item.VaccinationID = input.VaccinationID
	clinicID := input.ClinicID
	item.ClinicID = &clinicID
	return nil
}

// CreateItemForComplete は BUG-018 Complete 用の ambient-tx 明細作成。
// unbilled / post-close は command 側が一度だけ行う。totals 再計算は skip し caller が最後にまとめる。
func (s *billingItemService) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if err := validateCreateBillingItemInput(input); err != nil {
		return nil, err
	}
	return s.createItemInAmbientTx(ctx, input, createItemAmbientOpts{
		recalculate:     false,
		recordPostClose: false,
	})
}

// RecalculateTotalsForComplete は Complete 用に items から totals を再計算して billings に書く。
func (s *billingItemService) RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error) {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return 0, 0, 0, apperrors.Wrap(err, "failed to find billing items for complete totals")
	}
	subtotal, taxTotal, totalAmount = CalculateBillingTotals(items)
	if err := s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount); err != nil {
		return 0, 0, 0, apperrors.Wrap(err, "failed to update billing totals for complete")
	}
	return subtotal, taxTotal, totalAmount, nil
}
