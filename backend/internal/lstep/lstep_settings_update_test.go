package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock ClinicSettingsRepository (lstep_settings_update.go 専用) ----
//
// closing_settings_service_test.go の mockClinicSettingsRepository は
// UpdateCPMVersion 等の閾値更新系メソッドを実装していない（埋め込みインターフェースの
// ゼロ値のまま）ため、lstep_settings_update.go のテストには使えない。専用のフル実装モックを用意する。
type mockClinicSettingsUpdateRepository struct {
	findByClinicIDFn                       func(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	saveFn                                 func(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error)
	updateCPMVersionFn                     func(ctx context.Context, clinicID uint64, version string) error
	updateDormantThresholdsFn              func(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error
	updateCPMV2ThresholdsFn                func(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error
	updateCPMV1ThresholdsFn                func(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error
	updateHealthPreventionThresholdsFn     func(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error
	updateCPMVersionCalled                 bool
	updateDormantThresholdsCalled          bool
	updateCPMV2ThresholdsCalled            bool
	updateCPMV1ThresholdsCalled            bool
	updateHealthPreventionThresholdsCalled bool
}

func (m *mockClinicSettingsUpdateRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.ClinicSettings{ClinicID: clinicID}, nil
}

func (m *mockClinicSettingsUpdateRepository) Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	if m.saveFn != nil {
		return m.saveFn(ctx, clinicID, s)
	}
	return s, nil
}

func (m *mockClinicSettingsUpdateRepository) UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error {
	m.updateCPMVersionCalled = true
	if m.updateCPMVersionFn != nil {
		return m.updateCPMVersionFn(ctx, clinicID, version)
	}
	return nil
}

func (m *mockClinicSettingsUpdateRepository) UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error {
	m.updateDormantThresholdsCalled = true
	if m.updateDormantThresholdsFn != nil {
		return m.updateDormantThresholdsFn(ctx, clinicID, thresholds)
	}
	return nil
}

func (m *mockClinicSettingsUpdateRepository) UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error {
	m.updateCPMV2ThresholdsCalled = true
	if m.updateCPMV2ThresholdsFn != nil {
		return m.updateCPMV2ThresholdsFn(ctx, clinicID, thresholds)
	}
	return nil
}

func (m *mockClinicSettingsUpdateRepository) UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error {
	m.updateCPMV1ThresholdsCalled = true
	if m.updateCPMV1ThresholdsFn != nil {
		return m.updateCPMV1ThresholdsFn(ctx, clinicID, thresholds)
	}
	return nil
}

func (m *mockClinicSettingsUpdateRepository) UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error {
	m.updateHealthPreventionThresholdsCalled = true
	if m.updateHealthPreventionThresholdsFn != nil {
		return m.updateHealthPreventionThresholdsFn(ctx, clinicID, thresholds)
	}
	return nil
}

// ---- updateIntegrationCredentials ----

func TestLstepSettingsService_UpdateIntegrationCredentials(t *testing.T) {
	tests := []struct {
		name        string
		input       *UpdateLstepSettingsInput
		upsertErr   error
		wantErr     bool
		wantUpserts int
	}{
		{
			name:        "all fields empty skips upsert entirely",
			input:       &UpdateLstepSettingsInput{},
			wantErr:     false,
			wantUpserts: 0,
		},
		{
			name: "single field set upserts once",
			input: &UpdateLstepSettingsInput{
				LstepBaseURL: "https://api.example.com",
			},
			wantErr:     false,
			wantUpserts: 1,
		},
		{
			name: "multiple fields set upserts for each",
			input: &UpdateLstepSettingsInput{
				LstepAPIKey:            "secret-key",
				LstepBaseURL:           "https://api.example.com",
				LineChannelAccessToken: "line-token",
				LineChannelSecret:      "line-secret",
				LiffID:                 "liff-1234",
				LineAccountName:        "clinic-account",
			},
			wantErr:     false,
			wantUpserts: 6,
		},
		{
			name: "propagates upsert error and stops on first failure",
			input: &UpdateLstepSettingsInput{
				LstepAPIKey:  "secret-key",
				LstepBaseURL: "https://api.example.com",
			},
			upsertErr:   errors.New("db error"),
			wantErr:     true,
			wantUpserts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upsertCount := 0
			repo := &mockLstepSettingsRepository{
				upsertFn: func(_ context.Context, _ *model.ClinicIntegration) error {
					upsertCount++
					return tt.upsertErr
				},
			}
			svc := &lstepSettingsService{repo: repo}
			err := svc.updateIntegrationCredentials(context.Background(), 1, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantUpserts, upsertCount)
		})
	}
}

// ---- updateClinicSyncConfig ----

func TestLstepSettingsService_UpdateClinicSyncConfig(t *testing.T) {
	t.Run("nil clinicSettingsRepo skips entirely", func(t *testing.T) {
		svc := &lstepSettingsService{clinicSettingsRepo: nil}
		err := svc.updateClinicSyncConfig(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
	})

	t.Run("all nil fields is a no-op success", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateClinicSyncConfig(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
		assert.False(t, repo.updateCPMVersionCalled)
		assert.False(t, repo.updateDormantThresholdsCalled)
		assert.False(t, repo.updateCPMV2ThresholdsCalled)
		assert.False(t, repo.updateCPMV1ThresholdsCalled)
		assert.False(t, repo.updateHealthPreventionThresholdsCalled)
	})

	t.Run("error in first sub-update (cpm_version) stops the chain", func(t *testing.T) {
		invalidVersion := "v99"
		validDays := 10
		repo := &mockClinicSettingsUpdateRepository{}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateClinicSyncConfig(context.Background(), 1, &UpdateLstepSettingsInput{
			CPMVersion:               &invalidVersion,
			DormantPrevention180Days: &validDays,
		})
		assert.Error(t, err)
		assert.False(t, repo.updateDormantThresholdsCalled)
	})

	t.Run("success calls all five sub-updates in order", func(t *testing.T) {
		version := "v2"
		days := 200
		v2 := 3
		v1Days := 100
		v1LTV := int64(1000)
		lookback := 30

		repo := &mockClinicSettingsUpdateRepository{}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateClinicSyncConfig(context.Background(), 1, &UpdateLstepSettingsInput{
			CPMVersion:                   &version,
			DormantPrevention180Days:     &days,
			CPMV2ComingThreshold:         &v2,
			CPMV1DormantDays:             &v1Days,
			CPMV1NoahLTV:                 &v1LTV,
			HealthPreventionLookbackDays: &lookback,
		})
		assert.NoError(t, err)
		assert.True(t, repo.updateCPMVersionCalled)
		assert.True(t, repo.updateDormantThresholdsCalled)
		assert.True(t, repo.updateCPMV2ThresholdsCalled)
		assert.True(t, repo.updateCPMV1ThresholdsCalled)
		assert.True(t, repo.updateHealthPreventionThresholdsCalled)
	})
}

// ---- updateCPMVersion ----

func TestLstepSettingsService_UpdateCPMVersion(t *testing.T) {
	tests := []struct {
		name    string
		version *string
		repoErr error
		wantErr bool
	}{
		{name: "nil version is a no-op", version: nil, wantErr: false},
		{name: "v1 is valid", version: strPtr("v1"), wantErr: false},
		{name: "v2 is valid", version: strPtr("v2"), wantErr: false},
		{name: "invalid version returns invalid input error", version: strPtr("v3"), wantErr: true},
		{name: "propagates repository error", version: strPtr("v1"), repoErr: errors.New("db error"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicSettingsUpdateRepository{
				updateCPMVersionFn: func(_ context.Context, _ uint64, _ string) error {
					return tt.repoErr
				},
			}
			svc := &lstepSettingsService{clinicSettingsRepo: repo}
			err := svc.updateCPMVersion(context.Background(), 1, &UpdateLstepSettingsInput{CPMVersion: tt.version})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- updateDormantThresholds ----

func TestLstepSettingsService_UpdateDormantThresholds(t *testing.T) {
	valid := 200

	t.Run("all nil fields is a no-op without reading clinic settings", func(t *testing.T) {
		findCalled := false
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				findCalled = true
				return &model.ClinicSettings{}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateDormantThresholds(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
		assert.False(t, findCalled)
	})

	t.Run("propagates FindByClinicID error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateDormantThresholds(context.Background(), 1, &UpdateLstepSettingsInput{DormantPrevention180Days: &valid})
		assert.Error(t, err)
	})

	t.Run("merges with current values and calls UpdateDormantThresholds", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					DormantPrevention180Days: 180,
					DormantPrevention210Days: 210,
					DormantPrevention240Days: 240,
					DormantPrevention365Days: 365,
				}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		override := 199
		err := svc.updateDormantThresholds(context.Background(), 1, &UpdateLstepSettingsInput{DormantPrevention180Days: &override})
		assert.NoError(t, err)
		assert.True(t, repo.updateDormantThresholdsCalled)
	})

	t.Run("propagates repository write error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			updateDormantThresholdsFn: func(_ context.Context, _ uint64, _ model.DormantThresholds) error {
				return errors.New("write failed")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateDormantThresholds(context.Background(), 1, &UpdateLstepSettingsInput{DormantPrevention180Days: &valid})
		assert.Error(t, err)
	})

	invalid := 0
	fieldErrorCases := []struct {
		name   string
		mutate func(in *UpdateLstepSettingsInput)
	}{
		{"invalid 180 days", func(in *UpdateLstepSettingsInput) { in.DormantPrevention180Days = &invalid }},
		{"invalid 210 days", func(in *UpdateLstepSettingsInput) { in.DormantPrevention210Days = &invalid }},
		{"invalid 240 days", func(in *UpdateLstepSettingsInput) { in.DormantPrevention240Days = &invalid }},
		{"invalid 365 days", func(in *UpdateLstepSettingsInput) { in.DormantPrevention365Days = &invalid }},
	}
	for _, tt := range fieldErrorCases {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicSettingsUpdateRepository{}
			svc := &lstepSettingsService{clinicSettingsRepo: repo}
			input := &UpdateLstepSettingsInput{}
			tt.mutate(input)
			err := svc.updateDormantThresholds(context.Background(), 1, input)
			assert.Error(t, err)
			assert.False(t, repo.updateDormantThresholdsCalled)
		})
	}
}

// ---- updateCPMV2Thresholds ----

func TestLstepSettingsService_UpdateCPMV2Thresholds(t *testing.T) {
	valid := 5

	t.Run("all nil fields is a no-op", func(t *testing.T) {
		findCalled := false
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				findCalled = true
				return &model.ClinicSettings{}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV2Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
		assert.False(t, findCalled)
	})

	t.Run("propagates FindByClinicID error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV2Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{CPMV2ComingThreshold: &valid})
		assert.Error(t, err)
	})

	t.Run("merges with current values and calls UpdateCPMV2Thresholds", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					CPMV2ComingThreshold: 2,
					CPMV2GoodThreshold:   4,
					CPMV2FamilyThreshold: 8,
					CPMV2NoahThreshold:   13,
				}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV2Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{CPMV2ComingThreshold: &valid})
		assert.NoError(t, err)
		assert.True(t, repo.updateCPMV2ThresholdsCalled)
	})

	t.Run("propagates repository write error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			updateCPMV2ThresholdsFn: func(_ context.Context, _ uint64, _ model.CPMV2Thresholds) error {
				return errors.New("write failed")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV2Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{CPMV2ComingThreshold: &valid})
		assert.Error(t, err)
	})

	invalid := 0
	fieldErrorCases := []struct {
		name   string
		mutate func(in *UpdateLstepSettingsInput)
	}{
		{"invalid coming threshold", func(in *UpdateLstepSettingsInput) { in.CPMV2ComingThreshold = &invalid }},
		{"invalid good threshold", func(in *UpdateLstepSettingsInput) { in.CPMV2GoodThreshold = &invalid }},
		{"invalid family threshold", func(in *UpdateLstepSettingsInput) { in.CPMV2FamilyThreshold = &invalid }},
		{"invalid noah threshold", func(in *UpdateLstepSettingsInput) { in.CPMV2NoahThreshold = &invalid }},
	}
	for _, tt := range fieldErrorCases {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicSettingsUpdateRepository{}
			svc := &lstepSettingsService{clinicSettingsRepo: repo}
			input := &UpdateLstepSettingsInput{}
			tt.mutate(input)
			err := svc.updateCPMV2Thresholds(context.Background(), 1, input)
			assert.Error(t, err)
			assert.False(t, repo.updateCPMV2ThresholdsCalled)
		})
	}
}

// ---- updateCPMV1Thresholds ----

func TestLstepSettingsService_UpdateCPMV1Thresholds(t *testing.T) {
	validInt := 10
	validInt64 := int64(1000)

	t.Run("all nil fields is a no-op", func(t *testing.T) {
		findCalled := false
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				findCalled = true
				return &model.ClinicSettings{}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV1Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
		assert.False(t, findCalled)
	})

	t.Run("propagates FindByClinicID error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV1Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{CPMV1DormantDays: &validInt})
		assert.Error(t, err)
	})

	t.Run("merges with current values and calls UpdateCPMV1Thresholds", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					CPMV1DormantDays:      240,
					CPMV1NoahDays:         365,
					CPMV1NoahAnnualVisits: 3,
					CPMV1NoahLTV:          80000,
					CPMV1CoreDays:         180,
					CPMV1CoreAnnualVisits: 2,
					CPMV1CoreLTV:          50000,
					CPMV1SpotMinAmount:    30000,
					CPMV1SpotInactiveDays: 90,
					CPMV1GrowingMaxDays:   90,
					CPMV1GrowingMinVisits: 2,
					CPMV1GrowingMaxVisits: 3,
					CPMV1LTVBreakLow:      20000,
				}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV1Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{
			CPMV1DormantDays:      &validInt,
			CPMV1NoahDays:         &validInt,
			CPMV1NoahAnnualVisits: &validInt,
			CPMV1NoahLTV:          &validInt64,
			CPMV1CoreDays:         &validInt,
			CPMV1CoreAnnualVisits: &validInt,
			CPMV1CoreLTV:          &validInt64,
			CPMV1SpotMinAmount:    &validInt64,
			CPMV1SpotInactiveDays: &validInt,
			CPMV1GrowingMaxDays:   &validInt,
			CPMV1GrowingMinVisits: &validInt,
			CPMV1GrowingMaxVisits: &validInt,
			CPMV1LTVBreakLow:      &validInt64,
		})
		assert.NoError(t, err)
		assert.True(t, repo.updateCPMV1ThresholdsCalled)
	})

	t.Run("propagates repository write error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			updateCPMV1ThresholdsFn: func(_ context.Context, _ uint64, _ model.CPMV1Thresholds) error {
				return errors.New("write failed")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateCPMV1Thresholds(context.Background(), 1, &UpdateLstepSettingsInput{CPMV1DormantDays: &validInt})
		assert.Error(t, err)
	})

	invalidInt := 0
	invalidInt64 := int64(-1)
	fieldErrorCases := []struct {
		name   string
		mutate func(in *UpdateLstepSettingsInput)
	}{
		{"invalid dormant_days", func(in *UpdateLstepSettingsInput) { in.CPMV1DormantDays = &invalidInt }},
		{"invalid noah_days", func(in *UpdateLstepSettingsInput) { in.CPMV1NoahDays = &invalidInt }},
		{"invalid noah_annual_visits", func(in *UpdateLstepSettingsInput) { in.CPMV1NoahAnnualVisits = &invalidInt }},
		{"invalid noah_ltv", func(in *UpdateLstepSettingsInput) { in.CPMV1NoahLTV = &invalidInt64 }},
		{"invalid core_days", func(in *UpdateLstepSettingsInput) { in.CPMV1CoreDays = &invalidInt }},
		{"invalid core_annual_visits", func(in *UpdateLstepSettingsInput) { in.CPMV1CoreAnnualVisits = &invalidInt }},
		{"invalid core_ltv", func(in *UpdateLstepSettingsInput) { in.CPMV1CoreLTV = &invalidInt64 }},
		{"invalid spot_min_amount", func(in *UpdateLstepSettingsInput) { in.CPMV1SpotMinAmount = &invalidInt64 }},
		{"invalid spot_inactive_days", func(in *UpdateLstepSettingsInput) { in.CPMV1SpotInactiveDays = &invalidInt }},
		{"invalid growing_max_days", func(in *UpdateLstepSettingsInput) { in.CPMV1GrowingMaxDays = &invalidInt }},
		{"invalid growing_min_visits", func(in *UpdateLstepSettingsInput) { in.CPMV1GrowingMinVisits = &invalidInt }},
		{"invalid growing_max_visits", func(in *UpdateLstepSettingsInput) { in.CPMV1GrowingMaxVisits = &invalidInt }},
		{"invalid ltv_break_low", func(in *UpdateLstepSettingsInput) { in.CPMV1LTVBreakLow = &invalidInt64 }},
	}
	for _, tt := range fieldErrorCases {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicSettingsUpdateRepository{}
			svc := &lstepSettingsService{clinicSettingsRepo: repo}
			input := &UpdateLstepSettingsInput{}
			tt.mutate(input)
			err := svc.updateCPMV1Thresholds(context.Background(), 1, input)
			assert.Error(t, err)
			assert.False(t, repo.updateCPMV1ThresholdsCalled)
		})
	}
}

// ---- updateHealthPreventionThresholds ----

func TestLstepSettingsService_UpdateHealthPreventionThresholds(t *testing.T) {
	valid := 45

	t.Run("all nil fields is a no-op", func(t *testing.T) {
		findCalled := false
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				findCalled = true
				return &model.ClinicSettings{}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateHealthPreventionThresholds(context.Background(), 1, &UpdateLstepSettingsInput{})
		assert.NoError(t, err)
		assert.False(t, findCalled)
	})

	t.Run("propagates FindByClinicID error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return nil, errors.New("db error")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateHealthPreventionThresholds(context.Background(), 1, &UpdateLstepSettingsInput{HealthPreventionLookbackDays: &valid})
		assert.Error(t, err)
	})

	t.Run("merges with current values and calls UpdateHealthPreventionThresholds", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{
					HealthPreventionLookbackDays: 365,
					VaccineDeadlineDays:          60,
				}, nil
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateHealthPreventionThresholds(context.Background(), 1, &UpdateLstepSettingsInput{HealthPreventionLookbackDays: &valid})
		assert.NoError(t, err)
		assert.True(t, repo.updateHealthPreventionThresholdsCalled)
	})

	t.Run("propagates repository write error", func(t *testing.T) {
		repo := &mockClinicSettingsUpdateRepository{
			updateHealthPreventionThresholdsFn: func(_ context.Context, _ uint64, _ model.HealthPreventionThresholds) error {
				return errors.New("write failed")
			},
		}
		svc := &lstepSettingsService{clinicSettingsRepo: repo}
		err := svc.updateHealthPreventionThresholds(context.Background(), 1, &UpdateLstepSettingsInput{HealthPreventionLookbackDays: &valid})
		assert.Error(t, err)
	})

	invalid := 0
	fieldErrorCases := []struct {
		name   string
		mutate func(in *UpdateLstepSettingsInput)
	}{
		{"invalid lookback days", func(in *UpdateLstepSettingsInput) { in.HealthPreventionLookbackDays = &invalid }},
		{"invalid vaccine deadline days", func(in *UpdateLstepSettingsInput) { in.VaccineDeadlineDays = &invalid }},
	}
	for _, tt := range fieldErrorCases {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicSettingsUpdateRepository{}
			svc := &lstepSettingsService{clinicSettingsRepo: repo}
			input := &UpdateLstepSettingsInput{}
			tt.mutate(input)
			err := svc.updateHealthPreventionThresholds(context.Background(), 1, input)
			assert.Error(t, err)
			assert.False(t, repo.updateHealthPreventionThresholdsCalled)
		})
	}
}

// ---- applyPositiveInt / applyNonNegativeInt64 (pure helper functions) ----

func TestApplyPositiveInt(t *testing.T) {
	tests := []struct {
		name       string
		value      *int
		wantErr    bool
		wantTarget int
	}{
		{name: "nil value is a no-op", value: nil, wantErr: false, wantTarget: 5},
		{name: "zero is invalid (must be >= 1)", value: intPtr(0), wantErr: true, wantTarget: 5},
		{name: "negative is invalid", value: intPtr(-3), wantErr: true, wantTarget: 5},
		{name: "one is the minimum valid value", value: intPtr(1), wantErr: false, wantTarget: 1},
		{name: "positive value overwrites target", value: intPtr(100), wantErr: false, wantTarget: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := 5
			err := applyPositiveInt(tt.value, "test_field", &target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantTarget, target)
		})
	}
}

func TestApplyNonNegativeInt64(t *testing.T) {
	tests := []struct {
		name       string
		value      *int64
		wantErr    bool
		wantTarget int64
	}{
		{name: "nil value is a no-op", value: nil, wantErr: false, wantTarget: 5},
		{name: "negative is invalid", value: ptrInt64(-1), wantErr: true, wantTarget: 5},
		{name: "zero is the minimum valid value", value: ptrInt64(0), wantErr: false, wantTarget: 0},
		{name: "positive value overwrites target", value: ptrInt64(1000), wantErr: false, wantTarget: 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := int64(5)
			err := applyNonNegativeInt64(tt.value, "test_field", &target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantTarget, target)
		})
	}
}

// 注: strPtr/intPtr/ptrInt64 は medicine_service_test.go / aggregation_service_test.go /
// accounting_service_test.go に既存定義があるためここでは再定義しない（redeclared 回避）。
