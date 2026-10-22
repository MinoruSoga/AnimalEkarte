package medicalrecord

// EMR-456 F1: 親 soft-delete と子 Create/親変更 Update の実 DB 直列化 fixture。
// 2 接続の実 transaction と明示 barrier（channel）で両取得順序を検証する。
// test DB が無い環境では testdb.SetupTestDB が skip する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupRealDBCheckupParentLock(t *testing.T) (*gorm.DB, CheckupTypeService, CheckupTypeRepository, uint64) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Owner{},
		&model.AnimalSpecies{}, &model.Pet{}, &model.MedicalRecord{},
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.CheckupType{}, &model.Checkup{},
	))
	testdb.Truncate(t, db,
		"checkups", "checkup_types",
		"staff_clinic_assignments", "staffs",
		"medical_records", "pets", "animal_species", "owners",
	)
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr checkup-parent-lock")
	repo := NewCheckupTypeRepository(db)
	svc := NewCheckupTypeService(repo, persistenceTx(db), &recordingCheckupTypeAudit{})
	return db, svc, repo, fx.ClinicA
}

func seedRealDBCheckupType(t *testing.T, db *gorm.DB, clinicID uint64, name string, parentID *uint64) *model.CheckupType {
	t.Helper()
	row := &model.CheckupType{ClinicID: clinicID, Name: name, IsActive: true, ParentID: parentID}
	require.NoError(t, db.WithContext(context.Background()).Create(row).Error)
	return row
}

// holdCheckupTypeParentLock は「先に lock を取る側」の transaction を goroutine で開く。
// lock 取得を locked で通知し、release を受けてから write を実行して commit する。
func holdCheckupTypeParentLock(db *gorm.DB, ctx context.Context, clinicID uint64, ids []uint64,
	locked chan<- struct{}, release <-chan struct{}, write func(txCtx context.Context) error) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- persistenceTx(db).WithTx(ctx, func(txCtx context.Context) error {
			locker, ok := NewCheckupTypeRepository(db).(masterWriteRowsLocker)
			if !ok {
				return errors.New("checkup type repository must implement masterWriteRowsLocker")
			}
			if err := locker.LockWriteRows(txCtx, clinicID, ids); err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
			return write(txCtx)
		})
	}()
	return done
}

// assertCheckupTypeLockBlocks は lock 保持中に 2 番目の操作が完了しないことを
// 短い待機で観測する（reservation write gate realdb 試験と同じ手法）。
func assertCheckupTypeLockBlocks(t *testing.T, done <-chan error, op string) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("%s must wait on the held lock (finished early: %v)", op, err)
	case <-time.After(300 * time.Millisecond):
	}
}

func assertCheckupTypeSoftDeleted(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	var deleted bool
	require.NoError(t, db.Raw(`SELECT deleted_at IS NOT NULL FROM checkup_types WHERE id = ?`, id).Scan(&deleted).Error)
	require.True(t, deleted, "checkup type %d must be soft-deleted", id)
}

// 子 Create が先に親 lock を取った場合: Delete は lock 待ち後の snapshot で子を数え、
// conflict で失敗する。子保存と親 soft-delete は両立しない。
func TestCheckupTypeParentLock_RealDB_ChildCreateBeforeParentDelete(t *testing.T) {
	db, svc, repo, clinicID := setupRealDBCheckupParentLock(t)
	ctx := context.Background()
	parent := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-parent-A", nil)

	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := holdCheckupTypeParentLock(db, ctx, clinicID, []uint64{parent.ID}, locked, release,
		func(txCtx context.Context) error {
			if _, err := repo.FindByID(txCtx, clinicID, parent.ID); err != nil {
				return err
			}
			return repo.Create(txCtx, &model.CheckupType{ClinicID: clinicID, Name: "D3mr-lock-child-A", IsActive: true, ParentID: &parent.ID})
		})
	opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	secondDone := make(chan error, 1)
	go func() {
		<-locked
		secondDone <- svc.Delete(opCtx, clinicID, parent.ID)
	}()
	<-locked
	assertCheckupTypeLockBlocks(t, secondDone, "parent delete")
	close(release)
	require.NoError(t, <-firstDone)
	require.True(t, apperrors.IsConflict(<-secondDone), "committed child must conflict the delete")
	if _, err := repo.FindByID(ctx, clinicID, parent.ID); err != nil {
		t.Fatalf("parent must survive: %v", err)
	}
	children, err := repo.CountChildrenByParentID(ctx, clinicID, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), children)
}

// 親 Delete が先に lock を取った場合: 子 Create は lock 待ち後の snapshot で
// soft-delete 済みの親を読めず失敗する。親削除と子保存は両立しない。
func TestCheckupTypeParentLock_RealDB_ParentDeleteBeforeChildCreate(t *testing.T) {
	db, svc, repo, clinicID := setupRealDBCheckupParentLock(t)
	ctx := context.Background()
	parent := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-parent-B", nil)
	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := holdCheckupTypeParentLock(db, ctx, clinicID, []uint64{parent.ID}, locked, release,
		func(txCtx context.Context) error {
			return repo.Delete(txCtx, clinicID, parent.ID)
		})
	opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	secondDone := make(chan error, 1)
	go func() {
		<-locked
		_, err := svc.Create(opCtx, clinicID, &CreateCheckupTypeInput{
			Name: "D3mr-lock-child-B", ParentID: &parent.ID, ActorID: &testCheckupTypeActorID,
		})
		secondDone <- err
	}()
	<-locked
	assertCheckupTypeLockBlocks(t, secondDone, "child create")
	close(release)
	require.NoError(t, <-firstDone)
	require.Error(t, <-secondDone)
	assertCheckupTypeSoftDeleted(t, db, parent.ID)
	children, err := repo.CountChildrenByParentID(ctx, clinicID, parent.ID)
	require.NoError(t, err)
	assert.Zero(t, children)
}

// 親変更 Update が先に lock を取った場合: Delete は commit 済みの子を数えて conflict になる。
func TestCheckupTypeParentLock_RealDB_ParentUpdateBeforeParentDelete(t *testing.T) {
	db, svc, repo, clinicID := setupRealDBCheckupParentLock(t)
	ctx := context.Background()
	parent := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-parent-C", nil)
	child := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-child-C", nil)
	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := holdCheckupTypeParentLock(db, ctx, clinicID, []uint64{child.ID, parent.ID}, locked, release,
		func(txCtx context.Context) error {
			before, err := repo.FindByID(txCtx, clinicID, child.ID)
			if err != nil {
				return err
			}
			if _, err := repo.FindByID(txCtx, clinicID, parent.ID); err != nil {
				return err
			}
			version := before.Version
			_, err = repo.Update(txCtx, clinicID, child.ID, UpdateCheckupTypeInput{ParentID: &parent.ID, expectedVersion: &version})
			return err
		})
	opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	secondDone := make(chan error, 1)
	go func() {
		<-locked
		secondDone <- svc.Delete(opCtx, clinicID, parent.ID)
	}()
	<-locked
	assertCheckupTypeLockBlocks(t, secondDone, "parent delete")
	close(release)
	require.NoError(t, <-firstDone)
	require.True(t, apperrors.IsConflict(<-secondDone), "committed parent change must conflict the delete")
	got, err := repo.FindByID(ctx, clinicID, child.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ParentID)
	assert.Equal(t, parent.ID, *got.ParentID)
	if _, err := repo.FindByID(ctx, clinicID, parent.ID); err != nil {
		t.Fatalf("parent must survive: %v", err)
	}
}

// 親 Delete が先に lock を取った場合: 親変更 Update は lock 待ち後の snapshot で
// soft-delete 済みの親を読めず失敗し、子は変わらない。
func TestCheckupTypeParentLock_RealDB_ParentDeleteBeforeParentUpdate(t *testing.T) {
	db, svc, repo, clinicID := setupRealDBCheckupParentLock(t)
	ctx := context.Background()
	parent := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-parent-D", nil)
	child := seedRealDBCheckupType(t, db, clinicID, "D3mr-lock-child-D", nil)
	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := holdCheckupTypeParentLock(db, ctx, clinicID, []uint64{parent.ID}, locked, release,
		func(txCtx context.Context) error {
			return repo.Delete(txCtx, clinicID, parent.ID)
		})
	opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	secondDone := make(chan error, 1)
	go func() {
		<-locked
		_, err := svc.Update(opCtx, clinicID, child.ID, &UpdateCheckupTypeInput{
			ParentID: &parent.ID, Version: int64Ptr(1), ActorID: &testCheckupTypeActorID,
		})
		secondDone <- err
	}()
	<-locked
	assertCheckupTypeLockBlocks(t, secondDone, "parent update")
	close(release)
	require.NoError(t, <-firstDone)
	require.Error(t, <-secondDone)
	assertCheckupTypeSoftDeleted(t, db, parent.ID)
	got, err := repo.FindByID(ctx, clinicID, child.ID)
	require.NoError(t, err)
	assert.Nil(t, got.ParentID)
}
