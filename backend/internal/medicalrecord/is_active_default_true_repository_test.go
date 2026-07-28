package medicalrecord

// BUG-455-S5: gorm:"default:true" IsActive compensation — explicit false must persist.
// Mirrors AUTH-B (permission_group_repository) with in-memory + FindByID + raw column asserts.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCageRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupCageRepositoryTestDB(t)
	repo := NewCageRepository(db)
	ctx := context.Background()

	inactive := &model.Cage{
		ClinicID: 1, Name: "inactive cage",
		CageType: model.CageTypeGeneral, CageSize: model.CageSizeSmall,
		IsActive: false,
	}
	require.NoError(t, repo.Create(ctx, inactive))
	require.NotZero(t, inactive.ID)
	assert.False(t, inactive.IsActive)

	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.Cage{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.Cage{
		ClinicID: 1, Name: "active cage",
		CageType: model.CageTypeGeneral, CageSize: model.CageSizeSmall,
		IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	gotActive, err := repo.FindByID(ctx, 1, active.ID)
	require.NoError(t, err)
	assert.True(t, gotActive.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.Cage{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestCheckupTypeRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()

	inactive := &model.CheckupType{ClinicID: 1, Name: "inactive checkup type", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.CheckupType{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.CheckupType{ClinicID: 1, Name: "active checkup type", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.CheckupType{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestChiefComplaintTypeRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	ctx := context.Background()

	inactive := &model.ChiefComplaintType{ClinicID: 1, Name: "inactive chief", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.ChiefComplaintType{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.ChiefComplaintType{ClinicID: 1, Name: "active chief", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.ChiefComplaintType{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestConsultationRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	ctx := context.Background()

	inactive := &model.Consultation{ClinicID: 1, Name: "inactive consultation", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.Consultation{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.Consultation{ClinicID: 1, Name: "active consultation", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.Consultation{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestDiagnosisTypeRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	ctx := context.Background()

	inactive := &model.DiagnosisType{ClinicID: 1, Name: "inactive dx type", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.DiagnosisType{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.DiagnosisType{ClinicID: 1, Name: "active dx type", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.DiagnosisType{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestDiagnosisNameRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	typeRepo := NewDiagnosisTypeRepository(db)
	nameRepo := NewDiagnosisNameRepository(db)
	ctx := context.Background()

	category := &model.DiagnosisType{ClinicID: 1, Name: "dx parent for is_active", IsActive: true}
	require.NoError(t, typeRepo.Create(ctx, category))

	inactive := &model.DiagnosisName{
		ClinicID: 1, Name: "inactive dx name", DiagnosisTypeID: category.ID, IsActive: false,
	}
	require.NoError(t, nameRepo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := nameRepo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.DiagnosisName{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.DiagnosisName{
		ClinicID: 1, Name: "active dx name", DiagnosisTypeID: category.ID, IsActive: true,
	}
	require.NoError(t, nameRepo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.DiagnosisName{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestExamTypeRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()

	inactive := &model.ExaminationType{ClinicID: 1, Name: "inactive exam type", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.ExaminationType{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.ExaminationType{ClinicID: 1, Name: "active exam type", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.ExaminationType{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestHospitalizationPlanRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	ctx := context.Background()

	inactive := &model.HospitalizationPlan{ClinicID: 1, Name: "inactive plan", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.HospitalizationPlan{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.HospitalizationPlan{ClinicID: 1, Name: "active plan", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.HospitalizationPlan{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestInquiryTemplateRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()

	inactive := &model.InquiryTemplate{ClinicID: 1, Title: "inactive template", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.InquiryTemplate{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.InquiryTemplate{ClinicID: 1, Title: "active template", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.InquiryTemplate{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestMedicineRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupMedicineRepositoryTestDB(t)
	repo := NewMedicineRepository(db)
	ctx := context.Background()

	inactive := &model.Medicine{ClinicID: 1, Name: "inactive medicine", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.Medicine{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.Medicine{ClinicID: 1, Name: "active medicine", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.Medicine{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestProcedureRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupProcedureRepositoryTestDB(t)
	repo := NewProcedureRepository(db)
	ctx := context.Background()

	inactive := &model.Procedure{ClinicID: 1, Name: "inactive procedure", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.Procedure{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.Procedure{ClinicID: 1, Name: "active procedure", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.Procedure{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}

func TestVaccineRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	ctx := context.Background()

	inactive := &model.Vaccine{ClinicID: 1, Name: "inactive vaccine", IsActive: false}
	require.NoError(t, repo.Create(ctx, inactive))
	assert.False(t, inactive.IsActive)
	got, err := repo.FindByID(ctx, 1, inactive.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	var raw bool
	require.NoError(t, db.WithContext(ctx).Model(&model.Vaccine{}).Select("is_active").Where("id = ?", inactive.ID).Scan(&raw).Error)
	assert.False(t, raw)

	active := &model.Vaccine{ClinicID: 1, Name: "active vaccine", IsActive: true}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)
	require.NoError(t, db.WithContext(ctx).Model(&model.Vaccine{}).Select("is_active").Where("id = ?", active.ID).Scan(&raw).Error)
	assert.True(t, raw)
}
