package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// This file extends checkup_sync_service_test.go's PreviewCheckupSync coverage with branches not
// exercised there: nil input, repository error, thresholds lookup error, and tag-cache lookup
// error for line-linked owners.

func TestCheckupSyncService_PreviewCheckupSync_NilInput(t *testing.T) {
	svc := newCheckupSyncSvcForPreview(nil)
	result, err := svc.PreviewCheckupSync(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestCheckupSyncService_PreviewCheckupSync_RepositoryError(t *testing.T) {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *repository.FindCheckupSyncPreviewParams) ([]repository.CheckupSyncPreviewRow, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, &mockLstepTagCacheRepository{}, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// thresholdsFailingSettingsService embeds mockLstepSettingsService (stable) and overrides
// GetCPMV1Thresholds, which the base mock hardcodes to always succeed with no injection hook.
type thresholdsFailingSettingsService struct {
	*mockLstepSettingsService
}

func (m *thresholdsFailingSettingsService) GetCPMV1Thresholds(_ context.Context, _ uint64) (model.CPMV1Thresholds, error) {
	return model.CPMV1Thresholds{}, errors.New("db error")
}

func TestCheckupSyncService_PreviewCheckupSync_ThresholdsError(t *testing.T) {
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *repository.FindCheckupSyncPreviewParams) ([]repository.CheckupSyncPreviewRow, error) {
			return nil, nil
		},
	}
	svc := NewCheckupSyncService(
		repo,
		&mockOwnerRepository{},
		&mockPetRepository{},
		&mockLstepTagCacheRepository{},
		&thresholdsFailingSettingsService{mockLstepSettingsService: &mockLstepSettingsService{}},
		&mockAuditService{},
	)
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheckupSyncService_PreviewCheckupSync_TagCacheLookupError(t *testing.T) {
	lineID := "U_test"
	rows := []repository.CheckupSyncPreviewRow{
		{OwnerID: 1, OwnerName: "line-linked", LineUserID: &lineID, LstepOptOut: false, LivingPetCount: 1},
	}
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *repository.FindCheckupSyncPreviewParams) ([]repository.CheckupSyncPreviewRow, error) {
			return rows, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, tagCacheRepo, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	// Tag cache lookup failure for a line-linked owner is non-fatal: currentTags falls back to
	// an empty slice and preview continues.
	assert.NoError(t, err)
	if assert.NotNil(t, result) && assert.Len(t, result.Owners, 1) {
		assert.Empty(t, result.Owners[0].CurrentTags)
	}
}
