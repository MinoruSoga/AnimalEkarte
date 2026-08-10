package billing

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// Unbilled aggregation and same-day ungrouped summary helpers for BillingItemService.
// Split from billing_item_service.go (ARCH-A4-billing S1) for file cohesion only — behavior unchanged.

func (s *billingItemService) aggregateUnbilled(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error) {
	treatments, err := s.treatmentRepo.FindUnbilledByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find unbilled treatments", "error", err)
		return nil, apperrors.Wrap(err, "failed to find unbilled treatments")
	}
	items := make([]model.BillingItem, 0, len(treatments))
	for i := range treatments {
		items = append(items, treatmentToUnbilledBillingItem(&treatments[i]))
	}

	if trimmingFinder, ok := s.repo.(unbilledTrimmingItemFinder); ok {
		trimmingItems, err := trimmingFinder.FindUnbilledTrimmingItemsByPetID(ctx, clinicID, petID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find unbilled trimming items", "error", err)
			return nil, apperrors.Wrap(err, "failed to find unbilled trimming items")
		}
		items = append(items, trimmingItems...)
	}

	vaccinationItems, unbillableCount, err := s.repo.FindUnbilledVaccinationItemsByPetID(ctx, clinicID, petID)
	if err != nil {
		// infra / unexpected: keep 500 fail-closed (do not convert to warning)
		slog.ErrorContext(ctx, "failed to find unbilled vaccination items", "error", err)
		return nil, apperrors.Wrap(err, "failed to find unbilled vaccination items")
	}
	items = append(items, vaccinationItems...)

	warnings := make([]UnbilledWarning, 0, 1)
	if unbillableCount > 0 {
		warnings = append(warnings, UnbilledWarning{
			Source:   UnbilledWarningSourceVaccination,
			Code:     UnbilledWarningCodeVaccinationMasterUnbillable,
			Count:    unbillableCount,
			Blocking: true,
		})
	}
	return &UnbilledDetails{Items: items, Warnings: warnings}, nil
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

func (s *billingItemService) GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error) {
	mrCount, err := s.treatmentRepo.CountFinalizedUnconfirmedByPetAndDate(ctx, clinicID, petID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count ungrouped medical records", "error", err)
		return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped medical records")
	}
	var trimmingCount int64
	if counter, ok := s.repo.(ungroupedTrimmingCounter); ok {
		trimmingCount, err = counter.CountNonAccountingTrimmingByPetAndDate(ctx, clinicID, petID, date)
		if err != nil {
			slog.ErrorContext(ctx, "failed to count ungrouped trimming", "error", err)
			return UngroupedSameDaySummary{}, apperrors.Wrap(err, "failed to count ungrouped trimming")
		}
	}
	return UngroupedSameDaySummary{MedicalRecordCount: mrCount, TrimmingCount: trimmingCount}, nil
}

func treatmentToUnbilledBillingItem(t *model.Treatment) model.BillingItem {
	treatmentID := t.ID
	medicalRecordID := t.MedicalRecordID
	return model.BillingItem{
		ID:                    t.ID,
		BillingID:             0,
		Category:              treatmentTypeToItemCategory(t),
		Name:                  t.Content,
		UnitPrice:             t.UnitPrice,
		Quantity:              t.Quantity,
		TaxType:               model.TaxTypeExcluded,
		TaxRate:               sharedkernel.DefaultTaxRate,
		IsInsuranceApplicable: t.IsInsurance,
		Source:                model.ItemSourceMedicalRecord,
		TreatmentID:           &treatmentID,
		// BUG-011: complete が treatment 付き明細に medical_record_id を載せるため未請求候補で返す。
		MedicalRecordID: &medicalRecordID,
		SortOrder:       t.SortOrder,
	}
}

func treatmentTypeToItemCategory(t *model.Treatment) model.ItemCategory {
	return sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
		Source:            model.ItemSourceMedicalRecord,
		TreatmentItemType: t.ItemType,
		IsSurgery:         t.Procedure != nil && t.Procedure.IsSurgery,
	})
}
