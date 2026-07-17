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
	lineID1 := "U_test_1"
	lineID2 := "U_test_2"
	rows := []repository.CheckupSyncPreviewRow{
		{OwnerID: 1, OwnerName: "line-linked-1", LineUserID: &lineID1, LstepOptOut: false, LivingPetCount: 1},
		{OwnerID: 2, OwnerName: "line-linked-2", LineUserID: &lineID2, LstepOptOut: false, LivingPetCount: 1},
	}
	repo := &mockCheckupSyncRepository{
		findCheckupSyncPreviewFn: func(_ context.Context, _ *repository.FindCheckupSyncPreviewParams) ([]repository.CheckupSyncPreviewRow, error) {
			return rows, nil
		},
	}
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnersFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]*model.LstepTagCache, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupSyncService(repo, &mockOwnerRepository{}, &mockPetRepository{}, tagCacheRepo, &mockLstepSettingsService{}, &mockAuditService{})
	result, err := svc.PreviewCheckupSync(context.Background(), 1, &PreviewCheckupSyncInput{CheckupType: "annual"}, nil)
	// G7-2: タグキャッシュはバッチ取得(FindByOwners)に変更されたため、失敗は per-owner ではなく
	// 全体に及ぶ。line連携済み全員の currentTags が空スライスにフォールバックし、プレビュー自体は継続する
	// （この一括失敗時の挙動差は BE-refactor.md G7-2 の実装手順で仕様として固定されている）。
	assert.NoError(t, err)
	if assert.NotNil(t, result) && assert.Len(t, result.Owners, 2) {
		assert.Empty(t, result.Owners[0].CurrentTags)
		assert.Empty(t, result.Owners[1].CurrentTags)
	}
}
