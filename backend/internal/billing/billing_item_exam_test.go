package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeBillingExam(
	t *testing.T,
	f billingItemReferenceFixture,
	name string,
	price *int64,
) (*model.ExaminationType, *model.Examination) {
	t.Helper()
	require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.ExaminationType{}, &model.Examination{}, &model.BillingItem{}, &model.BillingConfirmation{}))
	ensureConfirmedMedicalRecord(t, f)
	examType := &model.ExaminationType{
		ClinicID: f.clinicID,
		Name:     name,
		Price:    price,
		IsActive: true,
	}
	require.NoError(t, f.db.Create(examType).Error)
	exam := &model.Examination{
		ClinicID:        f.clinicID,
		MedicalRecordID: &f.medicalRecord.ID,
		PetID:           &f.pet.ID,
		ExamTypeID:      examType.ID,
		Date:            time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		Status:          model.ExaminationStatusCompleted,
	}
	require.NoError(t, f.db.Create(exam).Error)
	return examType, exam
}

func TestBillingItemExamProvenance_UnbilledCandidates(t *testing.T) {
	t.Run("derives exam type price after billing confirmation", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(4200)
		_, exam := makeBillingExam(t, f, "血液検査", &price)

		items, unbillable, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Zero(t, unbillable)
		assert.Equal(t, model.ItemCategoryTest, items[0].Category)
		assert.Equal(t, "血液検査", items[0].Name)
		assert.Equal(t, price, items[0].UnitPrice)
		require.NotNil(t, items[0].ExamID)
		assert.Equal(t, exam.ID, *items[0].ExamID)
	})

	t.Run("skips unpriced exam type as blocking unbillable", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		_, _ = makeBillingExam(t, f, "価格未設定検査", nil)

		items, unbillable, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 1, unbillable)
	})

	t.Run("excludes exam before billing confirmation", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(3000)
		require.NoError(t, testdb.EnsureAutoMigrated(f.db, &model.ExaminationType{}, &model.Examination{}))
		examType := &model.ExaminationType{ClinicID: f.clinicID, Name: "未確認検査", Price: &price, IsActive: true}
		require.NoError(t, f.db.Create(examType).Error)
		exam := &model.Examination{
			ClinicID:        f.clinicID,
			MedicalRecordID: &f.medicalRecord.ID,
			PetID:           &f.pet.ID,
			ExamTypeID:      examType.ID,
			Date:            time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, f.db.Create(exam).Error)

		items, unbillable, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Zero(t, unbillable)
	})
}

func TestBillingItemExamProvenance_DeleteNullsExamAndClinicID(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(3500)
	_, exam := makeBillingExam(t, f, "削除テスト検査", &price)
	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.ExamID = &exam.ID

	created, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, created.ExamID)
	require.NotNil(t, created.ClinicID)

	staffID := uint64(42)
	require.NoError(t, svc.DeleteItem(context.Background(), f.clinicID, created.ID, &DeleteBillingItemInput{StaffID: &staffID}))

	var released struct {
		ExamID   *uint64
		ClinicID *uint64
	}
	require.NoError(t, f.db.Unscoped().
		Table("billing_items").
		Select("exam_id", "clinic_id").
		Where("id = ?", created.ID).
		Take(&released).Error)
	assert.Nil(t, released.ExamID, "exam_id must be NULLed on delete to release provenance claim")
	assert.Nil(t, released.ClinicID, "clinic_id must be NULLed on delete")

	items, _, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
	require.NoError(t, err)
	require.Len(t, items, 1, "exam must reappear as unbilled after provenance release")
	require.NotNil(t, items[0].ExamID)
	assert.Equal(t, exam.ID, *items[0].ExamID)
}

func TestBillingItemExamProvenance_UsesCanonicalExamTypeValues(t *testing.T) {
	f := setupBillingItemReferenceFixture(t)
	price := int64(4200)
	examType, exam := makeBillingExam(t, f, "血液検査", &price)
	svc := newBillingItemReferenceService(f, f.repo)
	input := billingItemReferenceCreateInput(f)
	input.ExamID = &exam.ID
	input.Name = "caller controlled name"
	input.UnitPrice = 0

	created, err := svc.CreateItem(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, examType.Name, created.Name)
	assert.Equal(t, price, created.UnitPrice)
}
