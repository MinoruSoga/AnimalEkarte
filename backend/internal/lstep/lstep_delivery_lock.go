package lstep

import (
	"context"
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// deliveryOwnerLocker は claim 済み配信を owner 直列化境界内で実行する
// narrow optional interface（D1）。具象 tag-cache repository が実装する。
// 境界は owner advisory → owner row UPDATE → recipient advisory →
// lstep_settings row SHARE の正規順序で取得し、コールバック（再検証〜provider
// 実行）終了まで保持する。
type deliveryOwnerLocker interface {
	WithDeliveryOwnerLock(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error
}

// errDeliveryOwnerLockUnavailable は locker を解決できない場合の fail-closed エラー。
var errDeliveryOwnerLockUnavailable = errors.New("delivery owner locker unavailable")

// deliveryOwnerLockerOf resolves the owner locker from the configured tag
// repository, unwrapping batch page caches whose inner repo carries the lock.
func deliveryOwnerLockerOf(repo deliveryTagCacheRepository) deliveryOwnerLocker {
	for repo != nil {
		if c, ok := repo.(*deliveryBatchTagCache); ok {
			repo = c.inner
			continue
		}
		locker, _ := repo.(deliveryOwnerLocker)
		return locker
	}
	return nil
}

// unwrapDeliveryOwnerRepo strips the batch page cache so in-lock reads always
// hit committed database state.
func unwrapDeliveryOwnerRepo(repo deliveryOwnerRepository) deliveryOwnerRepository {
	for {
		c, ok := repo.(*deliveryBatchOwnerCache)
		if !ok {
			return repo
		}
		repo = c.inner
	}
}

// unwrapDeliveryTagRepo strips the batch page cache so in-lock exclusion
// checks see freshly committed EXCL writes.
func unwrapDeliveryTagRepo(repo deliveryTagCacheRepository) deliveryTagCacheRepository {
	for {
		c, ok := repo.(*deliveryBatchTagCache)
		if !ok {
			return repo
		}
		repo = c.inner
	}
}

// deliveryOutcomeTimeout bounds detached outcome persistence after provider
// execution; it must never inherit the caller's cancelled context or ambient
// (already-committed/rolled-back) lock transaction.
const deliveryOutcomeTimeout = 5 * time.Second

func detachedDeliveryOutcomeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(persistence.DetachTx(ctx), deliveryOutcomeTimeout)
}

// markDeliveryKnownUnsent marks a claim that provably never reached the
// provider as retryable-failed on a bounded detached context (D2). The
// conditional update never overwrites fired/suppressed/completed outcomes.
func markDeliveryKnownUnsent(ctx context.Context, repo LstepDeliveryTriggerLogRepository, clinicID, logID uint64) error {
	outCtx, cancel := detachedDeliveryOutcomeContext(ctx)
	defer cancel()
	return repo.MarkKnownUnsent(outCtx, clinicID, logID, preSendRevalidationFailure)
}
