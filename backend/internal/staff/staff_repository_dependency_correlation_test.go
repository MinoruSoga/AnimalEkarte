package staff_test

// staff_repository_dependency_correlation_test.go — SEC-SWEEP-02-STAFF-B1
// CountBlockingReferencesByStaffID の「子 clinic_id が論理的な親と食い違う」破損 fixture 実測。
// 対象: medical_record_addenda / vital_records / payments join。
// 直接 clinics FK のみの check（medical_records 等）は原理的に SEC-SWEEP-02 クラス対象外
// （本ファイル末尾の SchemaDirectClinicFKEvidence コメントと doctor_id 実測を参照）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func setupStaffDependencyCorrelationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Account{}, &model.Occupation{},
		&model.Staff{}, &model.StaffClinicAssignment{}, &model.ShiftEntry{},
		&model.MedicalRecordAddendum{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.VitalRecord{},
		&model.DailyRecord{},
	))
	// 共有 test DB 汚染を避けるため依存テーブルを明示 TRUNCATE。
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			vital_records,
			medical_record_addenda,
			daily_records,
			payments,
			billings,
			medical_records,
			pets,
			owners,
			animal_species,
			staff_clinic_assignments,
			staffs,
			occupations,
			accounts,
			shift_entries
		CASCADE
	`).Error)
	return db
}

func depCountByLabel(deps []StaffDependencyCount, label string) int64 {
	for _, d := range deps {
		if d.Label == label {
			return d.Count
		}
	}
	return 0
}

func makeCorrelationOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	o := &model.Owner{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

func makeCorrelationSpecies(t *testing.T, db *gorm.DB, name string) *model.AnimalSpecies {
	t.Helper()
	s := &model.AnimalSpecies{Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func makeCorrelationPet(t *testing.T, db *gorm.DB, clinicID, ownerID, speciesID uint64, name string) *model.Pet {
	t.Helper()
	p := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: speciesID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

// TestStaffRepository_CountBlockingReferencesByStaffID_Addenda_CorrelatesMedicalRecordClinic
// 破損: addenda.clinic_id = A だが medical_record_id が clinic B の親を指す。
// 旧実装は子 clinic_id だけで数えるため clinic A で過カウントする。
// 修復後は親 medical_records.clinic_id 相関により破損行を除外する。
func TestStaffRepository_CountBlockingReferencesByStaffID_Addenda_CorrelatesMedicalRecordClinic(t *testing.T) {
	db := setupStaffDependencyCorrelationTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "追記著者スタッフ")
	date := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	mrA := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-ADD-A", Date: date, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrA).Error)
	mrB := &model.MedicalRecord{ClinicID: clinicB, RecordNo: "MR-ADD-B", Date: date, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrB).Error)

	// 正常: 子・親とも clinicA
	legit := &model.MedicalRecordAddendum{
		MedicalRecordID: mrA.ID,
		ClinicID:        clinicA,
		AuthorUserID:    staff.ID,
		BeforeText:      "old",
		AfterText:       "new",
		Reason:          "legit",
	}
	require.NoError(t, db.WithContext(ctx).Create(legit).Error)

	// 破損: 子 clinicA、親 medical_records は clinicB
	polluted := &model.MedicalRecordAddendum{
		MedicalRecordID: mrB.ID,
		ClinicID:        clinicA,
		AuthorUserID:    staff.ID,
		BeforeText:      "old",
		AfterText:       "new",
		Reason:          "polluted-parent-B",
	}
	require.NoError(t, db.WithContext(ctx).Create(polluted).Error)

	depsA, err := repo.CountBlockingReferencesByStaffID(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), depCountByLabel(depsA, "カルテ追記"),
		"parent clinic 相関: 正常1件のみ。破損(addenda.clinic_id=A, parent=B)は数えない")

	depsB, err := repo.CountBlockingReferencesByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), depCountByLabel(depsB, "カルテ追記"),
		"破損行の子 clinic_id は A のため clinic B からも見えない（漏出なし）")
}

// TestStaffRepository_CountBlockingReferencesByStaffID_Vitals_CorrelatesMedicalRecordClinic
// 破損: vital.clinic_id = A だが medical_record_id が clinic B の親を指す。
func TestStaffRepository_CountBlockingReferencesByStaffID_Vitals_CorrelatesMedicalRecordClinic(t *testing.T) {
	db := setupStaffDependencyCorrelationTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "バイタル担当スタッフ")
	owner := makeCorrelationOwner(t, db, clinicA, "飼主A")
	species := makeCorrelationSpecies(t, db, "犬-staff-b1")
	pet := makeCorrelationPet(t, db, clinicA, owner.ID, species.ID, "ペットA")
	date := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	mrA := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-VIT-A", Date: date, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrA).Error)
	mrB := &model.MedicalRecord{ClinicID: clinicB, RecordNo: "MR-VIT-B", Date: date, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(ctx).Create(mrB).Error)

	legit := &model.VitalRecord{
		ClinicID:        clinicA,
		PetID:           pet.ID,
		MedicalRecordID: &mrA.ID,
		StaffID:         &staff.ID,
		RecordedAt:      time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC),
		Notes:           "legit",
	}
	require.NoError(t, db.WithContext(ctx).Create(legit).Error)

	polluted := &model.VitalRecord{
		ClinicID:        clinicA,
		PetID:           pet.ID,
		MedicalRecordID: &mrB.ID,
		StaffID:         &staff.ID,
		RecordedAt:      time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC),
		Notes:           "polluted-parent-B",
	}
	require.NoError(t, db.WithContext(ctx).Create(polluted).Error)

	depsA, err := repo.CountBlockingReferencesByStaffID(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), depCountByLabel(depsA, "バイタル記録"),
		"parent clinic 相関: 正常1件のみ。破損(vital.clinic_id=A, parent=B)は数えない")

	depsB, err := repo.CountBlockingReferencesByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), depCountByLabel(depsB, "バイタル記録"),
		"破損行の子 clinic_id は A のため clinic B からも見えない（漏出なし）")
}

// TestStaffRepository_CountBlockingReferencesByStaffID_Payments_ScopesViaBillingClinic
// payments は子 clinic_id を持たず billings.clinic_id で join 済み。
// billing が他院なら対象 clinic では 0、billing の医院では 1 を実測する。
func TestStaffRepository_CountBlockingReferencesByStaffID_Payments_ScopesViaBillingClinic(t *testing.T) {
	db := setupStaffDependencyCorrelationTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "支払い担当スタッフ")
	billingB := &model.Billing{
		ClinicID:      clinicB,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.WithContext(ctx).Create(billingB).Error)
	payment := &model.Payment{BillingID: billingB.ID, PaidBy: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(payment).Error)

	depsA, err := repo.CountBlockingReferencesByStaffID(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), depCountByLabel(depsA, "支払い"),
		"billing.clinic_id=B の payment は clinic A から数えない")

	depsB, err := repo.CountBlockingReferencesByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), depCountByLabel(depsB, "支払い"),
		"billing.clinic_id 経由で clinic B から正しく数える")
}

// TestStaffRepository_CountBlockingReferencesByStaffID_MedicalRecordsDoctor_DirectClinicFK
// medical_records は clinic_id が clinics への直接 FK。親中間エンティティ経由の相関欠陥は成立しない。
// 実測: 対象 clinic の行のみ数え、別 clinic の行は見えない。
func TestStaffRepository_CountBlockingReferencesByStaffID_MedicalRecordsDoctor_DirectClinicFK(t *testing.T) {
	db := setupStaffDependencyCorrelationTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "直接FKカルテストaff")
	date := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)
	mrA := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-DIR-A", Date: date, DoctorID: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(mrA).Error)
	// 同 staff_id を別 clinic 行に載せた破損相当（アプリ経路では稀だが DB 上は可能）
	mrB := &model.MedicalRecord{ClinicID: clinicB, RecordNo: "MR-DIR-B", Date: date, DoctorID: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(mrB).Error)

	depsA, err := repo.CountBlockingReferencesByStaffID(ctx, clinicA, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), depCountByLabel(depsA, "カルテ"))

	depsB, err := repo.CountBlockingReferencesByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), depCountByLabel(depsB, "カルテ"),
		"各行が自身の clinic_id で数えられる（直接 FK。親相関欠如による漏出クラスではない）")
}

// SchemaDirectClinicFKEvidence (SEC-SWEEP-02-STAFF-B1 Probe B)
//
// 以下は backend/migrations/001_init.sql で clinic_id REFERENCES clinics(id) を直接持ち、
// Count 述語もその列を使うため、grandchild→parent clinic 相関欠如クラスは原理的に成立しない:
//   - medical_records.doctor_id / entered_by
//   - hospitalizations.doctor_id
//   - exams.doctor_id
//   - shift_entries.staff_id
//   - billing_refunds.refunded_by
//   - cash_register_closes.closed_by
// 書き込み経路はいずれも request の clinicID を child.clinic_id に stamp する。
// 複合 FK (parent_id, clinic_id) は不要（親中間エンティティを経由しない）。
