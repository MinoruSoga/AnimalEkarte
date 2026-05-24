package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicSettingsRepository は締め時間設定のデータアクセスインターフェース
type ClinicSettingsRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error)
	// UpdateCPMVersion は cpm_version のみを対象とした UPSERT。
	UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error
	// UpdateDormantThresholds は dormant_prevention_*_days 4 カラムを対象とした UPSERT。
	UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error
	// UpdateCPMV2Thresholds は cpm_v2_*_threshold 4 カラムを対象とした UPSERT。
	UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error
	// UpdateCPMV1Thresholds は cpm_v1_* 13 カラムを対象とした UPSERT。
	UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error
	// UpdateHealthPreventionThresholds は health_prevention_lookback_days / vaccine_deadline_days を対象とした UPSERT。
	UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error
}

type clinicSettingsRepository struct{ db *gorm.DB }

// NewClinicSettingsRepository は ClinicSettingsRepository を初期化して返す
func NewClinicSettingsRepository(db *gorm.DB) ClinicSettingsRepository {
	return &clinicSettingsRepository{db: db}
}

func (r *clinicSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	var s model.ClinicSettings
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).First(&s).Error
	if err != nil {
		wrapped := apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
		if errors.Is(wrapped, apperrors.ErrNotFound) {
			// レコードがなければデフォルト値を返す
			return &model.ClinicSettings{
				ClinicID:            clinicID,
				ClosingAmPmBoundary: "14:00",
				ClosingWeekdayEnd:   "18:30",
				ClosingSundayEnd:    "17:30",
			}, nil
		}
		return nil, wrapped
	}
	return &s, nil
}

func (r *clinicSettingsRepository) Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"closing_am_pm_boundary",
				"closing_weekday_end",
				"closing_sunday_end",
				"closed_weekdays",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", s.ClinicID))
	}
	return s, nil
}

func (r *clinicSettingsRepository) UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error {
	s := &model.ClinicSettings{
		ClinicID:            clinicID,
		ClosingAmPmBoundary: "14:00",
		ClosingWeekdayEnd:   "18:30",
		ClosingSundayEnd:    "17:30",
		CPMVersion:          version,
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"cpm_version", "updated_at"}),
		}).
		Create(s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error {
	s := &model.ClinicSettings{
		ClinicID:                 clinicID,
		ClosingAmPmBoundary:      "14:00",
		ClosingWeekdayEnd:        "18:30",
		ClosingSundayEnd:         "17:30",
		DormantPrevention180Days: thresholds.Stage180,
		DormantPrevention210Days: thresholds.Stage210,
		DormantPrevention240Days: thresholds.Stage240,
		DormantPrevention365Days: thresholds.Stage365,
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"dormant_prevention_180_days",
				"dormant_prevention_210_days",
				"dormant_prevention_240_days",
				"dormant_prevention_365_days",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error {
	s := &model.ClinicSettings{
		ClinicID:             clinicID,
		ClosingAmPmBoundary:  "14:00",
		ClosingWeekdayEnd:    "18:30",
		ClosingSundayEnd:     "17:30",
		CPMV2ComingThreshold: thresholds.Coming,
		CPMV2GoodThreshold:   thresholds.Good,
		CPMV2FamilyThreshold: thresholds.Family,
		CPMV2NoahThreshold:   thresholds.Noah,
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"cpm_v2_coming_threshold",
				"cpm_v2_good_threshold",
				"cpm_v2_family_threshold",
				"cpm_v2_noah_threshold",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

//nolint:gocritic // hugeParam: thresholds は immutable DTO として interface に揃える設計
func (r *clinicSettingsRepository) UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error {
	s := &model.ClinicSettings{
		ClinicID:              clinicID,
		ClosingAmPmBoundary:   "14:00",
		ClosingWeekdayEnd:     "18:30",
		ClosingSundayEnd:      "17:30",
		CPMV1DormantDays:      thresholds.DormantDays,
		CPMV1NoahDays:         thresholds.NoahDays,
		CPMV1NoahAnnualVisits: thresholds.NoahAnnualVisits,
		CPMV1NoahLTV:          thresholds.NoahLTV,
		CPMV1CoreDays:         thresholds.CoreDays,
		CPMV1CoreAnnualVisits: thresholds.CoreAnnualVisits,
		CPMV1CoreLTV:          thresholds.CoreLTV,
		CPMV1SpotMinAmount:    thresholds.SpotMinAmount,
		CPMV1SpotInactiveDays: thresholds.SpotInactiveDays,
		CPMV1GrowingMaxDays:   thresholds.GrowingMaxDays,
		CPMV1GrowingMinVisits: thresholds.GrowingMinVisits,
		CPMV1GrowingMaxVisits: thresholds.GrowingMaxVisits,
		CPMV1LTVBreakLow:      thresholds.LTVBreakLow,
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"cpm_v1_dormant_days",
				"cpm_v1_noah_days",
				"cpm_v1_noah_annual_visits",
				"cpm_v1_noah_ltv",
				"cpm_v1_core_days",
				"cpm_v1_core_annual_visits",
				"cpm_v1_core_ltv",
				"cpm_v1_spot_min_amount",
				"cpm_v1_spot_inactive_days",
				"cpm_v1_growing_max_days",
				"cpm_v1_growing_min_visits",
				"cpm_v1_growing_max_visits",
				"cpm_v1_ltv_break_low",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error {
	s := &model.ClinicSettings{
		ClinicID:                     clinicID,
		ClosingAmPmBoundary:          "14:00",
		ClosingWeekdayEnd:            "18:30",
		ClosingSundayEnd:             "17:30",
		HealthPreventionLookbackDays: thresholds.LookbackDays,
		VaccineDeadlineDays:          thresholds.VaccineDeadline,
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"health_prevention_lookback_days",
				"vaccine_deadline_days",
				"updated_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}
