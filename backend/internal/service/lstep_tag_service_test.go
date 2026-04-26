package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- tests ----
// Reuses: mockLstepSettingsService (lstep_lifecycle_service_test.go)
//         mockOwnerRepository       (owner_service_test.go)
//         mockLstepTagCacheRepository (lstep_lifecycle_service_test.go)
//         mockAuditService           (lstep_lifecycle_service_test.go)

func TestGetOwnerTags_NotLinked(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{})
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.False(t, res.IsLinked)
	assert.Empty(t, res.Tags)
}

func TestGetOwnerTags_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "1")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{})
	_, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestGetOwnerTags_CacheFallback(t *testing.T) {
	lineID := "U123"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "my_tag"}}, nil
		},
	}
	// GetRawCredentials returns empty apiKey → client is nil → falls back to cache
	settingsSvc := &mockLstepSettingsService{}
	svc := NewLstepTagService(settingsSvc, ownerRepo, tagCache, &mockAuditService{})
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.True(t, res.IsLinked)
	assert.Equal(t, []string{"my_tag"}, res.Tags)
}

func TestAddOwnerTag_AutoManagedTag(t *testing.T) {
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{})
	err := svc.AddOwnerTag(context.Background(), 1, 1, "cpm_stage_1", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestAddOwnerTag_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{})
	err := svc.AddOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_AutoManagedTag(t *testing.T) {
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{})
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "vaccine_rabies", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestRemoveOwnerTag_NotLinked(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{})
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.NoError(t, err)
}
