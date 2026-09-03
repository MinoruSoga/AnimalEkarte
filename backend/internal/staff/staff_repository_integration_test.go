package staff_test

// staff_repository_test.go — StaffRepository の統合テスト（実 Postgres テスト DB）。
//
// staff_occupation_preload_clinic_isolation_test.go / staff_preload_clinic_isolation_test.go が
// Staff.Occupation / Staff.Doctor(reservation) preload の clinic_id 隔離を既にカバーしているため、
// 本ファイルは StaffRepository 自体の CRUD・多医院所属ゲート（Update/Delete/UpdatePrimaryClinicID/
// Reorder が staff_clinic_assignments EXISTS で許可判定する挙動）・CountBlockingReferencesByStaffID を
// 対象とする。seedClinicsForFK / makeStaffClinicAssignment / makeDoctor は他ファイルで定義済みのため再利用する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

// setupStaffRepositoryTestDB は staff_repository.go のテスト用に FK 親（companies/clinics/accounts/
// occupations）と staffs/staff_clinic_assignments/shift_entries を整備する。
func setupStaffRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Account{}, &model.Occupation{},
		&model.Staff{}, &model.StaffClinicAssignment{}, &model.ShiftEntry{},
		&model.Reservation{},
		&model.MedicalRecord{}, &model.Hospitalization{}, &model.Examination{},
		&model.BillingRefund{}, &model.CashRegisterClose{},
		&model.Billing{}, &model.Payment{},
		&model.MedicalRecordAddendum{}, &model.VitalRecord{}, &model.DailyRecord{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs, occupations, accounts, shift_entries CASCADE").Error)
	return db
}

func makeStaffAccount(t *testing.T, db *gorm.DB, email string) *model.Account {
	t.Helper()
	a := &model.Account{Email: email, PasswordHash: "hash"}
	require.NoError(t, db.WithContext(context.Background()).Create(a).Error)
	return a
}

// ---- FindByID ----

func TestStaffRepository_FindByID_HappyPathPreloadsAccountAndOccupation(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc := makeStaffAccount(t, db, "staff-findbyid@example.com")
	occ := makeOccupation(t, db, clinicID, "獣医師")
	staff := &model.Staff{ClinicID: clinicID, Name: "検索対象スタッフ", StaffType: model.StaffTypeDoctor, AccountID: &acc.ID, OccupationID: &occ.ID}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, "検索対象スタッフ", got.Name)
	require.NotNil(t, got.Account)
	assert.Equal(t, acc.ID, got.Account.ID)
	require.NotNil(t, got.Occupation)
	assert.Equal(t, occ.ID, got.Occupation.ID)
}

func TestStaffRepository_FindByID_HidesOccupationOutsideAuthoritativeClinic(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	foreignOccupation := makeOccupation(t, db, clinicB, "医院B職種")
	staff := &model.Staff{
		ClinicID:     clinicA,
		Name:         "GetMe相当のidentity read対象",
		StaffType:    model.StaffTypeDoctor,
		OccupationID: &foreignOccupation.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)

	got, err := repo.FindByID(ctx, staff.ID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.Occupation, "authoritative clinic 外の職種を identity response に載せない")
	assert.Nil(t, got.OccupationID, "foreign master FK 自体も GetMe 相当 response から除外する")
}

func TestStaffRepository_FindByID_NotFoundForNonexistentID(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999999)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
}

func TestStaffRepository_FindByID_ExcludesSoftDeleted(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "削除予定スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	require.NoError(t, repo.Delete(ctx, clinicID, staff.ID))

	_, err := repo.FindByID(ctx, staff.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "ソフトデリート後は NotFound であるべき: %v", err)
}

func TestStaffRepository_FindByIDInClinic_HidesForeignOccupationForSharedStaff(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	foreignOccupation := makeOccupation(t, db, clinicB, "医院B職種")
	staff := &model.Staff{
		ClinicID:     clinicB,
		Name:         "多施設所属スタッフ",
		StaffType:    model.StaffTypeDoctor,
		OccupationID: &foreignOccupation.ID,
	}
	require.NoError(t, db.Create(staff).Error)
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)
	makeStaffClinicAssignment(t, db, staff.ID, clinicB)

	scoped, ok := repo.(interface {
		FindByIDInClinic(context.Context, uint64, uint64) (*model.Staff, error)
	})
	require.True(t, ok, "repository must expose a clinic-scoped staff detail read")

	got, err := scoped.FindByIDInClinic(ctx, clinicA, staff.ID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.Occupation)
	assert.Nil(t, got.OccupationID, "foreign master FK must not be serialized")
}

// ---- FindByAccountID ----

func TestStaffRepository_FindByAccountID_HappyPath(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	acc := makeStaffAccount(t, db, "staff-byaccount@example.com")
	staff := &model.Staff{ClinicID: clinicID, Name: "アカウント紐付きスタッフ", StaffType: model.StaffTypeDoctor, AccountID: &acc.ID}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)

	got, err := repo.FindByAccountID(ctx, acc.ID)
	require.NoError(t, err)
	assert.Equal(t, staff.ID, got.ID)
}

func TestStaffRepository_FindByAccountID_NotFoundForNonexistentAccountID(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	_, err := repo.FindByAccountID(ctx, 999999)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// ---- Create ----

func TestStaffRepository_Create_HappyPath(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := &model.Staff{ClinicID: clinicID, Name: "新規作成スタッフ", StaffType: model.StaffTypeDoctor}
	require.NoError(t, repo.Create(ctx, staff))
	require.NotZero(t, staff.ID)

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成スタッフ", got.Name)
}

// BUG-455-S7: gorm default:true omits zero bools from INSERT.
func TestStaffRepository_Create_ReservationVisibleFalsePersists(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := &model.Staff{
		ClinicID:           clinicID,
		Name:               "hidden staff",
		StaffType:          model.StaffTypeDoctor,
		ReservationVisible: false,
	}
	require.NoError(t, repo.Create(ctx, staff))
	assert.False(t, staff.ReservationVisible)

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.False(t, got.ReservationVisible)

	var raw bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Staff{}).
		Select("reservation_visible").
		Where("id = ?", staff.ID).
		Scan(&raw).Error)
	assert.False(t, raw, "raw reservation_visible must be false")
}

// ---- FindAll ----

func TestStaffRepository_FindAll_OnlyAssignedStaffVisible(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staffA := makeDoctor(t, db, clinicA, "医院A所属スタッフ")
	makeStaffClinicAssignment(t, db, staffA.ID, clinicA)
	// 医院Bのみに所属するスタッフ（clinicAには見えないはず）
	staffB := makeDoctor(t, db, clinicB, "医院B所属スタッフ")
	makeStaffClinicAssignment(t, db, staffB.ID, clinicB)

	got, total, err := repo.FindAll(ctx, clinicA, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, staffA.ID, got[0].ID)
}

func TestStaffRepository_FindAll_ExcludesSoftDeletedAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "退職スタッフ")
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}
	require.NoError(t, db.WithContext(ctx).Create(assignment).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assignment).Error) // ソフトデリート

	got, total, err := repo.FindAll(ctx, clinicID, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "ソフトデリート済み所属は INNER JOIN 条件から除外される")
	assert.Empty(t, got)
}

func TestStaffRepository_FindAll_ExcludesSoftDeletedStaff(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "削除予定スタッフ2")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	require.NoError(t, repo.Delete(ctx, clinicID, staff.ID))

	got, total, err := repo.FindAll(ctx, clinicID, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, got)
}

func TestStaffRepository_FindAll_Pagination(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	names := []string{"ページスタッフ1", "ページスタッフ2", "ページスタッフ3"}
	for _, name := range names {
		s := makeDoctor(t, db, clinicID, name)
		makeStaffClinicAssignment(t, db, s.ID, clinicID)
	}

	got, total, err := repo.FindAll(ctx, clinicID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, got, 2, "page1 limit2 → 2件")
}

// ---- Update ----

func TestStaffRepository_Update_HappyPathWithAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "更新前スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)

	require.NoError(t, repo.Update(ctx, clinicID, staff.ID, map[string]any{"name": "更新後スタッフ"}))

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新後スタッフ", got.Name)
}

func TestStaffRepository_Update_NotFoundWithoutAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	// clinicA のみに所属するスタッフを clinicB から更新しようとする
	staff := makeDoctor(t, db, clinicA, "無関係スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)

	err := repo.Update(ctx, clinicB, staff.ID, map[string]any{"name": "改ざん試行"})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "配属のないクリニックからの更新は NotFound: %v", err)

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, "無関係スタッフ", got.Name, "改ざんされていないこと")
}

// ---- UpdatePrimaryClinicID ----

func TestStaffRepository_UpdatePrimaryClinicID_HappyPath(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "主所属変更対象")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)
	makeStaffClinicAssignment(t, db, staff.ID, clinicB)

	require.NoError(t, repo.UpdatePrimaryClinicID(ctx, staff.ID, clinicB))

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Equal(t, clinicB, got.ClinicID)
}

func TestStaffRepository_UpdatePrimaryClinicID_NotFoundWithoutAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "未配属変更対象")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)

	err := repo.UpdatePrimaryClinicID(ctx, staff.ID, clinicB)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "clinicB への未配属では主所属変更できない: %v", err)
}

// ---- Delete ----

func TestStaffRepository_Delete_ConflictWhenShiftEntryExists(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "シフト使用中スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staff.ID,
		Date:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		ShiftType: model.ShiftTypeFull,
	}
	require.NoError(t, db.WithContext(ctx).Create(entry).Error)

	err := repo.Delete(ctx, clinicID, staff.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

	got, findErr := repo.FindByID(ctx, staff.ID)
	require.NoError(t, findErr)
	assert.Equal(t, staff.ID, got.ID)
}

func TestStaffRepository_Delete_HappyPathSoftDeletes(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "削除確認スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)

	require.NoError(t, repo.Delete(ctx, clinicID, staff.ID))

	_, err := repo.FindByID(ctx, staff.ID)
	assert.True(t, apperrors.IsNotFound(err))

	// 行自体は DB に残っている（ソフトデリート）
	var rawCount int64
	db.Unscoped().Model(&model.Staff{}).Where("id = ?", staff.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount, "ソフトデリートされた行はDBにまだ存在する")
}

func TestStaffRepository_Delete_NotFoundWithoutAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "削除拒否対象")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)

	err := repo.Delete(ctx, clinicB, staff.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	got, err := repo.FindByID(ctx, staff.ID)
	require.NoError(t, err, "配属のないクリニックからの削除で消えてはならない")
	assert.Equal(t, staff.ID, got.ID)
}

func TestStaffRepository_Delete_ConflictForMultipleActiveAssignmentsPreservesIdentityAndAssignments(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "複数医院所属スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)
	makeStaffClinicAssignment(t, db, staff.ID, clinicB)

	var staffBefore model.Staff
	require.NoError(t, db.WithContext(ctx).Unscoped().First(&staffBefore, "id = ?", staff.ID).Error)
	var assignmentsBefore []model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ?", staff.ID).
		Order("id ASC").
		Find(&assignmentsBefore).Error)
	require.Len(t, assignmentsBefore, 2)

	err := repo.Delete(ctx, clinicA, staff.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "複数の有効な所属がある場合は Conflict: %v", err)

	var staffAfter model.Staff
	require.NoError(t, db.WithContext(ctx).Unscoped().First(&staffAfter, "id = ?", staff.ID).Error)
	assert.Equal(t, staffBefore, staffAfter, "Conflict 時は staff identity を変更しない")
	assert.False(t, staffAfter.DeletedAt.Valid)

	var assignmentsAfter []model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ?", staff.ID).
		Order("id ASC").
		Find(&assignmentsAfter).Error)
	assert.Equal(t, assignmentsBefore, assignmentsAfter, "Conflict 時は全 clinic assignment を変更しない")
}

func TestStaffRepository_Delete_SoftDeletedOtherAssignmentDoesNotConflict(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "有効所属一件スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, clinicA)
	inactiveAssignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, db.WithContext(ctx).Create(inactiveAssignment).Error)
	require.NoError(t, db.WithContext(ctx).Delete(inactiveAssignment).Error)

	require.NoError(t, repo.Delete(ctx, clinicA, staff.ID))

	var deletedStaff model.Staff
	require.NoError(t, db.WithContext(ctx).Unscoped().First(&deletedStaff, "id = ?", staff.ID).Error)
	assert.True(t, deletedStaff.DeletedAt.Valid, "無効な別所属は active assignment 数に含めない")

	var assignments []model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ?", staff.ID).
		Order("id ASC").
		Find(&assignments).Error)
	require.Len(t, assignments, 2)
	assert.False(t, assignments[0].DeletedAt.Valid, "現在 clinic の assignment は変更しない")
	assert.True(t, assignments[1].DeletedAt.Valid, "既存の soft-delete 状態を変更しない")
}

// ---- Reorder ----

func TestStaffRepository_Reorder_HappyPath(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	s1 := makeDoctor(t, db, clinicID, "並び順スタッフ1")
	s2 := makeDoctor(t, db, clinicID, "並び順スタッフ2")
	s3 := makeDoctor(t, db, clinicID, "並び順スタッフ3")
	makeStaffClinicAssignment(t, db, s1.ID, clinicID)
	makeStaffClinicAssignment(t, db, s2.ID, clinicID)
	makeStaffClinicAssignment(t, db, s3.ID, clinicID)

	require.NoError(t, repo.Reorder(ctx, clinicID, []uint64{s3.ID, s1.ID, s2.ID}))

	got, _, err := repo.FindAll(ctx, clinicID, 1, 100)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, s3.ID, got[0].ID)
	assert.Equal(t, s1.ID, got[1].ID)
	assert.Equal(t, s2.ID, got[2].ID)
}

func TestStaffRepository_Reorder_FailsForIDOutsideClinic(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	s1 := makeDoctor(t, db, clinicA, "医院Aスタッフ")
	makeStaffClinicAssignment(t, db, s1.ID, clinicA)
	sOther := makeDoctor(t, db, clinicB, "医院Bスタッフ")
	makeStaffClinicAssignment(t, db, sOther.ID, clinicB)

	err := repo.Reorder(ctx, clinicA, []uint64{s1.ID, sOther.ID})
	require.Error(t, err, "clinicA に配属のない ID を含む Reorder は失敗すべき")
}

func TestStaffRepository_Reorder_RejectsSecondaryClinicAssignment(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const primaryClinic, secondaryClinic = uint64(2), uint64(1)
	seedClinicsForFK(t, db, primaryClinic, secondaryClinic)

	staff := makeDoctor(t, db, primaryClinic, "共有スタッフ")
	makeStaffClinicAssignment(t, db, staff.ID, primaryClinic)
	makeStaffClinicAssignment(t, db, staff.ID, secondaryClinic)

	err := repo.Reorder(ctx, secondaryClinic, []uint64{staff.ID})
	require.Error(t, err, "secondary clinic must not mutate the primary-owned sort_order")

	var reloaded model.Staff
	require.NoError(t, db.First(&reloaded, staff.ID).Error)
	assert.Equal(t, staff.SortOrder, reloaded.SortOrder)
}

// ---- CountBlockingReferencesByStaffID ----

func TestStaffRepository_CountBlockingReferencesByStaffID_ZeroWhenNoReferences(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "参照なしスタッフ")

	deps, err := repo.CountBlockingReferencesByStaffID(ctx, clinicID, staff.ID)
	require.NoError(t, err)
	assert.Empty(t, deps)
}

func TestStaffRepository_CountBlockingReferencesByStaffID_CountsMedicalRecordsExcludingSoftDeleted(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "カルテ参照スタッフ")

	active := &model.MedicalRecord{ClinicID: clinicID, RecordNo: "R-0001", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), DoctorID: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(active).Error)
	deleted := &model.MedicalRecord{ClinicID: clinicID, RecordNo: "R-0002", Date: time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC), DoctorID: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(deleted).Error)
	require.NoError(t, db.WithContext(ctx).Delete(deleted).Error)

	deps, err := repo.CountBlockingReferencesByStaffID(ctx, clinicID, staff.ID)
	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, "カルテ", deps[0].Label)
	assert.Equal(t, int64(1), deps[0].Count, "ソフトデリート済みカルテは数えない")
}

func TestStaffRepository_CountBlockingReferencesByStaffID_CountsShiftEntries(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "シフト参照スタッフ")
	entry := &model.ShiftEntry{ClinicID: clinicID, StaffID: staff.ID, Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), ShiftType: model.ShiftTypeFull}
	require.NoError(t, db.WithContext(ctx).Create(entry).Error)

	deps, err := repo.CountBlockingReferencesByStaffID(ctx, clinicID, staff.ID)
	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, "シフト", deps[0].Label)
	assert.Equal(t, int64(1), deps[0].Count)
}

func TestStaffRepository_CountBlockingReferencesByStaffID_CountsPaymentsViaBillingsJoin(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "支払い参照スタッフ")
	billing := &model.Billing{ClinicID: clinicID, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)
	payment := &model.Payment{BillingID: billing.ID, PaidBy: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(payment).Error)

	deps, err := repo.CountBlockingReferencesByStaffID(ctx, clinicID, staff.ID)
	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, "支払い", deps[0].Label)
	assert.Equal(t, int64(1), deps[0].Count)
}

func TestStaffRepository_CountBlockingReferencesByStaffID_ClinicIsolation(t *testing.T) {
	db := setupStaffRepositoryTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "隔離検証スタッフ")
	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "R-ISO", Date: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), DoctorID: &staff.ID}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	// clinicB から同じ staff_id を問い合わせても医院Aのカルテは見えない
	deps, err := repo.CountBlockingReferencesByStaffID(ctx, clinicB, staff.ID)
	require.NoError(t, err)
	assert.Empty(t, deps, "別クリニックからは参照が0件であるべき")
}
