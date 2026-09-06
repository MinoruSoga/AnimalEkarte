package main

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/trimming"
)

type billingCompositionDependencies struct {
	Transactor       billing.Transactor
	Audit            audit.Kernel
	MedicalRecords   medicalrecord.MedicalRecordRepository
	Hospitalizations medicalrecord.HospitalizationRepository
	Reservations     reservation.ReservationStore
	TagSync          lstep.LstepTagSyncService
	StaffAssignments staff.StaffClinicAssignmentRepository
	Treatments       medicalrecord.TreatmentRepository
	TrimmingCourses  trimming.TrimmingCourseRepository
	TrimmingOptions  trimming.TrimmingOptionRepository
	Owners           owner.Repository
	MerchandiseItems inventory.MerchandiseItemRepository
	ClosingSettings  clinic.ClosingSettingsService
	Clinics          *clinic.Repository
	ClinicHolidays   clinic.ClinicHolidayRepository
}

type billingComposition struct {
	Accounting   billing.AccountingService
	BillingItems billing.BillingItemService
	Insurance    billing.InsuranceService
	CashRegister billing.CashRegisterService
	services     billingServices
}

func newBillingComposition(
	repositories billingRepositories,
	dependencies billingCompositionDependencies,
) billingComposition {
	auditLogger := billingAuditBridge{logger: dependencies.Audit}
	auditTxLogger := billingAuditTxBridge{logger: dependencies.Audit}
	services := billingServices{
		core:       newBillingCoreServices(repositories, dependencies, auditTxLogger),
		documents:  newBillingDocumentServices(repositories, dependencies, auditLogger, auditTxLogger),
		references: newBillingReferenceServices(repositories, dependencies),
	}
	return billingComposition{
		Accounting:   services.core.accounting,
		BillingItems: services.core.billingItems,
		Insurance:    services.core.insurance,
		CashRegister: services.core.cashRegister,
		services:     services,
	}
}

func (c billingComposition) newHandler(
	requirePermission billing.PermissionMiddleware,
	hasPermission httpapi.PermissionChecker,
) *billing.Handler {
	s := c.services
	return billing.NewHandler(
		billing.NewInsuranceHandler(s.core.insurance),
		billing.NewCampaignHandler(s.references.campaigns),
		billing.NewPaymentMethodMasterHandler(s.references.paymentMethods),
		billing.NewEstimateHandler(s.documents.estimates, hasPermission),
		billing.NewBillingConfirmationHandler(s.documents.confirmations, requirePermission),
		billing.NewBillingItemHandler(s.core.billingItems, s.core.cashRegister, hasPermission, requirePermission),
		billing.NewRefundHandler(s.documents.refunds, requirePermission),
		billing.NewAccountingHandler(s.core.accounting, s.core.cashRegister, hasPermission),
		billing.NewCashRegisterHandler(s.core.cashRegister, requirePermission),
		billing.NewAccountingReportHandler(s.references.accountingReports, requirePermission),
		requirePermission,
	)
}

type billingAuditBridge struct {
	logger audit.Service
}

func (a billingAuditBridge) LogEntry(
	ctx context.Context,
	entry *billing.AuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("billing audit logger is required")
	}
	return a.logger.LogEntry(ctx, billingAuditEntry(entry))
}

type billingAuditTxBridge struct {
	logger audit.TxLogger
}

func (a billingAuditTxBridge) LogEntryTx(
	ctx context.Context,
	entry *billing.AuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("billing transaction audit logger is required")
	}
	return a.logger.LogEntryTx(ctx, billingAuditEntry(entry))
}

func billingAuditEntry(entry *billing.AuditEntry) *audit.Entry {
	return &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Metadata:   entry.Metadata,
	}
}
