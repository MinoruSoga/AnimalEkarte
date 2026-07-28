package main

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic"
)

type clinicRepositories struct {
	Clinics               clinic.ClinicRepository
	Holidays              clinic.ClinicHolidayRepository
	Settings              clinic.ClinicSettingsRepository
	ClosingSpecialPeriods clinic.ClosingSpecialPeriodRepository
	Company               clinic.CompanyRepository
}

type clinicComposition struct {
	Repositories    clinicRepositories
	Clinic          clinic.ClinicService
	ClosingSettings clinic.ClosingSettingsService
	holidays        clinic.ClinicHolidayService
	company         clinic.CompanyService
}

func newClinicRepositories(db *gorm.DB) clinicRepositories {
	return clinicRepositories{
		Clinics:               clinic.NewClinicRepository(db),
		Holidays:              clinic.NewClinicHolidayRepository(db),
		Settings:              clinic.NewClinicSettingsRepository(db),
		ClosingSpecialPeriods: clinic.NewClosingSpecialPeriodRepository(db),
		Company:               clinic.NewCompanyRepository(db),
	}
}

func newClinicComposition(
	repositories clinicRepositories,
	permissionGroups clinic.PermissionGroupWriter,
	transactor clinic.Transactor,
) clinicComposition {
	closingSettings := clinic.NewClosingSettingsService(
		repositories.Settings,
		repositories.ClosingSpecialPeriods,
		repositories.Holidays,
	)

	return clinicComposition{
		Repositories: repositories,
		Clinic: clinic.NewClinicService(
			repositories.Clinics,
			permissionGroups,
			transactor,
		),
		ClosingSettings: closingSettings,
		holidays:        clinic.NewClinicHolidayService(repositories.Holidays),
		company: clinic.NewCompanyService(
			repositories.Company,
		),
	}
}

func (c clinicComposition) newHandler(
	requirePermission clinic.PermissionMiddleware,
) *clinic.Handler {
	return clinic.NewHandler(
		c.Clinic,
		c.holidays,
		c.ClosingSettings,
		c.company,
		requirePermission,
	)
}
