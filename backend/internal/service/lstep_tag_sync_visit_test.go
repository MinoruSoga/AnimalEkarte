package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockMedRecordRepoForLstepVisit is a full repository.MedicalRecordRepository
// implementation used only by the lstep visit-tag sync tests (SyncVisitCompletionTags,
// SyncNextVisitTag). It exposes configurable hooks only for the two methods those
// sync functions actually call, so tests can control the FindOwnerVisitSummary /
// FindLatestByOwner return values — the shared mockMedicalRecordRepository in
// medical_record_service_test.go hardcodes FindLatestByOwner to (nil, nil) and cannot
// be reused for these branches.
type mockMedRecordRepoForLstepVisit struct {
	findOwnerVisitSummaryFn func(ctx context.Context, clinicID, ownerID uint64) (*repository.OwnerVisitSummary, error)
	findLatestByOwnerFn     func(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error)
}

func (m *mockMedRecordRepoForLstepVisit) FindAll(_ context.Context, _ []uint64, _ repository.MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindByID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) Create(_ context.Context, _ *model.MedicalRecord) error {
	return nil
}
func (m *mockMedRecordRepoForLstepVisit) Update(_ context.Context, _, _ uint64, _ map[string]any, _ *int) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockMedRecordRepoForLstepVisit) DeleteDraftByAppointmentID(_ context.Context, _, _ uint64) error {
	return nil
}

// LockByIDForUpdate は X-11 finalize-lock テスト用に FindByID と同じ挙動へ委譲する。
func (m *mockMedRecordRepoForLstepVisit) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}
func (m *mockMedRecordRepoForLstepVisit) CountByPetID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindFirstVisitDateByPetID(_ context.Context, _, _ uint64) (*time.Time, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) CountEstimatesByMedicalRecordID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*repository.OwnerVisitSummary, error) {
	if m.findOwnerVisitSummaryFn != nil {
		return m.findOwnerVisitSummaryFn(ctx, clinicID, ownerID)
	}
	return &repository.OwnerVisitSummary{}, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	if m.findLatestByOwnerFn != nil {
		return m.findLatestByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindDormantOwnerEntries(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindDormantOwnerEntriesCursor(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindOwnersByFirstVisitDate(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindOwnersByLastVisitDays(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) FindOwnersByNextVisitRecommended(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *mockMedRecordRepoForLstepVisit) CountByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// mockAccountingRepoForLstepVisit is a full repository.AccountingRepository
// implementation used only by lstep visit-tag sync tests, exposing a configurable
// hook for SumPaidByOwner (the shared mockAccountingRepository hardcodes it to
// (0, nil) and cannot exercise the repo-error branch).
type mockAccountingRepoForLstepVisit struct {
	sumPaidByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

func (m *mockAccountingRepoForLstepVisit) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepoForLstepVisit) FindAllForClinics(_ context.Context, _ []uint64, _, _ *uint64, _, _, _ *string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepoForLstepVisit) FindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepoForLstepVisit) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepoForLstepVisit) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepoForLstepVisit) Create(_ context.Context, _ uint64, _ *model.Billing) error {
	return nil
}
func (m *mockAccountingRepoForLstepVisit) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.Billing, error) {
	return nil, nil
}
func (m *mockAccountingRepoForLstepVisit) SavePayment(_ context.Context, _ *model.Payment) error {
	return nil
}
func (m *mockAccountingRepoForLstepVisit) SavePaymentSplits(_ context.Context, _ []model.PaymentSplit) error {
	return nil
}
func (m *mockAccountingRepoForLstepVisit) CompleteAccountingAppointments(_ context.Context, _ uint64, _, _, _ *uint64, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAccountingRepoForLstepVisit) FindUnpaidByBilling(_ context.Context, _ uint64, _, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *mockAccountingRepoForLstepVisit) FindUnpaidByOwner(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	return nil, 0, repository.UnpaidSummary{}, nil
}
func (m *mockAccountingRepoForLstepVisit) SumUnpaidByOwner(_ context.Context, _, _ uint64) (repository.OwnerUnpaidBalance, error) {
	return repository.OwnerUnpaidBalance{}, nil
}
func (m *mockAccountingRepoForLstepVisit) FindMonthlyUnpaidCarryover(_ context.Context, _ uint64, _, _ string, _, _ int) ([]repository.MonthlyUnpaidOwnerPet, int64, repository.MonthlyUnpaidSummary, error) {
	return nil, 0, repository.MonthlyUnpaidSummary{}, nil
}
func (m *mockAccountingRepoForLstepVisit) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*repository.DailySummaryResult, error) {
	return &repository.DailySummaryResult{}, nil
}
func (m *mockAccountingRepoForLstepVisit) GetCloseAggregate(_ context.Context, _ repository.GetCloseAggregateInput) (*repository.CloseAggregateResult, error) {
	return &repository.CloseAggregateResult{}, nil
}
func (m *mockAccountingRepoForLstepVisit) GetMonthlyReport(_ context.Context, _ uint64, _, _ int) (*repository.MonthlyReportResult, error) {
	return &repository.MonthlyReportResult{}, nil
}
func (m *mockAccountingRepoForLstepVisit) GetMonthlyReportByPeriod(_ context.Context, _ uint64, _, _ time.Time) (*repository.MonthlyReportResult, error) {
	return &repository.MonthlyReportResult{}, nil
}
func (m *mockAccountingRepoForLstepVisit) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.sumPaidByOwnerFn != nil {
		return m.sumPaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}
func (m *mockAccountingRepoForLstepVisit) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockAccountingRepoForLstepVisit) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]repository.OwnerAnnualRevenue, error) {
	return nil, nil
}

func TestSyncVisitCompletionTags(t *testing.T) {
	lineUID := "U_visit"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return nil, errors.New("owner db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("skips when owner opted out", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LstepOptOut: true}, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("skips when owner has no LINE user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2}, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("returns wrapped error when visit summary lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*repository.OwnerVisitSummary, error) {
					return nil, errors.New("summary db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when LTV sum lookup fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{},
			accountRepo: &mockAccountingRepoForLstepVisit{
				sumPaidByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return 0, errors.New("ltv db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("noop when buildClient returns nil client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{},
			accountRepo:   &mockAccountingRepoForLstepVisit{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
	})

	t.Run("success: adds fresh visit tags and clears dormant/legacy tags", func(t *testing.T) {
		var addedTags, removedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTags = append(removedTags, tagName)
				return nil
			},
		}
		first := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		last := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*repository.OwnerVisitSummary, error) {
					return &repository.OwnerVisitSummary{FirstVisitAt: &first, LastVisitAt: &last, AnnualCount: 6}, nil
				},
			},
			accountRepo: &mockAccountingRepoForLstepVisit{
				sumPaidByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 60_000, nil },
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{
						{TagName: "dormant_180d"},
						{TagName: "VISIT_120日超"},
						{TagName: "cpm_dormant"},
					}, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, "first_visit_2024-01-01")
		assert.Contains(t, addedTags, "last_visit_2026-04-01")
		assert.Contains(t, addedTags, "ltv_amount_5")
		assert.Contains(t, addedTags, "visit_count_annual_5")
		assert.Contains(t, removedTags, "dormant_180d")
		assert.Contains(t, removedTags, "VISIT_120日超")
		assert.Contains(t, removedTags, "cpm_dormant")
		assert.Contains(t, removedTags, "dormant")
		assert.Contains(t, removedTags, "noshow")
		assert.Contains(t, removedTags, "reserved")
	})

	t.Run("returns wrapped error when adding a visit tag fails", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("add failed") },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{},
			accountRepo:   &mockAccountingRepoForLstepVisit{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo:  &mockLstepTagCacheRepository{},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when loading tag cache for dormant cleanup fails", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{},
			accountRepo:   &mockAccountingRepoForLstepVisit{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return nil, errors.New("cache db error")
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("removing a dormant tag fails: recorded as api failure but does not propagate", func(t *testing.T) {
		incrementCalls := 0
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("remove failed") },
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedRecordRepoForLstepVisit{},
			accountRepo:   &mockAccountingRepoForLstepVisit{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
			tagCacheRepo: &mockLstepTagCacheRepository{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
					return []*model.LstepTagCache{{TagName: "dormant_180d"}}, nil
				},
			},
			errorCounterRepo: &mockErrorCounterRepo{
				incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
					incrementCalls++
					return 1, nil
				},
			},
		}
		err := svc.SyncVisitCompletionTags(context.Background(), 1, 2)
		assert.NoError(t, err)
		assert.Positive(t, incrementCalls)
	})
}
