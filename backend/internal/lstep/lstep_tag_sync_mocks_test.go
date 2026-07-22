package lstep

// Tag-sync tests moved with their implementation in BE9-2C L③a. These narrow
// doubles mirror the consumer-side dependency views instead of importing the
// repository aggregator back into the lstep package.

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockOwnerRepository struct {
	findByIDFn                    func(context.Context, uint64, uint64) (*model.Owner, error)
	updateFn                      func(context.Context, uint64, uint64, map[string]any) error
	findByIDsFn                   func(context.Context, uint64, []uint64) ([]*model.Owner, error)
	findAllWithLineUserIDFn       func(context.Context, uint64) ([]model.Owner, error)
	findAllWithLineUserIDCursorFn func(context.Context, uint64, uint64, int) ([]model.Owner, error)
}

func (m *mockOwnerRepository) Update(ctx context.Context, clinicID, ownerID uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, ownerID, fields)
	}
	return nil
}

func (m *mockOwnerRepository) RecordLstepOptOut(ctx context.Context, clinicID, ownerID uint64, at time.Time, reason string) error {
	return m.Update(ctx, clinicID, ownerID, map[string]any{
		"lstep_opt_out":        true,
		"lstep_opt_out_at":     at,
		"lstep_opt_out_reason": reason,
	})
}

func (m *mockOwnerRepository) ClearLstepOptOut(ctx context.Context, clinicID, ownerID uint64) error {
	return m.Update(ctx, clinicID, ownerID, map[string]any{
		"lstep_opt_out":        false,
		"lstep_opt_out_at":     nil,
		"lstep_opt_out_reason": nil,
	})
}

func (m *mockOwnerRepository) FindByIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error) {
	if m.findByIDsFn != nil {
		return m.findByIDsFn(ctx, clinicID, ownerIDs)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error) {
	if m.findAllWithLineUserIDFn != nil {
		return m.findAllWithLineUserIDFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockOwnerRepository) FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error) {
	if m.findAllWithLineUserIDCursorFn != nil {
		return m.findAllWithLineUserIDCursorFn(ctx, clinicID, afterID, limit)
	}
	return nil, nil
}

type mockPetRepository struct {
	findByIDFn              func(context.Context, uint64, uint64) (*model.Pet, error)
	updateFn                func(context.Context, uint64, uint64, map[string]any) error
	findLivingByOwnerFn     func(context.Context, uint64, uint64) ([]model.Pet, error)
	countByOwnerFn          func(context.Context, uint64, uint64) (int64, error)
	countLivingByOwnerFn    func(context.Context, uint64, uint64) (int64, error)
	countLivingByOwnerIDsFn func(context.Context, uint64, []uint64) (map[uint64]int64, error)
}

func (m *mockPetRepository) FindByID(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

func (m *mockPetRepository) Update(ctx context.Context, clinicID, petID uint64, fields map[string]any) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, petID, fields)
	}
	return nil
}

func (m *mockPetRepository) RecordDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	return m.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at":     deceasedAt,
		"deceased_reason": reason,
		"status":          model.PetStatusDeceased,
	})
}

func (m *mockPetRepository) ClearDeath(ctx context.Context, clinicID, petID uint64) error {
	return m.Update(ctx, clinicID, petID, map[string]any{
		"deceased_at":     nil,
		"deceased_reason": nil,
		"status":          model.PetStatusAlive,
	})
}

func (m *mockPetRepository) CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error) {
	if m.countLivingByOwnerIDsFn != nil {
		return m.countLivingByOwnerIDsFn(ctx, clinicID, ownerIDs)
	}
	return nil, nil
}

func (m *mockPetRepository) FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error) {
	if m.findLivingByOwnerFn != nil {
		return m.findLivingByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockPetRepository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.countByOwnerFn != nil {
		return m.countByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockPetRepository) CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.countLivingByOwnerFn != nil {
		return m.countLivingByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

type mockMedicalRecordRepository struct {
	findOwnerVisitSummaryFn     func(context.Context, uint64, uint64) (*medicalrecord.OwnerVisitSummary, error)
	findLatestByOwnerFn         func(context.Context, uint64, uint64) (*model.MedicalRecord, error)
	findOwnersByFirstVisitFn    func(context.Context, uint64, time.Time) ([]uint64, error)
	findOwnersByLastVisitDaysFn func(context.Context, uint64, int, time.Time) ([]uint64, error)
	findOwnersByNextVisitRecFn  func(context.Context, uint64, time.Time) ([]uint64, error)
}

func (m *mockMedicalRecordRepository) FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByFirstVisitFn != nil {
		return m.findOwnersByFirstVisitFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error) {
	if m.findOwnersByLastVisitDaysFn != nil {
		return m.findOwnersByLastVisitDaysFn(ctx, clinicID, exactDays, asOf)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByNextVisitRecFn != nil {
		return m.findOwnersByNextVisitRecFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*medicalrecord.OwnerVisitSummary, error) {
	if m.findOwnerVisitSummaryFn != nil {
		return m.findOwnerVisitSummaryFn(ctx, clinicID, ownerID)
	}
	return &medicalrecord.OwnerVisitSummary{}, nil
}

func (m *mockMedicalRecordRepository) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	if m.findLatestByOwnerFn != nil {
		return m.findLatestByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

type mockAccountingRepository struct {
	sumPaidByOwnerFn func(context.Context, uint64, uint64) (int64, error)
}

func (m *mockAccountingRepository) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.sumPaidByOwnerFn != nil {
		return m.sumPaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockAccountingRepository) MaxSingleVisitAmountByOwner(context.Context, uint64, uint64) (int64, error) {
	return 0, nil
}

func (m *mockAccountingRepository) FindOwnersByAnnualRevenue(context.Context, uint64) ([]billing.OwnerAnnualRevenue, error) {
	return nil, nil
}

type mockCheckupRepository struct {
	findByOwnerIDFn func(context.Context, uint64, uint64) ([]model.Checkup, error)
}

func (m *mockCheckupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	if m.findByOwnerIDFn != nil {
		return m.findByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

type mockVaccinationRepository struct {
	findByIDFn    func(context.Context, uint64, uint64) (*model.Vaccination, error)
	findByOwnerFn func(context.Context, uint64, uint64) ([]model.Vaccination, error)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

type mockPrescriptionRepository struct {
	findActiveByOwnerFn func(context.Context, uint64, uint64) ([]model.Prescription, error)
}

func (m *mockPrescriptionRepository) FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error) {
	if m.findActiveByOwnerFn != nil {
		return m.findActiveByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

type mockBillingItemRepository struct {
	hasItemByOwnerSinceFn         func(context.Context, uint64, uint64, time.Time, []string) (bool, error)
	hasFoodPurchaseByOwnerSinceFn func(context.Context, uint64, uint64, time.Time, []string) (bool, error)
}

func (m *mockBillingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasItemByOwnerSinceFn != nil {
		return m.hasItemByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}

func (m *mockBillingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasFoodPurchaseByOwnerSinceFn != nil {
		return m.hasFoodPurchaseByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}
