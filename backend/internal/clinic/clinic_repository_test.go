package clinic

// clinic_repository_test.go — ClinicRepository の統合テスト。
//
// clinics / companies テーブルはどの setupTestDB / setupXTestDB からも TRUNCATE されない
// （本番同様の永続シードデータが乗っている前提）。そのため各テストは makeClinicFixture で
// 都度あたらしい company + clinic を作成し、他テストのシードデータや実行順に依存しない
// 完全に隔離された clinic_id を得てから検証する。

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupClinicTestDB は clinic_repository のテスト用に DB を整備する。
//
// CountBlockingReferencesByClinicID（clinic_repository.go）は appointments/medical_records/
// hospitalizations/exams/vaccinations/checkups/billings/clinic_settings/clinic_integrations/
// lstep_settings/permission_groups の11テーブルを参照する。以前はこれらの一部（exams・
// clinic_settings 等）を本 setup で AutoMigrate せず、同package内の他テストファイル
// （exam_type_repository_test.go 等）が先に実行されて偶然テーブルが出来ていることに暗黙依存していた。
// go test はファイル名のアルファベット順でテスト関数を登録するため、"clinic_repository_test.go" は
// "exam_type_repository_test.go" より先に実行され、この暗黙依存は本来常に破綻する
// （2026-07-15 CI: "relation \"exams\" does not exist" / "relation \"clinic_settings\" does not exist"
// で TestClinicRepository_CountBlockingReferencesByClinicID が決定論的に fail した）。
// 実行順序に依存しないよう、参照される全テーブルをここで明示的に整備する。
func setupClinicTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{}, &model.PermissionGroup{},
		&model.Reservation{},
		&model.ExaminationType{}, &model.ExamTypeField{}, &model.Examination{},
		&model.Hospitalization{},
		&model.Vaccination{},
		&model.CheckupType{}, &model.Checkup{},
		&model.ClinicIntegration{},
		&model.LstepSettings{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	return db
}

func makeClinicBillingFixture(t *testing.T, db *gorm.DB, clinicID uint64, amount int64, status model.BillingStatus, scheduledDate time.Time) {
	t.Helper()
	ctx := context.Background()
	billing := &model.Billing{
		ClinicID:      clinicID,
		TotalAmount:   amount,
		Status:        status,
		ScheduledDate: scheduledDate,
	}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)
}

func TestClinicRepository_FindAll(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "Aテスト診療所_順序確認")
	clinicZ := makeClinicFixture(t, db, "Zテスト診療所_順序確認")

	got, err := repo.FindAll(ctx)
	require.NoError(t, err)

	idxA, idxZ := -1, -1
	for i, c := range got {
		if c.ID == clinicA.ID {
			idxA = i
		}
		if c.ID == clinicZ.ID {
			idxZ = i
		}
	}
	require.NotEqual(t, -1, idxA, "clinicA が結果に含まれること")
	require.NotEqual(t, -1, idxZ, "clinicZ が結果に含まれること")
	assert.Less(t, idxA, idxZ, "name ASC 順で A が Z より前に来ること")
}

func TestClinicRepository_FindByStaffID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "スタッフ検索用A")
	clinicB := makeClinicFixture(t, db, "スタッフ検索用B")
	clinicC := makeClinicFixture(t, db, "スタッフ検索用Cソフト削除")

	staff := &model.Staff{ClinicID: clinicA.ID, Name: "兼務スタッフ", StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)

	// 有効な割当: clinic A, B
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA.ID, IsMain: true}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB.ID}).Error)

	// ソフト削除済み割当: clinic C（結果に含まれてはならない）
	assignC := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicC.ID}
	require.NoError(t, db.WithContext(ctx).Create(assignC).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assignC).Error)

	got, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)

	ids := make([]uint64, 0, len(got))
	for _, c := range got {
		ids = append(ids, c.ID)
	}
	assert.Contains(t, ids, clinicA.ID)
	assert.Contains(t, ids, clinicB.ID)
	assert.NotContains(t, ids, clinicC.ID, "ソフト削除された割当のクリニックは含まれない")
}

func TestClinicRepository_FindByID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinic := makeClinicFixture(t, db, "単件取得テスト")

	t.Run("存在するIDは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinic.Name, got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 999888001)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestClinicRepository_LockActiveByID_SourceContract(t *testing.T) {
	source, err := os.ReadFile("../clinic/clinic_repository.go")
	require.NoError(t, err)

	const methodSignature = "func (r *Repository) LockActiveByID("
	const nextMethodSignature = "func (r *Repository) FindByID("
	methodStart := bytes.Index(source, []byte(methodSignature))
	require.NotEqual(t, -1, methodStart)
	methodEndOffset := bytes.Index(source[methodStart:], []byte(nextMethodSignature))
	require.NotEqual(t, -1, methodEndOffset)
	methodSource := string(source)[methodStart : methodStart+methodEndOffset]

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, `clause.Locking{Strength: "SHARE"}`)
	assert.Contains(t, methodSource, `Where("id = ? AND is_active = ?", id, true)`)
	assert.Contains(t, methodSource, `apperrors.FromGORM`)
}

func TestClinicRepository_LockActiveByID_RequiresAmbientTransaction(t *testing.T) {
	repo := NewClinicRepository(nil)

	clinic, err := repo.LockActiveByID(context.Background(), 1)

	assert.Nil(t, clinic)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestClinicRepository_LockActiveByID_ActiveAndNotFound(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()

	activeClinic := makeClinicFixture(t, db, "active clinic lock target")
	inactiveClinic := makeClinicFixture(t, db, "inactive clinic lock target")
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Clinic{}).
		Where("id = ?", inactiveClinic.ID).
		Update("is_active", false).Error)

	t.Run("active clinic is returned", func(t *testing.T) {
		var locked *model.Clinic
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			var lockErr error
			locked, lockErr = repo.LockActiveByID(txCtx, activeClinic.ID)
			return lockErr
		})

		require.NoError(t, err)
		require.NotNil(t, locked)
		assert.Equal(t, activeClinic.ID, locked.ID)
		assert.True(t, locked.IsActive)
	})

	t.Run("inactive clinic is NotFound", func(t *testing.T) {
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			_, lockErr := repo.LockActiveByID(txCtx, inactiveClinic.ID)
			return lockErr
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("missing clinic is NotFound", func(t *testing.T) {
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			_, lockErr := repo.LockActiveByID(txCtx, 999888004)
			return lockErr
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestClinicRepository_LockActiveByID_HoldsShareLockUntilTransactionEnds(t *testing.T) {
	tests := []struct {
		name           string
		transactionErr error
	}{
		{name: "commit"},
		{name: "rollback", transactionErr: errors.New("force lock holder rollback")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupClinicTestDB(t)
			repo := NewClinicRepository(db)
			transactor := persistence.NewTransactor(db)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			clinic := makeClinicFixture(t, db, "share lock target "+tt.name)

			locked := make(chan struct{})
			release := make(chan struct{})
			holderDone := make(chan error, 1)
			go func() {
				holderDone <- transactor.WithTx(ctx, func(txCtx context.Context) error {
					if _, err := repo.LockActiveByID(txCtx, clinic.ID); err != nil {
						return err
					}
					close(locked)
					select {
					case <-release:
						return tt.transactionErr
					case <-ctx.Done():
						return ctx.Err()
					}
				})
			}()

			select {
			case <-locked:
			case <-ctx.Done():
				t.Fatal("clinic SHARE lock was not acquired")
			}

			deleteDone := make(chan error, 1)
			go func() {
				deleteDone <- repo.Delete(ctx, clinic.ID)
			}()

			var deleteErr error
			completedBeforeRelease := false
			select {
			case deleteErr = <-deleteDone:
				completedBeforeRelease = true
			case <-time.After(100 * time.Millisecond):
			}
			close(release)

			holderErr := <-holderDone
			if tt.transactionErr == nil {
				require.NoError(t, holderErr)
			} else {
				require.ErrorIs(t, holderErr, tt.transactionErr)
			}
			if !completedBeforeRelease {
				select {
				case deleteErr = <-deleteDone:
				case <-ctx.Done():
					t.Fatal("clinic deletion did not resume after the lock transaction ended")
				}
			}
			require.NoError(t, deleteErr)
			assert.False(t, completedBeforeRelease, "clinic delete must wait for the SHARE lock")
		})
	}
}

func TestClinicRepository_LockByIDForUpdate_SourceContract(t *testing.T) {
	source, err := os.ReadFile("../clinic/clinic_repository.go")
	require.NoError(t, err)

	const methodSignature = "func (r *Repository) LockByIDForUpdate("
	const nextMethodSignature = "func (r *Repository) FindByID("
	methodStart := bytes.Index(source, []byte(methodSignature))
	require.NotEqual(t, -1, methodStart)
	methodEndOffset := bytes.Index(source[methodStart:], []byte(nextMethodSignature))
	require.NotEqual(t, -1, methodEndOffset)
	methodSource := string(source)[methodStart : methodStart+methodEndOffset]

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, `clause.Locking{Strength: "UPDATE"}`)
	assert.Contains(t, methodSource, `Where("id = ?", id)`)
	assert.NotContains(t, methodSource, "is_active")
	assert.Contains(t, methodSource, `apperrors.FromGORM`)
}

func TestClinicRepository_LockByIDForUpdate_RequiresAmbientTransaction(t *testing.T) {
	repo := NewClinicRepository(nil)

	clinic, err := repo.LockByIDForUpdate(context.Background(), 1)

	assert.Nil(t, clinic)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestClinicRepository_LockByIDForUpdate_IncludesInactiveAndReturnsNotFound(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()

	activeClinic := makeClinicFixture(t, db, "active clinic update lock target")
	inactiveClinic := makeClinicFixture(t, db, "inactive clinic update lock target")
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Clinic{}).
		Where("id = ?", inactiveClinic.ID).
		Update("is_active", false).Error)

	t.Run("active clinic is returned", func(t *testing.T) {
		var locked *model.Clinic
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			var lockErr error
			locked, lockErr = repo.LockByIDForUpdate(txCtx, activeClinic.ID)
			return lockErr
		})

		require.NoError(t, err)
		require.NotNil(t, locked)
		assert.Equal(t, activeClinic.ID, locked.ID)
		assert.True(t, locked.IsActive)
	})

	t.Run("inactive clinic is returned for deletion", func(t *testing.T) {
		var locked *model.Clinic
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			var lockErr error
			locked, lockErr = repo.LockByIDForUpdate(txCtx, inactiveClinic.ID)
			return lockErr
		})

		require.NoError(t, err)
		require.NotNil(t, locked)
		assert.Equal(t, inactiveClinic.ID, locked.ID)
		assert.False(t, locked.IsActive)
	})

	t.Run("missing clinic is NotFound", func(t *testing.T) {
		err := transactor.WithTx(ctx, func(txCtx context.Context) error {
			_, lockErr := repo.LockByIDForUpdate(txCtx, 999888005)
			return lockErr
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestClinicRepository_LockByIDForUpdate_HoldsExclusiveLockUntilTransactionEnds(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clinic := makeClinicFixture(t, db, "update lock target")

	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- transactor.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := repo.LockByIDForUpdate(txCtx, clinic.ID); err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()

	select {
	case <-locked:
	case <-ctx.Done():
		t.Fatal("clinic UPDATE lock was not acquired")
	}

	shareDone := make(chan error, 1)
	go func() {
		shareDone <- transactor.WithTx(ctx, func(txCtx context.Context) error {
			_, err := repo.LockActiveByID(txCtx, clinic.ID)
			return err
		})
	}()

	completedBeforeRelease := false
	var shareErr error
	select {
	case shareErr = <-shareDone:
		completedBeforeRelease = true
	case <-time.After(100 * time.Millisecond):
	}
	close(release)

	require.NoError(t, <-holderDone)
	if !completedBeforeRelease {
		select {
		case shareErr = <-shareDone:
		case <-ctx.Done():
			t.Fatal("clinic SHARE lock did not resume after the UPDATE lock transaction ended")
		}
	}
	require.NoError(t, shareErr)
	assert.False(t, completedBeforeRelease, "clinic SHARE lock must wait for the UPDATE lock")
}

func TestClinicRepository_FindCompany(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	// companies は永続テーブルのため、既存シードが無ければ最小限の1件を用意する。
	var existing int64
	db.Model(&model.Company{}).Count(&existing)
	if existing == 0 {
		require.NoError(t, db.WithContext(ctx).Create(&model.Company{Name: "初期法人"}).Error)
	}

	got, err := repo.FindCompany(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.Name)
}

func TestClinicRepository_Create(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	company := &model.Company{Name: "Create用法人"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	clinic := &model.Clinic{CompanyID: company.ID, Name: "新規作成クリニック"}
	err := repo.Create(ctx, clinic)
	require.NoError(t, err)
	assert.NotZero(t, clinic.ID)

	got, err := repo.FindByID(ctx, clinic.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成クリニック", got.Name)
	assert.InDelta(t, 0.10, got.StandardTaxRate, 0.0001, "未指定時は DB デフォルト税率が適用される")
}

func TestClinicRepository_UpdateClinic(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinic := makeClinicFixture(t, db, "更新前クリニック")

	t.Run("成功", func(t *testing.T) {
		name := "更新後クリニック"
		err := repo.UpdateClinic(ctx, clinic.ID, &UpdateClinicInput{Name: &name})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後クリニック", got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		name := "x"
		err := repo.UpdateClinic(ctx, 999888002, &UpdateClinicInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("空入力はno-op", func(t *testing.T) {
		err := repo.UpdateClinic(ctx, clinic.ID, &UpdateClinicInput{})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後クリニック", got.Name)
	})
}

func TestClinicRepository_Delete(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	t.Run("子データのないクリニックは削除できる", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "削除対象クリニック")
		err := repo.Delete(ctx, clinic.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinic.ID)
		assert.True(t, apperrors.IsNotFound(err), "削除後は NotFound になるべき")
	})

	t.Run("直接削除はPermissionGroup所有行を書き換えない", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "PG付き削除対象クリニック")
		pg := &model.PermissionGroup{ClinicID: clinic.ID, Name: "削除予定グループ"}
		require.NoError(t, db.WithContext(ctx).Create(pg).Error)
		require.NoError(t, db.WithContext(ctx).Delete(pg).Error) // ソフト削除

		rollbackErr := errors.New("rollback direct repository delete")
		err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			require.NoError(t, repo.Delete(txCtx, clinic.ID))

			var count int64
			require.NoError(t, persistence.DBOrTx(txCtx, db).Unscoped().Model(&model.PermissionGroup{}).Where("id = ?", pg.ID).Count(&count).Error)
			assert.Equal(t, int64(1), count, "clinic repository must not delete permission-group-owned rows")
			return rollbackErr
		})
		require.ErrorIs(t, err, rollbackErr)

		_, findErr := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, findErr, "test rollback must preserve the clinic")
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, 999888003)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("飼主が紐付いていれば Conflict で行は残る", func(t *testing.T) {
		require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Owner{}))
		clinic := makeClinicFixture(t, db, "飼主付き削除拒否クリニック")
		testdb.MakeTestOwner(t, db, clinic.ID, "削除阻止飼主")

		err := repo.Delete(ctx, clinic.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		got, findErr := repo.FindByID(ctx, clinic.ID)
		require.NoError(t, findErr)
		assert.Equal(t, clinic.ID, got.ID)
	})
}

func TestClinicPermissionGroupCleanupDelete_RollsBackTogether(t *testing.T) {
	db := setupClinicTestDB(t)
	clinicRepo := NewClinicRepository(db)
	permissionGroupRepo := auth.NewPermissionGroupRepository(db)
	transactor := persistence.NewTransactor(db)
	ctx := context.Background()

	clinic := makeClinicFixture(t, db, "PG cleanup/delete rollback clinic")
	group := &model.PermissionGroup{ClinicID: clinic.ID, Name: "rollback soft-deleted group"}
	require.NoError(t, db.WithContext(ctx).Create(group).Error)
	require.NoError(t, db.WithContext(ctx).Delete(group).Error)
	rollbackErr := errors.New("force cleanup/delete rollback")

	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		if cleanupErr := permissionGroupRepo.DeleteSoftDeletedByClinicID(txCtx, clinic.ID); cleanupErr != nil {
			return cleanupErr
		}
		if deleteErr := clinicRepo.Delete(txCtx, clinic.ID); deleteErr != nil {
			return deleteErr
		}
		return rollbackErr
	})

	require.ErrorIs(t, err, rollbackErr)
	_, findErr := clinicRepo.FindByID(ctx, clinic.ID)
	require.NoError(t, findErr, "clinic deletion must roll back")
	var groupCount int64
	require.NoError(t, db.Unscoped().
		Model(&model.PermissionGroup{}).
		Where("id = ?", group.ID).
		Count(&groupCount).Error)
	assert.Equal(t, int64(1), groupCount, "permission-group cleanup must roll back with clinic delete")
}

func TestClinicRepository_CountOwnersByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "飼主数A")
	clinicB := makeClinicFixture(t, db, "飼主数B")

	testdb.MakeTestOwner(t, db, clinicA.ID, "飼主1")
	testdb.MakeTestOwner(t, db, clinicA.ID, "飼主2")
	deletedOwner := testdb.MakeTestOwner(t, db, clinicA.ID, "削除済み飼主")
	require.NoError(t, db.WithContext(ctx).Delete(deletedOwner).Error)
	testdb.MakeTestOwner(t, db, clinicB.ID, "別クリニック飼主")

	got, err := repo.CountOwnersByClinicID(ctx, clinicA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got, "ソフト削除・別クリニックを除外して2件")
}

func TestClinicRepository_CountStaffByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	clinicA := makeClinicFixture(t, db, "スタッフ数A")
	clinicB := makeClinicFixture(t, db, "スタッフ数B")

	staff1 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ1", StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.WithContext(ctx).Create(staff1).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff1.ID, ClinicID: clinicA.ID, IsMain: true}).Error)

	staff2 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ2", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff2).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff2.ID, ClinicID: clinicA.ID}).Error)

	// ソフト削除された割当は除外される
	staff3 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ3(割当ソフト削除)", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff3).Error)
	assign3 := &model.StaffClinicAssignment{StaffID: staff3.ID, ClinicID: clinicA.ID}
	require.NoError(t, db.WithContext(ctx).Create(assign3).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assign3).Error)

	// 別クリニックのみに割当のスタッフは対象外
	staff4 := &model.Staff{ClinicID: clinicB.ID, Name: "スタッフ4(別クリニックのみ)", StaffType: model.StaffTypeNurse}
	require.NoError(t, db.WithContext(ctx).Create(staff4).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staff4.ID, ClinicID: clinicB.ID, IsMain: true}).Error)

	got, err := repo.CountStaffByClinicID(ctx, clinicA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got)
}

// Fixed (#236 root cause, 2026-07-13): model.ClinicSettings now carries explicit gorm
// "type:time"/"column:" tags matching backend/migrations/001_init.sql, so AutoMigrate no
// longer fails and this table dependency check runs normally. See
// TestClinicSettingsRepository_* in clinic_settings_repository_test.go for the same fix.
func TestClinicRepository_CountBlockingReferencesByClinicID(t *testing.T) {
	db := setupClinicTestDB(t)
	repo := NewClinicRepository(db)
	ctx := context.Background()

	t.Run("依存データが無ければ空スライス", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "依存なしクリニック")
		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("会計データがあれば件数付きでラベルが返る", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "会計依存クリニック")
		makeClinicBillingFixture(t, db, clinic.ID, 1000, model.BillingStatusWaiting, time.Now())
		makeClinicBillingFixture(t, db, clinic.ID, 2000, model.BillingStatusWaiting, time.Now())

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "会計", got[0].Label)
		assert.Equal(t, int64(2), got[0].Count)
	})

	t.Run("ソフト削除された会計は除外される(P2)", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "会計ソフト削除クリニック")
		makeClinicBillingFixture(t, db, clinic.ID, 1000, model.BillingStatusWaiting, time.Now())

		var b model.Billing
		require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", clinic.ID).First(&b).Error)
		require.NoError(t, db.WithContext(ctx).Delete(&b).Error)

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		assert.Empty(t, got, "唯一の会計がソフト削除されれば依存なしになる")
	})

	// Fixed (#236 root cause, 2026-07-13): see comment above TestClinicRepository_CountBlockingReferencesByClinicID.
	t.Run("clinic_settingsはソフトデリート対象外テーブルとして検出される", func(t *testing.T) {
		clinic := makeClinicFixture(t, db, "医院設定依存クリニック")
		require.NoError(t, db.WithContext(ctx).Create(&model.ClinicSettings{ClinicID: clinic.ID}).Error)

		got, err := repo.CountBlockingReferencesByClinicID(ctx, clinic.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "医院設定", got[0].Label)
		assert.Equal(t, int64(1), got[0].Count)
	})
}
