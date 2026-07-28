package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/staff"
)

// staffRepositories are created before reservation composition so the
// canonical staff writer can be injected into reservation-owned repositories.
type staffRepositories struct {
	Staff          staff.StaffRepository
	Assignments    staff.StaffClinicAssignmentRepository
	Occupations    staff.OccupationRepository
	ShiftEntries   staff.ShiftEntryRepository
	ShiftTemplates staff.ShiftTemplateRepository
}

func newStaffRepositories(db *gorm.DB) staffRepositories {
	return staffRepositories{
		Staff:          staff.NewStaffRepository(db),
		Assignments:    staff.NewStaffClinicAssignmentRepository(db),
		Occupations:    staff.NewOccupationRepository(db),
		ShiftEntries:   staff.NewShiftEntryRepository(db),
		ShiftTemplates: staff.NewShiftTemplateRepository(db),
	}
}

type staffCompositionDependencies struct {
	Transactor       staff.Transactor
	Accounts         staff.StaffAccountStore
	PermissionGroups staff.PermissionGroupRepository
	Reservations     staff.StaffAssignmentReservationUsage
	ReservationStaff staff.ReservationStaffRepository
	Clinics          staff.StaffAssignmentClinicLookup
	Audit            audit.TxLogger
}

type staffComposition struct {
	Staff          staff.StaffService
	Assignments    staff.StaffClinicAssignmentService
	Occupations    staff.OccupationService
	ShiftEntries   staff.ShiftEntryService
	ShiftTemplates staff.ShiftTemplateService
}

func newStaffComposition(
	repositories staffRepositories,
	dependencies staffCompositionDependencies,
) staffComposition {
	staffService := staff.NewStaffServiceWithAudits(
		repositories.Staff,
		dependencies.Accounts,
		repositories.Assignments,
		dependencies.Reservations,
		repositories.ShiftEntries,
		dependencies.PermissionGroups,
		dependencies.ReservationStaff,
		repositories.Occupations,
		dependencies.Clinics,
		dependencies.Transactor,
		staffCredentialAuditAdapter{logger: dependencies.Audit},
		staffPermissionAssignmentAuditAdapter{logger: dependencies.Audit},
	)
	return staffComposition{
		Staff: staffService,
		Assignments: staff.NewStaffClinicAssignmentService(
			repositories.Assignments,
		),
		Occupations: staff.NewOccupationService(
			repositories.Occupations,
		),
		ShiftEntries: staff.NewShiftEntryService(
			repositories.ShiftEntries,
			repositories.Staff,
			repositories.Assignments,
			dependencies.Transactor,
		),
		ShiftTemplates: staff.NewShiftTemplateService(
			repositories.ShiftTemplates,
		),
	}
}

type staffPermissionAssignmentAuditAdapter struct {
	logger audit.TxLogger
}

func (a staffPermissionAssignmentAuditAdapter) LogEntryTx(
	ctx context.Context,
	entry *staff.PermissionAssignmentAuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("staff permission assignment audit logger is required")
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
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}

func (c staffComposition) newHandler(
	requirePermission staff.PermissionMiddleware,
	permissionCheckers ...staff.PermissionChecker,
) *staff.Handler {
	var hasPermission staff.PermissionChecker
	if len(permissionCheckers) > 0 {
		hasPermission = permissionCheckers[0]
	}
	return staff.NewHandlerWithPermissionChecker(
		c.Staff,
		c.Assignments,
		c.Occupations,
		c.ShiftEntries,
		c.ShiftTemplates,
		requirePermission,
		hasPermission,
	)
}

type staffCredentialAuditAdapter struct {
	logger audit.TxLogger
}

func (a staffCredentialAuditAdapter) LogEntryTx(
	ctx context.Context,
	entry staff.CredentialAuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("staff credential audit logger is required")
	}
	return a.logger.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}
