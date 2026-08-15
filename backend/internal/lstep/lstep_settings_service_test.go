package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
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

func (m *mockLstepSettingsRepository) FindCredentialByClinicServiceKey(
	_ context.Context,
	clinicID uint64,
	service, keyName string,
) (*model.ClinicIntegration, error) {
	return &model.ClinicIntegration{
		ClinicID: clinicID,
		Service:  service,
		KeyName:  keyName,
	}, nil
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
			svc := NewLstepSettingsService(tt.repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
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
			svc := NewLstepSettingsService(tt.repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
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
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)
		_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{}, nil)
		assert.NoError(t, err)
		assert.False(t, upsertCalled)
	})
}

// TestGetDormantThresholds_DBValues: DB に値があれば DB 値を返す
func TestGetDormantThresholds_DBValues(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return &model.ClinicSettings{
				DormantPrevention180Days: 175,
				DormantPrevention210Days: 200,
				DormantPrevention240Days: 230,
				DormantPrevention365Days: 350,
			}, nil
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	got, err := svc.GetDormantThresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.DormantThresholds{Stage180: 175, Stage210: 200, Stage240: 230, Stage365: 350}, got)
}

// TestGetDormantThresholds_ZeroFallback: DB 値が 0 ならデフォルト補完される
func TestGetDormantThresholds_ZeroFallback(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return &model.ClinicSettings{
				DormantPrevention180Days: 0,
				DormantPrevention210Days: 0,
				DormantPrevention240Days: 0,
				DormantPrevention365Days: 0,
			}, nil
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	got, err := svc.GetDormantThresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.DormantThresholds{Stage180: 180, Stage210: 210, Stage240: 240, Stage365: 365}, got)
}

// TestGetDormantThresholds_NilRepo: clinicSettingsRepo が nil の場合はデフォルト値を返す
func TestGetDormantThresholds_NilRepo(t *testing.T) {
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
	got, err := svc.GetDormantThresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.DormantThresholds{Stage180: 180, Stage210: 210, Stage240: 240, Stage365: 365}, got)
}

// TestGetDormantThresholds_RepoError: リポジトリエラー時はエラーを返す
func TestGetDormantThresholds_RepoError(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	_, err := svc.GetDormantThresholds(context.Background(), 1)
	assert.Error(t, err)
}

// TestGetCPMV2Thresholds_DBValues: DB に値があれば DB 値を返す
func TestGetCPMV2Thresholds_DBValues(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return &model.ClinicSettings{
				CPMV2ComingThreshold: 3,
				CPMV2GoodThreshold:   6,
				CPMV2FamilyThreshold: 10,
				CPMV2NoahThreshold:   15,
			}, nil
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	got, err := svc.GetCPMV2Thresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.CPMV2Thresholds{Coming: 3, Good: 6, Family: 10, Noah: 15}, got)
}

// TestGetCPMV2Thresholds_ZeroFallback: DB 値が 0 ならデフォルト補完される
func TestGetCPMV2Thresholds_ZeroFallback(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return &model.ClinicSettings{}, nil
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	got, err := svc.GetCPMV2Thresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.CPMV2Thresholds{Coming: 2, Good: 4, Family: 8, Noah: 13}, got)
}

// TestGetCPMV2Thresholds_NilRepo: clinicSettingsRepo が nil の場合はデフォルト値を返す
func TestGetCPMV2Thresholds_NilRepo(t *testing.T) {
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, nil)
	got, err := svc.GetCPMV2Thresholds(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, model.CPMV2Thresholds{Coming: 2, Good: 4, Family: 8, Noah: 13}, got)
}

// TestGetCPMV2Thresholds_RepoError: リポジトリエラー時はエラーを返す
func TestGetCPMV2Thresholds_RepoError(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)
	_, err := svc.GetCPMV2Thresholds(context.Background(), 1)
	assert.Error(t, err)
}

// ================================================================
// GetSettings: syncSettingsRepo / clinicSettingsRepo 分岐の網羅
// ================================================================

func TestGetSettings_AppliesSyncSettingsWhenPresent(t *testing.T) {
	syncEnabledAt := time.Now()
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return &model.LstepSettings{IsSyncEnabled: true, SyncEnabledAt: &syncEnabledAt}, nil
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)

	res, err := svc.GetSettings(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, res.IsSyncEnabled)
	if assert.NotNil(t, res.SyncEnabledAt) {
		assert.True(t, res.SyncEnabledAt.Equal(syncEnabledAt))
	}
}

func TestGetSettings_SyncSettingsNotFoundIsNotError(t *testing.T) {
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return nil, apperrors.WrapNotFound("lstep_settings", "clinic_id=1")
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)

	res, err := svc.GetSettings(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, res.IsSyncEnabled)
}

func TestGetSettings_SyncSettingsRepoErrorPropagates(t *testing.T) {
	syncRepo := &mockLstepSyncSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LstepSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, syncRepo, nil, nil, nil)

	_, err := svc.GetSettings(context.Background(), 1)
	require.Error(t, err)
}

func TestGetSettings_ClinicSettingsRepoErrorPropagates(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{
		findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)

	_, err := svc.GetSettings(context.Background(), 1)
	require.Error(t, err)
}

func TestGetSettings_ClinicSettingsAppliesDefaultsAndCPMVersion(t *testing.T) {
	t.Run("CPMVersion 空文字は v1 にデフォルトされる", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{}, nil // 全フィールドゼロ値
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)

		res, err := svc.GetSettings(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "v1", res.CPMVersion)
		assert.Equal(t, 180, res.DormantPrevention180Days)
		assert.Equal(t, 2, res.CPMV2ComingThreshold)
		assert.NotZero(t, res.CPMV1NoahDays)
		assert.NotZero(t, res.HealthPreventionLookbackDays)
	})

	t.Run("CPMVersion は DB 値が優先される", func(t *testing.T) {
		csRepo := &mockClinicSettingsRepository{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.ClinicSettings, error) {
				return &model.ClinicSettings{CPMVersion: "v2"}, nil
			},
		}
		svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)

		res, err := svc.GetSettings(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "v2", res.CPMVersion)
	})
}

func TestGetSettings_DecryptErrorPropagates(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	repo := &mockLstepSettingsRepository{
		findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
			return []*model.ClinicIntegration{
				{KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "not-valid-base64!!", UpdatedAt: time.Now()},
			}, nil
		},
	}
	svc := NewLstepSettingsService(repo, nil, cipher, nil, nil)

	res, err := svc.GetSettings(context.Background(), 1)
	require.Error(t, err, "LSB-04: 復号失敗を握り潰さず呼び出し元へ返す")
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "decrypt")
}

// ================================================================
// buildLstepSettingsResponse: 直接単体テスト
// ================================================================

func TestBuildLstepSettingsResponse(t *testing.T) {
	now := time.Now()

	t.Run("空map -> IsConfigured=false・マスク値も空", func(t *testing.T) {
		resp := buildLstepSettingsResponse(map[string]string{}, &now)
		assert.False(t, resp.IsConfigured)
		assert.Empty(t, resp.LstepAPIKeyMasked)
		assert.Empty(t, resp.LineChannelAccessTokenMasked)
		assert.Empty(t, resp.LineChannelSecretMasked)
		if assert.NotNil(t, resp.LastUpdatedAt) {
			assert.True(t, resp.LastUpdatedAt.Equal(now))
		}
	})

	t.Run("apiKeyのみ設定 -> IsConfigured=true", func(t *testing.T) {
		resp := buildLstepSettingsResponse(map[string]string{model.IntegrationKeyLstepAPIKey: "secret"}, nil)
		assert.True(t, resp.IsConfigured)
		assert.NotEmpty(t, resp.LstepAPIKeyMasked)
		assert.Nil(t, resp.LastUpdatedAt)
	})

	t.Run("tokenのみ設定 -> IsConfigured=true", func(t *testing.T) {
		resp := buildLstepSettingsResponse(map[string]string{model.IntegrationKeyLineChannelAccessToken: "token"}, nil)
		assert.True(t, resp.IsConfigured)
		assert.NotEmpty(t, resp.LineChannelAccessTokenMasked)
	})

	t.Run("全フィールドが正しくマッピングされる", func(t *testing.T) {
		kv := map[string]string{
			model.IntegrationKeyLstepAPIKey:            "api-key",
			model.IntegrationKeyLstepBaseURL:           "https://example.com",
			model.IntegrationKeyLineChannelAccessToken: "token",
			model.IntegrationKeyLineChannelSecret:      "secret",
			model.IntegrationKeyLiffID:                 "liff-1",
			model.IntegrationKeyLineAccountName:        "アカウント名",
		}
		resp := buildLstepSettingsResponse(kv, nil)
		assert.Equal(t, "https://example.com", resp.LstepBaseURL)
		assert.Equal(t, "liff-1", resp.LiffID)
		assert.Equal(t, "アカウント名", resp.LineAccountName)
		assert.NotEmpty(t, resp.LineChannelSecretMasked)
	})
}

// ================================================================
// UpdateSettings: 未カバー分岐の網羅
// ================================================================

func TestUpdateSettings_IntegrationCredentialsError(t *testing.T) {
	repo := &mockLstepSettingsRepository{
		upsertFn: func(_ context.Context, _ *model.ClinicIntegration) error {
			return errors.New("db error")
		},
	}
	svc := NewLstepSettingsService(repo, &mockLstepSyncSettingsRepository{}, nil, nil, nil)

	_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{LstepAPIKey: "new-key"}, nil)
	require.Error(t, err)
}

func TestUpdateSettings_ClinicSyncConfigError(t *testing.T) {
	csRepo := &mockClinicSettingsRepository{}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, nil, csRepo)

	badVersion := "v99"
	_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{CPMVersion: &badVersion}, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestUpdateSettings_SkipsSyncEnabledWhenRepoNil(t *testing.T) {
	enabled := true
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, nil, nil, nil, nil)

	_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{IsSyncEnabled: &enabled}, nil)
	require.NoError(t, err, "syncSettingsRepo が nil なら IsSyncEnabled 更新をスキップする")
}

func TestUpdateSettings_AuditLogSuccess(t *testing.T) {
	audit := &mockAuditService{}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, audit, nil)

	staffID := uint64(5)
	_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{}, &staffID)
	assert.NoError(t, err)
}

func TestUpdateSettings_AuditLogFailureIsBestEffort(t *testing.T) {
	audit := &mockAuditService{logLstepOperationErr: errors.New("audit db down")}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, audit, nil)

	_, err := svc.UpdateSettings(context.Background(), 1, &UpdateLstepSettingsInput{}, nil)
	assert.NoError(t, err, "監査ログ失敗は UpdateSettings 全体を失敗させない（best-effort）")
}

// ================================================================
// DeleteSettings: auditSvc 分岐の網羅
// ================================================================

func TestDeleteSettings_AuditLogSuccess(t *testing.T) {
	audit := &mockAuditService{}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, audit, nil)

	err := svc.DeleteSettings(context.Background(), 1, nil)
	assert.NoError(t, err)
}

func TestDeleteSettings_AuditLogFailureIsBestEffort(t *testing.T) {
	audit := &mockAuditService{logLstepOperationErr: errors.New("audit db down")}
	svc := NewLstepSettingsService(&mockLstepSettingsRepository{}, &mockLstepSyncSettingsRepository{}, nil, audit, nil)

	err := svc.DeleteSettings(context.Background(), 1, nil)
	assert.NoError(t, err, "監査ログ失敗は DeleteSettings 全体を失敗させない（best-effort）")
}

// mockClinicSettingsRepository — lstepClinicSettingsRepo view の最小モック
// （service/closing_settings_service_test.go の同名 mock の view 型版複製）。
type mockClinicSettingsRepository struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
}

func (m *mockClinicSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.ClinicSettings{ClinicID: clinicID}, nil
}

func (m *mockClinicSettingsRepository) UpdateCPMVersion(_ context.Context, _ uint64, _ string) error {
	return nil
}

func (m *mockClinicSettingsRepository) UpdateCPMV1Thresholds(_ context.Context, _ uint64, _ model.CPMV1Thresholds) error {
	return nil
}

func (m *mockClinicSettingsRepository) UpdateCPMV2Thresholds(_ context.Context, _ uint64, _ model.CPMV2Thresholds) error {
	return nil
}

func (m *mockClinicSettingsRepository) UpdateDormantThresholds(_ context.Context, _ uint64, _ model.DormantThresholds) error {
	return nil
}

func (m *mockClinicSettingsRepository) UpdateHealthPreventionThresholds(_ context.Context, _ uint64, _ model.HealthPreventionThresholds) error {
	return nil
}

// mockAuditService — lstepAuditLogger view の最小モック。
type mockAuditService struct {
	logLstepOperationFn             func(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
	logLstepOperationErr            error
	logLstepOperationWithMetadataFn func(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
	logEntryTxErr                   error
	entries                         []*LifecycleAuditEntry
}

func (m *mockAuditService) LogEntryTx(_ context.Context, entry *LifecycleAuditEntry) error {
	if m.logEntryTxErr != nil {
		return m.logEntryTxErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditService) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error {
	if m.logLstepOperationWithMetadataFn != nil {
		return m.logLstepOperationWithMetadataFn(ctx, clinicID, actorID, action, resource, resourceID, metadata)
	}
	return m.logLstepOperationErr
}

func (m *mockAuditService) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
	if m.logLstepOperationFn != nil {
		return m.logLstepOperationFn(ctx, clinicID, actorID, action, resource, resourceID)
	}
	return m.logLstepOperationErr
}
