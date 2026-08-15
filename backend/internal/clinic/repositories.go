package clinic

import (
	"gorm.io/gorm"
)

func NewClinicHolidayRepository(db *gorm.DB) ClinicHolidayRepository {
	return &clinicHolidayRepository{db: db}
}

func NewClinicSettingsRepository(db *gorm.DB) ClinicSettingsRepository {
	return &clinicSettingsRepository{db: db}
}

func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return &closingSpecialPeriodRepository{db: db}
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}
