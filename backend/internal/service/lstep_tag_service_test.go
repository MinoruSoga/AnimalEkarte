package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mocks ----

type mockLstepTagConfigRepository struct {
	findAllAutoManagedPrefixesFn    func(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	findAllConditionTagMappingsFn   func(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	findAllSendPurposeTagPrefixesFn func(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
}

func (m *mockLstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.findAllAutoManagedPrefixesFn != nil {
		return m.findAllAutoManagedPrefixesFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) FindAllConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	if m.findAllConditionTagMappingsFn != nil {
		return m.findAllConditionTagMappingsFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) FindAllSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	if m.findAllSendPurposeTagPrefixesFn != nil {
		return m.findAllSendPurposeTagPrefixesFn(ctx)
	}
	return nil, nil
}

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
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
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
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
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
	svc := NewLstepTagService(settingsSvc, ownerRepo, tagCache, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.True(t, res.IsLinked)
	assert.Equal(t, []string{"my_tag"}, res.Tags)
}

func TestAddOwnerTag_AutoManagedTag(t *testing.T) {
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "cpm_stage_1", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestIsAutoManagedTag_NewSpecPrefixes(t *testing.T) {
	tags := []string{
		"CPM_01_出会い",
		"LTV_上位20",
		"LTV_フード購入あり",
		"VISIT_120日超",
		"PET_犬あり",
		"HLTH_健診あり",
		"PREV_ワクチン期限",
		"EXCL_配信停止",
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			assert.True(t, isSystemManagedTag(tag))
		})
	}
}

func TestAddOwnerTag_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_AutoManagedTag(t *testing.T) {
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return []*model.LstepAutoManagedPrefix{{Prefix: "vaccine_", Category: "C2"}}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, tagConfigRepo)
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
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.NoError(t, err)
}
