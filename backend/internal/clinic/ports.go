// Package clinic owns clinic, company, holiday, and closing-settings HTTP,
// use-case, and persistence as one ADR-006 vertical slice.
package clinic

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// Transactor is the transaction boundary consumed by clinic use cases.
type Transactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// PermissionGroupWriter is clinic's consumer-side auth port. It intentionally
// exposes only the writes required by clinic create/delete orchestration.
type PermissionGroupWriter interface {
	Create(ctx context.Context, group *model.PermissionGroup) error
	UpdateRules(ctx context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error
	DeleteSoftDeletedByClinicID(ctx context.Context, clinicID uint64) error
}

// ClinicDependencyCount describes a persisted reference that blocks clinic deletion.
type ClinicDependencyCount struct {
	Label string
	Count int64
}

type clinicServiceRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	FindByIDs(ctx context.Context, ids []uint64) ([]model.Clinic, error)
	FindActiveIDs(ctx context.Context, ids []uint64) ([]uint64, error)
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
	LockByIDForUpdate(ctx context.Context, id uint64) (*model.Clinic, error)
	FindCompany(ctx context.Context) (*model.Company, error)
	Create(ctx context.Context, clinic *model.Clinic) error
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) error
	Delete(ctx context.Context, id uint64) error
	CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error)
	CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error)
	CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error)
}

// ClinicHolidayRepository is the data port consumed by holiday use cases.
type ClinicHolidayRepository interface {
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
	FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	Delete(ctx context.Context, clinicID uint64, date time.Time) error
}

// ClinicSettingsRepository is the settings data port shared with LSTEP consumers.
type ClinicSettingsRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	Save(ctx context.Context, clinicID uint64, settings *model.ClinicSettings) (*model.ClinicSettings, error)
	UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error
	UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error
	UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error
	UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error
	UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error
}

// ClosingSpecialPeriodRepository is the data port for exceptional closing schedules.
type ClosingSpecialPeriodRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error)
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
	Create(ctx context.Context, period *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	// CreateCheckingOverlap serializes overlap check + insert under a clinic advisory lock (POC-05 / X-05).
	CreateCheckingOverlap(ctx context.Context, period *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	// UpdateCheckingOverlap serializes overlap check + update+reload under a clinic advisory lock (POC-05 / X-05).
	UpdateCheckingOverlap(ctx context.Context, clinicID, id uint64, startDate, endDate time.Time, cmd UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
}

// CompanyRepository is the global company-singleton data port.
type CompanyRepository interface {
	FindSingleton(ctx context.Context) (*model.Company, error)
	Update(ctx context.Context, cmd UpdateCompanyInput) error
}
