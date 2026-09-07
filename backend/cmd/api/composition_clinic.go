package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/clinic"
)

type clinicRepositories struct {
	Clinics               *clinic.Repository
	Holidays              clinic.ClinicHolidayRepository
	Settings              clinic.ClinicSettingsRepository
	ClosingSpecialPeriods clinic.ClosingSpecialPeriodRepository
	Company               clinic.CompanyRepository
}

type clinicComposition struct {
	Repositories    clinicRepositories
	Clinic          clinic.Service
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
	auditTx audit.TxLogger,
) clinicComposition {
	closingSettings := clinic.NewClosingSettingsService(
		repositories.Settings,
		repositories.ClosingSpecialPeriods,
		repositories.Holidays,
		&clinic.ClosingSettingsServiceDeps{
			Transactor:   transactor,
			ClinicLocker: repositories.Clinics,
			AuditTx:      clinicAuditTxBridge{logger: auditTx},
		},
	)

	return clinicComposition{
		Repositories: repositories,
		Clinic: clinic.NewService(
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

// clinicAuditTxBridge adapts audit.TxLogger to clinic.AuditTxLogger.
type clinicAuditTxBridge struct {
	logger audit.TxLogger
}

func (a clinicAuditTxBridge) LogEntryTx(ctx context.Context, entry *clinic.AuditEntry) error {
	if a.logger == nil {
		return fmt.Errorf("clinic transaction audit logger is required")
	}
	if entry == nil {
		return fmt.Errorf("clinic audit entry is required")
	}
	return a.logger.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Metadata:   entry.Metadata,
	})
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
