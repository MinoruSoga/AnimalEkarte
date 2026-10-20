// reservation_write_gate.go — reservation-graph:write:v1 共通gate（EMR-467）。
//
// reservations technical-spec §6.4（全writerの取得規約）で選定された DB 共通の
// transaction advisory gate `reservation-graph:write:v1` を、最外の業務物理
// transaction の先頭（全下位 lock より前）で取得するための共通 TX helper。
//
// 配置は internal/persistence（横断 package）。§6.4 の gate は reservation の
// writer 群だけでなく pet/owner 等の blocking writer（writer-reader matrix
// W19 以降）も「上位 gate から参加 / gate 取得済みを assert」するため、
// domain 間 import allowlist に依存しないこの package に置く。
// 鍵文字列の persist は AcquireLineRecipientLock の先例に従う。
//
// 契約（§6.4 / revisit-intake §5.2 / EMR-467 受入条件）:
//
//  1. 最外の業務物理 transaction 開始直後（receipt FOR UPDATE を含む全下位
//     lock・事前 read より前）に AcquireReservationGraphWriteGate を呼ぶ。
//     gate 前に業務 row / receipt / policy / close / staff / その他 advisory を
//     保持してはならない。事前 read の集合は gate 後に再確認する。
//  2. gate は commit / rollback まで保持される（pg_advisory_xact_lock）。
//  3. 同一 transaction の下位処理（ambient participant）は
//     AssertReservationGraphWriteGate で gate 参加を assert する。assert は
//     gate 未取得・ambient transaction 欠落・別 transaction / 入れ子
//     transaction への混入を fail-closed で拒否する。
//  4. 下位 lock 取得後の gate 後付け（逆順）は pg_locks 検査で検出して拒否する。
//     逆順取得を許すと gate 保持者と下位 lock 保持者が互いを待ち deadlock に
//     至るため、取得待ちの前に拒否する。
//
// 呼出例（canonical order）:
//
//	return transactor.WithTx(ctx, func(ctx context.Context) error {
//	    ctx, err := persistence.AcquireReservationGraphWriteGate(ctx)
//	    if err != nil {
//	        return err
//	    }
//	    if err := persistence.AssertReservationGraphWriteGate(ctx); err != nil {
//	        return err // ambient participant
//	    }
//	    // gate 後: receipt FOR UPDATE → 既存 owner 固有の下位順序 …
//	    return nil
//	})
//
// gate は他院データの読取権限を与えない（§6.4）。参加は直列化の順序保証だけを
// 意味し、認可・clinic 分離は各 writer の既存検査に従う。
package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// reservationGraphWriteGateKey は §6.4 で選定済みの固定 gate key。
// namespace を新設せず、既存の hashtextextended(text, 0) 方式で hash する。
const reservationGraphWriteGateKey = "reservation-graph:write:v1"

// writeGateState は gate を取得した transaction への束縛情報。
// gorm.DB の pointer 同一性で「gate を取得した同一 ambient transaction」を
// 判定する（入れ子 Transaction は別の *gorm.DB になるため混入を検出できる）。
type writeGateState struct {
	gateKey string
	tx      *gorm.DB
}

type writeGateContextKey struct{}

func writeGateFromContext(ctx context.Context) *writeGateState {
	state, _ := ctx.Value(writeGateContextKey{}).(*writeGateState)
	return state
}

// AcquireReservationGraphWriteGate は現在の ambient transaction 上で
// reservation-graph:write:v1 advisory gate を取得し、gate 状態を ctx へ束縛した
// 後続 ctx を返す。
//
// 必須 precondition:
//   - ambient transaction が存在すること（欠落は fail-closed）。
//   - 同一 transaction 内で gate より前に取得された lock / 実施された read が
//     ないこと。pg_locks で現在 transaction の保持 lock を検査し、自己
//     transactionid / virtualxid 以外の保持があれば後付け取得として拒否する。
//
// 同一 transaction での再呼出しは冪等 no-op（advisory xact lock は同一
// transaction 内で再入可能）。既に別 transaction 分の gate 状態を持つ ctx は
// 混入として拒否する。
func AcquireReservationGraphWriteGate(ctx context.Context) (context.Context, error) {
	tx := TxFromContext(ctx)
	if tx == nil {
		return ctx, apperrors.WrapInternalServerError(
			"reservation-graph write gate requires an ambient transaction")
	}
	if existing := writeGateFromContext(ctx); existing != nil {
		if existing.tx == tx {
			// 同一 transaction での再取得は冪等。gate は commit/rollback まで保持済み。
			return ctx, nil
		}
		return ctx, apperrors.WrapInternalServerError(
			"reservation-graph write gate is already held by a different transaction")
	}
	if err := assertGateIsFirstLock(ctx, tx); err != nil {
		return ctx, err
	}
	if err := tx.WithContext(ctx).
		Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", reservationGraphWriteGateKey).
		Error; err != nil {
		return ctx, apperrors.FromGORM(err, "reservation", reservationGraphWriteGateKey)
	}
	return context.WithValue(ctx, writeGateContextKey{}, &writeGateState{
		gateKey: reservationGraphWriteGateKey,
		tx:      tx,
	}), nil
}

// AssertReservationGraphWriteGate は gate 参加中の下位処理（ambient
// participant）のための assert。次の全てを満たさない場合 fail-closed で拒否する:
//
//   - ctx に gate 取得済みの状態がある（未取得の下位 write を検出）。
//   - 現在の ambient transaction が存在する（DetachTx 等の TX 抜けを検出）。
//   - gate を取得した transaction と現在の ambient transaction が同一である
//     （別 transaction / 入れ子 transaction への混入を検出）。
func AssertReservationGraphWriteGate(ctx context.Context) error {
	tx := TxFromContext(ctx)
	state := writeGateFromContext(ctx)
	switch {
	case state == nil:
		return apperrors.WrapInternalServerError(
			"reservation-graph write gate is not acquired: lower writes must run inside the gated transaction")
	case tx == nil:
		return apperrors.WrapInternalServerError(
			"reservation-graph write gate participant requires an ambient transaction")
	case state.tx != tx:
		return apperrors.WrapInternalServerError(
			"reservation-graph write gate is bound to a different transaction")
	default:
		return nil
	}
}

// assertGateIsFirstLock は gate 前付け契約（§6.4: gate 前に業務 row / receipt /
// policy / close / staff / その他 advisory を保持しない）を pg_locks で検証する。
//
// BEGIN 直後の transaction が保持するのは自己 transactionid / virtualxid の
// み。それ以外の relation / page / tuple / object / advisory 等の保持は、この
// transaction 内で gate より前に表 access・row lock・advisory lock が発生した
// こと（＝下位 lock 後の後付け gate）を意味するため拒否する。検査は advisory
// lock 取得の前に行うため、逆順呼出しは待機ではなく即時エラーで止まり、
// gate 保持者との相互待ち（deadlock）に至らない。
//
// この検査 query 自身が pg_locks view の AccessShareLock を同一 transaction 内
// で保持するため、その view 分だけは自己検出の誤りとして除外する。
func assertGateIsFirstLock(ctx context.Context, tx *gorm.DB) error {
	var priorLockTypes string
	if err := tx.WithContext(ctx).Raw(`
		SELECT coalesce(string_agg(DISTINCT l.locktype, ',' ORDER BY l.locktype), '')
		FROM pg_locks l
		WHERE l.pid = pg_backend_pid()
		  AND l.granted
		  AND l.locktype NOT IN ('virtualxid', 'transactionid')
		  AND l.relation IS DISTINCT FROM 'pg_locks'::regclass`).
		Scan(&priorLockTypes).Error; err != nil {
		return apperrors.FromGORM(err, "reservation", reservationGraphWriteGateKey)
	}
	if priorLockTypes != "" {
		return apperrors.WrapInternalServerError(fmt.Sprintf(
			"reservation-graph write gate must be acquired before any lower lock or read in the same transaction (prior locks: %s)",
			priorLockTypes))
	}
	return nil
}
