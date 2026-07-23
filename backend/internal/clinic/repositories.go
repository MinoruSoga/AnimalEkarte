package clinic

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/clinic/clinicholiday"
	"github.com/animal-ekarte/backend/internal/clinic/clinicsettings"
	"github.com/animal-ekarte/backend/internal/clinic/closingspecialperiod"
	"github.com/animal-ekarte/backend/internal/clinic/company"
)

func NewClinicHolidayRepository(db *gorm.DB) ClinicHolidayRepository {
	return clinicholiday.New(db)
}

func NewClinicSettingsRepository(db *gorm.DB) ClinicSettingsRepository {
	return clinicsettings.New(db)
}

func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return closingspecialperiod.New(db)
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return company.New(db)
}
