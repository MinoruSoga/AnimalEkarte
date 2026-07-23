// Package service keeps the legacy closing-settings service surface during BE9 migration.
// The implementation lives in internal/clinic.
package service

import (
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ClosingSettingsResponse = clinic.ClosingSettingsResponse
type DaySchedule = clinic.DaySchedule
type UpdateClinicSettingsInput = clinic.UpdateClinicSettingsInput
type CreateSpecialPeriodInput = clinic.CreateSpecialPeriodInput
type UpdateSpecialPeriodInput = clinic.UpdateSpecialPeriodInput
type ClosingSettingsService = clinic.ClosingSettingsService

func NewClosingSettingsService(
	settingsRepo repository.ClinicSettingsRepository,
	periodRepo repository.ClosingSpecialPeriodRepository,
	holidayRepo repository.ClinicHolidayRepository,
) ClosingSettingsService {
	return clinic.NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo)
}
