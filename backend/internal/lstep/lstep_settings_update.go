package lstep

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepSettingsService) updateIntegrationCredentials(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	pairs := []struct {
		keyName string
		value   string
	}{
		{model.IntegrationKeyLstepAPIKey, input.LstepAPIKey},
		{model.IntegrationKeyLstepBaseURL, input.LstepBaseURL},
		{model.IntegrationKeyLineChannelAccessToken, input.LineChannelAccessToken},
		{model.IntegrationKeyLineChannelSecret, input.LineChannelSecret},
		{model.IntegrationKeyLiffID, input.LiffID},
		{model.IntegrationKeyLineAccountName, input.LineAccountName},
	}

	if input.ClearLiffID {
		if err := s.repo.DeleteByClinicServiceAndKey(ctx, clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLiffID); err != nil {
			return err
		}
	}

	for _, pair := range pairs {
		if pair.value == "" {
			continue
		}
		encrypted, err := s.encrypt(pair.keyName, pair.value)
		if err != nil {
			return apperrors.Wrap(err, "failed to encrypt integration value")
		}
		record := &model.ClinicIntegration{
			ClinicID: clinicID,
			Service:  model.IntegrationServiceLstep,
			KeyName:  pair.keyName,
			KeyValue: encrypted,
		}
		if err := s.repo.Upsert(ctx, record); err != nil {
			return apperrors.Wrap(err, "failed to update lstep setting")
		}
	}
	return nil
}

func (s *lstepSettingsService) updateClinicSyncConfig(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if s.clinicSettingsRepo == nil {
		return nil
	}
	if err := s.updateCPMVersion(ctx, clinicID, input); err != nil {
		return err
	}
	if err := s.updateDormantThresholds(ctx, clinicID, input); err != nil {
		return err
	}
	if err := s.updateCPMV2Thresholds(ctx, clinicID, input); err != nil {
		return err
	}
	if err := s.updateCPMV1Thresholds(ctx, clinicID, input); err != nil {
		return err
	}
	return s.updateHealthPreventionThresholds(ctx, clinicID, input)
}

func (s *lstepSettingsService) updateCPMVersion(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if input.CPMVersion == nil {
		return nil
	}
	ver := *input.CPMVersion
	if ver != "v1" && ver != "v2" {
		return apperrors.WrapInvalidInput(fmt.Sprintf("cpm_version must be 'v1' or 'v2', got %q", ver))
	}
	if err := s.clinicSettingsRepo.UpdateCPMVersion(ctx, clinicID, ver); err != nil {
		return apperrors.Wrap(err, "failed to update cpm_version")
	}
	return nil
}

func (s *lstepSettingsService) updateDormantThresholds(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if input.DormantPrevention180Days == nil && input.DormantPrevention210Days == nil &&
		input.DormantPrevention240Days == nil && input.DormantPrevention365Days == nil {
		return nil
	}
	current, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find clinic settings")
	}
	thresholds := model.DormantThresholds{
		Stage180: current.DormantPrevention180Days,
		Stage210: current.DormantPrevention210Days,
		Stage240: current.DormantPrevention240Days,
		Stage365: current.DormantPrevention365Days,
	}.WithDefaults()
	if err := applyPositiveInt(input.DormantPrevention180Days, "dormant_prevention_180_days", &thresholds.Stage180); err != nil {
		return err
	}
	if err := applyPositiveInt(input.DormantPrevention210Days, "dormant_prevention_210_days", &thresholds.Stage210); err != nil {
		return err
	}
	if err := applyPositiveInt(input.DormantPrevention240Days, "dormant_prevention_240_days", &thresholds.Stage240); err != nil {
		return err
	}
	if err := applyPositiveInt(input.DormantPrevention365Days, "dormant_prevention_365_days", &thresholds.Stage365); err != nil {
		return err
	}
	if err := s.clinicSettingsRepo.UpdateDormantThresholds(ctx, clinicID, thresholds); err != nil {
		return apperrors.Wrap(err, "failed to update dormant thresholds")
	}
	return nil
}

func (s *lstepSettingsService) updateCPMV2Thresholds(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if input.CPMV2ComingThreshold == nil && input.CPMV2GoodThreshold == nil &&
		input.CPMV2FamilyThreshold == nil && input.CPMV2NoahThreshold == nil {
		return nil
	}
	current, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find clinic settings")
	}
	thresholds := model.CPMV2Thresholds{
		Coming: current.CPMV2ComingThreshold,
		Good:   current.CPMV2GoodThreshold,
		Family: current.CPMV2FamilyThreshold,
		Noah:   current.CPMV2NoahThreshold,
	}.WithDefaults()
	if err := applyPositiveInt(input.CPMV2ComingThreshold, "cpm_v2_coming_threshold", &thresholds.Coming); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV2GoodThreshold, "cpm_v2_good_threshold", &thresholds.Good); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV2FamilyThreshold, "cpm_v2_family_threshold", &thresholds.Family); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV2NoahThreshold, "cpm_v2_noah_threshold", &thresholds.Noah); err != nil {
		return err
	}
	if err := s.clinicSettingsRepo.UpdateCPMV2Thresholds(ctx, clinicID, thresholds); err != nil {
		return apperrors.Wrap(err, "failed to update cpm v2 thresholds")
	}
	return nil
}

func (s *lstepSettingsService) updateCPMV1Thresholds(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if input.CPMV1DormantDays == nil && input.CPMV1NoahDays == nil && input.CPMV1NoahAnnualVisits == nil &&
		input.CPMV1NoahLTV == nil && input.CPMV1CoreDays == nil && input.CPMV1CoreAnnualVisits == nil &&
		input.CPMV1CoreLTV == nil && input.CPMV1SpotMinAmount == nil && input.CPMV1SpotInactiveDays == nil &&
		input.CPMV1GrowingMaxDays == nil && input.CPMV1GrowingMinVisits == nil && input.CPMV1GrowingMaxVisits == nil &&
		input.CPMV1LTVBreakLow == nil {
		return nil
	}
	current, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find clinic settings")
	}
	thresholds := model.CPMV1Thresholds{
		DormantDays:      current.CPMV1DormantDays,
		NoahDays:         current.CPMV1NoahDays,
		NoahAnnualVisits: current.CPMV1NoahAnnualVisits,
		NoahLTV:          current.CPMV1NoahLTV,
		CoreDays:         current.CPMV1CoreDays,
		CoreAnnualVisits: current.CPMV1CoreAnnualVisits,
		CoreLTV:          current.CPMV1CoreLTV,
		SpotMinAmount:    current.CPMV1SpotMinAmount,
		SpotInactiveDays: current.CPMV1SpotInactiveDays,
		GrowingMaxDays:   current.CPMV1GrowingMaxDays,
		GrowingMinVisits: current.CPMV1GrowingMinVisits,
		GrowingMaxVisits: current.CPMV1GrowingMaxVisits,
		LTVBreakLow:      current.CPMV1LTVBreakLow,
	}.WithDefaults()
	if err := applyPositiveInt(input.CPMV1DormantDays, "cpm_v1_dormant_days", &thresholds.DormantDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1NoahDays, "cpm_v1_noah_days", &thresholds.NoahDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1NoahAnnualVisits, "cpm_v1_noah_annual_visits", &thresholds.NoahAnnualVisits); err != nil {
		return err
	}
	if err := applyNonNegativeInt64(input.CPMV1NoahLTV, "cpm_v1_noah_ltv", &thresholds.NoahLTV); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1CoreDays, "cpm_v1_core_days", &thresholds.CoreDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1CoreAnnualVisits, "cpm_v1_core_annual_visits", &thresholds.CoreAnnualVisits); err != nil {
		return err
	}
	if err := applyNonNegativeInt64(input.CPMV1CoreLTV, "cpm_v1_core_ltv", &thresholds.CoreLTV); err != nil {
		return err
	}
	if err := applyNonNegativeInt64(input.CPMV1SpotMinAmount, "cpm_v1_spot_min_amount", &thresholds.SpotMinAmount); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1SpotInactiveDays, "cpm_v1_spot_inactive_days", &thresholds.SpotInactiveDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1GrowingMaxDays, "cpm_v1_growing_max_days", &thresholds.GrowingMaxDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1GrowingMinVisits, "cpm_v1_growing_min_visits", &thresholds.GrowingMinVisits); err != nil {
		return err
	}
	if err := applyPositiveInt(input.CPMV1GrowingMaxVisits, "cpm_v1_growing_max_visits", &thresholds.GrowingMaxVisits); err != nil {
		return err
	}
	if err := applyNonNegativeInt64(input.CPMV1LTVBreakLow, "cpm_v1_ltv_break_low", &thresholds.LTVBreakLow); err != nil {
		return err
	}
	if err := s.clinicSettingsRepo.UpdateCPMV1Thresholds(ctx, clinicID, thresholds); err != nil {
		return apperrors.Wrap(err, "failed to update cpm v1 thresholds")
	}
	return nil
}

func (s *lstepSettingsService) updateHealthPreventionThresholds(ctx context.Context, clinicID uint64, input *UpdateLstepSettingsInput) error {
	if input.HealthPreventionLookbackDays == nil && input.VaccineDeadlineDays == nil {
		return nil
	}
	current, err := s.clinicSettingsRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find clinic settings")
	}
	thresholds := model.HealthPreventionThresholds{
		LookbackDays:    current.HealthPreventionLookbackDays,
		VaccineDeadline: current.VaccineDeadlineDays,
	}.WithDefaults()
	if err := applyPositiveInt(input.HealthPreventionLookbackDays, "health_prevention_lookback_days", &thresholds.LookbackDays); err != nil {
		return err
	}
	if err := applyPositiveInt(input.VaccineDeadlineDays, "vaccine_deadline_days", &thresholds.VaccineDeadline); err != nil {
		return err
	}
	if err := s.clinicSettingsRepo.UpdateHealthPreventionThresholds(ctx, clinicID, thresholds); err != nil {
		return apperrors.Wrap(err, "failed to update health prevention thresholds")
	}
	return nil
}

func applyPositiveInt(value *int, field string, target *int) error {
	if value == nil {
		return nil
	}
	if *value < 1 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s must be >= 1, got %d", field, *value))
	}
	*target = *value
	return nil
}

func applyNonNegativeInt64(value *int64, field string, target *int64) error {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s must be >= 0, got %d", field, *value))
	}
	*target = *value
	return nil
}
