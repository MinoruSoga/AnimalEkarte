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

	t.Run("carries non-insurance exam type setting to billing candidate", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(4200)
		examType, _ := makeBillingExam(t, f, "保険対象外検査", &price)
		require.NoError(t, f.db.Model(examType).Update("is_non_insurance", true).Error)

		items, unbillable, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Zero(t, unbillable)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].IsInsuranceApplicable)
		assert.False(t, *items[0].IsInsuranceApplicable)
	})

	t.Run("carries insurance-eligible exam type setting to billing candidate", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(4200)
		examType, _ := makeBillingExam(t, f, "保険対象検査", &price)
		require.NoError(t, f.db.Model(examType).Update("is_non_insurance", false).Error)

		items, _, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].IsInsuranceApplicable)
		assert.True(t, *items[0].IsInsuranceApplicable)
	})

	// ExaminationType.is_non_insurance is NOT NULL. A default false is an
	// explicit insured value, not a legacy NULL fixture (nullable kinds differ).
	// master 行が取得できない（削除済み）場合は候補化せず unbillable として止める。
	// 旧欠落を既定値へ落として会計行を作らない。
	t.Run("stops deleted master as blocking unbillable without speculation", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(4200)
		examType, _ := makeBillingExam(t, f, "削除済みマスタ検査", &price)
		require.NoError(t, f.db.Model(examType).Update("is_non_insurance", true).Error)
		require.NoError(t, f.db.Delete(examType).Error)

		items, unbillable, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 1, unbillable)
	})

	// EMR-454: 追加後の master 変更で保存済み会計行（snapshot）を書き換えない。
	// 会計済み exam は未請求候補から外れ、保存済み行の採用値（is_insurance_applicable）は
	// master 更新の前後で不変（参照値保持）。
	t.Run("keeps billed snapshot unchanged after master insurance setting changes", func(t *testing.T) {
		f := setupBillingItemReferenceFixture(t)
		price := int64(4200)
		examType, exam := makeBillingExam(t, f, "会計済み検査", &price)
		require.NoError(t, f.db.Model(examType).Update("is_non_insurance", true).Error)

		// 会計行として保存（候補の採用値 = 保険対象外 false を snapshot）
		require.NoError(t, f.db.Exec(
			`INSERT INTO billing_items (billing_id, clinic_id, category, name, unit_price, quantity, source, is_insurance_applicable, exam_id)
			 VALUES (?, ?, 'test', ?, ?, 1, 'medical_record', false, ?)`,
			f.billing.ID, f.clinicID, "会計済み検査", price, exam.ID,
		).Error)

		// 会計済みのため未請求候補には出ない
		items, _, err := f.repo.FindUnbilledExamItemsByPetID(context.Background(), f.clinicID, f.pet.ID)
		require.NoError(t, err)
		assert.Empty(t, items)

		// master を true → false へ変更しても、保存済み行の採用値は書き換わらない
		require.NoError(t, f.db.Model(examType).Update("is_non_insurance", false).Error)
		var after model.BillingItem
		require.NoError(t, f.db.Where("exam_id = ?", exam.ID).First(&after).Error)
		require.NotNil(t, after.IsInsuranceApplicable)
		assert.False(t, *after.IsInsuranceApplicable)
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
