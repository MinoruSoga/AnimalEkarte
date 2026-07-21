package lstep

// service_mocks_test.go — def残存（service 側 lifecycle test 等）の共有 mock を
// 移動先で再宣言する規約の複製（BE9-2C L②）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagCacheRepository struct {
	upsertTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
	deleteTagFn        func(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	deleteAllByOwnerFn func(ctx context.Context, clinicID, ownerID uint64) error
	findByOwnerFn      func(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	findByOwnersFn     func(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
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

// mockLstepTagConfigRepository — internal/service/lstep_tag_service_test.go の同名 mock の複製
// （LstepTagService は L③ 対象のため carrier を service 側に残す・BE9-2C L②）。
type mockLstepTagConfigRepository struct {
	findAllAutoManagedPrefixesFn    func(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	createAutoManagedPrefixFn       func(ctx context.Context, m *model.LstepAutoManagedPrefix) error
	deleteAutoManagedPrefixFn       func(ctx context.Context, id uint64) error
	findAllConditionTagMappingsFn   func(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	createConditionTagMappingFn     func(ctx context.Context, m *model.LstepConditionTagMapping) error
	deleteConditionTagMappingFn     func(ctx context.Context, id uint64) error
	findAllSendPurposeTagPrefixesFn func(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
	createSendPurposeTagPrefixFn    func(ctx context.Context, m *model.LstepSendPurposeTagPrefix) error
	deleteSendPurposeTagPrefixFn    func(ctx context.Context, id uint64) error
}

func (m *mockLstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.findAllAutoManagedPrefixesFn != nil {
		return m.findAllAutoManagedPrefixesFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateAutoManagedPrefix(ctx context.Context, rec *model.LstepAutoManagedPrefix) error {
	if m.createAutoManagedPrefixFn != nil {
		return m.createAutoManagedPrefixFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteAutoManagedPrefix(ctx context.Context, id uint64) error {
	if m.deleteAutoManagedPrefixFn != nil {
		return m.deleteAutoManagedPrefixFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) FindAllConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	if m.findAllConditionTagMappingsFn != nil {
		return m.findAllConditionTagMappingsFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateConditionTagMapping(ctx context.Context, rec *model.LstepConditionTagMapping) error {
	if m.createConditionTagMappingFn != nil {
		return m.createConditionTagMappingFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteConditionTagMapping(ctx context.Context, id uint64) error {
	if m.deleteConditionTagMappingFn != nil {
		return m.deleteConditionTagMappingFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) FindAllSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	if m.findAllSendPurposeTagPrefixesFn != nil {
		return m.findAllSendPurposeTagPrefixesFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateSendPurposeTagPrefix(ctx context.Context, rec *model.LstepSendPurposeTagPrefix) error {
	if m.createSendPurposeTagPrefixFn != nil {
		return m.createSendPurposeTagPrefixFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error {
	if m.deleteSendPurposeTagPrefixFn != nil {
		return m.deleteSendPurposeTagPrefixFn(ctx, id)
	}
	return nil
}
