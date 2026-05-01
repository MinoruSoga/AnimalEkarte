package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- mock LstepSettingsRepository ----

type mockLstepSettingsRepository struct {
	findByClinicAndServiceFn   func(ctx context.Context, clinicID uint64, svc string) ([]*model.ClinicIntegration, error)
	upsertFn                   func(ctx context.Context, integration *model.ClinicIntegration) error
	deleteByClinicAndServiceFn func(ctx context.Context, clinicID uint64, svc string) error
}

func (m *mockLstepSettingsRepository) FindByClinicAndService(ctx context.Context, clinicID uint64, svc string) ([]*model.ClinicIntegration, error) {
	if m.findByClinicAndServiceFn != nil {
		return m.findByClinicAndServiceFn(ctx, clinicID, svc)
	}
	return []*model.ClinicIntegration{}, nil
}
func (m *mockLstepSettingsRepository) Upsert(ctx context.Context, integration *model.ClinicIntegration) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, integration)
	}
	return nil
}
func (m *mockLstepSettingsRepository) DeleteByClinicAndService(ctx context.Context, clinicID uint64, svc string) error {
	if m.deleteByClinicAndServiceFn != nil {
		return m.deleteByClinicAndServiceFn(ctx, clinicID, svc)
	}
	return nil
}

// ---- mock LstepSyncSettingsRepository ----

type mockLstepSyncSettingsRepository struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LstepSettings, error)
	upsertFn         func(ctx context.Context, settings *model.LstepSettings) (*model.LstepSettings, error)
}

func (m *mockLstepSyncSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LstepSettings, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *mockLstepSyncSettingsRepository) Upsert(ctx context.Context, settings *model.LstepSettings) (*model.LstepSettings, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, settings)
	}
	return settings, nil
}

// ---- tests ----

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockLstepSettingsRepository
		wantErr bool
	}{
		{
			name: "success empty records",
			repo: &mockLstepSettingsRepository{},
		},
		{
			name: "repo error",
			repo: &mockLstepSettingsRepository{
				findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepSettingsService(tt.repo, &mockLstepSyncSettingsRepository{}, nil, nil)
			res, err := svc.GetSettings(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.False(t, res.IsConfigured)
			}
		})
	}
}

func TestDeleteSettings(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockLstepSettingsRepository
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockLstepSettingsRepository{},
		},
		{
			name: "repo error",
			repo: &mockLstepSettingsRepository{
				deleteByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewLstepSettingsService(tt.repo, &mockLstepSyncSettingsRepository{}, nil, nil)
			err := svc.DeleteSettings(context.Background(), 1, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIsSyncEnabled: レコード未作成は false、作成済みは IsSyncEnabled 値を返す
func TestIsSyncEnabled(t *testing.T) {
	t.Run("not found returns false without error", func(t *testing.T) {
		syncRepo := &mockLstepSyncSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
				return nil, apperrors.WrapNotFound("lstep_settings", "clinic_id=1")
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		enabled, err := svc.IsSyncEnabled(context.Background(), 1)
		assert.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("returns true when IsSyncEnabled=true", func(t *testing.T) {
		syncRepo := &mockLstepSyncSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
				return &model.LstepSettings{IsSyncEnabled: true}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		enabled, err := svc.IsSyncEnabled(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		syncRepo := &mockLstepSyncSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		_, err := svc.IsSyncEnabled(context.Background(), 1)
		assert.Error(t, err)
	})
}

// TestSyncEnabledAtLifecycle: false→true で SyncEnabledAt がセットされ、true→false では保持される
func TestSyncEnabledAtLifecycle(t *testing.T) {
	t.Run("false to true sets SyncEnabledAt", func(t *testing.T) {
		before := time.Now()
		var upserted *model.LstepSettings
		syncRepo := &mockLstepSyncSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
				return &model.LstepSettings{ClinicID: 1, IsSyncEnabled: false}, nil
			},
			upsertFn: func(_ context.Context, s *model.LstepSettings) (*model.LstepSettings, error) {
				upserted = s
				return s, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		enabled := true
		_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{IsSyncEnabled: &enabled}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, upserted)
		assert.True(t, upserted.IsSyncEnabled)
		assert.NotNil(t, upserted.SyncEnabledAt)
		assert.True(t, upserted.SyncEnabledAt.After(before))
	})

	t.Run("true to false preserves SyncEnabledAt", func(t *testing.T) {
		original := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		var upserted *model.LstepSettings
		syncRepo := &mockLstepSyncSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
				return &model.LstepSettings{ClinicID: 1, IsSyncEnabled: true, SyncEnabledAt: &original}, nil
			},
			upsertFn: func(_ context.Context, s *model.LstepSettings) (*model.LstepSettings, error) {
				upserted = s
				return s, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		disabled := false
		_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{IsSyncEnabled: &disabled}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, upserted)
		assert.False(t, upserted.IsSyncEnabled)
		assert.NotNil(t, upserted.SyncEnabledAt)
		assert.Equal(t, original, *upserted.SyncEnabledAt)
	})

	t.Run("nil IsSyncEnabled skips sync settings update", func(t *testing.T) {
		upsertCalled := false
		syncRepo := &mockLstepSyncSettingsRepository{
			upsertFn: func(_ context.Context, _ *model.LstepSettings) (*model.LstepSettings, error) {
				upsertCalled = true
				return nil, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil)
		_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{}, nil)
		assert.NoError(t, err)
		assert.False(t, upsertCalled)
	})
}

// TestAllClinicsFiltersBySyncEnabled: AllClinics バッチが is_sync_enabled=false のクリニックをスキップする
func TestAllClinicsFiltersBySyncEnabled(t *testing.T) {
	t.Run("RunNoShowCheckAllClinics skips disabled clinics", func(t *testing.T) {
		processed := make([]uint64, 0)
		clinicRepo := &mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 1}, {ID: 2}}, nil
			},
		}
		resRepo := &batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, clinicID uint64) ([]model.Reservation, error) {
				processed = append(processed, clinicID)
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, clinicID uint64) (bool, error) {
				return clinicID == 1, nil // clinic 1 のみ有効
			},
		}
		svc := NewLstepBatchService(resRepo, &batchMockTagSyncSvc{}, clinicRepo, &batchMockMedRecordRepo{}, &batchMockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []uint64{1}, processed)
	})

	t.Run("RunDormantDetectionAllClinics skips disabled clinics", func(t *testing.T) {
		processed := make([]uint64, 0)
		clinicRepo := &mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 10}, {ID: 20}}, nil
			},
		}
		medRepo := &batchMockMedRecordRepo{
			findDormantFn: func(_ context.Context, clinicID uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				processed = append(processed, clinicID)
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, clinicID uint64) (bool, error) {
				return clinicID == 20, nil // clinic 20 のみ有効
			},
		}
		svc := NewLstepBatchService(&batchMockReservationRepo{}, &batchMockTagSyncSvc{}, clinicRepo, medRepo, &batchMockAuditService{}, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []uint64{20}, processed)
	})
}
