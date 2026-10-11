package lstep

// service_mocks_test.go — def残存（service 側 lifecycle test 等）の共有 mock を
// 移動先で再宣言する規約の複製（BE9-2C L②）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagCacheRepository struct {
	enqueueNextVisitSyncFn     func(context.Context, uint64, uint64) (uint64, error)
	acknowledgeNextVisitSyncFn func(context.Context, uint64, uint64, uint64) error
	withNextVisitOwnerLockFn   func(context.Context, uint64, uint64, func(context.Context) error) error
	upsertTagFn                func(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
	deleteTagFn                func(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	deleteAllByOwnerFn         func(ctx context.Context, clinicID, ownerID uint64) error
	findByOwnerFn              func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	findByOwnersFn             func(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
}

func (m *mockLstepTagCacheRepository) EnqueueNextVisitSync(ctx context.Context, clinicID, ownerID uint64) (uint64, error) {
	if m.enqueueNextVisitSyncFn != nil {
		return m.enqueueNextVisitSyncFn(ctx, clinicID, ownerID)
	}
	return 1, nil
}

func (m *mockLstepTagCacheRepository) AcknowledgeNextVisitSync(ctx context.Context, clinicID, ownerID, revision uint64) error {
	if m.acknowledgeNextVisitSyncFn != nil {
		return m.acknowledgeNextVisitSyncFn(ctx, clinicID, ownerID, revision)
	}
	return nil
}

func (m *mockLstepTagCacheRepository) UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error {
	if m.upsertTagFn != nil {
		return m.upsertTagFn(ctx, clinicID, ownerID, tagName, category, reason)
	}
	return nil
}

func (m *mockLstepTagCacheRepository) DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error {
	if m.deleteTagFn != nil {
		return m.deleteTagFn(ctx, clinicID, ownerID, tagName)
	}
	return nil
}

func (m *mockLstepTagCacheRepository) DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error {
	if m.deleteAllByOwnerFn != nil {
		return m.deleteAllByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagCacheRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockLstepTagCacheRepository) FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
	if m.findByOwnersFn != nil {
		return m.findByOwnersFn(ctx, clinicID, ownerIDs)
	}
	return map[uint64][]*model.LstepTagCache{}, nil
}

func (m *mockLstepTagCacheRepository) TagSummary(ctx context.Context, clinicID uint64) ([]TagSummaryRow, int64, error) {
	return nil, 0, nil
}

func (m *mockLstepTagCacheRepository) FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error) {
	return nil, 0, nil
}

func (m *mockLstepTagCacheRepository) BulkReplaceOwnerTags(ctx context.Context, clinicID, ownerID uint64, tags []TagEntry) error {
	return nil
}

func (m *mockLstepTagCacheRepository) FindOwnerIDsByTag(_ context.Context, _ uint64, _ string) ([]uint64, error) {
	return nil, nil
}

func (m *mockLstepTagCacheRepository) WithNextVisitOwnerLock(ctx context.Context, clinicID, ownerID uint64, fn func(context.Context) error) error {
	if m.withNextVisitOwnerLockFn != nil {
		return m.withNextVisitOwnerLockFn(ctx, clinicID, ownerID, fn)
	}
	return fn(ctx)
}
