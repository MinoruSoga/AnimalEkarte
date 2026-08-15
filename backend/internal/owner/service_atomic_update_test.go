package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerService_UpdateUsesAtomicUpdateAndFind(t *testing.T) {
	atomicCalled := false
	legacyUpdateCalled := false
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Owner, error) {
			return &model.Owner{ID: id, ClinicID: clinicID}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			legacyUpdateCalled = true
			return nil
		},
		updateAndFindFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Owner, error) {
			atomicCalled = true
			return nil, errors.New("reload failed before commit")
		},
	}
	svc := NewService(repo, nil, nil, nil)
	name := "更新後"

	got, err := svc.Update(context.Background(), 1, 10, &UpdateOwnerInput{OwnerName: &name})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, atomicCalled)
	assert.False(t, legacyUpdateCalled)
}
