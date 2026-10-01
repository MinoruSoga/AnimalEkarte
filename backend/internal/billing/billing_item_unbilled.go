package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	UnbilledWarningSourceExam                   = "exam"
	UnbilledWarningCodeExamTypeMasterUnbillable = "exam_type_master_unbillable"
)

// Unbilled aggregation and same-day ungrouped summary helpers for BillingItemService.
// Split from billing_item_service.go (ARCH-A4-billing S1) for file cohesion only — behavior unchanged.

func (s *billingItemService) aggregateUnbilled(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
	treatments, err := s.treatmentRepo.FindUnbilledByPetID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find unbilled treatments")
	}
	items := make([]model.BillingItem, 0, len(treatments))
	for i := range treatments {
		items = append(items, treatmentToUnbilledBillingItem(&treatments[i]))
	}

	if trimmingFinder, ok := s.repo.(unbilledTrimmingItemFinder); ok {
		trimmingItems, err := trimmingFinder.FindUnbilledTrimmingItemsByPetID(ctx, clinicID, petID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to find unbilled trimming items")
		}
		items = append(items, trimmingItems...)
	}

	vaccinationItems, unbillableCount, err := s.repo.FindUnbilledVaccinationItemsByPetID(ctx, clinicID, petID)
	if err != nil {
		// infra / unexpected: keep 500 fail-closed (do not convert to warning)
		return nil, apperrors.Wrap(err, "failed to find unbilled vaccination items")
	}
	items = append(items, vaccinationItems...)

	warnings := make([]UnbilledWarning, 0, 2)
	if unbillableCount > 0 {
		warnings = append(warnings, UnbilledWarning{
			Source:   UnbilledWarningSourceVaccination,
			Code:     UnbilledWarningCodeVaccinationMasterUnbillable,
			Count:    unbillableCount,
			Blocking: true,
		})
	}

	examItems, examUnbillable, examErr := s.repo.FindUnbilledExamItemsByPetID(ctx, clinicID, petID)
	if examErr != nil {
		return nil, apperrors.Wrap(examErr, "failed to find unbilled exam items")
	}
	items = append(items, examItems...)
	if examUnbillable > 0 {
		warnings = append(warnings, UnbilledWarning{
			Source:   UnbilledWarningSourceExam,
			Code:     UnbilledWarningCodeExamTypeMasterUnbillable,
			Count:    examUnbillable,
			Blocking: true,
		})
	}
	// EMR-229: 飼主/キャンペーン割引を候補明細へ事前適用する。
	// revision は事前適用後の items を fingerprint するため、画面表示と確定時の
	// 版照合が一致する（client は discount_amount を complete payload に返送し、
	// create 側は DiscountAmount != 0 で自動割引をスキップして引き継ぐ）。
	ownerRate := s.applyUnbilledDiscounts(ctx, clinicID, petID, items)
	return &UnbilledDetails{
		Items:             items,
		Warnings:          warnings,
		Revision:          computeUnbilledRevision(items, warnings),
		OwnerDiscountRate: ownerRate,
	}, nil
}

// resolvePetOwnerDiscountRate は pet→主飼主の割引率を返す。finder/ownerRepo 未配線・
// 解決失敗は 0（best-effort: 割引なしの候補表示にフォールバックする）。
func (s *billingItemService) resolvePetOwnerDiscountRate(ctx context.Context, clinicID, petID uint64) float64 {
	if s.petOwnerFinder == nil {
		return 0
	}
	ownerID, err := s.petOwnerFinder.FindPetOwnerInClinic(ctx, clinicID, petID)
	if err != nil || ownerID == 0 {
		return 0
	}
	return s.resolveOwnerDiscountRate(ctx, clinicID, &ownerID)
}

// applyUnbilledDiscounts は未請求候補の各明細へ飼主/キャンペーン割引を事前適用し、
// 使用した飼主割引率を返す（EMR-229）。
//
// best-effort: campaign 検索失敗は該当カテゴリのキャンペーン割引なしで続行する。
// createItemInAmbientTx の自動割引と同じ除外規則（vaccination/exam 明細・既に明示
// 割引済みの明細は対象外）に従い、候補表示と確定結果が一致するようにする。
// キャンペーン適用日は会計予定日が未確定のため当日（time.Now）を使う。
func (s *billingItemService) applyUnbilledDiscounts(ctx context.Context, clinicID, petID uint64, items []model.BillingItem) float64 {
	ownerRate := s.resolvePetOwnerDiscountRate(ctx, clinicID, petID)
	if s.campaignRepo == nil && ownerRate <= 0 {
		return ownerRate
	}
	date := time.Now()
	// キャンペーン検索は (category, merchandiseItemID) 単位でキャッシュする。
	// category だけでキャッシュすると同一区分の別商品に対する個別商品キャンペーンを
	// 取りこぼす（FindApplicableForItem は商品ID指定キャンペーンを評価する）。
	type campaignCacheKey struct {
		category          model.ItemCategory
		merchandiseItemID uint64
	}
	campaigns := make(map[campaignCacheKey]*model.Campaign)
	for i := range items {
		if items[i].VaccinationID != nil || items[i].ExamID != nil {
			continue
		}
		if items[i].DiscountAmount != 0 {
			continue
		}
		var key campaignCacheKey
		key.category = items[i].Category
		if items[i].MerchandiseItemID != nil {
			key.merchandiseItemID = *items[i].MerchandiseItemID
		}
		campaign, cached := campaigns[key]
		if !cached && s.campaignRepo != nil {
			c, cerr := s.campaignRepo.FindApplicableForItem(ctx, clinicID, date, items[i].Category, items[i].MerchandiseItemID)
			if cerr != nil {
				slog.WarnContext(ctx, "campaign lookup failed for unbilled discount prefill", "error", cerr, "clinic_id", clinicID, "pet_id", petID)
			} else {
				campaign = c
			}
			campaigns[key] = campaign
		}
		itemSubtotal := int64(float64(items[i].UnitPrice) * items[i].Quantity)
		if amount, appliedRate := ResolveItemDiscount(itemSubtotal, campaign, ownerRate); amount > 0 {
			items[i].DiscountAmount = amount
			items[i].DiscountRate = appliedRate
		}
	}
	return ownerRate
}

func hasBlockingUnbilledWarning(warnings []UnbilledWarning) bool {
	for i := range warnings {
		if warnings[i].Blocking && warnings[i].Count > 0 {
			return true
		}
	}
	return false
}

// GetUnbilledItems は legacy raw-array 契約。全 source 成功時は items を返す。
// blocking data-quality warning がある場合は silent partial を避け fail-closed（従来どおり error）。
// 部分可視化は GetUnbilledItemDetails を使う（BUG-013）。
func (s *billingItemService) GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	details, err := s.aggregateUnbilled(ctx, clinicID, petID)
	if err != nil {
		return nil, err
	}
	if hasBlockingUnbilledWarning(details.Warnings) {
		return nil, apperrors.WrapInternalServerError("vaccination vaccine master is not billable")
	}
	return details.Items, nil
}

func (s *billingItemService) GetUnbilledItemDetails(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
	return s.aggregateUnbilled(ctx, clinicID, petID)
}

func (s *billingItemService) AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error {
	details, err := s.aggregateUnbilled(ctx, clinicID, petID)
	if err != nil {
		return err
	}
	if hasBlockingUnbilledWarning(details.Warnings) {
		return apperrors.WrapConflict("未請求候補に請求不能な予防接種が含まれるため会計を確定できません")
	}
	return nil
}

// AssertUnbilledForComplete は complete 確定時の write-time fail-closed 検証（EMR-196②）。
// blocking unbilled warning（BUG-013）に加え、画面表示した集約の版（expectedRevision）と
// tx 内再集計の版が一致しなければ Conflict を返す。expectedRevision は必須で、
// 空文字は必ず不一致になる（未指定での通過は service 層が 400 で拒否済み）。
func (s *billingItemService) AssertUnbilledForComplete(ctx context.Context, clinicID, petID uint64, expectedRevision string) error {
	details, err := s.aggregateUnbilled(ctx, clinicID, petID)
	if err != nil {
		return err
	}
	if hasBlockingUnbilledWarning(details.Warnings) {
		return apperrors.WrapConflict("未請求候補に請求不能な予防接種が含まれるため会計を確定できません")
	}
	if details.Revision != expectedRevision {
		return newUnbilledRevisionConflictError(details.Revision)
	}
	return nil
}

func (s *billingItemService) GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
	mrCount, err := s.treatmentRepo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicID, petID, date)
	if err != nil {
		return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped medical records")
	}
	var trimmingCount int64
	if counter, ok := s.repo.(ungroupedTrimmingCounter); ok {
		trimmingCount, err = counter.CountNonAccountingTrimmingByPetAndDate(ctx, clinicID, petID, date)
		if err != nil {
			return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped trimming")
		}
	}
	return UngroupedSameDaySummary{MedicalRecordCount: mrCount, TrimmingCount: trimmingCount}, nil
}

func treatmentToUnbilledBillingItem(t *model.Treatment) model.BillingItem {
	treatmentID := t.ID
	medicalRecordID := t.MedicalRecordID
	taxType, taxRate := treatmentMasterTax(t)
	return model.BillingItem{
		ID:        t.ID,
		BillingID: 0,
		Category:  treatmentTypeToItemCategory(t),
		Name:      t.Content,
		UnitPrice: t.UnitPrice,
		Quantity:  t.Quantity,
		// EMR-229: カルテ明細に記録された明示割引額（円）を候補へ引き継ぐ。
		// treatment.discount_rate は FE/BE で率の単位が不一致（% vs 小数）のため伝播しない。
		DiscountAmount:        t.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: t.IsInsurance,
		Source:                model.ItemSourceMedicalRecord,
		TreatmentID:           &treatmentID,
		// BUG-011: complete が treatment 付き明細に medical_record_id を載せるため未請求候補で返す。
		MedicalRecordID: &medicalRecordID,
		SortOrder:       t.SortOrder,
	}
}

// treatmentMasterTax はリンク済みマスタ（consultation / procedure / medicine）の
// 税区分・税率を返す（EMR-65）。税フィールドを持たない inventory 由来やマスタ
// 未リンクの治療は従来の外税既定を維持する（fail-closed、行スキップや0円化はしない）。
func treatmentMasterTax(t *model.Treatment) (model.TaxType, float64) {
	if t.Consultation != nil {
		return t.Consultation.TaxType, t.Consultation.TaxRate
	}
	if t.Procedure != nil {
		return t.Procedure.TaxType, t.Procedure.TaxRate
	}
	if t.Medicine != nil {
		return t.Medicine.TaxType, t.Medicine.TaxRate
	}
	return model.TaxTypeExcluded, sharedkernel.DefaultTaxRate
}

func treatmentTypeToItemCategory(t *model.Treatment) model.ItemCategory {
	return sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
		Source:            model.ItemSourceMedicalRecord,
		TreatmentItemType: t.ItemType,
		IsSurgery:         t.Procedure != nil && t.Procedure.IsSurgery,
	})
}
