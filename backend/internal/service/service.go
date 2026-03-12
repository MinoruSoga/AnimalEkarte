package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	Owner               OwnerService
	Pet                 PetService
	Reservation         ReservationService
	MedicalRecord       MedicalRecordService
	Hospitalization     HospitalizationService
	Accounting          AccountingService
	Trimming            TrimmingService
	Inventory           InventoryService
	Staff               StaffService
	Cage                CageService
	Medicine            MedicineService
	Vaccine             VaccineService
	Insurance           InsuranceService
	ServiceType         ServiceTypeService
	Consultation        ConsultationService
	Procedure           ProcedureService
	HospitalizationPlan HospitalizationPlanService
	TrimmingCourse      TrimmingCourseService
	TrimmingOption      TrimmingOptionService
	ExaminationType     ExamTypeService
	DiagnosisCategory   DiagnosisCategoryService
	DiagnosisName       DiagnosisNameService
	CheckupType         CheckupTypeService
	Clinic              ClinicService
	UserAccount         UserAccountService
}

// NewServices はリポジトリからすべてのサービスを初期化して返す
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		Owner:               NewOwnerService(repos.Owner),
		Pet:                 NewPetService(repos.Pet),
		Reservation:         NewReservationService(repos.Reservation),
		MedicalRecord:       NewMedicalRecordService(repos.MedicalRecord),
		Hospitalization:     NewHospitalizationService(repos.Hospitalization),
		Accounting:          NewAccountingService(repos.Accounting),
		Trimming:            NewTrimmingService(repos.Trimming),
		Inventory:           NewInventoryService(repos.Inventory),
		Staff:               NewStaffService(repos.Staff),
		Cage:                NewCageService(repos.Cage),
		Medicine:            NewMedicineService(repos.Medicine),
		Vaccine:             NewVaccineService(repos.Vaccine),
		Insurance:           NewInsuranceService(repos.Insurance),
		ServiceType:         NewServiceTypeService(repos.ServiceType),
		Consultation:        NewConsultationService(repos.Consultation),
		Procedure:           NewProcedureService(repos.Procedure),
		HospitalizationPlan: NewHospitalizationPlanService(repos.HospitalizationPlan),
		TrimmingCourse:      NewTrimmingCourseService(repos.TrimmingCourse),
		TrimmingOption:      NewTrimmingOptionService(repos.TrimmingOption),
		ExaminationType:     NewExamTypeService(repos.ExaminationType),
		DiagnosisCategory:   NewDiagnosisCategoryService(repos.DiagnosisCategory),
		DiagnosisName:       NewDiagnosisNameService(repos.DiagnosisName),
		CheckupType:         NewCheckupTypeService(repos.CheckupType),
		Clinic:              NewClinicService(repos.Clinic),
		UserAccount:         NewUserAccountService(repos.UserAccount),
	}
}
