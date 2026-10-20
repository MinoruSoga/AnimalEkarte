// reservation_write_gate_realdb_test.go — reservation-graph:write:v1 共通gate の
// 実 DB 検証（EMR-467 受入条件）。
//
// 使い捨て test DB（testdb.SetupTestDB）と合成 fixture のみを使用し、次を検証する:
//   - 取得順: gate が transaction 先頭で取得できること
//   - 後付取得拒否: 下位 read / advisory lock / row lock 後の gate は待機ではなく
//     即時拒否されること
//   - 異なる TX 混入拒否: gate 状態の別 TX 接ぎ替え・入れ子 transaction を検出すること
//   - 成功 / rollback / 競合: commit・rollback で gate が解放されること、
//     gate 先頭の 2 transaction が gate 順に直列化されること（逆順 lock 競合試験）
package persistence

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// acquireInTx は raw gorm tx を gate ctx へ変換して gate を取得するテスト補助。
func acquireInTx(ctx context.Context, raw *gorm.DB) (context.Context, error) {
	return AcquireReservationGraphWriteGate(WithTxValue(ctx, raw))
}

func assertInTx(ctx context.Context, raw *gorm.DB) error {
	return AssertReservationGraphWriteGate(WithTxValue(ctx, raw))
}

// lockSeedRowForUpdate は合成 fixture 行への FOR UPDATE（下位 row lock）。
func lockSeedRowForUpdate(ctx context.Context, db *gorm.DB, id uint64) error {
	var row model.AnimalSpecies
	return DBOrTx(ctx, db).
		Where("id = ?", id).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&row).Error
}

func TestReservationWriteGate_RealDB_AcquireFirst_InFreshTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)

	gateAcquired := false
	require.NoError(t, tx.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		txCtx, err = AcquireReservationGraphWriteGate(txCtx)
		if err != nil {
			return err
		}
		gateAcquired = true
		// gate 後の下位処理は参加 assert を通る。
		return AssertReservationGraphWriteGate(txCtx)
	}))
	assert.True(t, gateAcquired)
}

func TestReservationWriteGate_RealDB_SameTxReacquire_Idempotent(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)

	// 同一 transaction での再取得は冪等 no-op（advisory xact lock の再入に対応）。
	require.NoError(t, tx.WithTx(context.Background(), func(txCtx context.Context) error {
		first, err := AcquireReservationGraphWriteGate(txCtx)
		if err != nil {
			return err
		}
		second, err := AcquireReservationGraphWriteGate(first)
		if err != nil {
			return err
		}
		if first != second {
			t.Error("idempotent reacquire must return the same context")
		}
		return AssertReservationGraphWriteGate(second)
	}))
}

func TestReservationWriteGate_RealDB_RejectsGateAfterRead(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}))
	tx := NewTransactor(db)

	// gate 前の事前 read は同一 TX 内に relation lock を残すため後付けとして拒否。
	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		var count int64
		if err := DBOrTx(txCtx, db).Model(&model.AnimalSpecies{}).Count(&count).Error; err != nil {
			return err
		}
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "before any lower lock or read")
}

func TestReservationWriteGate_RealDB_RejectsGateAfterLowerAdvisoryLock(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)

	// 既存 advisory lock（例: next-visit-intent）の後での gate 取得は後付けとして拒否。
	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		if err := DBOrTx(txCtx, db).
			Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", "next-visit-intent:1:1").Error; err != nil {
			return err
		}
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "before any lower lock or read")
}

func TestReservationWriteGate_RealDB_RejectsGateAfterRowLock(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}))
	tx := NewTransactor(db)

	seed := model.AnimalSpecies{Name: "gate-fixture-rowlock"}
	require.NoError(t, DBOrTx(context.Background(), db).Create(&seed).Error)

	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		if err := lockSeedRowForUpdate(txCtx, db, seed.ID); err != nil {
			return err
		}
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "before any lower lock or read")
}

func TestReservationWriteGate_RealDB_RejectsForeignTxSplice(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	var gateCtx context.Context
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		gateCtx, err = AcquireReservationGraphWriteGate(txCtx)
		require.NoError(t, err)
		return nil
	}))

	// gate 状態を持つ ctx の ambient tx を別 TX へ接ぎ替えた呼出しは混入として拒否。
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		return assertInTx(gateCtx, TxFromContext(txCtx))
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bound to a different transaction")
}

func TestReservationWriteGate_RealDB_RejectsNestedTxInsideGate(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)

	// gate 保持中の入れ子 transaction（別 *gorm.DB）は ambient participant になれない。
	err := tx.WithTx(context.Background(), func(outerCtx context.Context) error {
		outerCtx, err := AcquireReservationGraphWriteGate(outerCtx)
		if err != nil {
			return err
		}
		return NewTransactor(db).WithTx(outerCtx, func(innerCtx context.Context) error {
			if err := AssertReservationGraphWriteGate(innerCtx); err == nil {
				return errors.New("nested transaction must be rejected by gate assert")
			}
			if _, err := AcquireReservationGraphWriteGate(innerCtx); err == nil {
				return errors.New("nested transaction must not acquire the gate")
			}
			return nil
		})
	})
	require.NoError(t, err)
}

func TestReservationWriteGate_RealDB_ReleaseOnCommitAndRollback(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	// 成功 commit で gate が解放され、後続 TX が待ちなしで取得できる。
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	}))
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	}))

	// rollback でも gate は解放される。
	require.Error(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := AcquireReservationGraphWriteGate(txCtx)
		if err != nil {
			return err
		}
		return errors.New("force rollback")
	}))
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		_, err := AcquireReservationGraphWriteGate(txCtx)
		return err
	}))
}

func TestReservationWriteGate_RealDB_ContentionSerializesInGateOrder(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()

	// TX-A が gate を保持している間、TX-B の gate 取得は直列化される。
	// TX-A の commit（解放）後に TX-B だけが取得に進む。
	bAcquired := make(chan struct{})
	bDone := make(chan error, 1)

	require.NoError(t, db.Transaction(func(rawA *gorm.DB) error {
		ctxA, err := acquireInTx(ctx, rawA)
		if err != nil {
			return err
		}
		if err := assertInTx(ctxA, rawA); err != nil {
			return err
		}

		go func() {
			bDone <- db.Transaction(func(rawB *gorm.DB) error {
				ctxB, err := acquireInTx(ctx, rawB)
				if err != nil {
					return err
				}
				close(bAcquired)
				return assertInTx(ctxB, rawB)
			})
		}()

		// A 保持中に B が取得できていないことを短時間確認してから commit する。
		select {
		case <-bAcquired:
			t.Error("TX-B acquired the gate while TX-A still holds it")
			return errors.New("gate not serialized")
		case <-time.After(300 * time.Millisecond):
		}
		return nil // TX-A commit → gate 解放
	}))

	select {
	case err := <-bDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("TX-B never acquired the gate after TX-A committed")
	}
}

func TestReservationWriteGate_RealDB_ReverseOrderAcquireFailsFastWithoutDeadlock(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}))
	ctx := context.Background()

	// 逆順 lock 競合試験:
	//   TX-A: gate → row lock（正順。commit まで保持）
	//   TX-B: 別行の write で relation lock を保持 → gate 取得（逆順の後付け）
	// gate の pg_locks 検査が TX-B を待機ではなく即時拒否にするため、
	// 「A は B の lock 待ち」「B は A の gate 待ち」の相互待ち（deadlock）に至らない。
	// B は A が保持しない行にのみ書込むことで、lock 待ちを gate 検査に分離する。
	rowA := model.AnimalSpecies{Name: "gate-fixture-a"}
	rowB := model.AnimalSpecies{Name: "gate-fixture-b"}
	require.NoError(t, DBOrTx(ctx, db).Create(&rowA).Error)
	require.NoError(t, DBOrTx(ctx, db).Create(&rowB).Error)

	bResult := make(chan error, 1)
	aDone := make(chan error, 1)

	go func() {
		aDone <- db.Transaction(func(rawA *gorm.DB) error {
			ctxA, err := acquireInTx(ctx, rawA)
			if err != nil {
				return err
			}
			if err := lockSeedRowForUpdate(ctxA, db, rowA.ID); err != nil {
				return err
			}
			// TX-B の逆順 gate 試行が即時拒否されるのを待つ。
			select {
			case err := <-bResult:
				if err == nil {
					t.Error("TX-B post-hoc gate acquisition was not rejected")
				}
			case <-time.After(5 * time.Second):
				t.Error("TX-B gate attempt did not finish")
			}
			return nil // TX-A commit
		})
	}()

	require.NoError(t, db.Transaction(func(rawB *gorm.DB) error {
		ctxB := WithTxValue(ctx, rawB)
		// B は rowB への write で relation lock のみを持つ（rowA には触れない）。
		if err := DBOrTx(ctxB, db).
			Model(&model.AnimalSpecies{}).
			Where("id = ?", rowB.ID).
			Update("name", "gate-fixture-b-updated").Error; err != nil {
			return err
		}
		_, err := AcquireReservationGraphWriteGate(ctxB)
		bResult <- err
		require.Error(t, err)
		assert.Contains(t, err.Error(), "before any lower lock or read")
		return nil
	}))

	select {
	case err := <-aDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("TX-A did not finish; possible deadlock")
	}
}

func TestReservationWriteGate_RealDB_ParallelAcquireAcrossTx(t *testing.T) {
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	// 並列 TX がそれぞれ gate を取得して commit する。直列化により全員が成功する。
	const workers = 8
	var wg sync.WaitGroup
	results := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(slot int) {
			defer wg.Done()
			results[slot] = tx.WithTx(ctx, func(txCtx context.Context) error {
				txCtx, err := AcquireReservationGraphWriteGate(txCtx)
				if err != nil {
					return err
				}
				return AssertReservationGraphWriteGate(txCtx)
			})
		}(i)
	}
	wg.Wait()
	for slot, err := range results {
		assert.NoError(t, err, "worker %d", slot)
	}
}

func TestReservationWriteGate_RealDB_GateKeyIsFixedNamespace(t *testing.T) {
	// gate key は選定済み固定値。誤って別 key を発明していないことを固定する。
	assert.Equal(t, "reservation-graph:write:v1", reservationGraphWriteGateKey)

	// 代表的な既存 advisory key と hash 衝突しないこと（§6.4: 新しい並列 gate を
	// 発明しない / 既存 key と重ならない）。
	db := testdb.SetupTestDB(t)
	tx := NewTransactor(db)
	require.NoError(t, tx.WithTx(context.Background(), func(txCtx context.Context) error {
		var hashes []string
		if err := DBOrTx(txCtx, db).Raw(`
			SELECT hashtextextended(key, 0)::text
			FROM (VALUES ('reservation-graph:write:v1'), ('next-visit-intent:1:1')) AS t(key)`).
			Scan(&hashes).Error; err != nil {
			return err
		}
		if len(hashes) != 2 || hashes[0] == hashes[1] {
			return errors.New("gate key must not collide with existing advisory keys")
		}
		return nil
	}))
}
