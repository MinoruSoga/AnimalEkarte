// reservation_write_gate_test.go — reservation-graph:write:v1 共通gate の単体テスト。
// ctx 束縛・TX 同一性・fail-closed 契約を検証する。DB 接続を要する取得経路
// （pg_locks 検査・advisory lock）は実 DB を使うため対象外で、
// reservation_write_gate_realdb_test.go で検証する。gate 状態は同 package から
// 直接構築する（white box）。
package persistence

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTx は TX 同一性判定だけに使う識別用 *gorm.DB。接続は開かず、
// DB 操作は行わない（行うテストは realdb 側へ置く）。
func fakeTx() *gorm.DB {
	return &gorm.DB{}
}

func ambCtx(tx *gorm.DB) context.Context {
	return WithTxValue(context.Background(), tx)
}

// gateCtxWith は取得済み gate 状態を手動で構築した ctx を返す。
func gateCtxWith(tx *gorm.DB) context.Context {
	return context.WithValue(ambCtx(tx), writeGateContextKey{}, &writeGateState{
		gateKey: reservationGraphWriteGateKey,
		tx:      tx,
	})
}

func TestAcquireReservationGraphWriteGate_NoAmbientTx_FailClosed(t *testing.T) {
	// DB 接続前に fail-closed する path。
	_, err := AcquireReservationGraphWriteGate(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires an ambient transaction")
}

func TestAssertReservationGraphWriteGate(t *testing.T) {
	t.Run("gate未取得は拒否", func(t *testing.T) {
		err := AssertReservationGraphWriteGate(ambCtx(fakeTx()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not acquired")
	})

	t.Run("同一txのgate参加は成功", func(t *testing.T) {
		tx := fakeTx()
		assert.NoError(t, AssertReservationGraphWriteGate(gateCtxWith(tx)))

		// gate ctx から派生した ctx（value chain 先）でも参加を検出できる。
		derived := context.WithValue(gateCtxWith(tx), struct{}{}, "unused")
		assert.NoError(t, AssertReservationGraphWriteGate(derived))
	})

	t.Run("ambient tx欠落は拒否", func(t *testing.T) {
		// DetachTx は txKey を typed-nil で置き換えるため ambient tx が消える。
		err := AssertReservationGraphWriteGate(DetachTx(gateCtxWith(fakeTx())))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires an ambient transaction")
	})

	t.Run("別txへの混入は拒否", func(t *testing.T) {
		txA, txB := fakeTx(), fakeTx()
		err := AssertReservationGraphWriteGate(WithTxValue(gateCtxWith(txA), txB))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bound to a different transaction")
	})
}

func TestAcquireReservationGraphWriteGate_ForeignGateState_RejectedBeforeDB(t *testing.T) {
	// 別 tx 分の gate 状態を持つ ctx での再取得は、pg_locks 検査より先に拒否される
	// （実 DB への問合せが発生しない path）。
	txA, txB := fakeTx(), fakeTx()
	_, err := AcquireReservationGraphWriteGate(WithTxValue(gateCtxWith(txA), txB))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already held by a different transaction")
}

func TestAssertReservationGraphWriteGate_ErrorsAreInternal(t *testing.T) {
	// gate 契約違反は caller の入力誤りではなく実装側契約違反として
	// internal error（fail-closed）に分類されることを固定する。
	err := AssertReservationGraphWriteGate(context.Background())
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "apperrors.AppError を返す: %v", err)
}
