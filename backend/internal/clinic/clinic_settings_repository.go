package clinic

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type clinicSettingsRepository struct{ db *gorm.DB }

// defaultClosing* は 001_init.sql clinic_settings の DDL DEFAULT と一致させる（POC-10）。
// FindByClinicID フォールバックと Upsert 時の INSERT 既定値で共有し、リテラル重複を禁止する。
const (
	defaultClosingAmPmBoundary = "14:00"
	defaultClosingWeekdayEnd   = "18:30"
	defaultClosingSundayEnd    = "17:30"
)

func defaultClosingTimes(clinicID uint64) model.ClinicSettings {
	return model.ClinicSettings{
		ClinicID:            clinicID,
		ClosingAmPmBoundary: defaultClosingAmPmBoundary,
		ClosingWeekdayEnd:   defaultClosingWeekdayEnd,
		ClosingSundayEnd:    defaultClosingSundayEnd,
	}
}

func (r *clinicSettingsRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	var s model.ClinicSettings
	err := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).First(&s).Error
	if err != nil {
		wrapped := apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
		if errors.Is(wrapped, apperrors.ErrNotFound) {
			// レコードがなければデフォルト値を返す
			d := defaultClosingTimes(clinicID)
			return &d, nil
		}
		return nil, wrapped
	}
	return &s, nil
}

func (r *clinicSettingsRepository) Save(ctx context.Context, clinicID uint64, s *model.ClinicSettings) (*model.ClinicSettings, error) {
	// INSERT ... ON CONFLICT に Scopes(WHERE) は効かないため、書き込み先は arg の clinicID を正とする（POC-09）。
	s.ClinicID = clinicID
	err := persistence.DBOrTx(ctx, r.db).
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
	s := defaultClosingTimes(clinicID)
	s.CPMVersion = version
	// INSERT ON CONFLICT: tenant は s.ClinicID（= clinicID）で確定。Scopes は no-op のため付けない。
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"cpm_version", "updated_at"}),
		}).
		Create(&s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error {
	s := defaultClosingTimes(clinicID)
	s.DormantPrevention180Days = thresholds.Stage180
	s.DormantPrevention210Days = thresholds.Stage210
	s.DormantPrevention240Days = thresholds.Stage240
	s.DormantPrevention365Days = thresholds.Stage365
	err := persistence.DBOrTx(ctx, r.db).
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
		Create(&s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error {
	s := defaultClosingTimes(clinicID)
	s.CPMV2ComingThreshold = thresholds.Coming
	s.CPMV2GoodThreshold = thresholds.Good
	s.CPMV2FamilyThreshold = thresholds.Family
	s.CPMV2NoahThreshold = thresholds.Noah
	err := persistence.DBOrTx(ctx, r.db).
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
		Create(&s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

//nolint:gocritic // hugeParam: thresholds は immutable DTO として interface に揃える設計
func (r *clinicSettingsRepository) UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error {
	s := defaultClosingTimes(clinicID)
	s.CPMV1DormantDays = thresholds.DormantDays
	s.CPMV1NoahDays = thresholds.NoahDays
	s.CPMV1NoahAnnualVisits = thresholds.NoahAnnualVisits
	s.CPMV1NoahLTV = thresholds.NoahLTV
	s.CPMV1CoreDays = thresholds.CoreDays
	s.CPMV1CoreAnnualVisits = thresholds.CoreAnnualVisits
	s.CPMV1CoreLTV = thresholds.CoreLTV
	s.CPMV1SpotMinAmount = thresholds.SpotMinAmount
	s.CPMV1SpotInactiveDays = thresholds.SpotInactiveDays
	s.CPMV1GrowingMaxDays = thresholds.GrowingMaxDays
	s.CPMV1GrowingMinVisits = thresholds.GrowingMinVisits
	s.CPMV1GrowingMaxVisits = thresholds.GrowingMaxVisits
	s.CPMV1LTVBreakLow = thresholds.LTVBreakLow
	err := persistence.DBOrTx(ctx, r.db).
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
		Create(&s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *clinicSettingsRepository) UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error {
	s := defaultClosingTimes(clinicID)
	s.HealthPreventionLookbackDays = thresholds.LookbackDays
	s.VaccineDeadlineDays = thresholds.VaccineDeadline
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"health_prevention_lookback_days",
				"vaccine_deadline_days",
				"updated_at",
			}),
		}).
		Create(&s).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_settings", fmt.Sprintf("%d", clinicID))
	}
	return nil
}
