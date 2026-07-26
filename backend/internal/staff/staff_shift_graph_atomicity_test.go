package staff

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupStaffShiftGraphAtomicityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ShiftEntry{},
		&model.ShiftEntryBreak{},
		&model.ShiftTemplate{},
		&model.ShiftTemplateBreak{},
	))
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_entries ALTER COLUMN start_time TYPE time USING start_time::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_entries ALTER COLUMN end_time TYPE time USING end_time::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_entry_breaks ALTER COLUMN break_start TYPE time USING break_start::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_entry_breaks ALTER COLUMN break_end TYPE time USING break_end::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_templates ALTER COLUMN start_time TYPE time USING start_time::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_templates ALTER COLUMN end_time TYPE time USING end_time::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_template_breaks ALTER COLUMN break_start TYPE time USING break_start::time",
	).Error)
	require.NoError(t, db.Exec(
		"ALTER TABLE shift_template_breaks ALTER COLUMN break_end TYPE time USING break_end::time",
	).Error)
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE shift_entry_breaks, shift_entries, shift_template_breaks, "+
			"shift_templates, staff_clinic_assignments, staffs CASCADE",
	).Error)
	return db
}

func makeShiftGraphStaff(t *testing.T, db *gorm.DB) (*model.Clinic, *model.Staff) {
	t.Helper()
	company := &model.Company{Name: "シフトグラフ原子性テスト法人"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: "シフトグラフ原子性テスト医院", IsActive: true}
	require.NoError(t, db.Create(clinic).Error)
	staff := &model.Staff{
		ClinicID:  clinic.ID,
		Name:      "シフトグラフ原子性テストスタッフ",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinic.ID, IsMain: true,
	}).Error)
	return clinic, staff
}

func makeShiftGraphEntry(
	t *testing.T,
	db *gorm.DB,
	clinicID, staffID uint64,
) *model.ShiftEntry {
	t.Helper()
	start, end := "09:00:00", "17:00:00"
	entry := &model.ShiftEntry{
		ClinicID: clinicID, StaffID: staffID,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: model.ShiftTypeFull,
		StartTime: &start, EndTime: &end, Notes: "更新前",
	}
	require.NoError(t, db.Create(entry).Error)
	require.NoError(t, db.Create(&model.ShiftEntryBreak{
		ShiftEntryID: entry.ID, BreakStart: "12:00:00", BreakEnd: "13:00:00",
	}).Error)
	return entry
}

type failingShiftEntryBreakRepository struct {
	ShiftEntryRepository
	err error
}

func (r failingShiftEntryBreakRepository) ReplaceBreaks(
	ctx context.Context,
	shiftEntryID uint64,
	breaks []model.ShiftEntryBreak,
) error {
	if err := r.ShiftEntryRepository.ReplaceBreaks(ctx, shiftEntryID, breaks); err != nil {
		return err
	}
	return r.err
}

func TestShiftEntryServiceUpdate_RollsBackParentAndBreaksDatabase(t *testing.T) {
	db := setupStaffShiftGraphAtomicityDB(t)
	clinic, staff := makeShiftGraphStaff(t, db)
	entry := makeShiftGraphEntry(t, db, clinic.ID, staff.ID)
	sentinel := errors.New("fail after shift break replacement")
	repo := failingShiftEntryBreakRepository{
		ShiftEntryRepository: NewShiftEntryRepository(db),
		err:                  sentinel,
	}
	service := NewShiftEntryService(repo, nil, nil, persistence.NewTransactor(db))
	notes := "更新後"
	breaks := []ShiftBreakInput{{BreakStart: "14:00", BreakEnd: "14:30"}}

	updated, err := service.Update(context.Background(), clinic.ID, entry.ID, &UpdateShiftEntryInput{
		Notes: &notes, Breaks: &breaks,
	})

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, updated)
	var reloaded model.ShiftEntry
	require.NoError(t, db.First(&reloaded, entry.ID).Error)
	assert.Equal(t, "更新前", reloaded.Notes)
	var persisted []model.ShiftEntryBreak
	require.NoError(t, db.Where("shift_entry_id = ?", entry.ID).Find(&persisted).Error)
	require.Len(t, persisted, 1)
	assert.Equal(t, "12:00:00", persisted[0].BreakStart)
	assert.Equal(t, "13:00:00", persisted[0].BreakEnd)
}

type pausingShiftEntryLockRepository struct {
	ShiftEntryRepository
	locked  chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *pausingShiftEntryLockRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.ShiftEntry, error) {
	entry, err := r.ShiftEntryRepository.LockActiveByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	r.once.Do(func() {
		close(r.locked)
		<-r.release
	})
	return entry, nil
}

func awaitStaffTestSignal(t *testing.T, signal <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}

func awaitStaffTestError(t *testing.T, result <-chan error, label string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
		return nil
	}
}

func TestShiftEntryServiceUpdate_SerializesEffectiveTimeValidationDatabase(t *testing.T) {
	db := setupStaffShiftGraphAtomicityDB(t)
	clinic, staff := makeShiftGraphStaff(t, db)
	entry := makeShiftGraphEntry(t, db, clinic.ID, staff.ID)
	repo := &pausingShiftEntryLockRepository{
		ShiftEntryRepository: NewShiftEntryRepository(db),
		locked:               make(chan struct{}),
		release:              make(chan struct{}),
	}
	service := NewShiftEntryService(repo, nil, nil, persistence.NewTransactor(db))
	firstResult := make(chan error, 1)
	secondResult := make(chan error, 1)
	endTime := "19:00"
	startTime := "20:00"

	go func() {
		_, err := service.Update(context.Background(), clinic.ID, entry.ID, &UpdateShiftEntryInput{
			EndTime: &endTime,
		})
		firstResult <- err
	}()
	awaitStaffTestSignal(t, repo.locked, "first shift-entry row lock")
	go func() {
		_, err := service.Update(context.Background(), clinic.ID, entry.ID, &UpdateShiftEntryInput{
			StartTime: &startTime,
		})
		secondResult <- err
	}()
	close(repo.release)

	require.NoError(t, awaitStaffTestError(t, firstResult, "first shift update"))
	secondErr := awaitStaffTestError(t, secondResult, "second shift update")
	require.Error(t, secondErr)
	assert.True(t, apperrors.IsInvalidInput(secondErr), "unexpected error: %v", secondErr)
	var reloaded model.ShiftEntry
	require.NoError(t, db.First(&reloaded, entry.ID).Error)
	require.NotNil(t, reloaded.StartTime)
	require.NotNil(t, reloaded.EndTime)
	assert.Equal(t, "09:00:00", *reloaded.StartTime)
	assert.Equal(t, "19:00:00", *reloaded.EndTime)
}

type failingShiftTemplateBreakRepository struct {
	ShiftTemplateRepository
	err error
}

func (r failingShiftTemplateBreakRepository) UpdateBreaks(
	ctx context.Context,
	templateID uint64,
	breaks []model.ShiftTemplateBreak,
) error {
	if err := r.ShiftTemplateRepository.UpdateBreaks(ctx, templateID, breaks); err != nil {
		return err
	}
	return r.err
}

func makeShiftGraphTemplate(t *testing.T, db *gorm.DB, clinicID uint64) *model.ShiftTemplate {
	t.Helper()
	start, end := "09:00:00", "17:00:00"
	template := &model.ShiftTemplate{
		ClinicID: clinicID, Name: "更新前テンプレート",
		ShiftType: model.ShiftTypeFull, StartTime: &start, EndTime: &end, IsActive: true,
	}
	require.NoError(t, db.Omit("Breaks").Create(template).Error)
	require.NoError(t, db.Create(&model.ShiftTemplateBreak{
		ShiftTemplateID: template.ID, BreakStart: "12:00:00", BreakEnd: "13:00:00",
	}).Error)
	return template
}

func TestShiftTemplateServiceCreate_RollsBackParentWhenBreaksFailDatabase(t *testing.T) {
	db := setupStaffShiftGraphAtomicityDB(t)
	clinic, _ := makeShiftGraphStaff(t, db)
	sentinel := errors.New("fail after template break replacement")
	repo := failingShiftTemplateBreakRepository{
		ShiftTemplateRepository: NewShiftTemplateRepository(db),
		err:                     sentinel,
	}
	service := NewShiftTemplateService(repo)

	created, err := service.Create(context.Background(), clinic.ID, &CreateShiftTemplateInput{
		Name: "ロールバック対象テンプレート", ShiftType: string(model.ShiftTypeFull),
		StartTime: "09:00", EndTime: "17:00",
		Breaks: []ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00"}},
	})

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, created)
	var count int64
	require.NoError(t, db.Unscoped().Model(&model.ShiftTemplate{}).
		Where("name = ?", "ロールバック対象テンプレート").Count(&count).Error)
	assert.Zero(t, count)
}

func TestShiftTemplateServiceUpdate_RollsBackParentAndBreaksDatabase(t *testing.T) {
	db := setupStaffShiftGraphAtomicityDB(t)
	clinic, _ := makeShiftGraphStaff(t, db)
	template := makeShiftGraphTemplate(t, db, clinic.ID)
	sentinel := errors.New("fail after template break replacement")
	repo := failingShiftTemplateBreakRepository{
		ShiftTemplateRepository: NewShiftTemplateRepository(db),
		err:                     sentinel,
	}
	service := NewShiftTemplateService(repo)
	name := "更新後テンプレート"
	breaks := []ShiftBreakTemplateInput{{BreakStart: "15:00", BreakEnd: "15:30"}}

	updated, err := service.Update(context.Background(), clinic.ID, template.ID, &UpdateShiftTemplateInput{
		Name: &name, Breaks: &breaks,
	})

	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, updated)
	var reloaded model.ShiftTemplate
	require.NoError(t, db.First(&reloaded, template.ID).Error)
	assert.Equal(t, "更新前テンプレート", reloaded.Name)
	var persisted []model.ShiftTemplateBreak
	require.NoError(t, db.Where("shift_template_id = ?", template.ID).Find(&persisted).Error)
	require.Len(t, persisted, 1)
	assert.Equal(t, "12:00:00", persisted[0].BreakStart)
	assert.Equal(t, "13:00:00", persisted[0].BreakEnd)
}

func TestShiftTemplateServiceDelete_AllowsOwnedBreakChildrenDatabase(t *testing.T) {
	db := setupStaffShiftGraphAtomicityDB(t)
	clinic, _ := makeShiftGraphStaff(t, db)
	template := makeShiftGraphTemplate(t, db, clinic.ID)
	service := NewShiftTemplateService(NewShiftTemplateRepository(db))

	require.NoError(t, service.Delete(context.Background(), clinic.ID, template.ID))

	var activeCount int64
	require.NoError(t, db.Model(&model.ShiftTemplate{}).
		Where("id = ?", template.ID).Count(&activeCount).Error)
	assert.Zero(t, activeCount)
	var childCount int64
	require.NoError(t, db.Model(&model.ShiftTemplateBreak{}).
		Where("shift_template_id = ?", template.ID).Count(&childCount).Error)
	assert.Equal(t, int64(1), childCount, "owned children must not be treated as external usage")
}
