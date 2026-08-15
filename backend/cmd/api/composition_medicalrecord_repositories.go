package main

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// medicalRecordRepositories is the persistence graph owned by the
// medicalrecord domain. Fields stay private to the composition root so
// downstream domains receive only the narrow capabilities they consume.
type medicalRecordRepositories struct {
	cages                  medicalrecord.CageRepository
	carePlanItems          medicalrecord.CarePlanItemRepository
	checkupTypeFields      medicalrecord.CheckupTypeFieldRepository
	checkupFieldResults    medicalrecord.CheckupFieldResultRepository
	checkups               medicalrecord.CheckupRepository
	checkupTypes           medicalrecord.CheckupTypeRepository
	chiefComplaints        medicalrecord.ChiefComplaintTypeRepository
	clinicalPlans          medicalrecord.ClinicalPlanRepository
	consultations          medicalrecord.ConsultationRepository
	dailyRecords           medicalrecord.DailyRecordRepository
	diagnosisNames         medicalrecord.DiagnosisNameRepository
	diagnosisTypes         medicalrecord.DiagnosisTypeRepository
	examinationTypes       medicalrecord.ExamTypeRepository
	examinations           medicalrecord.ExaminationRepository
	hospitalizationPlans   medicalrecord.HospitalizationPlanRepository
	hospitalizations       medicalrecord.HospitalizationRepository
	inquiries              medicalrecord.InquiryRepository
	inquiryTemplates       medicalrecord.InquiryTemplateRepository
	labDuplicateChecker    medicalrecord.LabImportDuplicateChecker
	labImportEvents        medicalrecord.LabImportEventRepository
	labImportJobs          medicalrecord.LabImportJobRepository
	labImportUsageReceipts medicalrecord.LabImportUsageReceiptRepository
	labImportRevertReceipts medicalrecord.LabImportRevertReceiptRepository
	labImportRetractions   medicalrecord.LabImportRetractionRepository
	medicalRecordAddenda   medicalrecord.MedicalRecordAddendumRepository
	medicalRecordImages    medicalrecord.MedicalRecordImageRepository
	medicalRecords         medicalrecord.MedicalRecordRepository
	medicineDoseParameters medicalrecord.MedicineDoseParamRepository
	medicines              medicalrecord.MedicineRepository
	prescriptions          medicalrecord.PrescriptionRepository
	procedures             medicalrecord.ProcedureRepository
	treatmentPlans         medicalrecord.TreatmentPlanRepository
	treatments             medicalrecord.TreatmentRepository
	vaccinations           medicalrecord.VaccinationRepository
	vaccines               medicalrecord.VaccineRepository
	vitals                 medicalrecord.VitalRepository
}

func newMedicalRecordRepositories(db *gorm.DB) medicalRecordRepositories {
	return medicalRecordRepositories{
		cages:                  medicalrecord.NewCageRepository(db),
		carePlanItems:          medicalrecord.NewCarePlanItemRepository(db),
		checkupTypeFields:      medicalrecord.NewCheckupTypeFieldRepository(db),
		checkupFieldResults:    medicalrecord.NewCheckupFieldResultRepository(db),
		checkups:               medicalrecord.NewCheckupRepository(db),
		checkupTypes:           medicalrecord.NewCheckupTypeRepository(db),
		chiefComplaints:        medicalrecord.NewChiefComplaintTypeRepository(db),
		clinicalPlans:          medicalrecord.NewClinicalPlanRepository(db),
		consultations:          medicalrecord.NewConsultationRepository(db),
		dailyRecords:           medicalrecord.NewDailyRecordRepository(db),
		diagnosisNames:         medicalrecord.NewDiagnosisNameRepository(db),
		diagnosisTypes:         medicalrecord.NewDiagnosisTypeRepository(db),
		examinationTypes:       medicalrecord.NewExamTypeRepository(db),
		examinations:           medicalrecord.NewExaminationRepository(db),
		hospitalizationPlans:   medicalrecord.NewHospitalizationPlanRepository(db),
		hospitalizations:       medicalrecord.NewHospitalizationRepository(db),
		inquiries:              medicalrecord.NewInquiryRepository(db),
		inquiryTemplates:       medicalrecord.NewInquiryTemplateRepository(db),
		labDuplicateChecker:     medicalrecord.NewLabImportDuplicateCheckerDB(db),
		labImportEvents:         medicalrecord.NewLabImportEventRepository(db),
		labImportJobs:           medicalrecord.NewLabImportJobRepository(db),
		labImportUsageReceipts:  medicalrecord.NewLabImportUsageReceiptRepository(db),
		labImportRevertReceipts: medicalrecord.NewLabImportRevertReceiptRepository(db),
		labImportRetractions:    medicalrecord.NewLabImportRetractionRepository(db),
		medicalRecordAddenda:    medicalrecord.NewMedicalRecordAddendumRepository(db),
		medicalRecordImages:    medicalrecord.NewMedicalRecordImageRepository(db),
		medicalRecords:         medicalrecord.NewMedicalRecordRepository(db),
		medicineDoseParameters: medicalrecord.NewMedicineDoseParamRepository(db),
		medicines:              medicalrecord.NewMedicineRepository(db),
		prescriptions:          medicalrecord.NewPrescriptionRepository(db),
		procedures:             medicalrecord.NewProcedureRepository(db),
		treatmentPlans:         medicalrecord.NewTreatmentPlanRepository(db),
		treatments:             medicalrecord.NewTreatmentRepository(db),
		vaccinations:           medicalrecord.NewVaccinationRepository(db),
		vaccines:               medicalrecord.NewVaccineRepository(db),
		vitals:                 medicalrecord.NewVitalRepository(db),
	}
}
