package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, role *string) ([]model.Staff, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, role *string) ([]model.Staff, error) {
	var staffs []model.Staff
	q := r.db.WithContext(ctx).Model(&model.Staff{})
	if role != nil {
		q = q.Where("staff_role = ?", *role)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&staffs).Error; err != nil {
		return nil, apperrors.Wrap(err, "find staffs")
	}
	return staffs, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	var staff model.Staff
	if err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("staff", id.String())
		}
		return nil, apperrors.Wrap(err, "find staff by id")
	}
	return &staff, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if err := r.db.WithContext(ctx).Create(staff).Error; err != nil {
		return apperrors.Wrap(err, "create staff")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, staff *model.Staff) error {
	if err := r.db.WithContext(ctx).Save(staff).Error; err != nil {
		return apperrors.Wrap(err, "update staff")
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Staff{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete staff")
	}
	return nil
}

// ---- Cage ----

type CageRepository interface {
	FindAll(ctx context.Context, cageType *string) ([]model.Cage, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Cage, error)
	Create(ctx context.Context, cage *model.Cage) error
	Update(ctx context.Context, cage *model.Cage) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type cageRepository struct{ db *gorm.DB }

func NewCageRepository(db *gorm.DB) CageRepository { return &cageRepository{db: db} }

func (r *cageRepository) FindAll(ctx context.Context, cageType *string) ([]model.Cage, error) {
	var cages []model.Cage
	q := r.db.WithContext(ctx).Model(&model.Cage{})
	if cageType != nil {
		q = q.Where("cage_type = ?", *cageType)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&cages).Error; err != nil {
		return nil, apperrors.Wrap(err, "find cages")
	}
	return cages, nil
}

func (r *cageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Cage, error) {
	var cage model.Cage
	if err := r.db.WithContext(ctx).First(&cage, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("cage", id.String())
		}
		return nil, apperrors.Wrap(err, "find cage by id")
	}
	return &cage, nil
}

func (r *cageRepository) Create(ctx context.Context, cage *model.Cage) error {
	if err := r.db.WithContext(ctx).Create(cage).Error; err != nil {
		return apperrors.Wrap(err, "create cage")
	}
	return nil
}

func (r *cageRepository) Update(ctx context.Context, cage *model.Cage) error {
	if err := r.db.WithContext(ctx).Save(cage).Error; err != nil {
		return apperrors.Wrap(err, "update cage")
	}
	return nil
}

func (r *cageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Cage{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete cage")
	}
	return nil
}

// ---- Medicine ----

type MedicineRepository interface {
	FindAll(ctx context.Context) ([]model.Medicine, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, medicine *model.Medicine) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

func (r *medicineRepository) FindAll(ctx context.Context) ([]model.Medicine, error) {
	var medicines []model.Medicine
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&medicines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find medicines")
	}
	return medicines, nil
}

func (r *medicineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Medicine, error) {
	var medicine model.Medicine
	if err := r.db.WithContext(ctx).First(&medicine, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medicine", id.String())
		}
		return nil, apperrors.Wrap(err, "find medicine by id")
	}
	return &medicine, nil
}

func (r *medicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	if err := r.db.WithContext(ctx).Create(medicine).Error; err != nil {
		return apperrors.Wrap(err, "create medicine")
	}
	return nil
}

func (r *medicineRepository) Update(ctx context.Context, medicine *model.Medicine) error {
	if err := r.db.WithContext(ctx).Save(medicine).Error; err != nil {
		return apperrors.Wrap(err, "update medicine")
	}
	return nil
}

func (r *medicineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Medicine{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete medicine")
	}
	return nil
}

// ---- Vaccine ----

type VaccineRepository interface {
	FindAll(ctx context.Context, species *string) ([]model.Vaccine, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Vaccine, error)
	Create(ctx context.Context, vaccine *model.Vaccine) error
	Update(ctx context.Context, vaccine *model.Vaccine) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type vaccineRepository struct{ db *gorm.DB }

func NewVaccineRepository(db *gorm.DB) VaccineRepository { return &vaccineRepository{db: db} }

func (r *vaccineRepository) FindAll(ctx context.Context, species *string) ([]model.Vaccine, error) {
	var vaccines []model.Vaccine
	q := r.db.WithContext(ctx).Model(&model.Vaccine{})
	if species != nil {
		q = q.Where("species = ?", *species)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&vaccines).Error; err != nil {
		return nil, apperrors.Wrap(err, "find vaccines")
	}
	return vaccines, nil
}

func (r *vaccineRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Vaccine, error) {
	var vaccine model.Vaccine
	if err := r.db.WithContext(ctx).First(&vaccine, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("vaccine", id.String())
		}
		return nil, apperrors.Wrap(err, "find vaccine by id")
	}
	return &vaccine, nil
}

func (r *vaccineRepository) Create(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Create(vaccine).Error; err != nil {
		return apperrors.Wrap(err, "create vaccine")
	}
	return nil
}

func (r *vaccineRepository) Update(ctx context.Context, vaccine *model.Vaccine) error {
	if err := r.db.WithContext(ctx).Save(vaccine).Error; err != nil {
		return apperrors.Wrap(err, "update vaccine")
	}
	return nil
}

func (r *vaccineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Vaccine{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete vaccine")
	}
	return nil
}

// ---- Insurance ----

type InsuranceRepository interface {
	FindAll(ctx context.Context) ([]model.Insurance, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error)
	Create(ctx context.Context, insurance *model.Insurance) error
	Update(ctx context.Context, insurance *model.Insurance) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type insuranceRepository struct{ db *gorm.DB }

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository { return &insuranceRepository{db: db} }

func (r *insuranceRepository) FindAll(ctx context.Context) ([]model.Insurance, error) {
	var insurances []model.Insurance
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&insurances).Error; err != nil {
		return nil, apperrors.Wrap(err, "find insurances")
	}
	return insurances, nil
}

func (r *insuranceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Insurance, error) {
	var insurance model.Insurance
	if err := r.db.WithContext(ctx).First(&insurance, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("insurance", id.String())
		}
		return nil, apperrors.Wrap(err, "find insurance by id")
	}
	return &insurance, nil
}

func (r *insuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	if err := r.db.WithContext(ctx).Create(insurance).Error; err != nil {
		return apperrors.Wrap(err, "create insurance")
	}
	return nil
}

func (r *insuranceRepository) Update(ctx context.Context, insurance *model.Insurance) error {
	if err := r.db.WithContext(ctx).Save(insurance).Error; err != nil {
		return apperrors.Wrap(err, "update insurance")
	}
	return nil
}

func (r *insuranceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Insurance{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete insurance")
	}
	return nil
}

// ---- ServiceType ----

type ServiceTypeRepository interface {
	FindAll(ctx context.Context) ([]model.ServiceType, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error)
	Create(ctx context.Context, serviceType *model.ServiceType) error
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type serviceTypeRepository struct{ db *gorm.DB }

func NewServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &serviceTypeRepository{db: db}
}

func (r *serviceTypeRepository) FindAll(ctx context.Context) ([]model.ServiceType, error) {
	var serviceTypes []model.ServiceType
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&serviceTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find service types")
	}
	return serviceTypes, nil
}

func (r *serviceTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ServiceType, error) {
	var serviceType model.ServiceType
	if err := r.db.WithContext(ctx).First(&serviceType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("service_type", id.String())
		}
		return nil, apperrors.Wrap(err, "find service type by id")
	}
	return &serviceType, nil
}

func (r *serviceTypeRepository) Create(ctx context.Context, serviceType *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Create(serviceType).Error; err != nil {
		return apperrors.Wrap(err, "create service type")
	}
	return nil
}

func (r *serviceTypeRepository) Update(ctx context.Context, serviceType *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Save(serviceType).Error; err != nil {
		return apperrors.Wrap(err, "update service type")
	}
	return nil
}

func (r *serviceTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.ServiceType{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete service type")
	}
	return nil
}

// ---- Consultation ----

type ConsultationRepository interface {
	FindAll(ctx context.Context) ([]model.Consultation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context) ([]model.Consultation, error) {
	var consultations []model.Consultation
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&consultations).Error; err != nil {
		return nil, apperrors.Wrap(err, "find consultations")
	}
	return consultations, nil
}

func (r *consultationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error) {
	var consultation model.Consultation
	if err := r.db.WithContext(ctx).First(&consultation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("consultation", id.String())
		}
		return nil, apperrors.Wrap(err, "find consultation by id")
	}
	return &consultation, nil
}

func (r *consultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	if err := r.db.WithContext(ctx).Create(consultation).Error; err != nil {
		return apperrors.Wrap(err, "create consultation")
	}
	return nil
}

func (r *consultationRepository) Update(ctx context.Context, consultation *model.Consultation) error {
	if err := r.db.WithContext(ctx).Save(consultation).Error; err != nil {
		return apperrors.Wrap(err, "update consultation")
	}
	return nil
}

func (r *consultationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Consultation{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete consultation")
	}
	return nil
}

// ---- Procedure ----

type ProcedureRepository interface {
	FindAll(ctx context.Context) ([]model.Procedure, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Procedure, error)
	Create(ctx context.Context, procedure *model.Procedure) error
	Update(ctx context.Context, procedure *model.Procedure) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type procedureRepository struct{ db *gorm.DB }

func NewProcedureRepository(db *gorm.DB) ProcedureRepository { return &procedureRepository{db: db} }

func (r *procedureRepository) FindAll(ctx context.Context) ([]model.Procedure, error) {
	var procedures []model.Procedure
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&procedures).Error; err != nil {
		return nil, apperrors.Wrap(err, "find procedures")
	}
	return procedures, nil
}

func (r *procedureRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Procedure, error) {
	var procedure model.Procedure
	if err := r.db.WithContext(ctx).First(&procedure, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("procedure", id.String())
		}
		return nil, apperrors.Wrap(err, "find procedure by id")
	}
	return &procedure, nil
}

func (r *procedureRepository) Create(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Create(procedure).Error; err != nil {
		return apperrors.Wrap(err, "create procedure")
	}
	return nil
}

func (r *procedureRepository) Update(ctx context.Context, procedure *model.Procedure) error {
	if err := r.db.WithContext(ctx).Save(procedure).Error; err != nil {
		return apperrors.Wrap(err, "update procedure")
	}
	return nil
}

func (r *procedureRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Procedure{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete procedure")
	}
	return nil
}

// ---- HospitalizationPlan ----

type HospitalizationPlanRepository interface {
	FindAll(ctx context.Context) ([]model.HospitalizationPlan, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HospitalizationPlan, error)
	Create(ctx context.Context, plan *model.HospitalizationPlan) error
	Update(ctx context.Context, plan *model.HospitalizationPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type hospitalizationPlanRepository struct{ db *gorm.DB }

func NewHospitalizationPlanRepository(db *gorm.DB) HospitalizationPlanRepository {
	return &hospitalizationPlanRepository{db: db}
}

func (r *hospitalizationPlanRepository) FindAll(ctx context.Context) ([]model.HospitalizationPlan, error) {
	var plans []model.HospitalizationPlan
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&plans).Error; err != nil {
		return nil, apperrors.Wrap(err, "find hospitalization plans")
	}
	return plans, nil
}

func (r *hospitalizationPlanRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HospitalizationPlan, error) {
	var plan model.HospitalizationPlan
	if err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("hospitalization_plan", id.String())
		}
		return nil, apperrors.Wrap(err, "find hospitalization plan by id")
	}
	return &plan, nil
}

func (r *hospitalizationPlanRepository) Create(ctx context.Context, plan *model.HospitalizationPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return apperrors.Wrap(err, "create hospitalization plan")
	}
	return nil
}

func (r *hospitalizationPlanRepository) Update(ctx context.Context, plan *model.HospitalizationPlan) error {
	if err := r.db.WithContext(ctx).Save(plan).Error; err != nil {
		return apperrors.Wrap(err, "update hospitalization plan")
	}
	return nil
}

func (r *hospitalizationPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.HospitalizationPlan{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete hospitalization plan")
	}
	return nil
}

// ---- TrimmingCourse ----

type TrimmingCourseRepository interface {
	FindAll(ctx context.Context) ([]model.TrimmingCourse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, course *model.TrimmingCourse) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type trimmingCourseRepository struct{ db *gorm.DB }

func NewTrimmingCourseRepository(db *gorm.DB) TrimmingCourseRepository {
	return &trimmingCourseRepository{db: db}
}

func (r *trimmingCourseRepository) FindAll(ctx context.Context) ([]model.TrimmingCourse, error) {
	var courses []model.TrimmingCourse
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&courses).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming courses")
	}
	return courses, nil
}

func (r *trimmingCourseRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingCourse, error) {
	var course model.TrimmingCourse
	if err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_course", id.String())
		}
		return nil, apperrors.Wrap(err, "find trimming course by id")
	}
	return &course, nil
}

func (r *trimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	if err := r.db.WithContext(ctx).Create(course).Error; err != nil {
		return apperrors.Wrap(err, "create trimming course")
	}
	return nil
}

func (r *trimmingCourseRepository) Update(ctx context.Context, course *model.TrimmingCourse) error {
	if err := r.db.WithContext(ctx).Save(course).Error; err != nil {
		return apperrors.Wrap(err, "update trimming course")
	}
	return nil
}

func (r *trimmingCourseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.TrimmingCourse{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete trimming course")
	}
	return nil
}

// ---- TrimmingOption ----

type TrimmingOptionRepository interface {
	FindAll(ctx context.Context) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, option *model.TrimmingOption) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type trimmingOptionRepository struct{ db *gorm.DB }

func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context) ([]model.TrimmingOption, error) {
	var options []model.TrimmingOption
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.Wrap(err, "find trimming options")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	if err := r.db.WithContext(ctx).First(&option, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_option", id.String())
		}
		return nil, apperrors.Wrap(err, "find trimming option by id")
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Create(option).Error; err != nil {
		return apperrors.Wrap(err, "create trimming option")
	}
	return nil
}

func (r *trimmingOptionRepository) Update(ctx context.Context, option *model.TrimmingOption) error {
	if err := r.db.WithContext(ctx).Save(option).Error; err != nil {
		return apperrors.Wrap(err, "update trimming option")
	}
	return nil
}

func (r *trimmingOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.TrimmingOption{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete trimming option")
	}
	return nil
}

// ---- ExaminationType ----

type ExaminationTypeRepository interface {
	FindAll(ctx context.Context) ([]model.ExaminationType, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, exType *model.ExaminationType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type examinationTypeRepository struct{ db *gorm.DB }

func NewExaminationTypeRepository(db *gorm.DB) ExaminationTypeRepository {
	return &examinationTypeRepository{db: db}
}

func (r *examinationTypeRepository) FindAll(ctx context.Context) ([]model.ExaminationType, error) {
	var exTypes []model.ExaminationType
	if err := r.db.WithContext(ctx).Preload("Items").Order("sort_order ASC, name ASC").Find(&exTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find examination types")
	}
	return exTypes, nil
}

func (r *examinationTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExaminationType, error) {
	var exType model.ExaminationType
	if err := r.db.WithContext(ctx).Preload("Items").First(&exType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("examination_type", id.String())
		}
		return nil, apperrors.Wrap(err, "find examination type by id")
	}
	return &exType, nil
}

func (r *examinationTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	if err := r.db.WithContext(ctx).Create(exType).Error; err != nil {
		return apperrors.Wrap(err, "create examination type")
	}
	return nil
}

func (r *examinationTypeRepository) Update(ctx context.Context, exType *model.ExaminationType) error {
	if err := r.db.WithContext(ctx).Save(exType).Error; err != nil {
		return apperrors.Wrap(err, "update examination type")
	}
	return nil
}

func (r *examinationTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.ExaminationType{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete examination type")
	}
	return nil
}

// ---- DiagnosisCategory ----

type DiagnosisCategoryRepository interface {
	FindAll(ctx context.Context) ([]model.DiagnosisCategory, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error)
	Create(ctx context.Context, category *model.DiagnosisCategory) error
	Update(ctx context.Context, category *model.DiagnosisCategory) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisCategoryRepository struct{ db *gorm.DB }

func NewDiagnosisCategoryRepository(db *gorm.DB) DiagnosisCategoryRepository {
	return &diagnosisCategoryRepository{db: db}
}

func (r *diagnosisCategoryRepository) FindAll(ctx context.Context) ([]model.DiagnosisCategory, error) {
	var categories []model.DiagnosisCategory
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis categories")
	}
	return categories, nil
}

func (r *diagnosisCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisCategory, error) {
	var category model.DiagnosisCategory
	if err := r.db.WithContext(ctx).Preload("Names").First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("diagnosis_category", id.String())
		}
		return nil, apperrors.Wrap(err, "find diagnosis category by id")
	}
	return &category, nil
}

func (r *diagnosisCategoryRepository) Create(ctx context.Context, category *model.DiagnosisCategory) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return apperrors.Wrap(err, "create diagnosis category")
	}
	return nil
}

func (r *diagnosisCategoryRepository) Update(ctx context.Context, category *model.DiagnosisCategory) error {
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		return apperrors.Wrap(err, "update diagnosis category")
	}
	return nil
}

func (r *diagnosisCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.DiagnosisCategory{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete diagnosis category")
	}
	return nil
}

// ---- DiagnosisName ----

type DiagnosisNameRepository interface {
	FindAll(ctx context.Context) ([]model.DiagnosisName, error)
	FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, name *model.DiagnosisName) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type diagnosisNameRepository struct{ db *gorm.DB }

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context) ([]model.DiagnosisName, error) {
	var names []model.DiagnosisName
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.DiagnosisName, error) {
	var names []model.DiagnosisName
	if err := r.db.WithContext(ctx).
		Where("diagnosis_category_id = ?", categoryID).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, apperrors.Wrap(err, "find diagnosis names by category id")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.DiagnosisName, error) {
	var name model.DiagnosisName
	if err := r.db.WithContext(ctx).First(&name, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("diagnosis_name", id.String())
		}
		return nil, apperrors.Wrap(err, "find diagnosis name by id")
	}
	return &name, nil
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	if err := r.db.WithContext(ctx).Create(name).Error; err != nil {
		return apperrors.Wrap(err, "create diagnosis name")
	}
	return nil
}

func (r *diagnosisNameRepository) Update(ctx context.Context, name *model.DiagnosisName) error {
	if err := r.db.WithContext(ctx).Save(name).Error; err != nil {
		return apperrors.Wrap(err, "update diagnosis name")
	}
	return nil
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.DiagnosisName{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete diagnosis name")
	}
	return nil
}

// ---- CheckupType ----

type CheckupTypeRepository interface {
	FindAll(ctx context.Context) ([]model.CheckupType, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, checkupType *model.CheckupType) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type checkupTypeRepository struct{ db *gorm.DB }

func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

func (r *checkupTypeRepository) FindAll(ctx context.Context) ([]model.CheckupType, error) {
	var checkupTypes []model.CheckupType
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error; err != nil {
		return nil, apperrors.Wrap(err, "find checkup types")
	}
	return checkupTypes, nil
}

func (r *checkupTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CheckupType, error) {
	var checkupType model.CheckupType
	if err := r.db.WithContext(ctx).First(&checkupType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("checkup_type", id.String())
		}
		return nil, apperrors.Wrap(err, "find checkup type by id")
	}
	return &checkupType, nil
}

func (r *checkupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	if err := r.db.WithContext(ctx).Create(checkupType).Error; err != nil {
		return apperrors.Wrap(err, "create checkup type")
	}
	return nil
}

func (r *checkupTypeRepository) Update(ctx context.Context, checkupType *model.CheckupType) error {
	if err := r.db.WithContext(ctx).Save(checkupType).Error; err != nil {
		return apperrors.Wrap(err, "update checkup type")
	}
	return nil
}

func (r *checkupTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.CheckupType{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete checkup type")
	}
	return nil
}
