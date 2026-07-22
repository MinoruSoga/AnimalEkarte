package service

// Residual lifecycle tests use only the prefix-reading part of the tag-config
// repository. The full tag-service test double moved to internal/lstep in L③a.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagConfigRepository struct {
	findAllAutoManagedPrefixesFn func(context.Context) ([]*model.LstepAutoManagedPrefix, error)
}

func (m *mockLstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.findAllAutoManagedPrefixesFn != nil {
		return m.findAllAutoManagedPrefixesFn(ctx)
	}
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateAutoManagedPrefix(context.Context, *model.LstepAutoManagedPrefix) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteAutoManagedPrefix(context.Context, uint64) error {
	return nil
}

func (*mockLstepTagConfigRepository) FindAllConditionTagMappings(context.Context) ([]*model.LstepConditionTagMapping, error) {
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateConditionTagMapping(context.Context, *model.LstepConditionTagMapping) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteConditionTagMapping(context.Context, uint64) error {
	return nil
}

func (*mockLstepTagConfigRepository) FindAllSendPurposeTagPrefixes(context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateSendPurposeTagPrefix(context.Context, *model.LstepSendPurposeTagPrefix) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteSendPurposeTagPrefix(context.Context, uint64) error {
	return nil
}
