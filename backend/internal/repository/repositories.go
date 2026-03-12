package repository

import (
	"gorm.io/gorm"
)

// Repositories はすべてのリポジトリを保持するDIコンテナ
type Repositories struct {
	Owner               OwnerRepository
	Pet                 PetRepository
	Reservation         ReservationRepository
	MedicalRecord       MedicalRecordRepository
	Hospitalization     HospitalizationRepository
	Accounting          AccountingRepository
	Trimming            TrimmingRepository
	Inventory           InventoryRepository
	Staff               StaffRepository
	Cage                CageRepository
	Medicine            MedicineRepository
	Vaccine             VaccineRepository
	Insurance           InsuranceRepository
	ServiceType         ServiceTypeRepository
	Consultation        ConsultationRepository
	Procedure           ProcedureRepository
	HospitalizationPlan HospitalizationPlanRepository
	TrimmingCourse      TrimmingCourseRepository
	TrimmingOption      TrimmingOptionRepository
	ExaminationType     ExamTypeRepository
	DiagnosisCategory   DiagnosisCategoryRepository
	DiagnosisName       DiagnosisNameRepository
	CheckupType         CheckupTypeRepository
	Clinic              ClinicRepository
	UserAccount         UserAccountRepository
}

// NewRepositories はすべてのリポジトリを初期化して返す
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Owner:               NewOwnerRepository(db),
		Pet:                 NewPetRepository(db),
		Reservation:         NewReservationRepository(db),
		MedicalRecord:       NewMedicalRecordRepository(db),
		Hospitalization:     NewHospitalizationRepository(db),
		Accounting:          NewAccountingRepository(db),
		Trimming:            NewTrimmingRepository(db),
		Inventory:           NewInventoryRepository(db),
		Staff:               NewStaffRepository(db),
		Cage:                NewCageRepository(db),
		Medicine:            NewMedicineRepository(db),
		Vaccine:             NewVaccineRepository(db),
		Insurance:           NewInsuranceRepository(db),
		ServiceType:         NewServiceTypeRepository(db),
		Consultation:        NewConsultationRepository(db),
		Procedure:           NewProcedureRepository(db),
		HospitalizationPlan: NewHospitalizationPlanRepository(db),
		TrimmingCourse:      NewTrimmingCourseRepository(db),
		TrimmingOption:      NewTrimmingOptionRepository(db),
		ExaminationType:     NewExamTypeRepository(db),
		DiagnosisCategory:   NewDiagnosisCategoryRepository(db),
		DiagnosisName:       NewDiagnosisNameRepository(db),
		CheckupType:         NewCheckupTypeRepository(db),
		Clinic:              NewClinicRepository(db),
		UserAccount:         NewUserAccountRepository(db),
	}
}
