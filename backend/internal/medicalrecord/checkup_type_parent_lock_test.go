package medicalrecord

// EMR-456 F1: 健診種別の親 soft-delete と子 Create/親変更を共通 parent row lock で直列化する。
// このファイルでは lock の要求順序と fail-closed を service/repository レベルで検証し、
// 実 DB の同時実行再現は realdb_checkup_type_parent_lock_test.go で行う。

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LockWriteRows の no-op 実装を mock へ追加する（既存テストの parent なし経路を壊さない）。
func (m *mockCheckupTypeRepository) LockWriteRows(context.Context, uint64, []uint64) error { return nil }

// checkupTypeLockSpyRepo は LockWriteRows の呼出しを記録する spy。
// onLock は実際の lock 呼出し時に発火する hook（順序検証用）。
type checkupTypeLockSpyRepo struct {
	CheckupTypeRepository
	locks  [][]uint64
	onLock func()
}

func (r *checkupTypeLockSpyRepo) LockWriteRows(_ context.Context, _ uint64, ids []uint64) error {
	r.locks = append(r.locks, append([]uint64(nil), ids...))
	if r.onLock != nil {
		r.onLock()
	}
	return nil
}

func TestCheckupTypeParentLock_CreateRequestsParentLockBeforeValidation(t *testing.T) {
	parent := uint64(2)
	var events []string
	spy := &checkupTypeLockSpyRepo{}
	spy.CheckupTypeRepository = &mockCheckupTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
			if len(spy.locks) == 0 {
				t.Fatal("parent read before lock")
			}
			events = append(events, "validate-parent")
			return &model.CheckupType{ID: parent, ClinicID: 1}, nil
		},
		createFn: func(_ context.Context, ct *model.CheckupType) error {
			events = append(events, "create")
			ct.ID = 9
			ct.Version = 1
			return nil
		},
	}
	svc := NewCheckupTypeService(spy, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})
	_, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "child", ParentID: &parent, ActorID: &testCheckupTypeActorID})
	if err != nil {
		t.Fatal(err)
	}
	if len(spy.locks) != 1 || len(spy.locks[0]) != 1 || spy.locks[0][0] != parent {
		t.Fatalf("create must lock the parent row: %v", spy.locks)
	}
	if len(events) != 2 || events[0] != "validate-parent" || events[1] != "create" {
		t.Fatalf("unexpected events: %v", events)
	}
}

func TestCheckupTypeParentLock_UpdateSettingParentRequestsLock(t *testing.T) {
	parent := uint64(2)
	spy := &checkupTypeLockSpyRepo{}
	spy.CheckupTypeRepository = &mockCheckupTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
			if len(spy.locks) == 0 {
				t.Fatal("before-image read before lock")
			}
			return &model.CheckupType{ID: 9, ClinicID: 1, Version: 3}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupTypeInput) (*model.CheckupType, error) {
			return &model.CheckupType{ID: 9, ClinicID: 1, Version: 4}, nil
		},
	}
	svc := NewCheckupTypeService(spy, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})
	if _, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{Name: strPtr("改名"), ParentID: &parent, Version: int64Ptr(3), ActorID: &testCheckupTypeActorID}); err != nil {
		t.Fatal(err)
	}
	if len(spy.locks) != 1 || len(spy.locks[0]) != 2 || spy.locks[0][0] != parent || spy.locks[0][1] != 9 {
		t.Fatalf("update that sets parent must lock parent and target in ascending order: %v", spy.locks)
	}
	t.Run("update without parent change locks the target only", func(t *testing.T) {
		spy2 := &checkupTypeLockSpyRepo{CheckupTypeRepository: &mockCheckupTypeRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 3}, nil
			},
			updateFieldsFn: func(context.Context, uint64, uint64, UpdateCheckupTypeInput) (*model.CheckupType, error) {
				return &model.CheckupType{ID: 9, ClinicID: 1, Version: 4}, nil
			},
		}}
		svc2 := NewCheckupTypeService(spy2, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})
		if _, err := svc2.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{Name: strPtr("改名"), Version: int64Ptr(3), ActorID: &testCheckupTypeActorID}); err != nil {
			t.Fatal(err)
		}
		if len(spy2.locks) != 1 || len(spy2.locks[0]) != 1 || spy2.locks[0][0] != 9 {
			t.Fatalf("parent-less update must lock only the target row: %v", spy2.locks)
		}
	})
}

func TestCheckupTypeParentLock_DeleteLocksBeforeChildCheck(t *testing.T) {
	locked, checked := false, false
	spy := &checkupTypeLockSpyRepo{}
	spy.onLock = func() { locked = true }
	spy.CheckupTypeRepository = &mockCheckupTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
			if !locked {
				t.Fatal("row read before lock")
			}
			return &model.CheckupType{ID: 2, ClinicID: 1}, nil
		},
		countChildrenByParentIDFn: func(context.Context, uint64, uint64) (int64, error) {
			if !locked {
				t.Fatal("child check before lock")
			}
			checked = true
			return 1, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error {
			t.Fatal("delete must not run when children exist")
			return nil
		},
	}
	svc := NewCheckupTypeService(spy, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})
	err := svc.Delete(context.Background(), 1, 2)
	if !apperrors.IsConflict(err) || !locked || !checked {
		t.Fatalf("delete ordering broken: %v locked=%v checked=%v", err, locked, checked)
	}
	if len(spy.locks) != 1 || len(spy.locks[0]) != 1 || spy.locks[0][0] != 2 {
		t.Fatalf("delete must lock the parent row first: %v", spy.locks)
	}
}

func TestCheckupTypeParentLock_RepositorySQLOrderAndFailClosed(t *testing.T) {
	var output bytes.Buffer
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	locker, ok := NewCheckupTypeRepository(db).(masterWriteRowsLocker)
	if !ok {
		t.Fatal("checkup type repository must implement masterWriteRowsLocker")
	}
	ctx := persistence.WithTxValue(context.Background(), db)
	if err := locker.LockWriteRows(ctx, 7, []uint64{9, 2}); err != nil {
		t.Fatal(err)
	}
	sql := output.String()
	if first, second := strings.Index(sql, "id = 2"), strings.Index(sql, "id = 9"); first < 0 || second < first || strings.Count(sql, "FOR UPDATE") != 2 || strings.Count(sql, "clinic_id = 7") != 2 {
		t.Fatalf("unsafe lock SQL: %s", sql)
	}
	if err := locker.LockWriteRows(context.Background(), 7, []uint64{2}); err == nil {
		t.Fatal("missing ambient TX must fail closed")
	}
}
