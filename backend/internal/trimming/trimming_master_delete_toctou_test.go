package trimming

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type trimmingMasterDeleteTOCTOUFixture struct {
	db                 *gorm.DB
	deleteBackendPID   <-chan int
	deleteMaster       func(context.Context) error
	lockAndInsertUsage func(context.Context) error
	findMaster         func() error
	countUsage         func() (int64, error)
}

func reportTrimmingDeleteBackendPID(ctx context.Context, db *gorm.DB, backendPID chan<- int) error {
	var pid int
	if err := persistence.DBOrTx(ctx, db).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
		return err
	}
	select {
	case backendPID <- pid:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func setupTrimmingCourseDeleteTOCTOUFixture(t *testing.T) trimmingMasterDeleteTOCTOUFixture {
	t.Helper()
	db := setupTrimmingCourseTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	course := &model.TrimmingCourse{ClinicID: clinicID, Name: "TOCTOU course", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(course).Error)
	appointment := makeReservation(t, db, clinicID)

	baseRepo := NewTrimmingCourseRepository(db)
	deleteBackendPID := make(chan int, 1)
	serviceRepo := &mockTrimmingCourseRepository{
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			if err := reportTrimmingDeleteBackendPID(ctx, db, deleteBackendPID); err != nil {
				return err
			}
			return baseRepo.Delete(ctx, clinicID, id)
		},
		countUsageByCourseIDFn: baseRepo.CountUsageByTrimmingCourseID,
	}
	service := NewTrimmingCourseService(serviceRepo, &mockMinimalCourseTypeRepo{}, newTestTransactor(db))

	return trimmingMasterDeleteTOCTOUFixture{
		db:               db,
		deleteBackendPID: deleteBackendPID,
		deleteMaster: func(deleteCtx context.Context) error {
			return service.Delete(deleteCtx, clinicID, course.ID)
		},
		lockAndInsertUsage: func(txCtx context.Context) error {
			if _, err := baseRepo.FindByID(txCtx, clinicID, course.ID); err != nil {
				return err
			}
			return persistence.DBOrTx(txCtx, db).Create(&model.AppointmentTrimmingDetail{
				ClinicID:      clinicID,
				AppointmentID: appointment.ID,
				CourseID:      &course.ID,
			}).Error
		},
		findMaster: func() error {
			_, err := baseRepo.FindByID(ctx, clinicID, course.ID)
			return err
		},
		countUsage: func() (int64, error) {
			return baseRepo.CountUsageByTrimmingCourseID(ctx, clinicID, course.ID)
		},
	}
}

func setupTrimmingOptionDeleteTOCTOUFixture(t *testing.T) trimmingMasterDeleteTOCTOUFixture {
	t.Helper()
	db := setupTrimmingOptionTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	option := &model.TrimmingOption{ClinicID: clinicID, Name: "TOCTOU option", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(option).Error)
	appointment := makeReservation(t, db, clinicID)

	baseRepo := NewTrimmingOptionRepository(db)
	deleteBackendPID := make(chan int, 1)
	serviceRepo := &mockTrimmingOptionRepository{
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			if err := reportTrimmingDeleteBackendPID(ctx, db, deleteBackendPID); err != nil {
				return err
			}
			return baseRepo.Delete(ctx, clinicID, id)
		},
		countRecordsByOptFn: baseRepo.CountUsageByTrimmingOptionID,
	}
	service := NewTrimmingOptionService(serviceRepo, newTestTransactor(db))

	return trimmingMasterDeleteTOCTOUFixture{
		db:               db,
		deleteBackendPID: deleteBackendPID,
		deleteMaster: func(deleteCtx context.Context) error {
			return service.Delete(deleteCtx, clinicID, option.ID)
		},
		lockAndInsertUsage: func(txCtx context.Context) error {
			if _, err := baseRepo.FindByID(txCtx, clinicID, option.ID); err != nil {
				return err
			}
			return persistence.DBOrTx(txCtx, db).Create(&model.AppointmentTrimmingOption{
				AppointmentID: appointment.ID,
				OptionID:      option.ID,
			}).Error
		},
		findMaster: func() error {
			_, err := baseRepo.FindByID(ctx, clinicID, option.ID)
			return err
		},
		countUsage: func() (int64, error) {
			return baseRepo.CountUsageByTrimmingOptionID(ctx, clinicID, option.ID)
		},
	}
}

func setupTrimmingCourseTypeDeleteTOCTOUFixture(t *testing.T) trimmingMasterDeleteTOCTOUFixture {
	t.Helper()
	db := setupTrimmingCourseTypeTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	courseType := &model.TrimmingCourseType{ClinicID: clinicID, Name: "TOCTOU course type", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(courseType).Error)

	baseRepo := NewTrimmingCourseTypeRepository(db)
	deleteBackendPID := make(chan int, 1)
	serviceRepo := &mockTrimmingCourseTypeRepository{
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			if err := reportTrimmingDeleteBackendPID(ctx, db, deleteBackendPID); err != nil {
				return err
			}
			return baseRepo.Delete(ctx, clinicID, id)
		},
		countUsageFn: baseRepo.CountUsageByCourseTypeID,
	}
	service := NewTrimmingCourseTypeService(serviceRepo, newTestTransactor(db))

	return trimmingMasterDeleteTOCTOUFixture{
		db:               db,
		deleteBackendPID: deleteBackendPID,
		deleteMaster: func(deleteCtx context.Context) error {
			return service.Delete(deleteCtx, clinicID, courseType.ID)
		},
		lockAndInsertUsage: func(txCtx context.Context) error {
			if _, err := baseRepo.FindByID(txCtx, clinicID, courseType.ID); err != nil {
				return err
			}
			return persistence.DBOrTx(txCtx, db).Create(&model.TrimmingCourse{
				ClinicID:     clinicID,
				Name:         "course using type",
				CourseTypeID: &courseType.ID,
				IsActive:     true,
			}).Error
		},
		findMaster: func() error {
			_, err := baseRepo.FindByID(ctx, clinicID, courseType.ID)
			return err
		},
		countUsage: func() (int64, error) {
			return baseRepo.CountUsageByCourseTypeID(ctx, clinicID, courseType.ID)
		},
	}
}

func requireTrimmingDeleteWaitsForLock(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
	backendPID int,
	deleteDone <-chan error,
) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-deleteDone:
			require.Failf(t, "delete completed before waiting for the usage writer", "err=%v", err)
		case <-ticker.C:
			var waiting bool
			err := db.WithContext(ctx).
				Raw("SELECT EXISTS (SELECT 1 FROM pg_locks WHERE pid = ? AND NOT granted)", backendPID).
				Scan(&waiting).Error
			require.NoError(t, err)
			if waiting {
				return
			}
		case <-ctx.Done():
			require.Failf(t, "delete did not enter a database lock wait", "pid=%d err=%v", backendPID, ctx.Err())
		}
	}
}

func receiveTrimmingDeleteResult(t *testing.T, ctx context.Context, deleteDone <-chan error) error {
	t.Helper()
	select {
	case err := <-deleteDone:
		return err
	case <-ctx.Done():
		require.Failf(t, "delete did not complete before its deadline", "err=%v", ctx.Err())
		return ctx.Err()
	}
}

func TestTrimmingMasterDelete_TOCTOU(t *testing.T) {
	masters := []struct {
		name  string
		setup func(*testing.T) trimmingMasterDeleteTOCTOUFixture
	}{
		{name: "course", setup: setupTrimmingCourseDeleteTOCTOUFixture},
		{name: "option", setup: setupTrimmingOptionDeleteTOCTOUFixture},
		{name: "course_type", setup: setupTrimmingCourseTypeDeleteTOCTOUFixture},
	}
	outcomes := []struct {
		name                string
		commitUsage         bool
		wantDeleteConflict  bool
		wantMasterNotFound  bool
		wantCommittedUsages int64
	}{
		{
			name:                "usage commit wins and delete rolls back with conflict",
			commitUsage:         true,
			wantDeleteConflict:  true,
			wantCommittedUsages: 1,
		},
		{
			name:               "usage rollback lets delete commit",
			commitUsage:        false,
			wantMasterNotFound: true,
		},
	}

	for _, master := range masters {
		t.Run(master.name, func(t *testing.T) {
			for _, outcome := range outcomes {
				t.Run(outcome.name, func(t *testing.T) {
					fixture := master.setup(t)
					writerTx := fixture.db.WithContext(context.Background()).Begin()
					require.NoError(t, writerTx.Error)
					t.Cleanup(func() {
						_ = writerTx.Rollback().Error
					})
					writerCtx := persistence.WithTxValue(context.Background(), writerTx)
					require.NoError(t, fixture.lockAndInsertUsage(writerCtx))

					deleteCtx, cancelDelete := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancelDelete()
					deleteDone := make(chan error, 1)
					go func() {
						deleteDone <- fixture.deleteMaster(deleteCtx)
					}()
					select {
					case backendPID := <-fixture.deleteBackendPID:
						lockWaitCtx, cancelLockWait := context.WithTimeout(context.Background(), time.Second)
						requireTrimmingDeleteWaitsForLock(t, lockWaitCtx, fixture.db, backendPID, deleteDone)
						cancelLockWait()
					case err := <-deleteDone:
						require.Failf(t, "delete completed before reporting its database connection", "err=%v", err)
					case <-deleteCtx.Done():
						require.Failf(t, "delete did not issue its database operation", "err=%v", deleteCtx.Err())
					}

					if outcome.commitUsage {
						require.NoError(t, writerTx.Commit().Error)
					} else {
						require.NoError(t, writerTx.Rollback().Error)
					}

					deleteErr := receiveTrimmingDeleteResult(t, deleteCtx, deleteDone)
					if outcome.wantDeleteConflict {
						require.Error(t, deleteErr)
						assert.True(t, apperrors.IsConflict(deleteErr), "expected Conflict, got %v", deleteErr)
					} else {
						require.NoError(t, deleteErr)
					}

					findErr := fixture.findMaster()
					if outcome.wantMasterNotFound {
						require.Error(t, findErr)
						assert.True(t, apperrors.IsNotFound(findErr), "expected NotFound, got %v", findErr)
					} else {
						require.NoError(t, findErr)
					}

					usageCount, countErr := fixture.countUsage()
					require.NoError(t, countErr)
					assert.Equal(t, outcome.wantCommittedUsages, usageCount)
				})
			}
		})
	}
}
