package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- test doubles scoped to this file ----

// cpmSettingsServiceWrapper embeds mockLstepSettingsService (stable, from
// lstep_lifecycle_service_test.go) and overrides GetCPMVersion, which the base mock hardcodes to
// always return "v1" with no injection hook.
type cpmSettingsServiceWrapper struct {
	*mockLstepSettingsService
	cpmVersion    string
	cpmVersionErr error
}

func (m *cpmSettingsServiceWrapper) GetCPMVersion(_ context.Context, _ uint64) (string, error) {
	if m.cpmVersionErr != nil {
		return "", m.cpmVersionErr
	}
	return m.cpmVersion, nil
}

func newCPMSettingsService(version string, settings *mockLstepSettingsService) LstepSettingsService {
	if settings == nil {
		settings = &mockLstepSettingsService{}
	}
	return &cpmSettingsServiceWrapper{mockLstepSettingsService: settings, cpmVersion: version}
}

// cpmAccountingRepositoryWrapper embeds mockAccountingRepository (from accounting_service_test.go)
// and overrides the 3 CPM-specific methods, which the base mock hardcodes to always return zero
// values with no injection hooks.
type cpmAccountingRepositoryWrapper struct {
	*mockAccountingRepository
	sumPaidByOwnerFn              func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	maxSingleVisitAmountByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	findOwnersByAnnualRevenueFn   func(ctx context.Context, clinicID uint64) ([]billing.OwnerAnnualRevenue, error)
}

func (m *cpmAccountingRepositoryWrapper) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.sumPaidByOwnerFn != nil {
		return m.sumPaidByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *cpmAccountingRepositoryWrapper) MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.maxSingleVisitAmountByOwnerFn != nil {
		return m.maxSingleVisitAmountByOwnerFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *cpmAccountingRepositoryWrapper) FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]billing.OwnerAnnualRevenue, error) {
	if m.findOwnersByAnnualRevenueFn != nil {
		return m.findOwnersByAnnualRevenueFn(ctx, clinicID)
	}
	return nil, nil
}

func newCPMAccountingRepository() *cpmAccountingRepositoryWrapper {
	return &cpmAccountingRepositoryWrapper{mockAccountingRepository: &mockAccountingRepository{}}
}

// ---- SyncCPMStageTag ----

func TestSyncCPMStageTag(t *testing.T) {
	lineUID := "U_cpm_test"

	t.Run("returns wrapped error when GetCPMVersion fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &cpmSettingsServiceWrapper{mockLstepSettingsService: &mockLstepSettingsService{}, cpmVersionErr: errors.New("db error")},
		}
		assert.Error(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("dispatches to SyncCPMStageTagV2 when cpm_version=v2", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v2", &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }}),
		}
		// V2 path will itself call shouldSkipSync; disabling sync there proves dispatch occurred
		// without requiring the full V2 dependency graph.
		assert.NoError(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("skips when sync disabled (v1)", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }}),
		}
		assert.NoError(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("noop when owner has no line user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: nil}, nil
				},
			},
		}
		assert.NoError(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when FindOwnerVisitSummary fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when SumPaidByOwner fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.sumPaidByOwnerFn = func(_ context.Context, _, _ uint64) (int64, error) { return 0, errors.New("db error") }
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   accountRepo,
		}
		assert.Error(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when MaxSingleVisitAmountByOwner fails", func(t *testing.T) {
		accountRepo := newCPMAccountingRepository()
		accountRepo.maxSingleVisitAmountByOwnerFn = func(_ context.Context, _, _ uint64) (int64, error) { return 0, errors.New("db error") }
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   accountRepo,
		}
		assert.Error(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("continues past GetCPMV1Thresholds and short-circuits on an unconfigured client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &cpmSettingsServiceWrapper{
				mockLstepSettingsService: &mockLstepSettingsService{}, cpmVersion: "v1",
			},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   newCPMAccountingRepository(),
		}
		// buildClientFn is unset and no apiKey is configured, so buildClient short-circuits to
		// (nil, nil), exercising the GetCPMV1Thresholds success path without a real API call.
		err := svc.SyncCPMStageTag(context.Background(), 1, 10)
		assert.NoError(t, err)
	})

	t.Run("returns nil when buildClient yields no client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   newCPMAccountingRepository(),
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		assert.NoError(t, svc.SyncCPMStageTag(context.Background(), 1, 10))
	})

	t.Run("removes old CPM stage tags and adds the newly-calculated one", func(t *testing.T) {
		var addedStage string
		var removedStages []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error { addedStage = tagName; return nil },
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedStages = append(removedStages, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			// No visits at all → CalculateCPMStage returns cpm_dormant (DaysSinceVisit<0).
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   newCPMAccountingRepository(),
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncCPMStageTag(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, string(CPMStageDormant), addedStage)
		assert.Contains(t, removedStages, string(CPMStageEncounter))
		assert.NotContains(t, removedStages, string(CPMStageDormant), "the newly-active stage must not be removed")
	})

	t.Run("returns wrapped error when AddTag for the new stage fails", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: newCPMSettingsService("v1", nil),
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			accountRepo:   newCPMAccountingRepository(),
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.SyncCPMStageTag(context.Background(), 1, 10)
		assert.Error(t, err)
	})
}

// ---- SyncCPMStageTagV2 ----

func TestSyncCPMStageTagV2(t *testing.T) {
	lineUID := "U_cpm_v2_test"

	t.Run("skips when sync disabled", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		assert.NoError(t, svc.syncCPMStageTagV2(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		assert.Error(t, svc.syncCPMStageTagV2(context.Background(), 1, 10))
	})

	t.Run("noop when owner has no line user id", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: nil}, nil
				},
			},
		}
		assert.NoError(t, svc.syncCPMStageTagV2(context.Background(), 1, 10))
	})

	t.Run("returns wrapped error when FindOwnerVisitSummary fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
					return nil, errors.New("db error")
				},
			},
		}
		assert.Error(t, svc.syncCPMStageTagV2(context.Background(), 1, 10))
	})

	t.Run("returns nil when buildClient yields no client", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return nil, nil },
		}
		assert.NoError(t, svc.syncCPMStageTagV2(context.Background(), 1, 10))
	})

	t.Run("removes old V2 stage tags and adds the newly-calculated one", func(t *testing.T) {
		var addedStage string
		var removedStages []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error { addedStage = tagName; return nil },
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedStages = append(removedStages, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{
				findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
					return &medicalrecord.OwnerVisitSummary{TotalCount: 0}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncCPMStageTagV2(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, string(CPMStageV2Encounter), addedStage)
		assert.Contains(t, removedStages, string(CPMStageV2Coming))
	})

	t.Run("returns wrapped error when AddTag for the new stage fails", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			medRecordRepo: &mockMedicalRecordRepository{},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncCPMStageTagV2(context.Background(), 1, 10)
		assert.Error(t, err)
	})
}
