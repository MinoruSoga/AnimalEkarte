package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- StaffService ----

type StaffService interface {
	List(ctx context.Context, role *string) ([]model.Staff, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type staffService struct{ repo repository.StaffRepository }

func NewStaffService(repo repository.StaffRepository) StaffService {
	return &staffService{repo: repo}
}

func (s *staffService) List(ctx context.Context, role *string) ([]model.Staff, error) {
	return s.repo.FindAll(ctx, role)
}
func (s *staffService) GetByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *staffService) Create(ctx context.Context, staff *model.Staff) error {
	return s.repo.Create(ctx, staff)
}
func (s *staffService) Update(ctx context.Context, staff *model.Staff) error {
	return s.repo.Update(ctx, staff)
}
func (s *staffService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- CageService ----

type CageService interface {
	List(ctx context.Context, cageType *string) ([]model.Cage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, cage *model.Cage) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type cageService struct{ repo repository.CageRepository }

func NewCageService(repo repository.CageRepository) CageService {
	return &cageService{repo: repo}
}

func (s *cageService) List(ctx context.Context, cageType *string) ([]model.Cage, error) {
	return s.repo.FindAll(ctx, cageType)
}
func (s *cageService) GetByID(ctx context.Context, id uuid.UUID) (*model.Cage, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *cageService) Create(ctx context.Context, cage *model.Cage) error {
	return s.repo.Create(ctx, cage)
}
func (s *cageService) Update(ctx context.Context, cage *model.Cage) error {
	return s.repo.Update(ctx, cage)
}
func (s *cageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- MedicineService ----

type MedicineService interface {
	List(ctx context.Context) ([]model.Medicine, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, medicine *model.Medicine) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type medicineService struct{ repo repository.MedicineRepository }

func NewMedicineService(repo repository.MedicineRepository) MedicineService {
	return &medicineService{repo: repo}
}

func (s *medicineService) List(ctx context.Context) ([]model.Medicine, error) {
	return s.repo.FindAll(ctx)
}
func (s *medicineService) GetByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *medicineService) Create(ctx context.Context, medicine *model.Medicine) error {
	return s.repo.Create(ctx, medicine)
}
func (s *medicineService) Update(ctx context.Context, medicine *model.Medicine) error {
	return s.repo.Update(ctx, medicine)
}
func (s *medicineService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- VaccineService ----

type VaccineService interface {
	List(ctx context.Context, species *string) ([]model.Vaccine, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, vaccine *model.Vaccine) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type vaccineService struct{ repo repository.VaccineRepository }

func NewVaccineService(repo repository.VaccineRepository) VaccineService {
	return &vaccineService{repo: repo}
}

func (s *vaccineService) List(ctx context.Context, species *string) ([]model.Vaccine, error) {
	return s.repo.FindAll(ctx, species)
}
func (s *vaccineService) GetByID(ctx context.Context, id uuid.UUID) (*model.Vaccine, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *vaccineService) Create(ctx context.Context, vaccine *model.Vaccine) error {
	return s.repo.Create(ctx, vaccine)
}
func (s *vaccineService) Update(ctx context.Context, vaccine *model.Vaccine) error {
	return s.repo.Update(ctx, vaccine)
}
func (s *vaccineService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- InsuranceService ----

type InsuranceService interface {
	List(ctx context.Context) ([]model.Insurance, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, insurance *model.Insurance) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type insuranceService struct{ repo repository.InsuranceRepository }

func NewInsuranceService(repo repository.InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

func (s *insuranceService) List(ctx context.Context) ([]model.Insurance, error) {
	return s.repo.FindAll(ctx)
}
func (s *insuranceService) GetByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *insuranceService) Create(ctx context.Context, insurance *model.Insurance) error {
	return s.repo.Create(ctx, insurance)
}
func (s *insuranceService) Update(ctx context.Context, insurance *model.Insurance) error {
	return s.repo.Update(ctx, insurance)
}
func (s *insuranceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- ServiceTypeService ----

type ServiceTypeService interface {
	List(ctx context.Context) ([]model.ServiceType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error)
	Create(ctx context.Context, serviceType *model.ServiceType) error
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type serviceTypeService struct{ repo repository.ServiceTypeRepository }

func NewServiceTypeService(repo repository.ServiceTypeRepository) ServiceTypeService {
	return &serviceTypeService{repo: repo}
}

func (s *serviceTypeService) List(ctx context.Context) ([]model.ServiceType, error) {
	return s.repo.FindAll(ctx)
}
func (s *serviceTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *serviceTypeService) Create(ctx context.Context, serviceType *model.ServiceType) error {
	return s.repo.Create(ctx, serviceType)
}
func (s *serviceTypeService) Update(ctx context.Context, serviceType *model.ServiceType) error {
	return s.repo.Update(ctx, serviceType)
}
func (s *serviceTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- ConsultationService ----

type ConsultationService interface {
	List(ctx context.Context) ([]model.Consultation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type consultationService struct{ repo repository.ConsultationRepository }

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) List(ctx context.Context) ([]model.Consultation, error) {
	return s.repo.FindAll(ctx)
}
func (s *consultationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *consultationService) Create(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Create(ctx, consultation)
}
func (s *consultationService) Update(ctx context.Context, consultation *model.Consultation) error {
	return s.repo.Update(ctx, consultation)
}
func (s *consultationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- ProcedureService ----

type ProcedureService interface {
	List(ctx context.Context) ([]model.Procedure, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, procedure *model.Procedure) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type procedureService struct{ repo repository.ProcedureRepository }

func NewProcedureService(repo repository.ProcedureRepository) ProcedureService {
	return &procedureService{repo: repo}
}

func (s *procedureService) List(ctx context.Context) ([]model.Procedure, error) {
	return s.repo.FindAll(ctx)
}
func (s *procedureService) GetByID(ctx context.Context, id uuid.UUID) (*model.Procedure, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *procedureService) Create(ctx context.Context, procedure *model.Procedure) error {
	return s.repo.Create(ctx, procedure)
}
func (s *procedureService) Update(ctx context.Context, procedure *model.Procedure) error {
	return s.repo.Update(ctx, procedure)
}
func (s *procedureService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- HospitalizationPlanService ----

type HospitalizationPlanService interface {
	List(ctx context.Context) ([]model.HospitalizationPlan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, plan *model.HospitalizationPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type hospitalizationPlanService struct{ repo repository.HospitalizationPlanRepository }

func NewHospitalizationPlanService(repo repository.HospitalizationPlanRepository) HospitalizationPlanService {
	return &hospitalizationPlanService{repo: repo}
}

func (s *hospitalizationPlanService) List(ctx context.Context) ([]model.HospitalizationPlan, error) {
	return s.repo.FindAll(ctx)
}
func (s *hospitalizationPlanService) GetByID(ctx context.Context, id uuid.UUID) (*model.HospitalizationPlan, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *hospitalizationPlanService) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	return s.repo.Create(ctx, plan)
}
func (s *hospitalizationPlanService) Update(ctx context.Context, plan *model.HospitalizationPlan) error {
	return s.repo.Update(ctx, plan)
}
func (s *hospitalizationPlanService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- TrimmingCourseService ----

type TrimmingCourseService interface {
	List(ctx context.Context) ([]model.TrimmingCourse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, course *model.TrimmingCourse) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type trimmingCourseService struct{ repo repository.TrimmingCourseRepository }

func NewTrimmingCourseService(repo repository.TrimmingCourseRepository) TrimmingCourseService {
	return &trimmingCourseService{repo: repo}
}

func (s *trimmingCourseService) List(ctx context.Context) ([]model.TrimmingCourse, error) {
	return s.repo.FindAll(ctx)
}
func (s *trimmingCourseService) GetByID(ctx context.Context, id uuid.UUID) (*model.TrimmingCourse, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *trimmingCourseService) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return s.repo.Create(ctx, course)
}
func (s *trimmingCourseService) Update(ctx context.Context, course *model.TrimmingCourse) error {
	return s.repo.Update(ctx, course)
}
func (s *trimmingCourseService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- TrimmingOptionService ----

type TrimmingOptionService interface {
	List(ctx context.Context) ([]model.TrimmingOption, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, option *model.TrimmingOption) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type trimmingOptionService struct{ repo repository.TrimmingOptionRepository }

func NewTrimmingOptionService(repo repository.TrimmingOptionRepository) TrimmingOptionService {
	return &trimmingOptionService{repo: repo}
}

func (s *trimmingOptionService) List(ctx context.Context) ([]model.TrimmingOption, error) {
	return s.repo.FindAll(ctx)
}
func (s *trimmingOptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.TrimmingOption, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *trimmingOptionService) Create(ctx context.Context, option *model.TrimmingOption) error {
	return s.repo.Create(ctx, option)
}
func (s *trimmingOptionService) Update(ctx context.Context, option *model.TrimmingOption) error {
	return s.repo.Update(ctx, option)
}
func (s *trimmingOptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- ExaminationTypeService ----

type ExaminationTypeService interface {
	List(ctx context.Context) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, exType *model.ExaminationType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type examinationTypeService struct{ repo repository.ExaminationTypeRepository }

func NewExaminationTypeService(repo repository.ExaminationTypeRepository) ExaminationTypeService {
	return &examinationTypeService{repo: repo}
}

func (s *examinationTypeService) List(ctx context.Context) ([]model.ExaminationType, error) {
	return s.repo.FindAll(ctx)
}
func (s *examinationTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.ExaminationType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *examinationTypeService) Create(ctx context.Context, exType *model.ExaminationType) error {
	return s.repo.Create(ctx, exType)
}
func (s *examinationTypeService) Update(ctx context.Context, exType *model.ExaminationType) error {
	return s.repo.Update(ctx, exType)
}
func (s *examinationTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- DiagnosisCategoryService ----

type DiagnosisCategoryService interface {
	List(ctx context.Context) ([]model.DiagnosisCategory, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, category *model.DiagnosisCategory) error
	Update(ctx context.Context, category *model.DiagnosisCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisCategoryService struct{ repo repository.DiagnosisCategoryRepository }

func NewDiagnosisCategoryService(repo repository.DiagnosisCategoryRepository) DiagnosisCategoryService {
	return &diagnosisCategoryService{repo: repo}
}

func (s *diagnosisCategoryService) List(ctx context.Context) ([]model.DiagnosisCategory, error) {
	return s.repo.FindAll(ctx)
}
func (s *diagnosisCategoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *diagnosisCategoryService) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	return s.repo.Create(ctx, category)
}
func (s *diagnosisCategoryService) Update(ctx context.Context, category *model.DiagnosisCategory) error {
	return s.repo.Update(ctx, category)
}
func (s *diagnosisCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- DiagnosisNameService ----

type DiagnosisNameService interface {
	List(ctx context.Context) ([]model.DiagnosisName, error)
	ListByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, name *model.DiagnosisName) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisNameService struct{ repo repository.DiagnosisNameRepository }

func NewDiagnosisNameService(repo repository.DiagnosisNameRepository) DiagnosisNameService {
	return &diagnosisNameService{repo: repo}
}

func (s *diagnosisNameService) List(ctx context.Context) ([]model.DiagnosisName, error) {
	return s.repo.FindAll(ctx)
}
func (s *diagnosisNameService) ListByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error) {
	return s.repo.FindByCategoryID(ctx, categoryID)
}
func (s *diagnosisNameService) GetByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *diagnosisNameService) Create(ctx context.Context, name *model.DiagnosisName) error {
	return s.repo.Create(ctx, name)
}
func (s *diagnosisNameService) Update(ctx context.Context, name *model.DiagnosisName) error {
	return s.repo.Update(ctx, name)
}
func (s *diagnosisNameService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ---- CheckupTypeService ----

type CheckupTypeService interface {
	List(ctx context.Context) ([]model.CheckupType, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, checkupType *model.CheckupType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type checkupTypeService struct{ repo repository.CheckupTypeRepository }

func NewCheckupTypeService(repo repository.CheckupTypeRepository) CheckupTypeService {
	return &checkupTypeService{repo: repo}
}

func (s *checkupTypeService) List(ctx context.Context) ([]model.CheckupType, error) {
	return s.repo.FindAll(ctx)
}
func (s *checkupTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *checkupTypeService) Create(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Create(ctx, checkupType)
}
func (s *checkupTypeService) Update(ctx context.Context, checkupType *model.CheckupType) error {
	return s.repo.Update(ctx, checkupType)
}
func (s *checkupTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
