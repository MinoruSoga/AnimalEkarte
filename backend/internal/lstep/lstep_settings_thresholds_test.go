package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// This file tests the threshold-getter and sync-enabled methods implemented in
// lstep_settings_thresholds.go. It reuses mockLstepSettingsRepository / mockLstepSyncSettingsRepository
// (lstep_settings_service_test.go) and mockClinicSettingsRepository (closing_settings_service_test.go),
// which are already defined elsewhere in this package.

// ---- GetCPMVersion ----

func TestLstepSettingsService_GetCPMVersion(t *testing.T) {
	tests := []struct {
		name        string
		nilRepo     bool
		csRepo      *mockClinicSettingsRepository
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "nil clinicSettingsRepo returns v1",
			nilRepo:     true,
			wantVersion: "v1",
		},
		{
			name: "empty CPMVersion in DB returns v1",
			csRepo: &mockClinicSettingsRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
					return &model.ClinicSettings{CPMVersion: ""}, nil
				},
			},
			wantVersion: "v1",
		},
		{
			name: "non-empty CPMVersion returns DB value",
			csRepo: &mockClinicSettingsRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
					return &model.ClinicSettings{CPMVersion: "v2"}, nil
				},
			},
			wantVersion: "v2",
		},
		{
			name: "repo error is wrapped",
			csRepo: &mockClinicSettingsRepository{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var svc LstepSettingsService
			if tt.nilRepo {
				svc = NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
			} else {
				svc = NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, tt.csRepo)
			}

			version, err := svc.GetCPMVersion(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, version)
			}
		})
	}
}

// ---- GetCPMV1Thresholds ----

func TestLstepSettingsService_GetCPMV1Thresholds(t *testing.T) {
	t.Run("nil clinicSettingsRepo returns defaults", func(t *testing.T) {
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		got, err := svc.GetCPMV1Thresholds(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, model.CPMV1Thresholds{}.WithDefaults(), got)
	})

	t.Run("zero DB values are backfilled with defaults", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		got, err := svc.GetCPMV1Thresholds(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, model.CPMV1Thresholds{}.WithDefaults(), got)
	})

	t.Run("DB values are used as-is when set", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					CPMV1DormantDays:      300,
					CPMV1NoahDays:         400,
					CPMV1NoahAnnualVisits: 5,
					CPMV1NoahLTV:          100_000,
					CPMV1CoreDays:         200,
					CPMV1CoreAnnualVisits: 3,
					CPMV1CoreLTV:          60_000,
					CPMV1SpotMinAmount:    40_000,
					CPMV1SpotInactiveDays: 120,
					CPMV1GrowingMaxDays:   100,
					CPMV1GrowingMinVisits: 3,
					CPMV1GrowingMaxVisits: 4,
					CPMV1LTVBreakLow:      25_000,
				}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		got, err := svc.GetCPMV1Thresholds(context.Background(), 1)
		assert.NoError(t, err)
		want := model.CPMV1Thresholds{
			DormantDays:      300,
			NoahDays:         400,
			NoahAnnualVisits: 5,
			NoahLTV:          100_000,
			CoreDays:         200,
			CoreAnnualVisits: 3,
			CoreLTV:          60_000,
			SpotMinAmount:    40_000,
			SpotInactiveDays: 120,
			GrowingMaxDays:   100,
			GrowingMinVisits: 3,
			GrowingMaxVisits: 4,
			LTVBreakLow:      25_000,
		}
		assert.Equal(t, want, got)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		_, err := svc.GetCPMV1Thresholds(context.Background(), 1)
		assert.Error(t, err)
	})
}

// ---- GetHealthPreventionThresholds ----

func TestLstepSettingsService_GetHealthPreventionThresholds(t *testing.T) {
	t.Run("nil clinicSettingsRepo returns defaults", func(t *testing.T) {
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
		got, err := svc.GetHealthPreventionThresholds(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, model.HealthPreventionThresholds{}.WithDefaults(), got)
	})

	t.Run("zero DB values are backfilled with defaults", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		got, err := svc.GetHealthPreventionThresholds(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, model.HealthPreventionThresholds{}.WithDefaults(), got)
	})

	t.Run("DB values are used as-is when set", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					HealthPreventionLookbackDays: 400,
					VaccineDeadlineDays:          45,
				}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		got, err := svc.GetHealthPreventionThresholds(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, model.HealthPreventionThresholds{LookbackDays: 400, VaccineDeadline: 45}, got)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
		_, err := svc.GetHealthPreventionThresholds(context.Background(), 1)
		assert.Error(t, err)
	})
}

// ---- IsSyncEnabled: additional branch (nil syncSettingsRepo) ----
// TestIsSyncEnabled in lstep_settings_service_test.go already covers the not-found,
// true, and repo-error branches; this adds the nil-repo branch that was still uncovered.

func TestLstepSettingsService_IsSyncEnabled_NilSyncSettingsRepo(t *testing.T) {
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, nil, nil, nil, nil)
	enabled, err := svc.IsSyncEnabled(context.Background(), 1)
	assert.NoError(t, err)
	assert.False(t, enabled)
}

// ---- updateSyncEnabled: additional branches ----
// updateSyncEnabled is unexported, so it is exercised directly via a struct literal rather
// than through the constructor (consistent with lstepLifecycleService tests elsewhere in
// this package that construct the concrete struct directly for unexported-method coverage).

func TestLstepSettingsService_UpdateSyncEnabled_FindByClinicIDError(t *testing.T) {
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepSettingsService{syncSettingsRepo: syncRepo}

	err := svc.updateSyncEnabled(context.Background(), 1, true)

	assert.Error(t, err)
}

func TestLstepSettingsService_UpdateSyncEnabled_UpsertError(t *testing.T) {
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return nil, apperrors.WrapNotFound("lstep_settings", "clinic_id=1")
		},
		upsertFn: func(_ context.Context, _ *model.LstepSettings) (*model.LstepSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepSettingsService{syncSettingsRepo: syncRepo}

	err := svc.updateSyncEnabled(context.Background(), 1, true)

	assert.Error(t, err)
}

func TestLstepSettingsService_UpdateSyncEnabled_NotFoundRecordCreatesNew(t *testing.T) {
	var upserted *model.LstepSettings
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return nil, apperrors.WrapNotFound("lstep_settings", "clinic_id=1")
		},
		upsertFn: func(_ context.Context, s *model.LstepSettings) (*model.LstepSettings, error) {
			upserted = s
			return s, nil
		},
	}
	svc := &lstepSettingsService{syncSettingsRepo: syncRepo}

	err := svc.updateSyncEnabled(context.Background(), 1, true)

	assert.NoError(t, err)
	if assert.NotNil(t, upserted) {
		assert.True(t, upserted.IsSyncEnabled)
		assert.NotNil(t, upserted.SyncEnabledAt)
	}
}
