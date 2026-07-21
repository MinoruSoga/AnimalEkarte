package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetCPMVersion はクリニックの CPM 判定方式を返す。レコード未存在または空文字時は "v1" を返す。
func (s *lstepSettingsService) GetCPMVersion(ctx context.Context, clinicID uint64) (string, error) {
	if s.clinicSettingsRepo == nil {
		return "v1", nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to find clinic settings")
	}
	if settings.CPMVersion == "" {
		return "v1", nil
	}
	return settings.CPMVersion, nil
}

// GetDormantThresholds はクリニックの dormant prevention 4 段階閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.DormantThresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for dormant thresholds", "clinic_id", clinicID, "error", err)
		return model.DormantThresholds{}, apperrors.Wrap(err, "failed to find clinic settings for dormant thresholds")
	}
	return model.DormantThresholds{
		Stage180: settings.DormantPrevention180Days,
		Stage210: settings.DormantPrevention210Days,
		Stage240: settings.DormantPrevention240Days,
		Stage365: settings.DormantPrevention365Days,
	}.WithDefaults(), nil
}

// GetCPMV2Thresholds はクリニックの CPM V2 来院回数閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetCPMV2Thresholds(ctx context.Context, clinicID uint64) (model.CPMV2Thresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.CPMV2Thresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for cpm v2 thresholds", "clinic_id", clinicID, "error", err)
		return model.CPMV2Thresholds{}, apperrors.Wrap(err, "failed to find clinic settings for cpm v2 thresholds")
	}
	return model.CPMV2Thresholds{
		Coming: settings.CPMV2ComingThreshold,
		Good:   settings.CPMV2GoodThreshold,
		Family: settings.CPMV2FamilyThreshold,
		Noah:   settings.CPMV2NoahThreshold,
	}.WithDefaults(), nil
}

// GetCPMV1Thresholds はクリニックの CPM V1 判定閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.CPMV1Thresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for cpm v1 thresholds", "clinic_id", clinicID, "error", err)
		return model.CPMV1Thresholds{}, apperrors.Wrap(err, "failed to find clinic settings for cpm v1 thresholds")
	}
	return model.CPMV1Thresholds{
		DormantDays:      settings.CPMV1DormantDays,
		NoahDays:         settings.CPMV1NoahDays,
		NoahAnnualVisits: settings.CPMV1NoahAnnualVisits,
		NoahLTV:          settings.CPMV1NoahLTV,
		CoreDays:         settings.CPMV1CoreDays,
		CoreAnnualVisits: settings.CPMV1CoreAnnualVisits,
		CoreLTV:          settings.CPMV1CoreLTV,
		SpotMinAmount:    settings.CPMV1SpotMinAmount,
		SpotInactiveDays: settings.CPMV1SpotInactiveDays,
		GrowingMaxDays:   settings.CPMV1GrowingMaxDays,
		GrowingMinVisits: settings.CPMV1GrowingMinVisits,
		GrowingMaxVisits: settings.CPMV1GrowingMaxVisits,
		LTVBreakLow:      settings.CPMV1LTVBreakLow,
	}.WithDefaults(), nil
}

// GetHealthPreventionThresholds はクリニックの健診・予防タグ判定閾値を返す。DB 値が 0 以下なら default で補完する。
func (s *lstepSettingsService) GetHealthPreventionThresholds(ctx context.Context, clinicID uint64) (model.HealthPreventionThresholds, error) {
	if s.clinicSettingsRepo == nil {
		return model.HealthPreventionThresholds{}.WithDefaults(), nil
	}
	settings, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get clinic settings for health prevention thresholds", "clinic_id", clinicID, "error", err)
		return model.HealthPreventionThresholds{}, apperrors.Wrap(err, "failed to find clinic settings for health prevention thresholds")
	}
	return model.HealthPreventionThresholds{
		LookbackDays:    settings.HealthPreventionLookbackDays,
		VaccineDeadline: settings.VaccineDeadlineDays,
	}.WithDefaults(), nil
}

// IsSyncEnabled はクリニックの同期有効フラグを返す。レコード未作成時は false を返す。
func (s *lstepSettingsService) IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error) {
	if s.syncSettingsRepo == nil {
		return false, nil
	}
	settings, err := s.syncSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return false, nil
		}
		return false, apperrors.Wrap(err, "failed to find lstep sync settings")
	}
	return settings.IsSyncEnabled, nil
}

// updateSyncEnabled は IsSyncEnabled の変化に応じて lstep_settings を Upsert する。
// false → true の場合のみ SyncEnabledAt を現在時刻でセットする。true → false では保持する。
func (s *lstepSettingsService) updateSyncEnabled(ctx context.Context, clinicID uint64, newEnabled bool) error {
	current, err := s.syncSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil && !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to find lstep sync settings", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to find lstep sync settings")
	}

	record := &model.LstepSettings{
		ClinicID:      clinicID,
		IsSyncEnabled: newEnabled,
	}

	if current != nil {
		record.SyncEnabledAt = current.SyncEnabledAt
	}

	// false → true の遷移時のみ SyncEnabledAt をセット
	currentEnabled := current != nil && current.IsSyncEnabled
	if !currentEnabled && newEnabled {
		now := time.Now()
		record.SyncEnabledAt = &now
	}

	if _, err := s.syncSettingsRepo.Upsert(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to upsert lstep sync settings", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to update lstep sync settings")
	}
	return nil
}
