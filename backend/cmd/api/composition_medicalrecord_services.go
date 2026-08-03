package main

import "github.com/animal-ekarte/backend/internal/medicalrecord"

type medicalRecordServices struct {
	reference  medicalRecordReferenceServices
	masters    medicalRecordMasterServices
	preventive medicalRecordPreventiveServices
	lab        medicalRecordLabServices
	clinical   medicalRecordClinicalServices
	hospital   medicalRecordHospitalServices
	treatment  medicalRecordTreatmentServices
	core       medicalRecordCoreServices
}

type medicalRecordReferenceServices struct {
	diagnosisTypes   medicalrecord.DiagnosisTypeService
	diagnosisNames   medicalrecord.DiagnosisNameService
	examinationTypes medicalrecord.ExaminationTypeService
	chiefComplaints  medicalrecord.ChiefComplaintTypeService
	checkupTypes     medicalrecord.CheckupTypeService
	vaccines         medicalrecord.VaccineService
	inquiryTemplates medicalrecord.InquiryTemplateService
}

func newMedicalRecordReferenceServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
) medicalRecordReferenceServices {
	return medicalRecordReferenceServices{
		diagnosisTypes:   medicalrecord.NewDiagnosisTypeService(r.diagnosisTypes),
		diagnosisNames:   medicalrecord.NewDiagnosisNameService(r.diagnosisNames, r.diagnosisTypes),
		examinationTypes: medicalrecord.NewExamTypeService(r.examinationTypes, d.Transactor),
		chiefComplaints:  medicalrecord.NewChiefComplaintTypeService(r.chiefComplaints),
		checkupTypes:     medicalrecord.NewCheckupTypeService(r.checkupTypes),
		vaccines:         medicalrecord.NewVaccineService(r.vaccines),
		inquiryTemplates: medicalrecord.NewInquiryTemplateService(r.inquiryTemplates),
	}
}

type medicalRecordMasterServices struct {
	consultations          medicalrecord.ConsultationService
	procedures             medicalrecord.ProcedureService
	medicines              medicalrecord.MedicineService
	medicineDoseParameters medicalrecord.MedicineDoseParamService
	cages                  medicalrecord.CageService
	treatmentPlans         medicalrecord.TreatmentPlanService
}

func newMedicalRecordMasterServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordMasterServices {
	return medicalRecordMasterServices{
		consultations: medicalrecord.NewConsultationService(r.consultations),
		procedures:    medicalrecord.NewProcedureService(r.procedures, d.Transactor),
		medicines: medicalrecord.NewMedicineServiceWithAudit(
			r.medicines,
			d.Inventory,
			d.Transactor,
			auditTx,
		),
		medicineDoseParameters: medicalrecord.NewMedicineDoseParamService(
			r.medicineDoseParameters,
			r.medicines,
			d.Transactor,
			auditTx,
		),
		cages:          medicalrecord.NewCageService(r.cages, d.Transactor),
		treatmentPlans: medicalrecord.NewTreatmentPlanService(r.treatmentPlans, d.Transactor),
	}
}

type medicalRecordPreventiveServices struct {
	checkups            medicalrecord.CheckupService
	checkupFieldResults medicalrecord.CheckupFieldResultService
	vaccinations        medicalrecord.VaccinationService
	prescriptions       medicalrecord.PrescriptionService
	inquiries           medicalrecord.InquiryService
}

func newMedicalRecordPreventiveServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordPreventiveServices {
	return medicalRecordPreventiveServices{
		checkups: medicalrecord.NewCheckupService(
			r.checkups,
			r.medicalRecords,
			r.checkupTypes,
			d.DeliveryTrigger,
			d.TagSync,
			medicalrecord.CheckupWriteDependencies{
				RelationVerifier: d.Reservations,
				Transactor:       d.Transactor,
			},
		),
		checkupFieldResults: medicalrecord.NewCheckupFieldResultService(
			r.checkups,
			r.medicalRecords,
			r.checkupTypeFields,
			r.checkupFieldResults,
			auditTx,
			d.Transactor,
		),
		vaccinations: medicalrecord.NewVaccinationService(
			r.vaccinations,
			r.vaccines,
			d.TagSync,
			d.Reservations,
			r.medicalRecords,
			d.Transactor,
		),
		prescriptions: medicalrecord.NewPrescriptionService(
			r.prescriptions,
			r.medicalRecords,
			d.TagSync,
			d.Transactor,
		),
		inquiries: medicalrecord.NewInquiryService(r.inquiries, r.chiefComplaints),
	}
}

type medicalRecordLabServices struct {
	resultImport medicalrecord.LabResultImportService
	jobs         medicalrecord.LabImportJobService
	audit        medicalrecord.LabAuditLogger
	reports      medicalrecord.LabReportQueryService
}

func newMedicalRecordLabServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
) medicalRecordLabServices {
	jobs := medicalrecord.NewLabImportJobService(r.labImportJobs, r.labImportEvents)
	return medicalRecordLabServices{
		resultImport: medicalrecord.NewLabResultImportService(
			jobs,
			medicalrecord.NewLabImportExaminationService(
				r.examinations,
				r.labDuplicateChecker,
				r.examinationTypes,
				d.Pets,
				r.medicalRecords,
				d.Transactor,
			),
		),
		jobs:    jobs,
		audit:   medicalrecord.NewLabAuditLogger(medicalRecordAuditBridge{logger: d.Audit}),
		reports: medicalrecord.NewLabReportQueryService(r.examinations),
	}
}

type medicalRecordClinicalServices struct {
	vitals medicalrecord.VitalService
	plans  medicalrecord.ClinicalPlanService
	images medicalrecord.MedicalRecordImageService
}

func newMedicalRecordClinicalServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordClinicalServices {
	return medicalRecordClinicalServices{
		vitals: medicalrecord.NewVitalServiceWithRelationValidation(
			r.vitals,
			r.medicalRecords,
			d.Audit,
			d.Reservations,
			d.Staff,
			d.StaffAssignments,
			d.Transactor,
		),
		plans: medicalrecord.NewClinicalPlanService(
			r.clinicalPlans,
			r.medicalRecords,
			r.diagnosisTypes,
			r.diagnosisNames,
			d.Transactor,
			auditTx,
		),
		images: medicalrecord.NewMedicalRecordImageServiceWithRelationValidation(
			r.medicalRecordImages,
			r.medicalRecords,
			d.Pets,
			r.examinations,
			d.Staff,
			d.StaffAssignments,
			d.Transactor,
		),
	}
}

type medicalRecordHospitalServices struct {
	hospitalizations medicalrecord.HospitalizationService
	plans            medicalrecord.HospitalizationPlanService
	dailyRecords     medicalrecord.DailyRecordService
	carePlanItems    medicalrecord.CarePlanItemService
}

func newMedicalRecordHospitalServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordHospitalServices {
	return medicalRecordHospitalServices{
		hospitalizations: medicalrecord.NewHospitalizationServiceWithAudit(
			r.hospitalizations,
			d.Reservations,
			d.Pets,
			r.cages,
			r.carePlanItems,
			d.Accounting,
			d.BillingItems,
			d.Transactor,
			auditTx,
			medicalrecord.WithTreatmentPlanRepository(r.treatmentPlans),
		),
		plans: medicalrecord.NewHospitalizationPlanService(r.hospitalizationPlans),
		dailyRecords: medicalrecord.NewDailyRecordServiceWithRelationValidation(
			r.dailyRecords,
			r.hospitalizations,
			d.Reservations,
			d.Staff,
			d.StaffAssignments,
			d.Transactor,
		),
		carePlanItems: medicalrecord.NewCarePlanItemService(
			r.carePlanItems,
			r.hospitalizations,
			r.medicines,
			r.procedures,
			r.hospitalizationPlans,
			d.Transactor,
			auditTx,
		),
	}
}

type medicalRecordTreatmentServices struct {
	treatments medicalrecord.TreatmentService
}

func newMedicalRecordTreatmentServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordTreatmentServices {
	return medicalRecordTreatmentServices{
		treatments: medicalrecord.NewTreatmentServiceWithAudit(
			r.treatments,
			r.medicalRecords,
			r.medicines,
			r.procedures,
			r.consultations,
			d.Inventory,
			r.vitals,
			r.medicineDoseParameters,
			d.Transactor,
			auditTx,
		),
	}
}

type medicalRecordCoreServices struct {
	medicalRecords medicalrecord.MedicalRecordService
	addenda        medicalrecord.MedicalRecordAddendumService
	examinations   medicalrecord.ExaminationService
}

func newMedicalRecordCoreServices(
	r medicalRecordRepositories,
	d medicalRecordCompositionDependencies,
	auditTx medicalrecord.AuditTxLogger,
) medicalRecordCoreServices {
	medicalRecords := medicalrecord.NewMedicalRecordServiceWithTxAudit(
		r.medicalRecords,
		r.inquiries,
		r.clinicalPlans,
		r.chiefComplaints,
		r.diagnosisTypes,
		r.diagnosisNames,
		d.LineCustomers,
		d.Reservations,
		d.DeliveryTrigger,
		d.Audit,
		auditTx,
		d.Transactor,
		d.TagSync,
	)
	return medicalRecordCoreServices{
		medicalRecords: medicalRecords,
		addenda: medicalrecord.NewMedicalRecordAddendumService(
			r.medicalRecordAddenda,
			r.medicalRecords,
			auditTx,
			d.Transactor,
		),
		examinations: medicalrecord.NewExaminationService(
			r.examinations,
			r.medicalRecords,
			r.examinationTypes,
			auditTx,
			d.Transactor,
			d.Reservations,
		),
	}
}
