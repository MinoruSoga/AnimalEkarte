package staff_test

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - FindAll / FindByID / ExistsByStaffID は clinic_id でテナント隔離される。
//   - FindAll / FindByID の Staff preload は requested clinic の有効な所属があるスタッフだけを返す。
//   - FindAll は YearMonth / StaffID フィルタが正しく機能する。
//   - Create は (clinic_id, staff_id, date) の重複で AlreadyExists を返す（001_init.sql の
//     uk_shift_staff_date を模した複合 UNIQUE を AutoMigrate 後に明示追加して検証する）。
//   - Update / Delete は対象なしで NotFound を返す。
//   - ReplaceBreaks は既存 breaks を削除してから新しい breaks を作成する（空スライスなら全削除のみ）。
//   - FindOnDutyStaffs は退職(deleted_at)・非アクティブ(is_active=false)スタッフを除外する。
//
// makeDoctor は親 repository パッケージの isolation_test_helpers_test.go 内の同名ヘルパーの複製
// （BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupShiftEntryTestDB は shift_entries / shift_entry_breaks / staffs /
// staff_clinic_assignments と FK 親を整備する。
// 本番マイグレーションの複合 UNIQUE(clinic_id, staff_id, date) は GORM モデルの
// uniqueIndex タグが無いため AutoMigrate では再現されない。Create の重複検知を
// 意味のある形で検証するため、明示的に複合 UNIQUE INDEX を追加する。
func setupShiftEntryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.ShiftEntry{}, &model.ShiftEntryBreak{},
	))
	// shift_entry_breaks.break_start/break_end can be left as "timestamp with time zone" in a
	// pre-existing ekarte_db_test if the table was ever created before model.ShiftEntryBreak's
	// `gorm:"type:time"` tag existed — AutoMigrate never ALTERs an existing column's type, only
	// adds missing ones. Force the correct type every run (no-op cast if already "time").
	db.Exec(`ALTER TABLE shift_entry_breaks ALTER COLUMN break_start TYPE time USING break_start::time`)
	db.Exec(`ALTER TABLE shift_entry_breaks ALTER COLUMN break_end TYPE time USING break_end::time`)
	db.Exec("TRUNCATE TABLE shift_entry_breaks CASCADE")
	db.Exec("TRUNCATE TABLE shift_entries CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_shift_entry_staff_date
		ON shift_entries (clinic_id, staff_id, date)`)
	return db
}

func makeShiftEntryDoctor(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Staff {
	t.Helper()
	seedClinicsForFK(t, db, clinicID)
	staff := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		StaffType: model.StaffTypeDoctor,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	makeStaffClinicAssignment(t, db, staff.ID, clinicID)
	return staff
}

func makeShiftEntryWithType(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time, shiftType model.ShiftType) *model.ShiftEntry {
	t.Helper()
	e := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: shiftType,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e
}

func TestShiftEntryRepository_FindAll(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staffA1 := makeShiftEntryDoctor(t, db, clinicA, "医師A1")
	staffA2 := makeShiftEntryDoctor(t, db, clinicA, "医師A2")
	staffB := makeShiftEntryDoctor(t, db, clinicB, "医師B")

	june1 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	june2 := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	july1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	makeShiftEntryWithType(t, db, clinicA, staffA1.ID, june1, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, staffA2.ID, june2, model.ShiftTypeMorning)
	makeShiftEntryWithType(t, db, clinicA, staffA1.ID, july1, model.ShiftTypeOff)
	makeShiftEntryWithType(t, db, clinicB, staffB.ID, june1, model.ShiftTypeFull)

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{})
		require.NoError(t, err)
		require.Len(t, got, 3)
		for _, e := range got {
			assert.Equal(t, clinicA, e.ClinicID)
		}
	})

	t.Run("filters by YearMonth", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{YearMonth: "2026-06"})
		require.NoError(t, err)
		require.Len(t, got, 2)
		for _, e := range got {
			assert.Equal(t, 6, int(e.Date.Month()))
		}
	})

	t.Run("filters by StaffID", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{StaffID: &staffA1.ID})
		require.NoError(t, err)
		require.Len(t, got, 2)
		for _, e := range got {
			assert.Equal(t, staffA1.ID, e.StaffID)
		}
	})

	t.Run("invalid YearMonth returns invalid input error", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{YearMonth: "not-a-date"})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("preloads Staff", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{StaffID: &staffA1.ID})
		require.NoError(t, err)
		require.NotEmpty(t, got)
		require.NotNil(t, got[0].Staff)
		assert.Equal(t, "医師A1", got[0].Staff.Name)
	})
}

func TestShiftEntryRepository_FindAll_HidesForeignStaffWithoutRequestedClinicAssignment(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	foreignStaff := makeShiftEntryDoctor(t, db, clinicB, "別医院スタッフ")
	entry := makeShiftEntryWithType(
		t,
		db,
		clinicA,
		foreignStaff.ID,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		model.ShiftTypeFull,
	)

	got, err := repo.FindAll(ctx, clinicA, ShiftEntryFilter{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, entry.ID, got[0].ID, "clinic-scoped shift entry 自体は取得する")
	assert.Nil(t, got[0].Staff, "requested clinic に所属しない staff identity は preload しない")
}

func TestShiftEntryRepository_FindByID(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	staff := makeShiftEntryDoctor(t, db, clinicA, "医師A")
	entry := makeShiftEntryWithType(t, db, clinicA, staff.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), model.ShiftTypeFull)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, entry.ID)
		require.NoError(t, err)
		assert.Equal(t, entry.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, entry.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftEntryRepository_FindByID_HidesStaffWithSoftDeletedRequestedClinicAssignment(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	staff := makeShiftEntryDoctor(t, db, clinicID, "所属解除済みスタッフ")
	require.NoError(t, db.WithContext(ctx).
		Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
		Delete(&model.StaffClinicAssignment{}).Error)
	entry := makeShiftEntryWithType(
		t,
		db,
		clinicID,
		staff.ID,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		model.ShiftTypeFull,
	)

	got, err := repo.FindByID(ctx, clinicID, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, got.ID, "clinic-scoped shift entry 自体は取得する")
	assert.Nil(t, got.Staff, "soft-delete 済み assignment は有効な所属として扱わない")
}

func TestShiftEntryRepository_Create(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	staff := makeShiftEntryDoctor(t, db, clinicA, "医師A")
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	t.Run("creates successfully", func(t *testing.T) {
		entry := &model.ShiftEntry{ClinicID: clinicA, StaffID: staff.ID, Date: date, ShiftType: model.ShiftTypeFull}
		require.NoError(t, repo.Create(ctx, entry))
		assert.NotZero(t, entry.ID)
	})

	t.Run("duplicate (clinic_id, staff_id, date) returns AlreadyExists", func(t *testing.T) {
		dup := &model.ShiftEntry{ClinicID: clinicA, StaffID: staff.ID, Date: date, ShiftType: model.ShiftTypeMorning}
		err := repo.Create(ctx, dup)
		assert.Error(t, err)
		assert.True(t, apperrors.IsAlreadyExists(err), "重複エラーは AlreadyExists であるべき: %v", err)
	})
}

func TestShiftEntryRepository_Update(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	staff := makeShiftEntryDoctor(t, db, clinicA, "医師A")
	entry := makeShiftEntryWithType(t, db, clinicA, staff.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), model.ShiftTypeFull)

	t.Run("updates successfully", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, entry.ID, map[string]any{"notes": "更新済み"}))
		got, err := repo.FindByID(ctx, clinicA, entry.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新済み", got.Notes)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, uint64(999999), map[string]any{"notes": "x"})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, entry.ID, map[string]any{"notes": "hacked"})
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftEntryRepository_Delete(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	staff := makeShiftEntryDoctor(t, db, clinicA, "医師A")
	entry := makeShiftEntryWithType(t, db, clinicA, staff.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), model.ShiftTypeFull)

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, entry.ID)
		assert.True(t, apperrors.IsNotFound(err))

		var count int64
		require.NoError(t, db.Model(&model.ShiftEntry{}).Where("id = ?", entry.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count, "別クリニックからの削除は反映されない")
	})

	t.Run("deletes successfully (hard delete)", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, entry.ID))

		var count int64
		require.NoError(t, db.Model(&model.ShiftEntry{}).Where("id = ?", entry.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, entry.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftEntryRepository_ExistsByStaffID(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	staffWithEntry := makeShiftEntryDoctor(t, db, clinicA, "医師あり")
	staffWithoutEntry := makeShiftEntryDoctor(t, db, clinicA, "医師なし")
	makeShiftEntryWithType(t, db, clinicA, staffWithEntry.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), model.ShiftTypeFull)

	t.Run("true when staff has shift entries", func(t *testing.T) {
		got, err := repo.ExistsByStaffID(ctx, clinicA, staffWithEntry.ID)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("false when staff has no shift entries", func(t *testing.T) {
		got, err := repo.ExistsByStaffID(ctx, clinicA, staffWithoutEntry.ID)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("clinic isolation: other clinic's entry does not count", func(t *testing.T) {
		got, err := repo.ExistsByStaffID(ctx, clinicB, staffWithEntry.ID)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestShiftEntryRepository_ReplaceBreaks(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	staff := makeShiftEntryDoctor(t, db, clinicA, "医師A")
	entry := makeShiftEntryWithType(t, db, clinicA, staff.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), model.ShiftTypeFull)

	t.Run("creates breaks", func(t *testing.T) {
		breaks := []model.ShiftEntryBreak{
			{BreakStart: "12:00", BreakEnd: "13:00"},
		}
		require.NoError(t, repo.ReplaceBreaks(ctx, entry.ID, breaks))

		var count int64
		require.NoError(t, db.Model(&model.ShiftEntryBreak{}).Where("shift_entry_id = ?", entry.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("replaces existing breaks with new ones", func(t *testing.T) {
		breaks := []model.ShiftEntryBreak{
			{BreakStart: "10:00", BreakEnd: "10:15"},
			{BreakStart: "15:00", BreakEnd: "15:15"},
		}
		require.NoError(t, repo.ReplaceBreaks(ctx, entry.ID, breaks))

		var got []model.ShiftEntryBreak
		require.NoError(t, db.Where("shift_entry_id = ?", entry.ID).Order("break_start ASC").Find(&got).Error)
		require.Len(t, got, 2)
		// Postgres の time 型カラムは文字列スキャン時 "HH:MM:SS" 形式で返る（入力が "HH:MM" でも同様）。
		assert.Equal(t, "10:00:00", got[0].BreakStart)
	})

	t.Run("empty slice deletes all breaks without creating new ones", func(t *testing.T) {
		require.NoError(t, repo.ReplaceBreaks(ctx, entry.ID, []model.ShiftEntryBreak{}))

		var count int64
		require.NoError(t, db.Model(&model.ShiftEntryBreak{}).Where("shift_entry_id = ?", entry.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}

func TestShiftEntryRepository_FindOnDutyStaffs(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	otherDate := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	onDuty := makeShiftEntryDoctor(t, db, clinicA, "当番医")
	offDuty := makeShiftEntryDoctor(t, db, clinicA, "非番医")
	inactive := makeShiftEntryDoctor(t, db, clinicA, "非アクティブ医")
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", inactive.ID).Update("is_active", false).Error)
	retired := makeShiftEntryDoctor(t, db, clinicA, "退職医")

	makeShiftEntryWithType(t, db, clinicA, onDuty.ID, date, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, offDuty.ID, otherDate, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, inactive.ID, date, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, retired.ID, date, model.ShiftTypeFull)
	require.NoError(t, db.Delete(retired).Error)

	otherClinicStaff := makeShiftEntryDoctor(t, db, clinicB, "別クリニック医")
	makeShiftEntryWithType(t, db, clinicB, otherClinicStaff.ID, date, model.ShiftTypeFull)

	got, err := repo.FindOnDutyStaffs(ctx, clinicA, date)
	require.NoError(t, err)
	require.Len(t, got, 1, "非番・非アクティブ・退職・別クリニックは除外される")
	assert.Equal(t, onDuty.ID, got[0].ID)
}

func TestShiftEntryRepository_FindOnDutyStaffs_RequiresActiveClinicAssignment(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	date := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	valid := makeShiftEntryDoctor(t, db, clinicA, "有効所属スタッフ")
	corrupt := makeShiftEntryDoctor(t, db, clinicB, "別医院単独スタッフ")
	revoked := makeShiftEntryDoctor(t, db, clinicA, "所属解除スタッフ")
	require.NoError(t, db.WithContext(ctx).
		Where("staff_id = ? AND clinic_id = ?", revoked.ID, clinicA).
		Delete(&model.StaffClinicAssignment{}).Error)
	makeShiftEntryWithType(t, db, clinicA, valid.ID, date, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, corrupt.ID, date, model.ShiftTypeFull)
	makeShiftEntryWithType(t, db, clinicA, revoked.ID, date, model.ShiftTypeFull)

	got, err := repo.FindOnDutyStaffs(ctx, clinicA, date)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, valid.ID, got[0].ID)
}

func TestShiftEntryRepository_DeleteByStaffDate_JoinsAmbientTransaction(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	date := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	staff := makeShiftEntryDoctor(t, db, clinicID, "削除rollbackスタッフ")
	entry := makeShiftEntryWithType(t, db, clinicID, staff.ID, date, model.ShiftTypeFull)
	sentinel := errors.New("rollback delete by staff date")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.DeleteByStaffDate(txCtx, clinicID, staff.ID, date))
		return sentinel
	})

	require.ErrorIs(t, err, sentinel)
	var count int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("id = ?", entry.ID).Count(&count).Error)
	assert.EqualValues(t, 1, count, "outer rollback must restore the shift entry")
}
