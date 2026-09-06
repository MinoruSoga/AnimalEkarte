package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
)

type medicalRecordCompositionDependencies struct {
	DB               *gorm.DB
	Transactor       medicalrecord.Transactor
	Audit            audit.Kernel
	TagSync          lstep.LstepTagSyncService
	DeliveryTrigger  lstep.LstepDeliveryTriggerService
	LineCustomers    lstep.LineCustomerRepository
	Reservations     reservation.ReservationStore
	Pets             pet.CompleteRepository
	Staff            staff.Repository
	StaffAssignments staff.StaffClinicAssignmentRepository
	Inventory        inventory.Repository
	Accounting       billing.AccountingRepository
	BillingItems     billing.BillingItemRepository
}

type medicalRecordHTTPDependencies struct {
	Uploader          infra.FileUploader
	DB                *gorm.DB
	HasPermission     medicalrecord.PermissionChecker
	RequirePermission medicalrecord.PermissionMiddleware
}

type medicalRecordComposition struct {
	MedicalRecords medicalrecord.MedicalRecordService
	Checkups       medicalrecord.CheckupService
	DrainCheckups  func()
	services       medicalRecordServices
	auditTx        medicalRecordAuditTxBridge
}

func newMedicalRecordComposition(
	repositories medicalRecordRepositories,
	dependencies medicalRecordCompositionDependencies,
) medicalRecordComposition {
	auditTx := medicalRecordAuditTxBridge{logger: dependencies.Audit}
	lab := newMedicalRecordLabServices(repositories, dependencies)
	core := newMedicalRecordCoreServices(repositories, dependencies, auditTx)
	// TASK-032: examination clinical reads/mutations share the lab usage receipt sink.
	core.examinations = medicalrecord.AttachLabImportUsageTracker(core.examinations, lab.usageTracker)
	services := medicalRecordServices{
		reference:  newMedicalRecordReferenceServices(repositories, dependencies),
		masters:    newMedicalRecordMasterServices(repositories, dependencies, auditTx),
		preventive: newMedicalRecordPreventiveServices(repositories, dependencies, auditTx),
		lab:        lab,
		clinical:   newMedicalRecordClinicalServices(repositories, dependencies, auditTx),
		hospital:   newMedicalRecordHospitalServices(repositories, dependencies, auditTx),
		treatment:  newMedicalRecordTreatmentServices(repositories, dependencies, auditTx),
		core:       core,
	}

	return medicalRecordComposition{
		MedicalRecords: services.core.medicalRecords,
		Checkups:       services.preventive.checkups,
		DrainCheckups:  nilSafeDrain(services.preventive.checkups.Wait),
		services:       services,
		auditTx:        auditTx,
	}
}

func (c medicalRecordComposition) newHandler(
	dependencies medicalRecordHTTPDependencies,
) *medicalrecord.Handler {
	s := c.services
	return medicalrecord.NewHandler(
		medicalrecord.NewDiagnosisHandler(s.reference.diagnosisTypes, s.reference.diagnosisNames),
		medicalrecord.NewExamTypeHandler(s.reference.examinationTypes),
		medicalrecord.NewChiefComplaintHandler(s.reference.chiefComplaints),
		medicalrecord.NewCheckupHandler(s.preventive.checkups, s.preventive.checkupFieldResults),
		medicalrecord.NewCheckupTypeHandler(s.reference.checkupTypes),
		medicalrecord.NewVaccineHandler(s.reference.vaccines),
		medicalrecord.NewVaccinationHandler(s.preventive.vaccinations),
		medicalrecord.NewPrescriptionHandler(s.preventive.prescriptions),
		medicalrecord.NewInquiryHandler(s.preventive.inquiries),
		medicalrecord.NewInquiryTemplateHandler(s.reference.inquiryTemplates),
		medicalrecord.NewLabImportHandler(s.lab.resultImport, s.lab.jobs, s.lab.audit, s.lab.revert).
			WithDeviceMasters(s.lab.deviceMasters).
			WithDeviceReceive(s.lab.deviceReceive),
		medicalrecord.NewLabReportHandler(s.lab.reports),
		medicalrecord.NewVitalHandler(s.clinical.vitals, s.core.medicalRecords),
		medicalrecord.NewClinicalPlanHandler(s.clinical.plans),
		medicalrecord.NewMedicalRecordImageHandler(
			s.clinical.images,
			s.core.medicalRecords,
			dependencies.Uploader,
			medicalrecord.NewPostgresMedicalRecordImageUploadQuotaStore(dependencies.DB),
		),
		medicalrecord.NewTreatmentHandler(s.treatment.treatments, dependencies.HasPermission),
		medicalrecord.NewHospitalizationHandler(s.hospital.hospitalizations, dependencies.HasPermission),
		medicalrecord.NewHospitalizationPlanHandler(s.hospital.plans),
		medicalrecord.NewDailyRecordHandler(s.hospital.dailyRecords),
		medicalrecord.NewCarePlanItemHandler(s.hospital.carePlanItems),
		medicalrecord.NewConsultationHandler(s.masters.consultations),
		medicalrecord.NewProcedureHandler(s.masters.procedures),
		medicalrecord.NewMedicineHandler(s.masters.medicines),
		medicalrecord.NewMedicineDoseParamHandler(s.masters.medicineDoseParameters),
		medicalrecord.NewCageHandler(s.masters.cages),
		medicalrecord.NewTreatmentPlanHandler(s.masters.treatmentPlans, s.hospital.hospitalizations, s.core.medicalRecords, dependencies.HasPermission),
		medicalrecord.NewMedicalRecordHandler(s.core.medicalRecords),
		medicalrecord.NewMedicalRecordAddendumHandler(s.core.addenda),
		medicalrecord.NewExaminationHandler(s.core.examinations),
		medicalrecord.NewCheckupPackageImportHandler(
			medicalrecord.NewCheckupPackageImportService(
				dependencies.DB,
				persistence.NewTransactor(dependencies.DB),
				c.auditTx,
			),
		),
		dependencies.RequirePermission,
	)
}

type medicalRecordAuditBridge struct {
	logger audit.Service
}

func (a medicalRecordAuditBridge) LogEntry(
	ctx context.Context,
	entry *medicalrecord.AuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("medical record audit logger is required")
	}
	return a.logger.LogEntry(ctx, medicalRecordAuditEntry(entry))
}

type medicalRecordAuditTxBridge struct {
	logger audit.TxLogger
}

func (a medicalRecordAuditTxBridge) LogEntryTx(
	ctx context.Context,
	entry *medicalrecord.AuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("medical record transaction audit logger is required")
	}
	return a.logger.LogEntryTx(ctx, medicalRecordAuditEntry(entry))
}

func medicalRecordAuditEntry(entry *medicalrecord.AuditEntry) *audit.Entry {
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
